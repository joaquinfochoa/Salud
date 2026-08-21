#!/usr/bin/env python3
"""Small local research assistant for the health startup.

The script intentionally uses only Python's standard library. It can run without
an AI API, but can optionally call the OpenAI Responses API when OPENAI_API_KEY
is set.
"""

from __future__ import annotations

import argparse
import csv
import datetime as dt
import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import textwrap
import unicodedata
import urllib.parse
import urllib.error
import urllib.request
import xml.etree.ElementTree as ET


ROOT = Path(__file__).resolve().parent
DATA = ROOT / "data"
REPORTS = DATA / "reports"
INTERVIEWS = DATA / "interviews.csv"
EVIDENCE = DATA / "evidence.csv"
EXPERIMENTS = DATA / "experiments.csv"
VERTICALS = DATA / "vertical_scores.csv"
PAPERS = DATA / "papers.csv"
QUESTIONS = DATA / "research_questions.csv"
SOURCES = DATA / "sources.csv"
COMPETITORS = DATA / "competitors.csv"
REGULATIONS = DATA / "regulations.csv"
TOOLS = DATA / "tools.csv"
AUTO_RUNS = DATA / "auto_runs.csv"
DECISIONS = DATA / "decisions.csv"
FOUNDER_QUESTIONS = DATA / "founder_questions.csv"


CSV_HEADERS = {
    INTERVIEWS: [
        "date",
        "kind",
        "segment",
        "name",
        "pain",
        "urgency",
        "willingness",
        "notes",
    ],
    EVIDENCE: ["date", "claim", "source", "confidence", "notes"],
    EXPERIMENTS: [
        "date",
        "name",
        "hypothesis",
        "metric",
        "target",
        "status",
        "result",
        "notes",
    ],
    VERTICALS: [
        "vertical",
        "patient_pain",
        "professional_pain",
        "recurrence",
        "regulatory_friction",
        "founder_edge",
        "notes",
    ],
    PAPERS: [
        "date",
        "source",
        "query",
        "title",
        "authors",
        "year",
        "url",
        "abstract",
        "notes",
    ],
    QUESTIONS: [
        "id",
        "category",
        "question",
        "priority",
        "status",
        "decision_if_true",
        "decision_if_false",
        "notes",
    ],
    SOURCES: [
        "id",
        "date",
        "category",
        "source_type",
        "title",
        "url",
        "credibility",
        "finding",
        "next_action",
    ],
    COMPETITORS: [
        "name",
        "geography",
        "model",
        "target",
        "pricing_signal",
        "strengths",
        "weaknesses",
        "relevance",
        "url",
        "notes",
    ],
    REGULATIONS: [
        "id",
        "jurisdiction",
        "norm",
        "topic",
        "impact",
        "risk_level",
        "url",
        "status",
        "notes",
    ],
    TOOLS: [
        "name",
        "category",
        "use_case",
        "build_or_buy",
        "cost_signal",
        "url",
        "notes",
    ],
    AUTO_RUNS: [
        "date",
        "run_id",
        "question_id",
        "question",
        "model",
        "domain_bundle",
        "status",
        "report_path",
        "summary",
    ],
    DECISIONS: [
        "date",
        "id",
        "status",
        "decision",
        "reason",
        "evidence",
        "confidence",
        "next_action",
        "review_date",
        "notes",
    ],
    FOUNDER_QUESTIONS: [
        "date",
        "id",
        "status",
        "priority",
        "question",
        "why_it_matters",
        "decision_unlocked",
        "source",
        "answer",
        "answered_at",
        "next_action",
    ],
}


DOMAIN_BUNDLES = {
    "all": [],
    "official_ar": [
        "argentina.gob.ar",
        "indec.gob.ar",
        "sssalud.gob.ar",
        "datos.gob.ar",
        "gba.gob.ar",
        "ioma.gba.gob.ar",
        "buenosaires.gob.ar",
    ],
    "academic": [
        "pubmed.ncbi.nlm.nih.gov",
        "ncbi.nlm.nih.gov",
        "arxiv.org",
        "who.int",
        "paho.org",
    ],
    "healthtech": [
        "doctoralia.com.ar",
        "docturno.com",
        "zocdoc.com",
        "headway.co",
        "helloalma.com",
    ],
}

PROVIDERS = {"openai", "claude_cli", "local"}

STOPWORDS = {
    "abre",
    "argentina",
    "como",
    "con",
    "cual",
    "cuales",
    "cuando",
    "cuanto",
    "cuantas",
    "cuantos",
    "desde",
    "donde",
    "esta",
    "este",
    "estos",
    "para",
    "pero",
    "plataforma",
    "plataformas",
    "por",
    "privada",
    "privado",
    "que",
    "real",
    "reales",
    "salud",
    "sobre",
    "tiene",
    "una",
    "unos",
}


def claude_command() -> list[str]:
    path = shutil.which("claude") or shutil.which("claude.cmd") or shutil.which("claude.exe")
    if not path:
        ps1_path = shutil.which("claude.ps1")
        if ps1_path:
            return [
                "powershell",
                "-NoProfile",
                "-ExecutionPolicy",
                "Bypass",
                "-File",
                ps1_path,
            ]
        raise FileNotFoundError("No encontre el comando `claude` en PATH.")
    return [path]


def load_local_env() -> None:
    env_path = ROOT / ".env"
    if not env_path.exists():
        return

    for raw_line in env_path.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        key = key.strip()
        value = value.strip().strip('"').strip("'")
        if key and key not in os.environ:
            os.environ[key] = value


DEFAULT_VERTICALS = [
    {
        "vertical": "salud_mental",
        "patient_pain": "5",
        "professional_pain": "4",
        "recurrence": "5",
        "regulatory_friction": "3",
        "founder_edge": "3",
        "notes": "Alta demanda, teleconsulta viable, cuidado con crisis y datos sensibles.",
    },
    {
        "vertical": "kinesiologia",
        "patient_pain": "4",
        "professional_pain": "4",
        "recurrence": "4",
        "regulatory_friction": "2",
        "founder_edge": "2",
        "notes": "Frecuencia por tratamiento, posible domicilio, suele requerir prescripcion.",
    },
    {
        "vertical": "odontologia",
        "patient_pain": "4",
        "professional_pain": "5",
        "recurrence": "3",
        "regulatory_friction": "3",
        "founder_edge": "5",
        "notes": "Ventaja del fundador y tickets altos, mas presencial y cobertura variable.",
    },
]


DEFAULT_QUESTIONS = [
    {
        "id": "MKT-001",
        "category": "mercado",
        "question": "Cuantas personas en AMBA tienen cobertura y sufren friccion de acceso?",
        "priority": "5",
        "status": "open",
        "decision_if_true": "Arrancar en AMBA con propuesta de transparencia y acceso.",
        "decision_if_false": "Buscar nicho de pago privado o B2B empresa.",
        "notes": "Cruzar INDEC, SSS y entrevistas.",
    },
    {
        "id": "MKT-002",
        "category": "dinero",
        "question": "Cuanto gastan obras sociales/prepagas por salud mental, rehabilitacion y odontologia?",
        "priority": "5",
        "status": "open",
        "decision_if_true": "Priorizar vertical con dinero, dolor y friccion administrativa.",
        "decision_if_false": "Validar solo privado/reintegro asistido.",
        "notes": "Usar plan de cuentas SSS, transferencias y pedido de informacion.",
    },
    {
        "id": "USR-001",
        "category": "pacientes",
        "question": "Pacientes con cobertura pagarian privado por turno rapido, precio claro y profesional verificado?",
        "priority": "5",
        "status": "open",
        "decision_if_true": "Lanzar MVP sin convenios, con comprobante/reintegro asistido.",
        "decision_if_false": "No construir marketplace privado; buscar convenios primero.",
        "notes": "Validar con entrevistas y pagos reales.",
    },
    {
        "id": "PRO-001",
        "category": "profesionales",
        "question": "Profesionales aceptan comision si la plataforma trae pacientes y cobra online?",
        "priority": "5",
        "status": "open",
        "decision_if_true": "Modelo take-rate por consulta.",
        "decision_if_false": "Probar SaaS mensual o servicio administrativo.",
        "notes": "Separar psicologia, kine y odontologia.",
    },
    {
        "id": "LEG-001",
        "category": "regulacion",
        "question": "Que requisitos reales tiene operar como Red de Prestadores digital?",
        "priority": "5",
        "status": "open",
        "decision_if_true": "Planificar roadmap legal Etapa 2/3.",
        "decision_if_false": "Mantener rol de facilitador privado y no tocar financiadores.",
        "notes": "Confirmar con abogado sanitario y SSS.",
    },
    {
        "id": "LEG-002",
        "category": "regulacion",
        "question": "Que oportunidades y limites abre Mi Argentina y la receta electronica para una plataforma privada de salud?",
        "priority": "5",
        "status": "open",
        "decision_if_true": "Explorar integraciones futuras con receta electronica y consentimiento del paciente.",
        "decision_if_false": "Mantener el MVP como orientacion, turnos y comprobantes sin tocar recetas.",
        "notes": "Separar visualizacion ciudadana en Mi Argentina de acceso privado a datos de prescripcion.",
    },
    {
        "id": "FIN-001",
        "category": "finanzas",
        "question": "Cuanto capital de trabajo requiere pagar al profesional en el dia?",
        "priority": "4",
        "status": "open",
        "decision_if_true": "Diseñar limite de adelantos y financiamiento.",
        "decision_if_false": "No prometer cobro mismo dia salvo pago privado confirmado.",
        "notes": "Formula: consultas x honorario x demora / 30 x porcentaje adelantado.",
    },
    {
        "id": "AI-001",
        "category": "ia",
        "question": "Que orientacion por sintomas es segura sin convertirse en diagnostico?",
        "priority": "4",
        "status": "open",
        "decision_if_true": "Construir orientador no diagnostico con derivacion y disclaimer.",
        "decision_if_false": "Usar formulario manual y derivacion humana.",
        "notes": "Buscar evidencia de triage digital y riesgos de LLM clinico.",
    },
    {
        "id": "GTM-001",
        "category": "go_to_market",
        "question": "Que canal permite conseguir 20 profesionales verificados sin pagar marketing?",
        "priority": "4",
        "status": "open",
        "decision_if_true": "Concentrar lanzamiento por comunidad/canal.",
        "decision_if_false": "Cambiar vertical o propuesta al profesional.",
        "notes": "Colegios, asociaciones, grupos, recomendaciones.",
    },
]


DEFAULT_SOURCES = [
    {
        "id": "SRC-INDEC-CENSO-2022",
        "date": "2026-05-18",
        "category": "mercado",
        "source_type": "oficial",
        "title": "INDEC Censo 2022 - cobertura de salud",
        "url": "https://biblioteca.indec.gob.ar/bases/minde/1c2022_2.pdf",
        "credibility": "5",
        "finding": "Argentina tenia 45,9M personas; 60,9% con obra social/prepaga/PAMI y 35,8% solo sistema publico.",
        "next_action": "Usar como base TAM nacional.",
    },
    {
        "id": "SRC-INDEC-EPH-2025",
        "date": "2026-05-18",
        "category": "mercado",
        "source_type": "oficial",
        "title": "INDEC EPH primer semestre 2025 - cobertura urbana",
        "url": "https://www.indec.gob.ar/uploads/informesdeprensa/eph_indicadores_hogares_11_2514558E0A45.pdf",
        "credibility": "5",
        "finding": "En 31 aglomerados, 66,5% tiene obra social/prepaga/mutual/emergencia.",
        "next_action": "Cruzar con AMBA y segmento 18-64.",
    },
    {
        "id": "SRC-SSS-BENEFICIARIOS-2026",
        "date": "2026-05-18",
        "category": "financiadores",
        "source_type": "oficial",
        "title": "SSS poblacion beneficiaria por provincia",
        "url": "https://www.sssalud.gob.ar/index.php?cat=consultas&page=poblacion",
        "credibility": "5",
        "finding": "Al 01-04-2026, SSS informa 20,25M beneficiarios; Buenos Aires + CABA concentran mas de 10,5M.",
        "next_action": "Dimensionar SAM AMBA.",
    },
    {
        "id": "SRC-PBA-EPAUSS-2023",
        "date": "2026-05-18",
        "category": "uso_servicios",
        "source_type": "oficial/provincial",
        "title": "EPAUSS PBA 2023 - acceso y uso de servicios de salud",
        "url": "https://www.ioma.gba.gob.ar/wp-content/uploads/2024/07/Informe-EPAUSS-Ministerio-de-Salud-1.pdf",
        "credibility": "4",
        "finding": "61,1% de la poblacion bonaerense solicito consultas medicas en los ultimos 12 meses.",
        "next_action": "Usar como proxy hasta hallar fuente nacional de consultas.",
    },
    {
        "id": "SRC-GASTO-SALUD-2022",
        "date": "2026-05-18",
        "category": "dinero",
        "source_type": "oficial",
        "title": "Informe gasto en salud Argentina 2022",
        "url": "https://www.argentina.gob.ar/sites/default/files/2025/05/informe_gasto_en_salud_2022_dnesa.pdf",
        "credibility": "5",
        "finding": "Gasto total estimado USD 66.367M; obras sociales USD 21.203M; gasto privado USD 27.740M.",
        "next_action": "Estimar oportunidad por flujo privado y seguridad social.",
    },
    {
        "id": "SRC-SSS-PLAN-CUENTAS-1430",
        "date": "2026-05-18",
        "category": "dinero",
        "source_type": "normativa",
        "title": "Resolucion 1430/2010 - plan de cuentas Agentes del Seguro de Salud",
        "url": "https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-1430-2010-176808/texto",
        "credibility": "5",
        "finding": "El plan de cuentas desagrega gasto en salud mental, rehabilitacion, odontologia, alta complejidad, medicamentos y otros rubros.",
        "next_action": "Pedir/analisar estados contables o agregados por cuenta.",
    },
    {
        "id": "SRC-MSAL-MIARG-RECETA-2026",
        "date": "2026-05-20",
        "category": "regulacion",
        "source_type": "oficial",
        "title": "Ministerio de Salud - Receta Electronica en Mi Argentina",
        "url": "https://www.argentina.gob.ar/noticias/la-receta-electronica-se-incorpora-la-app-mi-argentina",
        "credibility": "5",
        "finding": "Mi Argentina permite consultar prescripciones electronicas de medicamentos emitidas por plataformas ReNaPDiS durante los 60 dias previos, con autorizacion del usuario.",
        "next_action": "Evaluar si el MVP debe mostrar solo informacion educativa o planificar integracion futura con consentimiento explicito.",
    },
    {
        "id": "SRC-MSAL-RECETA-ELECTRONICA",
        "date": "2026-05-20",
        "category": "regulacion",
        "source_type": "oficial",
        "title": "Ministerio de Salud - Receta Electronica",
        "url": "https://www.argentina.gob.ar/salud/digital/renapdis/receta-electronica",
        "credibility": "5",
        "finding": "La receta electronica es la modalidad vigente para medicamentos y las plataformas deben estar inscriptas y aprobadas por ReNaPDiS.",
        "next_action": "No incorporar prescripcion en el MVP sin confirmar encuadre legal, tecnico y operativo.",
    },
]


DEFAULT_COMPETITORS = [
    {
        "name": "Doctoralia",
        "geography": "Argentina / global",
        "model": "Directorio, agenda y visibilidad para profesionales",
        "target": "Pacientes y profesionales",
        "pricing_signal": "Modelo pago al profesional por presencia/funcionalidades",
        "strengths": "Marca, SEO, base de profesionales, reseñas",
        "weaknesses": "No resuelve liquidacion con financiadores ni autonomia economica completa",
        "relevance": "5",
        "url": "https://www.doctoralia.com.ar/",
        "notes": "Benchmark directo de busqueda y reserva.",
    },
    {
        "name": "Docturno",
        "geography": "Argentina",
        "model": "Turnos medicos online",
        "target": "Pacientes y centros/profesionales",
        "pricing_signal": "No confirmado",
        "strengths": "Foco local en turnos",
        "weaknesses": "Diferencial limitado si no gestiona pago/cobertura",
        "relevance": "4",
        "url": "https://www.docturno.com/",
        "notes": "Competidor local de experiencia de agenda.",
    },
    {
        "name": "Zocdoc",
        "geography": "Estados Unidos",
        "model": "Marketplace de busqueda y reserva con seguros",
        "target": "Pacientes y providers",
        "pricing_signal": "Fees a providers",
        "strengths": "UX de busqueda, disponibilidad, insurance filters",
        "weaknesses": "Modelo depende del sistema de seguros de EEUU",
        "relevance": "4",
        "url": "https://www.zocdoc.com/",
        "notes": "Benchmark de marketplace maduro.",
    },
    {
        "name": "Headway",
        "geography": "Estados Unidos",
        "model": "Red de salud mental que ayuda a terapeutas a aceptar seguros",
        "target": "Terapeutas y pacientes",
        "pricing_signal": "Intermediacion con aseguradoras",
        "strengths": "Resuelve credencializacion, seguros y cobro",
        "weaknesses": "Sistema regulatorio distinto",
        "relevance": "5",
        "url": "https://headway.co/",
        "notes": "Analogico fuerte para salud mental + financiadores.",
    },
    {
        "name": "Alma",
        "geography": "Estados Unidos",
        "model": "Red/comunidad para terapeutas independientes con seguros",
        "target": "Terapeutas",
        "pricing_signal": "Membresia/servicios",
        "strengths": "Propuesta al profesional independiente",
        "weaknesses": "Adaptacion local incierta",
        "relevance": "4",
        "url": "https://helloalma.com/",
        "notes": "Referencia de autonomia profesional en salud mental.",
    },
]


DEFAULT_REGULATIONS = [
    {
        "id": "REG-25326",
        "jurisdiction": "Argentina",
        "norm": "Ley 25.326",
        "topic": "Datos personales sensibles",
        "impact": "Datos de salud requieren consentimiento y controles fuertes.",
        "risk_level": "5",
        "url": "https://www.argentina.gob.ar/normativa/nacional/64790/actualizacion",
        "status": "vigente",
        "notes": "No guardar historia clinica completa en MVP si no es necesario.",
    },
    {
        "id": "REG-26529",
        "jurisdiction": "Argentina",
        "norm": "Ley 26.529",
        "topic": "Derechos del paciente e historia clinica",
        "impact": "Define consentimiento, confidencialidad e historia clinica.",
        "risk_level": "5",
        "url": "https://www.argentina.gob.ar/normativa/nacional/160432/actualizacion",
        "status": "vigente",
        "notes": "Separar orientacion administrativa de acto clinico.",
    },
    {
        "id": "REG-27553",
        "jurisdiction": "Argentina",
        "norm": "Ley 27.553",
        "topic": "Receta electronica y teleasistencia",
        "impact": "Marco para teleasistencia y recetas digitales/electronicas.",
        "risk_level": "4",
        "url": "https://www.argentina.gob.ar/normativa/nacional/ley-27553-340919/texto",
        "status": "vigente",
        "notes": "Evitar receta en MVP salvo integracion formal.",
    },
    {
        "id": "REG-1959-2024",
        "jurisdiction": "Argentina",
        "norm": "Resolucion 1959/2024",
        "topic": "ReNaPDiS",
        "impact": "Registro para plataformas digitales sanitarias.",
        "risk_level": "4",
        "url": "https://www.argentina.gob.ar/normativa/nacional/400825/texto",
        "status": "vigente",
        "notes": "Analizar si aplica al alcance MVP.",
    },
    {
        "id": "REG-5744-2024",
        "jurisdiction": "Argentina",
        "norm": "Resolucion 5744/2024",
        "topic": "Repositorios de receta electronica",
        "impact": "Exige interoperabilidad y APIs para que plataformas y farmacias accedan a repositorios segun cobertura del paciente.",
        "risk_level": "4",
        "url": "https://www.argentina.gob.ar/normativa/nacional/406757/texto",
        "status": "vigente",
        "notes": "No asumir acceso privado a recetas; revisar alcance por cobertura, farmacia, repositorio y consentimiento.",
    },
    {
        "id": "REG-2214-2025",
        "jurisdiction": "Argentina",
        "norm": "Resolucion 2214/2025",
        "topic": "Extension de receta electronica",
        "impact": "Extiende la receta electronica a indicaciones medicas, practicas, procedimientos y dispositivos.",
        "risk_level": "4",
        "url": "https://www.argentina.gob.ar/normativa/nacional/415349/texto",
        "status": "vigente",
        "notes": "Impacta roadmap si la plataforma evoluciona hacia ordenes, estudios o red prestacional.",
    },
    {
        "id": "REG-DISP-1-2025",
        "jurisdiction": "Argentina",
        "norm": "Disposicion 1/2025 DNSISA",
        "topic": "CUIR y visualizacion en Mi Argentina",
        "impact": "Define CUIR, datos minimos enviados al Ministerio y endpoint de repositorios para visualizacion de recetas en Mi Argentina.",
        "risk_level": "5",
        "url": "https://www.argentina.gob.ar/normativa/nacional/disposici%C3%B3n-1-2025-415504/texto",
        "status": "vigente",
        "notes": "Cualquier integracion requiere consentimiento, minimizacion de datos, autenticacion fuerte y revision legal.",
    },
    {
        "id": "REG-1430-2010",
        "jurisdiction": "Argentina",
        "norm": "Resolucion 1430/2010",
        "topic": "Plan de cuentas obras sociales",
        "impact": "Permite mapear rubros economicos de cobertura.",
        "risk_level": "3",
        "url": "https://www.argentina.gob.ar/normativa/nacional/resoluci%C3%B3n-1430-2010-176808/texto",
        "status": "vigente",
        "notes": "Base para pedir datos por gasto/rubro.",
    },
    {
        "id": "REG-23660-23661",
        "jurisdiction": "Argentina",
        "norm": "Leyes 23.660 y 23.661",
        "topic": "Obras sociales y Sistema Nacional del Seguro de Salud",
        "impact": "Marco de financiadores, prestadores y Fondo Solidario.",
        "risk_level": "5",
        "url": "https://www.sssalud.gob.ar/index/normativas/consulta/000437.pdf",
        "status": "vigente",
        "notes": "Confirmar Red de Prestadores con experto legal.",
    },
]


DEFAULT_TOOLS = [
    {
        "name": "Google Forms / Typeform",
        "category": "validacion",
        "use_case": "Capturar entrevistas, demanda y solicitudes de turno sin construir app.",
        "build_or_buy": "buy/free",
        "cost_signal": "bajo",
        "url": "https://www.google.com/forms/about/",
        "notes": "Suficiente para etapa concierge.",
    },
    {
        "name": "Calendly / Google Calendar",
        "category": "agenda",
        "use_case": "Coordinar disponibilidad inicial con profesionales.",
        "build_or_buy": "buy/free",
        "cost_signal": "bajo",
        "url": "https://calendly.com/",
        "notes": "Evita construir agenda propia al inicio.",
    },
    {
        "name": "Mercado Pago",
        "category": "pagos",
        "use_case": "Cobrar consultas privadas/coseguros y emitir comprobante operativo.",
        "build_or_buy": "buy",
        "cost_signal": "variable por transaccion",
        "url": "https://www.mercadopago.com.ar/",
        "notes": "Revisar split payments y obligaciones fiscales.",
    },
    {
        "name": "WhatsApp Business",
        "category": "concierge",
        "use_case": "Atencion humana, orientacion inicial y coordinacion de beta.",
        "build_or_buy": "buy/free",
        "cost_signal": "bajo",
        "url": "https://www.whatsapp.com/business/",
        "notes": "Canal natural para MVP manual.",
    },
    {
        "name": "Jitsi Meet",
        "category": "teleconsulta",
        "use_case": "Videollamada simple para beta sin construir video propio.",
        "build_or_buy": "buy/free/open",
        "cost_signal": "bajo",
        "url": "https://jitsi.org/",
        "notes": "Validar privacidad y consentimiento antes de uso clinico.",
    },
    {
        "name": "n8n",
        "category": "automatizacion",
        "use_case": "Automatizar carga de formularios, recordatorios y reportes.",
        "build_or_buy": "buy/open",
        "cost_signal": "bajo/medio",
        "url": "https://n8n.io/",
        "notes": "Util para operar solo.",
    },
]


def today() -> str:
    return dt.datetime.now().strftime("%Y-%m-%d")


def timestamp() -> str:
    return dt.datetime.now().strftime("%Y-%m-%d_%H%M%S")


def ensure_file(path: Path) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    if not path.exists():
        with path.open("w", newline="", encoding="utf-8") as f:
            writer = csv.DictWriter(f, fieldnames=CSV_HEADERS[path])
            writer.writeheader()


def append_row(path: Path, row: dict[str, str]) -> None:
    ensure_file(path)
    with path.open("a", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=CSV_HEADERS[path])
        writer.writerow(row)


def write_rows(path: Path, rows: list[dict[str, str]]) -> None:
    ensure_file(path)
    with path.open("w", newline="", encoding="utf-8") as f:
        writer = csv.DictWriter(f, fieldnames=CSV_HEADERS[path])
        writer.writeheader()
        writer.writerows(rows)


def read_rows(path: Path) -> list[dict[str, str]]:
    ensure_file(path)
    with path.open("r", newline="", encoding="utf-8") as f:
        return list(csv.DictReader(f))


def append_unique_rows(path: Path, rows: list[dict[str, str]], key: str) -> int:
    existing = {row.get(key, "") for row in read_rows(path)}
    created = 0
    for row in rows:
        if row.get(key, "") in existing:
            continue
        append_row(path, row)
        existing.add(row.get(key, ""))
        created += 1
    return created


def init(_: argparse.Namespace) -> None:
    DATA.mkdir(exist_ok=True)
    REPORTS.mkdir(parents=True, exist_ok=True)
    for path in CSV_HEADERS:
        ensure_file(path)
    if len(read_rows(VERTICALS)) == 0:
        for row in DEFAULT_VERTICALS:
            append_row(VERTICALS, row)
    print(f"Listo. Base creada en: {DATA}")


def seed_research(_: argparse.Namespace) -> None:
    init(argparse.Namespace())
    created = {
        "questions": append_unique_rows(QUESTIONS, DEFAULT_QUESTIONS, "id"),
        "sources": append_unique_rows(SOURCES, DEFAULT_SOURCES, "id"),
        "competitors": append_unique_rows(COMPETITORS, DEFAULT_COMPETITORS, "name"),
        "regulations": append_unique_rows(REGULATIONS, DEFAULT_REGULATIONS, "id"),
        "tools": append_unique_rows(TOOLS, DEFAULT_TOOLS, "name"),
    }
    print("Semilla de investigacion cargada:")
    for name, count in created.items():
        print(f"- {name}: {count} nuevos")


def add_interview(args: argparse.Namespace) -> None:
    append_row(
        INTERVIEWS,
        {
            "date": today(),
            "kind": args.kind,
            "segment": args.segment,
            "name": args.name,
            "pain": str(args.pain),
            "urgency": str(args.urgency),
            "willingness": str(args.willingness),
            "notes": args.notes,
        },
    )
    print("Entrevista registrada.")


def add_evidence(args: argparse.Namespace) -> None:
    append_row(
        EVIDENCE,
        {
            "date": today(),
            "claim": args.claim,
            "source": args.source,
            "confidence": str(args.confidence),
            "notes": args.notes,
        },
    )
    print("Evidencia registrada.")


def add_experiment(args: argparse.Namespace) -> None:
    append_row(
        EXPERIMENTS,
        {
            "date": today(),
            "name": args.name,
            "hypothesis": args.hypothesis,
            "metric": args.metric,
            "target": args.target,
            "status": args.status,
            "result": args.result,
            "notes": args.notes,
        },
    )
    print("Experimento registrado.")


def add_question(args: argparse.Namespace) -> None:
    append_row(
        QUESTIONS,
        {
            "id": args.id,
            "category": args.category,
            "question": args.question,
            "priority": str(args.priority),
            "status": args.status,
            "decision_if_true": args.decision_if_true,
            "decision_if_false": args.decision_if_false,
            "notes": args.notes,
        },
    )
    print("Pregunta registrada.")


def list_questions(args: argparse.Namespace) -> None:
    rows = read_rows(QUESTIONS)
    if args.status != "all":
        rows = [row for row in rows if row["status"] == args.status]
    rows = sorted(rows, key=lambda row: parse_int(row["priority"]), reverse=True)
    print("Preguntas de investigacion")
    print("==========================")
    for row in rows:
        print(f"{row['id']} [{row['category']}] P{row['priority']} {row['status']}")
        print(f"  {row['question']}")
        if row["notes"]:
            print(f"  Nota: {row['notes']}")


def add_source(args: argparse.Namespace) -> None:
    append_row(
        SOURCES,
        {
            "id": args.id,
            "date": today(),
            "category": args.category,
            "source_type": args.source_type,
            "title": args.title,
            "url": args.url,
            "credibility": str(args.credibility),
            "finding": args.finding,
            "next_action": args.next_action,
        },
    )
    print("Fuente registrada.")


def add_competitor(args: argparse.Namespace) -> None:
    append_row(
        COMPETITORS,
        {
            "name": args.name,
            "geography": args.geography,
            "model": args.model,
            "target": args.target,
            "pricing_signal": args.pricing_signal,
            "strengths": args.strengths,
            "weaknesses": args.weaknesses,
            "relevance": str(args.relevance),
            "url": args.url,
            "notes": args.notes,
        },
    )
    print("Competidor registrado.")


def add_regulation(args: argparse.Namespace) -> None:
    append_row(
        REGULATIONS,
        {
            "id": args.id,
            "jurisdiction": args.jurisdiction,
            "norm": args.norm,
            "topic": args.topic,
            "impact": args.impact,
            "risk_level": str(args.risk_level),
            "url": args.url,
            "status": args.status,
            "notes": args.notes,
        },
    )
    print("Norma registrada.")


def add_tool(args: argparse.Namespace) -> None:
    append_row(
        TOOLS,
        {
            "name": args.name,
            "category": args.category,
            "use_case": args.use_case,
            "build_or_buy": args.build_or_buy,
            "cost_signal": args.cost_signal,
            "url": args.url,
            "notes": args.notes,
        },
    )
    print("Herramienta registrada.")


def fetch_json(url: str) -> dict:
    request = urllib.request.Request(url, headers={"User-Agent": "salud-startup-lab/0.1"})
    with urllib.request.urlopen(request, timeout=45) as response:
        return json.loads(response.read().decode("utf-8"))


def fetch_text(url: str) -> str:
    request = urllib.request.Request(url, headers={"User-Agent": "salud-startup-lab/0.1"})
    with urllib.request.urlopen(request, timeout=45) as response:
        return response.read().decode("utf-8")


def extract_response_text(body: dict) -> str:
    if "output_text" in body:
        return str(body["output_text"]).strip()

    chunks: list[str] = []
    for item in body.get("output", []):
        for content in item.get("content", []):
            if content.get("type") in {"output_text", "text"}:
                chunks.append(content.get("text", ""))
    return "\n".join(chunks).strip()


def extract_web_sources(body: dict) -> list[dict[str, str]]:
    sources: list[dict[str, str]] = []
    seen: set[str] = set()

    for item in body.get("output", []):
        action = item.get("action") if isinstance(item, dict) else None
        if isinstance(action, dict):
            for source in action.get("sources", []) or []:
                url = source.get("url") or source.get("uri") or ""
                if url and url not in seen:
                    sources.append(
                        {
                            "title": source.get("title", ""),
                            "url": url,
                        }
                    )
                    seen.add(url)

        for content in item.get("content", []) if isinstance(item, dict) else []:
            for annotation in content.get("annotations", []) or []:
                if annotation.get("type") != "url_citation":
                    continue
                url = annotation.get("url", "")
                if url and url not in seen:
                    sources.append(
                        {
                            "title": annotation.get("title", ""),
                            "url": url,
                        }
                    )
                    seen.add(url)

    return sources


def call_openai_with_web(
    prompt: str,
    model: str,
    domain_bundle: str,
    search_context_size: str,
    reasoning_effort: str,
) -> dict:
    api_key = os.getenv("OPENAI_API_KEY")
    if not api_key:
        raise RuntimeError("OPENAI_API_KEY no configurada.")

    domains = DOMAIN_BUNDLES.get(domain_bundle, [])
    tool: dict[str, object] = {
        "type": "web_search",
        "search_context_size": search_context_size,
        "user_location": {
            "type": "approximate",
            "country": "AR",
            "city": "Buenos Aires",
            "region": "Buenos Aires",
            "timezone": "America/Argentina/Buenos_Aires",
        },
    }
    if domains:
        tool["filters"] = {"allowed_domains": domains}

    payload: dict[str, object] = {
        "model": model,
        "reasoning": {"effort": reasoning_effort},
        "tools": [tool],
        "tool_choice": "auto",
        "include": ["web_search_call.action.sources"],
        "input": prompt,
    }

    request = urllib.request.Request(
        "https://api.openai.com/v1/responses",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=180) as response:
        return json.loads(response.read().decode("utf-8"))


def decode_output(raw: bytes | None) -> str:
    if not raw:
        return ""
    for encoding in ("utf-8", "cp1252", "latin-1"):
        try:
            return raw.decode(encoding)
        except UnicodeDecodeError:
            continue
    return raw.decode("utf-8", errors="replace")


def call_claude_cli(
    prompt: str,
    model: str | None,
    max_turns: int,
    timeout_seconds: int,
) -> str:
    command = [
        *claude_command(),
        "-p",
        "Ejecuta la tarea indicada en el contenido recibido por stdin. Usa busqueda web si esta disponible y cita URLs.",
        "--max-turns",
        str(max_turns),
        "--output-format",
        "text",
        "--permission-mode",
        "auto",
        "--allowedTools",
        "WebSearch,WebFetch",
    ]
    if model:
        command.extend(["--model", model])

    result = subprocess.run(
        command,
        input=prompt.encode("utf-8"),
        capture_output=True,
        cwd=str(ROOT),
        timeout=timeout_seconds,
        check=False,
    )
    if result.returncode != 0:
        detail = (decode_output(result.stderr) or decode_output(result.stdout)).strip()
        raise RuntimeError(f"Claude CLI fallo con codigo {result.returncode}: {detail}")
    return decode_output(result.stdout).strip()


def mask_secret(value: str) -> str:
    if len(value) <= 12:
        return "***"
    return f"{value[:7]}...{value[-4:]}"


def check_api(args: argparse.Namespace) -> None:
    api_key = os.getenv("OPENAI_API_KEY", "")
    model = os.getenv("OPENAI_MODEL", "gpt-5")
    env_path = ROOT / ".env"

    if not api_key or api_key == "pega_tu_api_key_aca":
        print("OPENAI_API_KEY todavia no esta configurada.")
        print(f"Edita este archivo y pega tu key: {env_path}")
        print("Formato esperado:")
        print("OPENAI_API_KEY=sk-proj-...")
        print("OPENAI_MODEL=gpt-5")
        return

    print(f"OPENAI_API_KEY detectada: {mask_secret(api_key)}")
    print(f"Modelo configurado: {model}")

    if not args.live:
        print("Chequeo local OK. Para probar contra OpenAI: python .\\startup_ai_lab.py check-api --live")
        return

    payload = {
        "model": model,
        "input": "Responde solamente: ok",
    }
    request = urllib.request.Request(
        "https://api.openai.com/v1/responses",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )

    try:
        with urllib.request.urlopen(request, timeout=60) as response:
            body = json.loads(response.read().decode("utf-8"))
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        print(f"La API respondio con error HTTP {exc.code}.")
        print(detail[:1000])
        return
    except urllib.error.URLError as exc:
        print(f"No pude conectar con OpenAI: {exc}")
        return

    text = extract_response_text(body)
    print(f"Prueba live OK. Respuesta: {text}")


def check_claude(args: argparse.Namespace) -> None:
    try:
        version = subprocess.run(
            [*claude_command(), "--version"],
            text=True,
            capture_output=True,
            cwd=str(ROOT),
            timeout=30,
            check=False,
        )
    except (FileNotFoundError, subprocess.TimeoutExpired) as exc:
        print(f"No pude encontrar Claude Code CLI: {exc}")
        return

    if version.returncode != 0:
        print("Claude Code CLI existe, pero no pude leer version.")
        print((version.stderr or version.stdout).strip())
        return

    print(f"Claude Code CLI: {version.stdout.strip()}")

    auth = subprocess.run(
        [*claude_command(), "auth", "status"],
        text=True,
        capture_output=True,
        cwd=str(ROOT),
        timeout=30,
        check=False,
    )
    if auth.returncode == 0:
        print("Auth status:")
        print(auth.stdout.strip())
    else:
        print("No pude leer auth status:")
        print((auth.stderr or auth.stdout).strip())

    if not args.live:
        print("Chequeo local OK. Para probar respuesta: python .\\startup_ai_lab.py check-claude --live")
        return

    try:
        text = call_claude_cli(
            prompt="Responde solamente: ok",
            model=args.model,
            max_turns=1,
            timeout_seconds=90,
        )
    except (RuntimeError, subprocess.TimeoutExpired) as exc:
        print(f"No pude ejecutar Claude CLI en modo live: {exc}")
        return

    print(f"Prueba live OK. Respuesta: {text}")


def search_pubmed(query: str, limit: int) -> list[dict[str, str]]:
    params = urllib.parse.urlencode(
        {"db": "pubmed", "term": query, "retmode": "json", "retmax": str(limit)}
    )
    search_url = f"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esearch.fcgi?{params}"
    search = fetch_json(search_url)
    ids = search.get("esearchresult", {}).get("idlist", [])
    if not ids:
        return []

    summary_params = urllib.parse.urlencode(
        {"db": "pubmed", "id": ",".join(ids), "retmode": "json"}
    )
    summary_url = f"https://eutils.ncbi.nlm.nih.gov/entrez/eutils/esummary.fcgi?{summary_params}"
    summary = fetch_json(summary_url).get("result", {})

    results = []
    for uid in ids:
        item = summary.get(uid, {})
        authors = ", ".join(a.get("name", "") for a in item.get("authors", [])[:5])
        pubdate = item.get("pubdate", "")
        year = pubdate[:4] if pubdate else ""
        results.append(
            {
                "source": "pubmed",
                "query": query,
                "title": item.get("title", "").strip(),
                "authors": authors,
                "year": year,
                "url": f"https://pubmed.ncbi.nlm.nih.gov/{uid}/",
                "abstract": "",
                "notes": item.get("fulljournalname", ""),
            }
        )
    return results


def search_arxiv(query: str, limit: int) -> list[dict[str, str]]:
    params = urllib.parse.urlencode(
        {
            "search_query": f"all:{query}",
            "start": "0",
            "max_results": str(limit),
            "sortBy": "relevance",
            "sortOrder": "descending",
        }
    )
    url = f"https://export.arxiv.org/api/query?{params}"
    raw = fetch_text(url)
    root = ET.fromstring(raw)
    ns = {"atom": "http://www.w3.org/2005/Atom"}

    results = []
    for entry in root.findall("atom:entry", ns):
        title = " ".join((entry.findtext("atom:title", "", ns) or "").split())
        abstract = " ".join((entry.findtext("atom:summary", "", ns) or "").split())
        published = entry.findtext("atom:published", "", ns) or ""
        authors = ", ".join(
            a.findtext("atom:name", "", ns) or "" for a in entry.findall("atom:author", ns)[:5]
        )
        link = entry.findtext("atom:id", "", ns) or ""
        results.append(
            {
                "source": "arxiv",
                "query": query,
                "title": title,
                "authors": authors,
                "year": published[:4],
                "url": link,
                "abstract": abstract,
                "notes": "",
            }
        )
    return results


def paper_search(args: argparse.Namespace) -> None:
    ensure_file(PAPERS)
    sources = ["pubmed", "arxiv"] if args.source == "all" else [args.source]
    results: list[dict[str, str]] = []

    for source in sources:
        try:
            if source == "pubmed":
                results.extend(search_pubmed(args.query, args.limit))
            elif source == "arxiv":
                results.extend(search_arxiv(args.query, args.limit))
        except (urllib.error.URLError, TimeoutError, json.JSONDecodeError, ET.ParseError) as exc:
            print(f"No pude consultar {source}: {exc}")

    seen_urls = {row["url"] for row in read_rows(PAPERS)}
    saved = 0
    for result in results:
        if not result["title"] or result["url"] in seen_urls:
            continue
        append_row(PAPERS, {"date": today(), **result})
        seen_urls.add(result["url"])
        saved += 1

    print(f"Resultados nuevos guardados: {saved}")
    for result in results[: args.limit]:
        print(f"- [{result['source']}] {result['year']} - {result['title']}")
        print(f"  {result['url']}")


def parse_int(value: str, default: int = 0) -> int:
    try:
        return int(value)
    except (TypeError, ValueError):
        return default


def score_vertical(row: dict[str, str]) -> int:
    """Higher is better. Regulatory friction subtracts from the score."""

    return (
        parse_int(row["patient_pain"]) * 3
        + parse_int(row["professional_pain"]) * 3
        + parse_int(row["recurrence"]) * 2
        + parse_int(row["founder_edge"]) * 2
        - parse_int(row["regulatory_friction"]) * 2
    )


def score(_: argparse.Namespace) -> None:
    rows = sorted(read_rows(VERTICALS), key=score_vertical, reverse=True)
    print("Ranking de verticales")
    print("====================")
    for row in rows:
        print(f"{row['vertical']}: {score_vertical(row)} puntos")
        print(f"  {row['notes']}")


def build_scout_plan() -> str:
    questions = sorted(
        [row for row in read_rows(QUESTIONS) if row["status"] == "open"],
        key=lambda row: parse_int(row["priority"]),
        reverse=True,
    )
    sources = sorted(read_rows(SOURCES), key=lambda row: parse_int(row["credibility"]), reverse=True)
    competitors = sorted(
        read_rows(COMPETITORS), key=lambda row: parse_int(row["relevance"]), reverse=True
    )
    regulations = sorted(
        read_rows(REGULATIONS), key=lambda row: parse_int(row["risk_level"]), reverse=True
    )

    lines = [
        f"# Scout plan - {today()}",
        "",
        "## Objetivo",
        "",
        "Encontrar evidencia que reduzca incertidumbre antes de construir producto.",
        "",
        "## Preguntas prioritarias",
        "",
    ]
    for row in questions[:6]:
        lines.append(f"- {row['id']} / P{row['priority']}: {row['question']}")

    lines.extend(
        [
            "",
            "## Busquedas sugeridas",
            "",
            "### Mercado y dinero",
            "",
            '- "Superintendencia Servicios de Salud gasto prestaciones salud mental odontologia rehabilitacion"',
            '- "INDEC cobertura salud obra social prepaga AMBA edad ingresos"',
            '- "reclamos Superintendencia Servicios de Salud reintegros turnos copagos"',
            '- "estado contable obra social prestaciones salud mental rehabilitacion odontologia"',
            "",
            "### Papers y seguridad IA",
            "",
            'python .\\startup_ai_lab.py paper-search --source pubmed --query "digital triage primary care patient navigation" --limit 5',
            'python .\\startup_ai_lab.py paper-search --source pubmed --query "telemedicine mental health access waiting times" --limit 5',
            'python .\\startup_ai_lab.py paper-search --source arxiv --query "large language models clinical triage safety" --limit 5',
            "",
            "### Competidores y analogos",
            "",
        ]
    )
    for row in competitors[:5]:
        lines.append(f"- {row['name']}: estudiar modelo, pricing, onboarding y promesa. {row['url']}")

    lines.extend(
        [
            "",
            "### Regulacion critica",
            "",
        ]
    )
    for row in regulations[:5]:
        lines.append(f"- {row['norm']} / {row['topic']}: {row['impact']} {row['url']}")

    lines.extend(
        [
            "",
            "## Fuentes ya cargadas para no perder",
            "",
        ]
    )
    for row in sources[:5]:
        lines.append(f"- {row['title']}: {row['finding']}")

    lines.extend(
        [
            "",
            "## Checklist de esta semana",
            "",
            "- Cargar 5 entrevistas de pacientes.",
            "- Cargar 5 entrevistas de profesionales.",
            "- Guardar 5 fuentes nuevas sobre dinero/coberturas.",
            "- Revisar 3 competidores/analogos y registrar huecos.",
            "- Convertir 1 hallazgo en experimento concreto.",
        ]
    )

    return "\n".join(lines) + "\n"


def scout(args: argparse.Namespace) -> None:
    report = build_scout_plan()
    REPORTS.mkdir(parents=True, exist_ok=True)
    path = REPORTS / f"scout_plan_{today()}.md"
    path.write_text(report, encoding="utf-8")
    print(report)
    print(f"Plan guardado en: {path}")


def get_question(question_id: str | None, question_text: str | None) -> tuple[str, str]:
    if question_text:
        return question_id or "custom", question_text

    questions = read_rows(QUESTIONS)
    if question_id:
        for row in questions:
            if row["id"] == question_id:
                return row["id"], row["question"]
        raise ValueError(f"No encontre la pregunta {question_id}. Proba `questions`.")

    open_questions = [row for row in questions if row["status"] == "open"]
    if not open_questions:
        raise ValueError("No hay preguntas abiertas. Ejecuta `seed` o carga una pregunta.")
    row = sorted(open_questions, key=lambda item: parse_int(item["priority"]), reverse=True)[0]
    return row["id"], row["question"]


def build_auto_research_prompt(question_id: str, question: str, domain_bundle: str) -> str:
    program = (ROOT / "program.md").read_text(encoding="utf-8")
    known_sources = read_rows(SOURCES)
    known_regs = read_rows(REGULATIONS)
    known_competitors = read_rows(COMPETITORS)

    source_lines = "\n".join(
        f"- {row['title']}: {row['finding']} ({row['url']})" for row in known_sources[-8:]
    )
    regulation_lines = "\n".join(
        f"- {row['norm']} / {row['topic']}: {row['impact']} ({row['url']})"
        for row in known_regs[-8:]
    )
    competitor_lines = "\n".join(
        f"- {row['name']}: {row['model']}; hueco: {row['weaknesses']} ({row['url']})"
        for row in known_competitors[-8:]
    )

    return textwrap.dedent(
        f"""
        Sos un research agent autonomo para una startup healthtech argentina.
        Investiga usando internet y razonamiento. No inventes datos. Cuando un
        dato sea regulatorio, financiero o de mercado, citalo con URL. Si una
        fuente no alcanza para concluir, decilo.

        Pregunta de investigacion:
        {question_id}: {question}

        Dominio de busqueda configurado:
        {domain_bundle}

        Contexto estrategico del proyecto:
        {program}

        Fuentes ya conocidas:
        {source_lines or "- Sin fuentes estructuradas cargadas."}

        Regulacion ya conocida:
        {regulation_lines or "- Sin regulacion cargada."}

        Competidores/analogos ya conocidos:
        {competitor_lines or "- Sin competidores cargados."}

        Entregá un reporte en Markdown con esta estructura exacta:

        # Auto research - {question_id}

        ## Veredicto
        Una respuesta breve y honesta a la pregunta.

        ## Hallazgos con fuentes
        Bullets con dato, fuente y por que importa.

        ## Implicancias para la startup
        Que cambia en producto, mercado, legal, finanzas o go-to-market.

        ## Riesgos e incertidumbres
        Lo que sigue sin estar probado.

        ## Proximo experimento
        Una accion concreta de bajo costo para validar o refutar.

        ## Fuentes consultadas
        Lista de URLs usadas.
        """
    ).strip()


def normalize_text(value: str) -> str:
    decomposed = unicodedata.normalize("NFKD", value)
    ascii_text = "".join(char for char in decomposed if not unicodedata.combining(char))
    return ascii_text.lower()


def question_tokens(question: str) -> set[str]:
    normalized = normalize_text(question)
    tokens = {
        token
        for token in re.findall(r"[a-z0-9]+", normalized)
        if len(token) >= 4 and token not in STOPWORDS
    }
    return tokens


def row_blob(row: dict[str, str]) -> str:
    return normalize_text(" ".join(str(value) for value in row.values()))


def compact_text(value: str, limit: int = 220) -> str:
    cleaned = re.sub(r"\s+", " ", value).strip(" -\t\r\n")
    cleaned = re.sub(r"\[([^\]]+)\]\([^)]+\)", r"\1", cleaned)
    if len(cleaned) <= limit:
        return cleaned
    return cleaned[: limit - 3].rstrip() + "..."


def rank_rows_for_question(
    rows: list[dict[str, str]],
    tokens: set[str],
    weight_field: str | None = None,
) -> list[dict[str, str]]:
    scored: list[tuple[int, dict[str, str]]] = []
    for row in rows:
        blob = row_blob(row)
        score = sum(1 for token in tokens if token in blob)
        if score == 0 and tokens:
            continue
        if weight_field:
            score += parse_int(row.get(weight_field, ""), 0)
        scored.append((score, row))
    return [row for _, row in sorted(scored, key=lambda item: item[0], reverse=True)]


def bullet_line(title: str, finding: str, url: str) -> str:
    suffix = f" ({url})" if url else ""
    return f"- {title}: {finding}{suffix}"


def build_local_auto_research_report(
    question_id: str,
    question: str,
    domain_bundle: str,
    provider_error: str | None = None,
) -> str:
    tokens = question_tokens(question)
    sources = rank_rows_for_question(read_rows(SOURCES), tokens, "credibility")[:6]
    regulations = rank_rows_for_question(read_rows(REGULATIONS), tokens, "risk_level")[:6]
    competitors = rank_rows_for_question(read_rows(COMPETITORS), tokens, "relevance")[:4]
    papers = rank_rows_for_question(read_rows(PAPERS), tokens)[:4]
    evidence = rank_rows_for_question(read_rows(EVIDENCE), tokens, "confidence")[:4]

    has_material = any([sources, regulations, competitors, papers, evidence])
    veredicto = (
        "Reporte local preliminar: hay material cargado para orientar la decision, "
        "pero falta una busqueda web nueva antes de cerrar la pregunta."
        if has_material
        else "Reporte local preliminar: no hay evidencia cargada suficiente para responder esta pregunta."
    )

    lines = [
        f"# Auto research - {question_id}",
        "",
        "## Veredicto",
        "",
        veredicto,
        "",
    ]
    if provider_error:
        lines.extend(
            [
                "> Nota operativa: se genero este fallback local porque fallo el proveedor con internet/IA.",
                f"> Motivo: {provider_error}",
                "",
            ]
        )

    lines.extend(["## Hallazgos con fuentes", ""])
    if sources:
        for row in sources:
            lines.append(bullet_line(row["title"], row["finding"], row["url"]))
    if regulations:
        for row in regulations:
            title = f"{row['norm']} / {row['topic']}"
            lines.append(bullet_line(title, row["impact"], row["url"]))
    if papers:
        for row in papers:
            title = f"{row['source']} {row['year']} - {row['title']}"
            lines.append(bullet_line(title, row["abstract"][:220], row["url"]))
    if evidence:
        for row in evidence:
            title = f"Evidencia local ({row['source']})"
            lines.append(bullet_line(title, row["claim"], ""))
    if not has_material:
        lines.append("- Sin fuentes locales directamente vinculadas. Conviene cargar fuentes o correr con OpenAI/Claude.")

    lines.extend(["", "## Implicancias para la startup", ""])
    if regulations:
        lines.append("- Tratar cualquier dato de salud como sensible: consentimiento, minimizacion, trazabilidad y revision legal antes de integrar.")
    if sources:
        lines.append("- Usar las fuentes oficiales cargadas para definir alcance del MVP y separar lo posible ahora de lo que requiere integracion formal.")
    if competitors:
        lines.append("- Comparar el hueco competitivo contra analogos antes de construir funcionalidad propia.")
    if not has_material:
        lines.append("- La decision no deberia moverse todavia; primero falta evidencia externa y entrevistas.")
    lines.append("- Convertir la pregunta en una prueba chica: una landing, entrevista, pedido de datos o consulta legal concreta.")

    lines.extend(["", "## Riesgos e incertidumbres", ""])
    lines.extend(
        [
            "- Este modo local no navega internet ni verifica cambios recientes.",
            "- Las fuentes cargadas pueden estar incompletas o desactualizadas.",
            "- No reemplaza validacion legal, clinica ni entrevistas con usuarios reales.",
        ]
    )

    lines.extend(["", "## Proximo experimento", ""])
    lines.append(
        "Correr la misma pregunta con `--provider openai` o `--provider claude_cli` y contrastar el reporte con una entrevista a un profesional y una consulta legal breve."
    )

    urls = []
    for row in [*sources, *regulations, *competitors, *papers]:
        url = row.get("url", "")
        if url and url not in urls:
            urls.append(url)

    lines.extend(["", "## Fuentes consultadas", ""])
    if urls:
        lines.extend(f"- {url}" for url in urls)
    else:
        lines.append("- Sin URLs cargadas para esta pregunta.")

    lines.extend(["", "## Configuracion", "", f"- Dominio pedido: {domain_bundle}", "- Proveedor: local"])
    return "\n".join(lines) + "\n"


def auto_research(args: argparse.Namespace) -> None:
    question_id, question = get_question(args.question_id, args.question)
    provider = args.provider or os.getenv("AI_PROVIDER", "openai")
    if provider not in PROVIDERS:
        raise ValueError(f"Proveedor invalido: {provider}. Opciones: {', '.join(sorted(PROVIDERS))}")

    model = args.model
    if not model and provider == "openai":
        model = os.getenv("OPENAI_MODEL", "gpt-5")
    if not model and provider == "claude_cli":
        model = os.getenv("CLAUDE_MODEL", "")

    run_id = f"{timestamp()}_{question_id}".replace("/", "_").replace(" ", "_")
    prompt = build_auto_research_prompt(question_id, question, args.domain_bundle)

    REPORTS.mkdir(parents=True, exist_ok=True)
    path = REPORTS / f"auto_research_{run_id}.md"

    try:
        if provider == "local":
            text = build_local_auto_research_report(
                question_id=question_id,
                question=question,
                domain_bundle=args.domain_bundle,
            )
        elif provider == "openai":
            body = call_openai_with_web(
                prompt=prompt,
                model=model or "gpt-5",
                domain_bundle=args.domain_bundle,
                search_context_size=args.context_size,
                reasoning_effort=args.reasoning_effort,
            )
            text = extract_response_text(body)
            sources = extract_web_sources(body)
            if sources:
                source_block = "\n\n## Fuentes recuperadas por la herramienta\n\n"
                source_block += "\n".join(
                    f"- {source.get('title') or source['url']}: {source['url']}"
                    for source in sources
                )
                text = text.rstrip() + source_block + "\n"
        else:
            text = call_claude_cli(
                prompt=prompt,
                model=model or None,
                max_turns=args.max_turns,
                timeout_seconds=args.timeout,
            )

        path.write_text(text + "\n", encoding="utf-8")
        append_row(
            AUTO_RUNS,
            {
                "date": today(),
                "run_id": run_id,
                "question_id": question_id,
                "question": question,
                "model": f"{provider}:{model or 'default'}",
                "domain_bundle": args.domain_bundle,
                "status": "ok",
                "report_path": str(path),
                "summary": text[:500].replace("\n", " "),
            },
        )
        print(text)
        print(f"Reporte guardado en: {path}")
    except (
        RuntimeError,
        urllib.error.URLError,
        TimeoutError,
        json.JSONDecodeError,
        subprocess.TimeoutExpired,
    ) as exc:
        message = build_local_auto_research_report(
            question_id=question_id,
            question=question,
            domain_bundle=args.domain_bundle,
            provider_error=str(exc),
        )
        path.write_text(message, encoding="utf-8")
        append_row(
            AUTO_RUNS,
            {
                "date": today(),
                "run_id": run_id,
                "question_id": question_id,
                "question": question,
                "model": f"{provider}:{model or 'default'}",
                "domain_bundle": args.domain_bundle,
                "status": "fallback",
                "report_path": str(path),
                "summary": str(exc),
            },
        )
        print(message)


def auto_scout(args: argparse.Namespace) -> None:
    questions = sorted(
        [row for row in read_rows(QUESTIONS) if row["status"] == "open"],
        key=lambda row: parse_int(row["priority"]),
        reverse=True,
    )[: args.limit]
    if not questions:
        print("No hay preguntas abiertas para investigar.")
        return

    for row in questions:
        print(f"\n=== Investigando {row['id']}: {row['question']} ===\n")
        child_args = argparse.Namespace(
            question_id=row["id"],
            question=None,
            model=args.model,
            provider=args.provider,
            domain_bundle=args.domain_bundle,
            context_size=args.context_size,
            reasoning_effort=args.reasoning_effort,
            max_turns=args.max_turns,
            timeout=args.timeout,
        )
        auto_research(child_args)


def safe_read_text(path_value: str, limit: int = 12000) -> str:
    try:
        path = Path(path_value)
        if not path.exists():
            return ""
        text = path.read_text(encoding="utf-8")
    except (OSError, UnicodeDecodeError):
        return ""
    if len(text) <= limit:
        return text
    return text[:limit] + "\n\n[recortado]\n"


def recent_auto_reports(limit: int) -> list[dict[str, str]]:
    rows = [row for row in read_rows(AUTO_RUNS) if row.get("report_path")]
    return list(reversed(rows[-limit:]))


def section_text(markdown: str, names: set[str]) -> str:
    wanted = {normalize_text(name) for name in names}
    active = False
    collected: list[str] = []
    for line in markdown.splitlines():
        if line.startswith("## "):
            heading = normalize_text(line.strip("# "))
            active = any(name in heading for name in wanted)
            continue
        if active and line.startswith("## "):
            break
        if active:
            collected.append(line)
    return "\n".join(collected).strip()


def list_items(text: str, limit: int) -> list[str]:
    items: list[str] = []
    for raw_line in text.splitlines():
        line = raw_line.strip()
        if not line:
            continue
        if re.match(r"^[-*]\s+", line):
            line = re.sub(r"^[-*]\s+", "", line)
        elif re.match(r"^\d+[.)]\s+", line):
            line = re.sub(r"^\d+[.)]\s+", "", line)
        elif len(items) > 0:
            continue
        items.append(compact_text(line, 260))
        if len(items) >= limit:
            break
    return items


def next_generated_question_id(existing_ids: set[str]) -> str:
    base = f"AUTO-{today().replace('-', '')}"
    for index in range(1, 1000):
        candidate = f"{base}-{index:03d}"
        if candidate not in existing_ids:
            existing_ids.add(candidate)
            return candidate
    raise RuntimeError("No pude generar un ID unico de pregunta.")


def infer_category(text: str) -> str:
    normalized = normalize_text(text)
    if any(token in normalized for token in ["ley", "norma", "regul", "renapdis", "privacidad"]):
        return "regulacion"
    if any(token in normalized for token in ["paciente", "usuario", "demanda", "dolor"]):
        return "pacientes"
    if any(token in normalized for token in ["profesional", "medico", "matricula"]):
        return "profesionales"
    if any(token in normalized for token in ["pago", "cobro", "reintegro", "capital"]):
        return "finanzas"
    if any(token in normalized for token in ["canal", "adquisicion", "go-to-market", "marketing"]):
        return "go_to_market"
    return "mercado"


def next_decision_id(existing_ids: set[str]) -> str:
    base = f"DEC-{today().replace('-', '')}"
    for index in range(1, 1000):
        candidate = f"{base}-{index:03d}"
        if candidate not in existing_ids:
            existing_ids.add(candidate)
            return candidate
    raise RuntimeError("No pude generar un ID unico de decision.")


def next_founder_question_id(existing_ids: set[str]) -> str:
    base = f"FQ-{today().replace('-', '')}"
    for index in range(1, 1000):
        candidate = f"{base}-{index:03d}"
        if candidate not in existing_ids:
            existing_ids.add(candidate)
            return candidate
    raise RuntimeError("No pude generar un ID unico de pregunta al fundador.")


def local_reflection_payload(args: argparse.Namespace) -> dict[str, object]:
    reports = recent_auto_reports(args.reports)
    question_candidates: list[dict[str, str]] = []
    experiment_candidates: list[dict[str, str]] = []
    source_candidates: list[str] = []

    for report in reports:
        text = safe_read_text(report["report_path"])
        risks = list_items(section_text(text, {"Riesgos e incertidumbres"}), args.max_questions)
        experiments = list_items(section_text(text, {"Proximo experimento", "Próximo experimento"}), 2)
        sources = list_items(section_text(text, {"Fuentes consultadas"}), 8)

        for risk in risks:
            question_candidates.append(
                {
                    "category": infer_category(risk),
                    "question": f"Que evidencia falta para resolver esta incertidumbre: {compact_text(risk, 170)}?",
                    "priority": "4",
                    "decision_if_true": "Convertir el hallazgo en criterio de producto o compliance.",
                    "decision_if_false": "Mantener la hipotesis como riesgo abierto y no construir sobre ella.",
                    "notes": f"Generada por reflexion local desde {report['question_id']}.",
                }
            )

        for experiment in experiments:
            experiment_candidates.append(
                {
                    "name": f"Validar {compact_text(experiment, 70)}",
                    "hypothesis": f"Si ejecutamos este experimento, reducimos incertidumbre sobre: {compact_text(experiment, 180)}",
                    "metric": "respuesta concreta / decision tomada",
                    "target": "1 respuesta accionable en 7-14 dias",
                    "status": "planned",
                    "result": "",
                    "notes": f"Generado por reflexion local desde {report['question_id']}.",
                }
            )

        for source in sources:
            if source.startswith("http"):
                source_candidates.append(source)

    return {
        "questions": question_candidates[: args.max_questions],
        "experiments": experiment_candidates[: args.max_experiments],
        "sources_to_verify": source_candidates[:10],
        "strategy_notes": [
            "La reflexion local solo reorganiza reportes existentes; para crecer mejor usar `--provider claude_cli` u OpenAI.",
            "Las preguntas nuevas deben competir por prioridad contra entrevistas reales y experimentos pagos.",
        ],
    }


def build_reflection_prompt(args: argparse.Namespace) -> str:
    program = (ROOT / "program.md").read_text(encoding="utf-8")
    latest_brief = build_local_brief()
    questions = read_rows(QUESTIONS)
    experiments = read_rows(EXPERIMENTS)
    reports = recent_auto_reports(args.reports)
    report_blocks = []
    for report in reports:
        report_blocks.append(
            textwrap.dedent(
                f"""
                --- REPORTE {report['run_id']} / {report['question_id']} ---
                {safe_read_text(report['report_path'], limit=10000)}
                """
            ).strip()
        )

    open_questions = "\n".join(
        f"- {row['id']} / {row['category']} / P{row['priority']}: {row['question']}"
        for row in questions
        if row["status"] == "open"
    )
    active_experiments = "\n".join(
        f"- {row['name']} / {row['status']}: {row['hypothesis']}"
        for row in experiments
        if row["status"] in {"planned", "running"}
    )

    return textwrap.dedent(
        f"""
        Sos el meta-research agent de una startup healthtech argentina.
        Tu trabajo no es responder una pregunta, sino hacer que el sistema aprenda:
        detectar incertidumbres, crear nuevas preguntas, proponer experimentos
        concretos y señalar fuentes a verificar. No inventes datos.

        Reglas:
        - No dupliques preguntas existentes.
        - Cada pregunta nueva debe poder cambiar una decision.
        - Cada experimento debe ser ejecutable en 7 a 14 dias.
        - Prioriza entrevistas, consultas regulatorias, pruebas de pago y datos oficiales.
        - La prioridad usa escala 1 a 5, donde 5 es maxima prioridad y 1 es baja prioridad.
        - Separa datos verificados de inferencias.
        - Usa ASCII simple: sin tildes, emojis ni guiones largos.
        - Devuelve SOLO JSON valido, sin markdown ni comentarios.

        Schema exacto:
        {{
          "questions": [
            {{
              "category": "mercado|dinero|pacientes|profesionales|regulacion|finanzas|ia|go_to_market",
              "question": "...",
              "priority": 1,
              "decision_if_true": "...",
              "decision_if_false": "...",
              "notes": "..."
            }}
          ],
          "experiments": [
            {{
              "name": "...",
              "hypothesis": "...",
              "metric": "...",
              "target": "...",
              "notes": "..."
            }}
          ],
          "sources_to_verify": [
            "..."
          ],
          "strategy_notes": [
            "..."
          ]
        }}

        Limites:
        - max questions: {args.max_questions}
        - max experiments: {args.max_experiments}

        PROGRAMA:
        {program}

        BRIEF LOCAL:
        {latest_brief}

        PREGUNTAS ABIERTAS:
        {open_questions or "- Sin preguntas abiertas."}

        EXPERIMENTOS ACTIVOS:
        {active_experiments or "- Sin experimentos activos."}

        REPORTES RECIENTES:
        {chr(10).join(report_blocks) or "- Sin reportes recientes."}
        """
    ).strip()


def extract_json_object(text: str) -> dict[str, object]:
    cleaned = text.strip()
    if cleaned.startswith("```"):
        cleaned = re.sub(r"^```(?:json)?", "", cleaned).strip()
        cleaned = re.sub(r"```$", "", cleaned).strip()
    start = cleaned.find("{")
    end = cleaned.rfind("}")
    if start == -1 or end == -1 or end <= start:
        raise ValueError("La respuesta no contiene JSON.")
    return json.loads(cleaned[start : end + 1])


def call_openai_plain(prompt: str, model: str) -> str:
    api_key = os.getenv("OPENAI_API_KEY")
    if not api_key:
        raise RuntimeError("OPENAI_API_KEY no configurada.")

    request = urllib.request.Request(
        "https://api.openai.com/v1/responses",
        data=json.dumps({"model": model, "input": prompt}).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )
    with urllib.request.urlopen(request, timeout=120) as response:
        body = json.loads(response.read().decode("utf-8"))
    return extract_response_text(body)


def reflection_payload(args: argparse.Namespace) -> dict[str, object]:
    provider = args.provider or os.getenv("AI_PROVIDER", "local")
    if provider == "local":
        return local_reflection_payload(args)
    prompt = build_reflection_prompt(args)
    if provider == "openai":
        model = args.model or os.getenv("OPENAI_MODEL", "gpt-5")
        text = call_openai_plain(prompt, model)
    elif provider == "claude_cli":
        model = args.model or os.getenv("CLAUDE_MODEL", "") or None
        text = call_claude_cli(prompt, model, args.max_turns, args.timeout)
    else:
        raise ValueError(f"Proveedor invalido: {provider}. Opciones: {', '.join(sorted(PROVIDERS))}")
    return extract_json_object(text)


def normalize_payload_list(payload: dict[str, object], key: str) -> list[dict[str, object]]:
    value = payload.get(key, [])
    if not isinstance(value, list):
        return []
    return [item for item in value if isinstance(item, dict)]


def apply_reflection(payload: dict[str, object]) -> dict[str, int]:
    existing_questions = read_rows(QUESTIONS)
    existing_question_ids = {row["id"] for row in existing_questions}
    existing_question_texts = {normalize_text(row["question"]) for row in existing_questions}
    existing_experiments = read_rows(EXPERIMENTS)
    existing_experiment_names = {normalize_text(row["name"]) for row in existing_experiments}
    created_questions = 0
    created_experiments = 0

    for item in normalize_payload_list(payload, "questions"):
        question = compact_text(str(item.get("question", "")), 260)
        if not question or normalize_text(question) in existing_question_texts:
            continue
        priority = parse_int(str(item.get("priority", "3")), 3)
        priority = max(1, min(priority, 5))
        append_row(
            QUESTIONS,
            {
                "id": next_generated_question_id(existing_question_ids),
                "category": compact_text(str(item.get("category", "mercado")), 40),
                "question": question,
                "priority": str(priority),
                "status": "open",
                "decision_if_true": compact_text(str(item.get("decision_if_true", "")), 220),
                "decision_if_false": compact_text(str(item.get("decision_if_false", "")), 220),
                "notes": compact_text(str(item.get("notes", "Generada por reflect.")), 220),
            },
        )
        existing_question_texts.add(normalize_text(question))
        created_questions += 1

    for item in normalize_payload_list(payload, "experiments"):
        name = compact_text(str(item.get("name", "")), 90)
        if not name or normalize_text(name) in existing_experiment_names:
            continue
        append_row(
            EXPERIMENTS,
            {
                "date": today(),
                "name": name,
                "hypothesis": compact_text(str(item.get("hypothesis", "")), 260),
                "metric": compact_text(str(item.get("metric", "")), 140),
                "target": compact_text(str(item.get("target", "")), 140),
                "status": "planned",
                "result": "",
                "notes": compact_text(str(item.get("notes", "Generado por reflect.")), 220),
            },
        )
        existing_experiment_names.add(normalize_text(name))
        created_experiments += 1

    return {"questions": created_questions, "experiments": created_experiments}


def reflection_report(payload: dict[str, object], applied: dict[str, int] | None) -> str:
    lines = [f"# Reflection loop - {today()}", ""]
    if applied:
        lines.extend(
            [
                "## Cambios aplicados",
                "",
                f"- Preguntas nuevas: {applied['questions']}",
                f"- Experimentos nuevos: {applied['experiments']}",
                "",
            ]
        )
    else:
        lines.extend(["## Modo", "", "- Dry run: no se aplicaron cambios.", ""])

    lines.extend(["## Preguntas propuestas", ""])
    questions = normalize_payload_list(payload, "questions")
    if questions:
        for item in questions:
            lines.append(
                f"- P{item.get('priority', 3)} / {item.get('category', 'mercado')}: {item.get('question', '')}"
            )
    else:
        lines.append("- Sin preguntas propuestas.")

    lines.extend(["", "## Experimentos propuestos", ""])
    experiments = normalize_payload_list(payload, "experiments")
    if experiments:
        for item in experiments:
            lines.append(f"- {item.get('name', '')}: {item.get('hypothesis', '')}")
    else:
        lines.append("- Sin experimentos propuestos.")

    lines.extend(["", "## Fuentes a verificar", ""])
    sources = payload.get("sources_to_verify", [])
    if isinstance(sources, list) and sources:
        for source in sources[:12]:
            lines.append(f"- {source}")
    else:
        lines.append("- Sin fuentes propuestas.")

    lines.extend(["", "## Notas estrategicas", ""])
    notes = payload.get("strategy_notes", [])
    if isinstance(notes, list) and notes:
        for note in notes:
            lines.append(f"- {note}")
    else:
        lines.append("- Sin notas.")

    return "\n".join(lines) + "\n"


def reflect(args: argparse.Namespace) -> None:
    payload = reflection_payload(args)
    applied = apply_reflection(payload) if args.apply else None
    report = reflection_report(payload, applied)
    REPORTS.mkdir(parents=True, exist_ok=True)
    path = REPORTS / f"reflection_{timestamp()}.md"
    path.write_text(report, encoding="utf-8")
    print(report)
    print(f"Reflection guardada en: {path}")


def auto_loop(args: argparse.Namespace) -> None:
    for cycle in range(1, args.cycles + 1):
        print(f"\n=== Ciclo {cycle}/{args.cycles}: research ===\n")
        scout_args = argparse.Namespace(
            limit=args.research_limit,
            provider=args.provider,
            model=args.model,
            domain_bundle=args.domain_bundle,
            context_size=args.context_size,
            reasoning_effort=args.reasoning_effort,
            max_turns=args.max_turns,
            timeout=args.timeout,
        )
        auto_scout(scout_args)

        print(f"\n=== Ciclo {cycle}/{args.cycles}: reflection ===\n")
        reflect_args = argparse.Namespace(
            provider=args.provider,
            model=args.model,
            reports=args.reports,
            max_questions=args.max_questions,
            max_experiments=args.max_experiments,
            max_turns=args.max_turns,
            timeout=args.timeout,
            apply=True,
        )
        reflect(reflect_args)

        if not args.skip_self_analysis:
            print(f"\n=== Ciclo {cycle}/{args.cycles}: self-analysis ===\n")
            self_args = argparse.Namespace(
                provider=args.provider,
                model=args.model,
                reports=args.reports,
                max_turns=args.max_turns,
                timeout=args.timeout,
                apply_candidates=True,
            )
            self_analyze(self_args)

        if not args.skip_founder_questions:
            print(f"\n=== Ciclo {cycle}/{args.cycles}: founder questions ===\n")
            founder_args = argparse.Namespace(
                provider=args.provider,
                model=args.model,
                limit=args.max_founder_questions,
                max_turns=args.max_turns,
                timeout=args.timeout,
                apply=True,
            )
            founder_questions(founder_args)

    print("\n=== Brief actualizado ===\n")
    brief(argparse.Namespace(ai=False))


def latest_report_files(prefix: str, limit: int) -> list[Path]:
    if not REPORTS.exists():
        return []
    return sorted(REPORTS.glob(f"{prefix}_*.md"), key=lambda path: path.stat().st_mtime)[-limit:]


def local_self_analysis_payload(args: argparse.Namespace) -> dict[str, object]:
    interviews = read_rows(INTERVIEWS)
    founder_questions = read_rows(FOUNDER_QUESTIONS)
    experiments = read_rows(EXPERIMENTS)
    questions = read_rows(QUESTIONS)
    auto_runs = read_rows(AUTO_RUNS)
    decisions = read_rows(DECISIONS)
    answered_founder_questions = [row for row in founder_questions if row["status"] == "answered"]
    open_founder_questions = [row for row in founder_questions if row["status"] == "open"]
    open_questions = [row for row in questions if row["status"] == "open"]
    generated_questions = [row for row in open_questions if row["id"].startswith("AUTO-")]
    active_experiments = [row for row in experiments if row["status"] in {"planned", "running"}]
    completed_experiments = [row for row in experiments if row["status"] in {"done", "closed"}]
    successful_runs = [row for row in auto_runs if row["status"] == "ok"]

    blind_spots: list[str] = []
    self_corrections: list[str] = []
    next_focus: list[str] = []
    decision_candidates: list[dict[str, object]] = []

    if len(answered_founder_questions) == 0 and len(auto_runs) >= 3:
        blind_spots.append(
            "El sistema acumulo varias corridas de research, pero todavia no tiene respuestas del fundador. Riesgo: investigar sin criterio operativo propio."
        )
        self_corrections.append("Congelar research no critico hasta que el fundador responda las preguntas abiertas prioritarias.")
        next_focus.append("Generar y responder preguntas al fundador con `founder-questions --apply`.")
        decision_candidates.append(
            {
                "decision": "Priorizar respuestas del fundador antes de nuevas investigaciones no bloqueantes.",
                "reason": "Hay research acumulado y falta input estrategico del fundador; el mayor riesgo actual es decidir sin preferencias, restricciones y experiencia propia.",
                "evidence": f"{len(auto_runs)} corridas registradas, {len(answered_founder_questions)} respuestas del fundador.",
                "confidence": 4,
                "next_action": "Responder las preguntas abiertas de `founder_questions.csv` y regenerar brief.",
                "review_date": (dt.datetime.now() + dt.timedelta(days=7)).strftime("%Y-%m-%d"),
                "notes": "Decision candidata generada por self-analysis local.",
            }
        )

    if any("ReNaPDiS" in row["question"] for row in generated_questions):
        blind_spots.append(
            "ReNaPDiS aparece como cuello de botella recurrente; falta evidencia directa del tramite TAD y/o abogado sanitario."
        )
        next_focus.append("Iniciar prueba TAD ReNaPDiS o consulta legal focalizada.")
        decision_candidates.append(
            {
                "decision": "No hacer que receta electronica sea requisito del MVP hasta resolver ReNaPDiS y alcance legal.",
                "reason": "La oportunidad es atractiva, pero la inscripcion, direccion tecnica y privacidad son bloqueos no resueltos.",
                "evidence": "LEG-002 y preguntas AUTO sobre ReNaPDiS, Ley 26.657 y MVP sin receta.",
                "confidence": 3,
                "next_action": "Ejecutar `Probe TAD ReNaPDiS` y `Consulta legal focalizada salud mental + receta`.",
                "review_date": (dt.datetime.now() + dt.timedelta(days=14)).strftime("%Y-%m-%d"),
                "notes": "Decision candidata, no decision final.",
            }
        )

    if active_experiments and not completed_experiments:
        self_corrections.append(
            "Hay experimentos planeados sin resultados; el proximo ciclo debe medir ejecucion, no solo generar nuevos pendientes."
        )
    if open_founder_questions:
        next_focus.append("Responder preguntas abiertas del fundador antes de generar nuevas preguntas.")
    if len(open_questions) > 12:
        self_corrections.append(
            "El backlog de preguntas crecio; conviene cerrar o aparcar preguntas antes de generar muchas mas."
        )

    loop_score = 3
    if answered_founder_questions:
        loop_score += 1
    if completed_experiments:
        loop_score += 1
    if len(auto_runs) >= 5 and not answered_founder_questions:
        loop_score -= 1
    loop_score = max(1, min(loop_score, 5))

    status = "research-heavy" if len(auto_runs) >= 5 and not answered_founder_questions else "balanced"
    diagnosis = (
        "El agente esta aprendiendo y generando buenas preguntas, pero esta cargado hacia investigacion documental y necesita respuestas del fundador."
        if status == "research-heavy"
        else "El agente mantiene un balance razonable entre investigacion y accion."
    )

    learned: list[str] = []
    for report in recent_auto_reports(args.reports):
        text = safe_read_text(report["report_path"], limit=6000)
        verdict = compact_text(section_text(text, {"Veredicto"}), 320)
        if verdict:
            learned.append(f"{report['question_id']}: {verdict}")

    return {
        "loop_health": {"status": status, "score": loop_score, "diagnosis": diagnosis},
        "what_we_learned": learned[:6],
        "blind_spots": blind_spots,
        "self_corrections": self_corrections,
        "decision_candidates": decision_candidates,
        "next_focus": next_focus or ["Ejecutar el experimento activo mas cercano a una decision de MVP."],
        "notes": [
            f"Preguntas abiertas: {len(open_questions)}.",
            f"Preguntas generadas por reflexion: {len(generated_questions)}.",
            f"Preguntas al fundador abiertas: {len(open_founder_questions)}.",
            f"Respuestas del fundador: {len(answered_founder_questions)}.",
            f"Experimentos activos: {len(active_experiments)}.",
            f"Decisiones candidatas existentes: {len(decisions)}.",
            f"Corridas exitosas: {len(successful_runs)}.",
        ],
    }


def build_self_analysis_prompt(args: argparse.Namespace) -> str:
    program = (ROOT / "program.md").read_text(encoding="utf-8")
    latest_brief = build_local_brief()
    questions = read_rows(QUESTIONS)
    experiments = read_rows(EXPERIMENTS)
    decisions = read_rows(DECISIONS)
    founder_questions = read_rows(FOUNDER_QUESTIONS)
    reports = recent_auto_reports(args.reports)
    reflections = latest_report_files("reflection", args.reports)

    report_blocks = []
    for report in reports:
        report_blocks.append(
            textwrap.dedent(
                f"""
                --- AUTO REPORT {report['run_id']} / {report['question_id']} ---
                {safe_read_text(report['report_path'], limit=9000)}
                """
            ).strip()
        )
    for path in reflections:
        report_blocks.append(
            textwrap.dedent(
                f"""
                --- REFLECTION {path.name} ---
                {safe_read_text(str(path), limit=7000)}
                """
            ).strip()
        )

    open_questions = "\n".join(
        f"- {row['id']} / {row['category']} / P{row['priority']}: {row['question']}"
        for row in questions
        if row["status"] == "open"
    )
    active_experiments = "\n".join(
        f"- {row['name']} / {row['status']}: {row['metric']} -> {row['target']}"
        for row in experiments
        if row["status"] in {"planned", "running"}
    )
    existing_decisions = "\n".join(
        f"- {row['id']} / {row['status']} / C{row['confidence']}: {row['decision']}"
        for row in decisions
    )
    founder_questions_text = "\n".join(
        f"- {row['id']} / {row['status']} / P{row['priority']}: {row['question']} | answer: {row['answer']}"
        for row in founder_questions[-12:]
    )

    return textwrap.dedent(
        f"""
        Sos el autoanalista del research agent de una startup healthtech argentina.
        Tu tarea es mirar lo que el agente viene investigando y evaluar su propio
        estado: que aprendio, donde esta sesgado, que decisiones ya estan maduras,
        que decisiones NO debe tomar aun, y que accion concreta debe priorizar.

        Reglas:
        - No inventes datos.
        - No conviertas una hipotesis legal/clinica en decision final.
        - Las decisiones deben salir como candidatos con status implicito candidate.
        - El usuario pidio no hacer entrevistas externas ahora: prioriza preguntas al fundador, consultas regulatorias y experimentos ejecutables.
        - Si hay demasiado research y pocas respuestas del fundador, dilo sin suavizarlo.
        - Usa ASCII simple: sin tildes, emojis ni guiones largos.
        - Devuelve SOLO JSON valido.

        Schema exacto:
        {{
          "loop_health": {{
            "status": "balanced|research-heavy|action-heavy|blocked",
            "score": 1,
            "diagnosis": "..."
          }},
          "what_we_learned": ["..."],
          "blind_spots": ["..."],
          "self_corrections": ["..."],
          "decision_candidates": [
            {{
              "decision": "...",
              "reason": "...",
              "evidence": "...",
              "confidence": 1,
              "next_action": "...",
              "review_date": "YYYY-MM-DD",
              "notes": "..."
            }}
          ],
          "next_focus": ["..."],
          "notes": ["..."]
        }}

        Escala confidence: 1 bajo, 5 alto.

        PROGRAMA:
        {program}

        BRIEF ACTUAL:
        {latest_brief}

        PREGUNTAS ABIERTAS:
        {open_questions or "- Sin preguntas abiertas."}

        EXPERIMENTOS ACTIVOS:
        {active_experiments or "- Sin experimentos activos."}

        DECISIONES CANDIDATAS EXISTENTES:
        {existing_decisions or "- Sin decisiones guardadas."}

        PREGUNTAS Y RESPUESTAS DEL FUNDADOR:
        {founder_questions_text or "- Sin preguntas al fundador guardadas."}

        REPORTES Y REFLEXIONES RECIENTES:
        {chr(10).join(report_blocks) or "- Sin reportes recientes."}
        """
    ).strip()


def self_analysis_payload(args: argparse.Namespace) -> dict[str, object]:
    provider = args.provider or os.getenv("AI_PROVIDER", "local")
    if provider == "local":
        return local_self_analysis_payload(args)
    prompt = build_self_analysis_prompt(args)
    if provider == "openai":
        model = args.model or os.getenv("OPENAI_MODEL", "gpt-5")
        text = call_openai_plain(prompt, model)
    elif provider == "claude_cli":
        model = args.model or os.getenv("CLAUDE_MODEL", "") or None
        text = call_claude_cli(prompt, model, args.max_turns, args.timeout)
    else:
        raise ValueError(f"Proveedor invalido: {provider}. Opciones: {', '.join(sorted(PROVIDERS))}")
    return extract_json_object(text)


def payload_strings(payload: dict[str, object], key: str) -> list[str]:
    value = payload.get(key, [])
    if not isinstance(value, list):
        return []
    return [compact_text(str(item), 320) for item in value if str(item).strip()]


def apply_decision_candidates(payload: dict[str, object]) -> int:
    existing = read_rows(DECISIONS)
    existing_ids = {row["id"] for row in existing}
    existing_texts = {normalize_text(row["decision"]) for row in existing}
    created = 0
    for item in normalize_payload_list(payload, "decision_candidates"):
        decision = compact_text(str(item.get("decision", "")), 240)
        if not decision or normalize_text(decision) in existing_texts:
            continue
        confidence = max(1, min(parse_int(str(item.get("confidence", "2")), 2), 5))
        review_date = str(item.get("review_date", "")).strip()
        append_row(
            DECISIONS,
            {
                "date": today(),
                "id": next_decision_id(existing_ids),
                "status": "candidate",
                "decision": decision,
                "reason": compact_text(str(item.get("reason", "")), 260),
                "evidence": compact_text(str(item.get("evidence", "")), 260),
                "confidence": str(confidence),
                "next_action": compact_text(str(item.get("next_action", "")), 220),
                "review_date": compact_text(review_date, 40),
                "notes": compact_text(str(item.get("notes", "Generada por self-analysis.")), 220),
            },
        )
        existing_texts.add(normalize_text(decision))
        created += 1
    return created


def self_analysis_report(payload: dict[str, object], applied: int | None) -> str:
    health = payload.get("loop_health", {})
    if not isinstance(health, dict):
        health = {}

    lines = [f"# Self-analysis - {today()}", ""]
    lines.extend(
        [
            "## Salud del loop",
            "",
            f"- Estado: {health.get('status', 'sin_datos')}",
            f"- Score: {health.get('score', 'sin_datos')}/5",
            f"- Diagnostico: {health.get('diagnosis', 'Sin diagnostico.')}",
            "",
        ]
    )
    if applied is not None:
        lines.extend(["## Memoria actualizada", "", f"- Decisiones candidatas nuevas: {applied}", ""])
    else:
        lines.extend(["## Modo", "", "- Dry run: no se guardaron decisiones candidatas.", ""])

    sections = [
        ("Que aprendio", "what_we_learned"),
        ("Puntos ciegos", "blind_spots"),
        ("Autocorrecciones", "self_corrections"),
        ("Proximo foco", "next_focus"),
        ("Notas", "notes"),
    ]
    for title, key in sections:
        lines.extend([f"## {title}", ""])
        items = payload_strings(payload, key)
        if items:
            lines.extend(f"- {item}" for item in items)
        else:
            lines.append("- Sin datos.")
        lines.append("")

    lines.extend(["## Decisiones candidatas", ""])
    candidates = normalize_payload_list(payload, "decision_candidates")
    if candidates:
        for item in candidates:
            lines.append(f"- C{item.get('confidence', '?')}: {item.get('decision', '')}")
            reason = compact_text(str(item.get("reason", "")), 260)
            next_action = compact_text(str(item.get("next_action", "")), 220)
            if reason:
                lines.append(f"  Razon: {reason}")
            if next_action:
                lines.append(f"  Proximo paso: {next_action}")
    else:
        lines.append("- Sin decisiones candidatas.")
    lines.append("")
    return "\n".join(lines)


def self_analyze(args: argparse.Namespace) -> None:
    payload = self_analysis_payload(args)
    applied = apply_decision_candidates(payload) if args.apply_candidates else None
    report = self_analysis_report(payload, applied)
    REPORTS.mkdir(parents=True, exist_ok=True)
    path = REPORTS / f"self_analysis_{timestamp()}.md"
    path.write_text(report, encoding="utf-8")
    print(report)
    print(f"Self-analysis guardado en: {path}")


def local_founder_questions_payload(args: argparse.Namespace) -> dict[str, object]:
    decisions = sorted(
        [row for row in read_rows(DECISIONS) if row["status"] == "candidate"],
        key=lambda row: parse_int(row["confidence"]),
        reverse=True,
    )
    open_research_questions = sorted(
        [row for row in read_rows(QUESTIONS) if row["status"] == "open"],
        key=lambda row: parse_int(row["priority"]),
        reverse=True,
    )
    existing = read_rows(FOUNDER_QUESTIONS)
    existing_questions = {normalize_text(row["question"]) for row in existing}

    candidates: list[dict[str, object]] = []

    for row in decisions:
        decision = row["decision"]
        if "entrevista" in normalize_text(decision):
            question = (
                "Antes de validar afuera: que sabes hoy, por experiencia propia, sobre el dolor de los profesionales "
                "independientes? Contame casos concretos, no opiniones generales."
            )
        elif "receta electronica" in normalize_text(decision) and "mvp" in normalize_text(decision):
            question = (
                "Queremos que la receta electronica quede fuera del MVP inicial hasta resolver ReNaPDiS? "
                "Responde si/no y que riesgo aceptarias asumir."
            )
        elif "tad" in normalize_text(decision) or "renapdis" in normalize_text(decision):
            question = (
                "Estas dispuesto a iniciar el tramite TAD/ReNaPDiS esta semana solo para aprender requisitos, "
                "aunque no lo usemos en el MVP? Responde si/no y quien lo haria."
            )
        elif "salud mental" in normalize_text(decision):
            question = (
                "Por que salud mental deberia ser el vertical inicial frente a odontologia o kinesiologia? "
                "Dame tus 3 razones mas fuertes y tu principal duda."
            )
        else:
            question = f"Sobre esta decision candidata: {decision}. Que elegis hoy y que evidencia te haria cambiar de opinion?"

        if normalize_text(question) in existing_questions:
            continue
        candidates.append(
            {
                "priority": row["confidence"],
                "question": question,
                "why_it_matters": row["reason"],
                "decision_unlocked": row["decision"],
                "source": row["id"],
                "next_action": row["next_action"],
            }
        )
        existing_questions.add(normalize_text(question))
        if len(candidates) >= args.limit:
            break

    if len(candidates) < args.limit:
        for row in open_research_questions:
            question = (
                f"Para la pregunta `{row['id']}`: {row['question']} "
                "Cual es tu intuicion actual, que evidencia tenes de primera mano y que decision tomarias si tuvieras que decidir hoy?"
            )
            if normalize_text(question) in existing_questions:
                continue
            candidates.append(
                {
                    "priority": row["priority"],
                    "question": question,
                    "why_it_matters": row["notes"],
                    "decision_unlocked": row["decision_if_true"] or row["decision_if_false"],
                    "source": row["id"],
                    "next_action": "Responder para alimentar el proximo self-analysis.",
                }
            )
            existing_questions.add(normalize_text(question))
            if len(candidates) >= args.limit:
                break

    return {
        "questions": candidates,
        "notes": [
            "Estas preguntas reemplazan entrevistas externas como siguiente sensor de realidad.",
            "Responderlas alimenta el brief y el self-analysis del agente.",
        ],
    }


def build_founder_questions_prompt(args: argparse.Namespace) -> str:
    program = (ROOT / "program.md").read_text(encoding="utf-8")
    brief_text = build_local_brief()
    decisions = read_rows(DECISIONS)
    existing_founder_questions = read_rows(FOUNDER_QUESTIONS)
    unanswered = [row for row in existing_founder_questions if row["status"] == "open"]

    decisions_text = "\n".join(
        f"- {row['id']} / C{row['confidence']} / {row['status']}: {row['decision']} -> {row['next_action']}"
        for row in decisions
        if row["status"] == "candidate"
    )
    founder_questions_text = "\n".join(
        f"- {row['id']} / {row['status']} / P{row['priority']}: {row['question']}"
        for row in existing_founder_questions[-12:]
    )

    return textwrap.dedent(
        f"""
        Sos el agente que le hace preguntas al fundador de una startup healthtech.
        El usuario pidio explicitamente NO hacer entrevistas externas ahora.
        Tu trabajo es generar preguntas para que el fundador responda y asi el
        research agent pueda reducir incertidumbre sin salir todavia a campo.

        Reglas:
        - Genera preguntas directas, concretas y respondibles por el fundador.
        - Cada pregunta debe desbloquear una decision de producto, legal, mercado o go-to-market.
        - No hagas preguntas genericas tipo "que opinas".
        - No pidas entrevistas externas.
        - Usa ASCII simple: sin tildes, emojis ni guiones largos.
        - Devuelve SOLO JSON valido.

        Schema:
        {{
          "questions": [
            {{
              "priority": 5,
              "question": "...",
              "why_it_matters": "...",
              "decision_unlocked": "...",
              "source": "...",
              "next_action": "..."
            }}
          ],
          "notes": ["..."]
        }}

        Max questions: {args.limit}

        PROGRAMA:
        {program}

        BRIEF:
        {brief_text}

        DECISIONES CANDIDATAS:
        {decisions_text or "- Sin decisiones candidatas."}

        PREGUNTAS AL FUNDADOR YA EXISTENTES:
        {founder_questions_text or "- Sin preguntas previas."}

        PREGUNTAS ABIERTAS SIN RESPONDER:
        {len(unanswered)}
        """
    ).strip()


def founder_questions_payload(args: argparse.Namespace) -> dict[str, object]:
    provider = args.provider or os.getenv("AI_PROVIDER", "local")
    if provider == "local":
        return local_founder_questions_payload(args)
    prompt = build_founder_questions_prompt(args)
    if provider == "openai":
        model = args.model or os.getenv("OPENAI_MODEL", "gpt-5")
        text = call_openai_plain(prompt, model)
    elif provider == "claude_cli":
        model = args.model or os.getenv("CLAUDE_MODEL", "") or None
        text = call_claude_cli(prompt, model, args.max_turns, args.timeout)
    else:
        raise ValueError(f"Proveedor invalido: {provider}. Opciones: {', '.join(sorted(PROVIDERS))}")
    return extract_json_object(text)


def apply_founder_questions(payload: dict[str, object]) -> int:
    rows = read_rows(FOUNDER_QUESTIONS)
    existing_ids = {row["id"] for row in rows}
    existing_questions = {normalize_text(row["question"]) for row in rows}
    created = 0
    for item in normalize_payload_list(payload, "questions"):
        question = compact_text(str(item.get("question", "")), 320)
        if not question or normalize_text(question) in existing_questions:
            continue
        priority = max(1, min(parse_int(str(item.get("priority", "3")), 3), 5))
        append_row(
            FOUNDER_QUESTIONS,
            {
                "date": today(),
                "id": next_founder_question_id(existing_ids),
                "status": "open",
                "priority": str(priority),
                "question": question,
                "why_it_matters": compact_text(str(item.get("why_it_matters", "")), 260),
                "decision_unlocked": compact_text(str(item.get("decision_unlocked", "")), 260),
                "source": compact_text(str(item.get("source", "")), 80),
                "answer": "",
                "answered_at": "",
                "next_action": compact_text(str(item.get("next_action", "")), 220),
            },
        )
        existing_questions.add(normalize_text(question))
        created += 1
    return created


def founder_questions_report(payload: dict[str, object], applied: int | None) -> str:
    lines = [f"# Founder questions - {today()}", ""]
    if applied is None:
        lines.extend(["## Modo", "", "- Dry run: no se guardaron preguntas.", ""])
    else:
        lines.extend(["## Memoria actualizada", "", f"- Preguntas nuevas: {applied}", ""])

    lines.extend(["## Preguntas para responder", ""])
    questions = normalize_payload_list(payload, "questions")
    if questions:
        for index, item in enumerate(questions, start=1):
            lines.append(f"{index}. P{item.get('priority', 3)} - {item.get('question', '')}")
            why = compact_text(str(item.get("why_it_matters", "")), 240)
            decision = compact_text(str(item.get("decision_unlocked", "")), 240)
            if why:
                lines.append(f"   Por que importa: {why}")
            if decision:
                lines.append(f"   Decision que desbloquea: {decision}")
    else:
        lines.append("- Sin preguntas nuevas.")

    notes = payload.get("notes", [])
    if isinstance(notes, list) and notes:
        lines.extend(["", "## Notas", ""])
        lines.extend(f"- {compact_text(str(note), 220)}" for note in notes)
    return "\n".join(lines) + "\n"


def founder_questions(args: argparse.Namespace) -> None:
    payload = founder_questions_payload(args)
    applied = apply_founder_questions(payload) if args.apply else None
    report = founder_questions_report(payload, applied)
    REPORTS.mkdir(parents=True, exist_ok=True)
    path = REPORTS / f"founder_questions_{timestamp()}.md"
    path.write_text(report, encoding="utf-8")
    print(report)
    print(f"Preguntas guardadas en: {path}")


def founder_answer(args: argparse.Namespace) -> None:
    rows = read_rows(FOUNDER_QUESTIONS)
    updated = False
    for row in rows:
        if row["id"] != args.id:
            continue
        row["status"] = "answered"
        row["answer"] = args.answer
        row["answered_at"] = today()
        if args.next_action:
            row["next_action"] = args.next_action
        updated = True
        break
    if not updated:
        raise ValueError(f"No encontre la pregunta {args.id}.")
    write_rows(FOUNDER_QUESTIONS, rows)
    print(f"Respuesta registrada para {args.id}.")


def build_local_brief() -> str:
    interviews = read_rows(INTERVIEWS)
    evidence = read_rows(EVIDENCE)
    experiments = read_rows(EXPERIMENTS)
    verticals = sorted(read_rows(VERTICALS), key=score_vertical, reverse=True)
    papers = read_rows(PAPERS)
    questions = read_rows(QUESTIONS)
    sources = read_rows(SOURCES)
    competitors = read_rows(COMPETITORS)
    regulations = read_rows(REGULATIONS)
    tools = read_rows(TOOLS)
    auto_runs = read_rows(AUTO_RUNS)
    decisions = read_rows(DECISIONS)
    founder_questions = read_rows(FOUNDER_QUESTIONS)

    patient_interviews = [r for r in interviews if r["kind"] == "paciente"]
    pro_interviews = [r for r in interviews if r["kind"] == "profesional"]
    high_pain = [r for r in interviews if parse_int(r["pain"]) >= 4]
    high_willingness = [r for r in interviews if parse_int(r["willingness"]) >= 4]
    active_experiments = [r for r in experiments if r["status"] in {"planned", "running"}]
    open_questions = [r for r in questions if r["status"] == "open"]
    high_risk_regulations = [r for r in regulations if parse_int(r["risk_level"]) >= 4]
    candidate_decisions = [r for r in decisions if r["status"] == "candidate"]
    open_founder_questions = [r for r in founder_questions if r["status"] == "open"]
    answered_founder_questions = [r for r in founder_questions if r["status"] == "answered"]

    top_vertical = verticals[0]["vertical"] if verticals else "sin_datos"

    lines = [
        f"# Brief de investigacion - {today()}",
        "",
        "## Lectura rapida",
        "",
        f"- Entrevistas totales: {len(interviews)}",
        f"- Pacientes entrevistados: {len(patient_interviews)}",
        f"- Profesionales entrevistados: {len(pro_interviews)}",
        f"- Entrevistas con dolor alto: {len(high_pain)}",
        f"- Senales de disposicion a pagar/probar: {len(high_willingness)}",
        f"- Evidencias registradas: {len(evidence)}",
        f"- Papers/fuentes academicas guardadas: {len(papers)}",
        f"- Preguntas abiertas: {len(open_questions)}",
        f"- Fuentes estructuradas: {len(sources)}",
        f"- Competidores/analogos mapeados: {len(competitors)}",
        f"- Normas de riesgo alto: {len(high_risk_regulations)}",
        f"- Herramientas reutilizables mapeadas: {len(tools)}",
        f"- Corridas de research registradas: {len(auto_runs)}",
        f"- Decisiones candidatas: {len(candidate_decisions)}",
        f"- Preguntas al fundador abiertas: {len(open_founder_questions)}",
        f"- Respuestas del fundador: {len(answered_founder_questions)}",
        f"- Experimentos activos o planeados: {len(active_experiments)}",
        f"- Vertical sugerido por score actual: {top_vertical}",
        "",
        "## Ranking de verticales",
        "",
    ]

    for row in verticals:
        lines.append(f"- {row['vertical']}: {score_vertical(row)} puntos. {row['notes']}")

    lines.extend(
        [
            "",
            "## Senales recientes",
            "",
        ]
    )
    for row in interviews[-5:]:
        lines.append(
            f"- {row['kind']} / {row['segment']} / dolor {row['pain']} / "
            f"urgencia {row['urgency']} / pago {row['willingness']}: {row['notes']}"
        )
    if not interviews and not answered_founder_questions:
        lines.append("- Todavia no hay respuestas del fundador ni senales de campo registradas.")
    elif answered_founder_questions:
        for row in answered_founder_questions[-3:]:
            lines.append(f"- Fundador / {row['id']}: {compact_text(row['answer'], 220)}")

    lines.extend(
        [
            "",
            "## Papers/fuentes academicas recientes",
            "",
        ]
    )
    for row in papers[-5:]:
        lines.append(f"- {row['source']} / {row['year']}: {row['title']} ({row['url']})")
    if not papers:
        lines.append(
            "- Todavia no hay papers guardados. Probar con `paper-search --source pubmed`."
        )

    lines.extend(
        [
            "",
            "## Preguntas abiertas prioritarias",
            "",
        ]
    )
    prioritized_questions = sorted(
        open_questions, key=lambda row: parse_int(row["priority"]), reverse=True
    )
    for row in prioritized_questions[:5]:
        lines.append(f"- {row['id']} / {row['category']} / P{row['priority']}: {row['question']}")
    if not prioritized_questions:
        lines.append("- No hay preguntas abiertas cargadas.")

    generated_questions = [row for row in open_questions if row["id"].startswith("AUTO-")]
    if generated_questions:
        lines.extend(["", "## Preguntas generadas por reflexion", ""])
        for row in sorted(generated_questions, key=lambda item: parse_int(item["priority"]), reverse=True)[
            :5
        ]:
            lines.append(
                f"- {row['id']} / {row['category']} / P{row['priority']}: {row['question']}"
            )

    if active_experiments:
        lines.extend(["", "## Experimentos activos o planeados", ""])
        for row in active_experiments[:5]:
            lines.append(f"- {row['name']}: {row['metric']} -> {row['target']}")

    if candidate_decisions:
        lines.extend(["", "## Decisiones candidatas", ""])
        for row in sorted(candidate_decisions, key=lambda item: parse_int(item["confidence"]), reverse=True)[
            :5
        ]:
            lines.append(
                f"- {row['id']} / C{row['confidence']}: {row['decision']} -> {row['next_action']}"
            )

    if open_founder_questions:
        lines.extend(["", "## Preguntas abiertas para el fundador", ""])
        for row in sorted(open_founder_questions, key=lambda item: parse_int(item["priority"]), reverse=True)[
            :7
        ]:
            lines.append(f"- {row['id']} / P{row['priority']}: {row['question']}")
    if answered_founder_questions:
        lines.extend(["", "## Respuestas recientes del fundador", ""])
        for row in answered_founder_questions[-5:]:
            lines.append(f"- {row['id']}: {compact_text(row['answer'], 220)}")

    latest_self_analysis = latest_report_files("self_analysis", 1)
    if latest_self_analysis:
        self_text = safe_read_text(str(latest_self_analysis[-1]), limit=7000)
        health_items = list_items(section_text(self_text, {"Salud del loop"}), 3)
        focus_items = list_items(section_text(self_text, {"Proximo foco"}), 5)
        lines.extend(["", "## Autoanalisis reciente", ""])
        for item in health_items:
            lines.append(f"- {item}")
        if focus_items:
            lines.append("")
            lines.append("Foco recomendado:")
            for item in focus_items:
                lines.append(f"- {item}")

    lines.extend(
        [
            "",
            "## Fuentes estructuradas clave",
            "",
        ]
    )
    for row in sorted(sources, key=lambda item: parse_int(item["credibility"]), reverse=True)[:5]:
        lines.append(f"- {row['title']}: {row['finding']}")
    if not sources:
        lines.append("- No hay fuentes estructuradas cargadas. Ejecutar `seed`.")

    lines.extend(
        [
            "",
            "## Competidores/analogos a mirar",
            "",
        ]
    )
    for row in sorted(competitors, key=lambda item: parse_int(item["relevance"]), reverse=True)[:5]:
        lines.append(f"- {row['name']}: {row['model']}. Hueco: {row['weaknesses']}")
    if not competitors:
        lines.append("- No hay competidores cargados.")

    lines.extend(
        [
            "",
            "## Riesgos regulatorios principales",
            "",
        ]
    )
    for row in sorted(regulations, key=lambda item: parse_int(item["risk_level"]), reverse=True)[:5]:
        lines.append(f"- {row['norm']} / {row['topic']}: {row['impact']}")
    if not regulations:
        lines.append("- No hay regulacion cargada.")

    lines.extend(
        [
            "",
            "## Corridas autonomas recientes",
            "",
        ]
    )
    recent_auto_runs: list[dict[str, str]] = []
    seen_runs: set[tuple[str, str]] = set()
    for row in reversed(auto_runs):
        key = (row["question_id"], row["status"])
        if key in seen_runs:
            continue
        recent_auto_runs.append(row)
        seen_runs.add(key)
        if len(recent_auto_runs) == 5:
            break
    for row in reversed(recent_auto_runs):
        lines.append(f"- {row['date']} / {row['question_id']} / {row['status']}: {row['report_path']}")
    if not auto_runs:
        lines.append("- Todavia no hay corridas autonomas con IA/web.")

    lines.extend(
        [
            "",
            "## Proximo movimiento recomendado",
            "",
        ]
    )

    if open_founder_questions:
        lines.append("Responder las preguntas abiertas del fundador antes de seguir investigando.")
    elif len(answered_founder_questions) < 5:
        lines.append("Generar y responder una tanda de preguntas al fundador para destrabar decisiones.")
    elif not active_experiments:
        lines.append(
            "Definir un experimento de venta concierge: orientacion + turno + pago para un vertical."
        )
    else:
        lines.append(
            "Ejecutar el experimento activo y medir pago real, agenda tomada y repeticion."
        )

    lines.extend(
        [
            "",
            "## Riesgos a no olvidar",
            "",
            "- No prometer reintegro automatico sin convenio real.",
            "- No diagnosticar con IA; solo orientar y derivar.",
            "- Verificar matricula y consentimiento.",
            "- Evitar historia clinica completa en el MVP salvo necesidad clara.",
        ]
    )

    return "\n".join(lines) + "\n"


def ai_brief(local_brief: str) -> str:
    api_key = os.getenv("OPENAI_API_KEY")
    if not api_key:
        return local_brief

    model = os.getenv("OPENAI_MODEL", "gpt-5")
    program = (ROOT / "program.md").read_text(encoding="utf-8")
    prompt = textwrap.dedent(
        f"""
        Sos asesor estrategico de una startup healthtech argentina.
        Usa el programa y los datos locales para producir un brief accionable,
        conciso y sin vender humo. Separar hechos, hipotesis y proximos pasos.

        PROGRAMA:
        {program}

        BRIEF LOCAL:
        {local_brief}
        """
    ).strip()

    payload = {
        "model": model,
        "input": prompt,
    }
    request = urllib.request.Request(
        "https://api.openai.com/v1/responses",
        data=json.dumps(payload).encode("utf-8"),
        headers={
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        },
        method="POST",
    )

    try:
        with urllib.request.urlopen(request, timeout=90) as response:
            body = json.loads(response.read().decode("utf-8"))
    except urllib.error.URLError as exc:
        return local_brief + f"\n> IA no disponible: {exc}\n"

    return (extract_response_text(body) or local_brief) + "\n"


def brief(args: argparse.Namespace) -> None:
    local = build_local_brief()
    report = ai_brief(local) if args.ai else local
    REPORTS.mkdir(parents=True, exist_ok=True)
    path = REPORTS / f"brief_{today()}.md"
    path.write_text(report, encoding="utf-8")
    print(report)
    print(f"Reporte guardado en: {path}")


def prompts(_: argparse.Namespace) -> None:
    content = textwrap.dedent(
        """
        # Prompts de entrevistas

        ## Paciente

        1. Contame la ultima vez que necesitaste sacar un turno de salud.
        2. Que fue lo mas dificil: saber a quien ir, conseguir turno, pagar,
           cobertura, confianza, distancia u otra cosa?
        3. Antes de atenderte, sabias cuanto ibas a pagar?
        4. Usaste obra social/prepaga o pagaste privado? Por que?
        5. Si una plataforma te dijera especialidad probable, precio,
           disponibilidad y profesional verificado, la usarias?
        6. Que tendria que pasar para que pagues por ahi?
        7. Que te daria desconfianza?

        ## Profesional

        1. Como conseguis pacientes hoy?
        2. Que porcentaje viene por obra social, privado, institucion o
           recomendacion?
        3. Cuanto tardas en cobrar cada canal?
        4. Que tarea administrativa te molesta mas?
        5. Que te haria probar una plataforma nueva?
        6. Preferis pagar comision por consulta o suscripcion mensual?
        7. Que no aceptarias delegar nunca?

        ## Busquedas academicas sugeridas

        python .\\startup_ai_lab.py paper-search --source pubmed --query "telemedicine mental health access appointment scheduling" --limit 5
        python .\\startup_ai_lab.py paper-search --source pubmed --query "patient navigation digital health primary care" --limit 5
        python .\\startup_ai_lab.py paper-search --source arxiv --query "clinical triage large language models safety" --limit 5
        """
    ).strip()
    print(content)


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Startup AI/research lab")
    sub = parser.add_subparsers(dest="command", required=True)

    sub.add_parser("init").set_defaults(func=init)
    sub.add_parser("seed").set_defaults(func=seed_research)
    sub.add_parser("score").set_defaults(func=score)
    sub.add_parser("prompts").set_defaults(func=prompts)
    sub.add_parser("scout").set_defaults(func=scout)

    p_check = sub.add_parser("check-api")
    p_check.add_argument("--live", action="store_true")
    p_check.set_defaults(func=check_api)

    p_check_claude = sub.add_parser("check-claude")
    p_check_claude.add_argument("--live", action="store_true")
    p_check_claude.add_argument("--model", default=None)
    p_check_claude.set_defaults(func=check_claude)

    p_interview = sub.add_parser("interview")
    p_interview.add_argument("--kind", choices=["paciente", "profesional"], required=True)
    p_interview.add_argument("--segment", required=True)
    p_interview.add_argument("--name", default="anonimo")
    p_interview.add_argument("--pain", type=int, choices=range(1, 6), required=True)
    p_interview.add_argument("--urgency", type=int, choices=range(1, 6), required=True)
    p_interview.add_argument("--willingness", type=int, choices=range(1, 6), required=True)
    p_interview.add_argument("--notes", required=True)
    p_interview.set_defaults(func=add_interview)

    p_evidence = sub.add_parser("evidence")
    p_evidence.add_argument("--claim", required=True)
    p_evidence.add_argument("--source", required=True)
    p_evidence.add_argument("--confidence", type=int, choices=range(1, 6), required=True)
    p_evidence.add_argument("--notes", default="")
    p_evidence.set_defaults(func=add_evidence)

    p_experiment = sub.add_parser("experiment")
    p_experiment.add_argument("--name", required=True)
    p_experiment.add_argument("--hypothesis", required=True)
    p_experiment.add_argument("--metric", required=True)
    p_experiment.add_argument("--target", required=True)
    p_experiment.add_argument("--status", default="planned")
    p_experiment.add_argument("--result", default="")
    p_experiment.add_argument("--notes", default="")
    p_experiment.set_defaults(func=add_experiment)

    p_question = sub.add_parser("question")
    p_question.add_argument("--id", required=True)
    p_question.add_argument("--category", required=True)
    p_question.add_argument("--question", required=True)
    p_question.add_argument("--priority", type=int, choices=range(1, 6), default=3)
    p_question.add_argument("--status", default="open")
    p_question.add_argument("--decision-if-true", dest="decision_if_true", default="")
    p_question.add_argument("--decision-if-false", dest="decision_if_false", default="")
    p_question.add_argument("--notes", default="")
    p_question.set_defaults(func=add_question)

    p_questions = sub.add_parser("questions")
    p_questions.add_argument("--status", choices=["open", "answered", "parked", "all"], default="open")
    p_questions.set_defaults(func=list_questions)

    p_source = sub.add_parser("source")
    p_source.add_argument("--id", required=True)
    p_source.add_argument("--category", required=True)
    p_source.add_argument("--source-type", dest="source_type", default="web")
    p_source.add_argument("--title", required=True)
    p_source.add_argument("--url", required=True)
    p_source.add_argument("--credibility", type=int, choices=range(1, 6), default=3)
    p_source.add_argument("--finding", required=True)
    p_source.add_argument("--next-action", dest="next_action", default="")
    p_source.set_defaults(func=add_source)

    p_competitor = sub.add_parser("competitor")
    p_competitor.add_argument("--name", required=True)
    p_competitor.add_argument("--geography", default="")
    p_competitor.add_argument("--model", required=True)
    p_competitor.add_argument("--target", default="")
    p_competitor.add_argument("--pricing-signal", dest="pricing_signal", default="")
    p_competitor.add_argument("--strengths", default="")
    p_competitor.add_argument("--weaknesses", default="")
    p_competitor.add_argument("--relevance", type=int, choices=range(1, 6), default=3)
    p_competitor.add_argument("--url", default="")
    p_competitor.add_argument("--notes", default="")
    p_competitor.set_defaults(func=add_competitor)

    p_regulation = sub.add_parser("regulation")
    p_regulation.add_argument("--id", required=True)
    p_regulation.add_argument("--jurisdiction", default="Argentina")
    p_regulation.add_argument("--norm", required=True)
    p_regulation.add_argument("--topic", required=True)
    p_regulation.add_argument("--impact", required=True)
    p_regulation.add_argument("--risk-level", dest="risk_level", type=int, choices=range(1, 6), default=3)
    p_regulation.add_argument("--url", default="")
    p_regulation.add_argument("--status", default="vigente")
    p_regulation.add_argument("--notes", default="")
    p_regulation.set_defaults(func=add_regulation)

    p_tool = sub.add_parser("tool")
    p_tool.add_argument("--name", required=True)
    p_tool.add_argument("--category", required=True)
    p_tool.add_argument("--use-case", dest="use_case", required=True)
    p_tool.add_argument("--build-or-buy", dest="build_or_buy", default="buy")
    p_tool.add_argument("--cost-signal", dest="cost_signal", default="")
    p_tool.add_argument("--url", default="")
    p_tool.add_argument("--notes", default="")
    p_tool.set_defaults(func=add_tool)

    p_brief = sub.add_parser("brief")
    p_brief.add_argument("--ai", action="store_true")
    p_brief.set_defaults(func=brief)

    p_auto = sub.add_parser("auto-research")
    p_auto.add_argument("--question-id", dest="question_id")
    p_auto.add_argument("--question")
    p_auto.add_argument("--provider", choices=sorted(PROVIDERS), default=None)
    p_auto.add_argument("--model", default=None)
    p_auto.add_argument(
        "--domain-bundle",
        choices=sorted(DOMAIN_BUNDLES.keys()),
        default="official_ar",
    )
    p_auto.add_argument("--context-size", choices=["low", "medium", "high"], default="medium")
    p_auto.add_argument("--reasoning-effort", choices=["low", "medium", "high"], default="low")
    p_auto.add_argument("--max-turns", type=int, default=6)
    p_auto.add_argument("--timeout", type=int, default=600)
    p_auto.set_defaults(func=auto_research)

    p_auto_scout = sub.add_parser("auto-scout")
    p_auto_scout.add_argument("--limit", type=int, default=3)
    p_auto_scout.add_argument("--provider", choices=sorted(PROVIDERS), default=None)
    p_auto_scout.add_argument("--model", default=None)
    p_auto_scout.add_argument(
        "--domain-bundle",
        choices=sorted(DOMAIN_BUNDLES.keys()),
        default="official_ar",
    )
    p_auto_scout.add_argument("--context-size", choices=["low", "medium", "high"], default="medium")
    p_auto_scout.add_argument("--reasoning-effort", choices=["low", "medium", "high"], default="low")
    p_auto_scout.add_argument("--max-turns", type=int, default=6)
    p_auto_scout.add_argument("--timeout", type=int, default=600)
    p_auto_scout.set_defaults(func=auto_scout)

    p_reflect = sub.add_parser("reflect")
    p_reflect.add_argument("--provider", choices=sorted(PROVIDERS), default="local")
    p_reflect.add_argument("--model", default=None)
    p_reflect.add_argument("--reports", type=int, default=3)
    p_reflect.add_argument("--max-questions", type=int, default=5)
    p_reflect.add_argument("--max-experiments", type=int, default=3)
    p_reflect.add_argument("--max-turns", type=int, default=8)
    p_reflect.add_argument("--timeout", type=int, default=900)
    p_reflect.add_argument("--apply", action="store_true")
    p_reflect.set_defaults(func=reflect)

    p_self = sub.add_parser("self-analyze")
    p_self.add_argument("--provider", choices=sorted(PROVIDERS), default="local")
    p_self.add_argument("--model", default=None)
    p_self.add_argument("--reports", type=int, default=4)
    p_self.add_argument("--max-turns", type=int, default=8)
    p_self.add_argument("--timeout", type=int, default=900)
    p_self.add_argument("--apply-candidates", action="store_true")
    p_self.set_defaults(func=self_analyze)

    p_founder_q = sub.add_parser("founder-questions")
    p_founder_q.add_argument("--provider", choices=sorted(PROVIDERS), default="local")
    p_founder_q.add_argument("--model", default=None)
    p_founder_q.add_argument("--limit", type=int, default=7)
    p_founder_q.add_argument("--max-turns", type=int, default=8)
    p_founder_q.add_argument("--timeout", type=int, default=900)
    p_founder_q.add_argument("--apply", action="store_true")
    p_founder_q.set_defaults(func=founder_questions)

    p_founder_a = sub.add_parser("founder-answer")
    p_founder_a.add_argument("--id", required=True)
    p_founder_a.add_argument("--answer", required=True)
    p_founder_a.add_argument("--next-action", default="")
    p_founder_a.set_defaults(func=founder_answer)

    p_loop = sub.add_parser("auto-loop")
    p_loop.add_argument("--cycles", type=int, default=1)
    p_loop.add_argument("--research-limit", type=int, default=2)
    p_loop.add_argument("--provider", choices=sorted(PROVIDERS), default="local")
    p_loop.add_argument("--model", default=None)
    p_loop.add_argument(
        "--domain-bundle",
        choices=sorted(DOMAIN_BUNDLES.keys()),
        default="official_ar",
    )
    p_loop.add_argument("--context-size", choices=["low", "medium", "high"], default="medium")
    p_loop.add_argument("--reasoning-effort", choices=["low", "medium", "high"], default="low")
    p_loop.add_argument("--reports", type=int, default=4)
    p_loop.add_argument("--max-questions", type=int, default=5)
    p_loop.add_argument("--max-experiments", type=int, default=3)
    p_loop.add_argument("--max-founder-questions", type=int, default=5)
    p_loop.add_argument("--max-turns", type=int, default=10)
    p_loop.add_argument("--timeout", type=int, default=1200)
    p_loop.add_argument("--skip-self-analysis", action="store_true")
    p_loop.add_argument("--skip-founder-questions", action="store_true")
    p_loop.set_defaults(func=auto_loop)

    p_papers = sub.add_parser("paper-search")
    p_papers.add_argument("--query", required=True)
    p_papers.add_argument("--source", choices=["pubmed", "arxiv", "all"], default="pubmed")
    p_papers.add_argument("--limit", type=int, default=5)
    p_papers.set_defaults(func=paper_search)

    return parser


def main() -> None:
    load_local_env()
    parser = build_parser()
    args = parser.parse_args()
    args.func(args)


if __name__ == "__main__":
    main()
