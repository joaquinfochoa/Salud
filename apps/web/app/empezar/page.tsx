"use client";

import Link from "next/link";
import { useRouter, useSearchParams } from "next/navigation";
import { Suspense, useEffect, useState } from "react";
import { Campo } from "@/componentes/campo";
import { Chips } from "@/componentes/chips";
import { EditorSemana } from "@/componentes/editor-semana";
import { Paso } from "@/componentes/paso";
import {
  VistaPreviaPerfil,
  type Borrador,
  type Foco,
} from "@/componentes/vista-previa-perfil";
import { ESPECIALIDADES } from "@/componentes/tarjeta-profesional";
import {
  ErrorAPI,
  pedir,
  type Especialidad,
  type HorarioSemanal,
  type Profesional,
  type UsuarioActual,
} from "@/lib/api";
import { enCentavos, formatearPrecio } from "@/lib/formato";
import { aPeticionProfesional } from "@/lib/perfil";

const MODALIDADES = ["presencial", "telemedicina", "domicilio"] as const;

const NOMBRE_MODALIDAD: Record<string, string> = {
  presencial: "En consultorio",
  telemedicina: "Videollamada",
  domicilio: "A domicilio",
};

// Texto libre en el contrato, así que es una lista de sugerencias y no un
// catálogo: se puede agregar cualquier otra.
const OBRAS_SOCIALES = [
  "OSDE",
  "Swiss Medical",
  "Galeno",
  "Medifé",
  "OMINT",
  "IOMA",
  "PAMI",
  "Unión Personal",
] as const;

/**
 * El alta de un profesional, de a un paso por vez.
 *
 * Antes eran siete campos juntos en una tarjeta, con la matrícula —lo más
 * intimidante— primero. Acá cada pantalla pide una cosa.
 *
 * **El perfil se crea al terminar el paso 4**, que es el último obligatorio.
 * Todo lo que sigue son cambios sobre algo que ya existe, así que abandonar en
 * la bio o en los horarios no pierde nada y el nudge de `/panel` lo retoma. El
 * contrato exige matrícula, especialidad, precio, modalidades y zona juntos, así
 * que antes de ese punto no hay nada que guardar.
 *
 * **Termina cuando podés recibir turnos**, no cuando el perfil existe: sin
 * horarios cargados nadie te puede reservar, y cerrar antes sería dejar al
 * profesional a mitad de camino.
 */
function Onboarding() {
  const router = useRouter();
  const parametros = useSearchParams();

  // El paso viaja en la URL para que el botón de atrás del browser vuelva al
  // paso anterior en vez de salir del flujo. El componente no se desmonta al
  // cambiar solo la query, así que lo tipeado sobrevive.
  const pedido = Number(parametros.get("paso") ?? "1");

  const [tieneSesion, setTieneSesion] = useState<boolean | null>(null);
  const [perfil, setPerfil] = useState<Profesional | null>(null);

  const [nombre, setNombre] = useState("");
  const [apellido, setApellido] = useState("");
  const [email, setEmail] = useState("");
  const [contrasena, setContrasena] = useState("");
  const [matricula, setMatricula] = useState("");
  const [especialidad, setEspecialidad] = useState<Especialidad>("psicologia");
  const [modalidades, setModalidades] = useState<string[]>([]);
  const [zona, setZona] = useState("");
  const [precio, setPrecio] = useState("");
  const [obrasSociales, setObrasSociales] = useState<string[]>([]);
  const [bio, setBio] = useState("");
  const [horarios, setHorarios] = useState<HorarioSemanal[]>([]);

  const [aviso, setAviso] = useState<React.ReactNode>(null);
  const [error, setError] = useState<ErrorAPI | null>(null);
  const [enviando, setEnviando] = useState(false);

  // Quién sos antes de empezar: con sesión el paso de la cuenta sobra, y con
  // perfil ya creado esta pantalla no tiene sentido.
  useEffect(() => {
    let vigente = true;
    pedir<UsuarioActual>("/api/v1/usuarios/yo")
      .then((yo) => {
        if (!vigente) return;
        if (yo.perfilProfesionalId) {
          router.replace("/panel");
          return;
        }
        setNombre(yo.nombre);
        setApellido(yo.apellido);
        setTieneSesion(true);
      })
      .catch(() => vigente && setTieneSesion(false));
    return () => {
      vigente = false;
    };
  }, [router]);

  if (tieneSesion === null) {
    return (
      <main className="mx-auto w-full max-w-lg px-4 py-20 text-center sm:px-6">
        <p className="text-tinta-suave">Cargando…</p>
      </main>
    );
  }

  // Con sesión el flujo arranca en la matrícula, así que hay un paso menos.
  const primero = tieneSesion ? 2 : 1;
  const total = tieneSesion ? 6 : 7;
  const numeroVisible = (n: number) => (tieneSesion ? n - 1 : n);

  const completo: Record<number, boolean> = {
    1: Boolean(nombre.trim() && apellido.trim() && email.trim() && contrasena.length >= 8),
    2: Boolean(matricula.trim()),
    3: modalidades.length > 0 && Boolean(zona.trim()),
    4: enCentavos(precio) !== null,
    5: true,
    6: true,
  };

  // Nadie entra a un paso salteando los anteriores, ni escribiendo la URL a
  // mano: se cae al primero que le falta.
  let paso = Math.min(Math.max(pedido, primero), 7);
  for (let n = primero; n < paso; n++) {
    if (!completo[n] || (n === 4 && !perfil)) {
      paso = n;
      break;
    }
  }
  // Y la URL se corrige, no solo lo que se dibuja: dejarla en el paso pedido
  // haría que refrescar vuelva al mismo lugar equivocado, y que el botón de
  // atrás cuente pasos que nunca se vieron.
  if (paso !== pedido) router.replace(`/empezar?paso=${paso}`);

  const irA = (n: number) => {
    setAviso(null);
    setError(null);
    router.push(`/empezar?paso=${n}`);
  };

  async function crearCuenta() {
    setEnviando(true);
    setAviso(null);
    setError(null);
    try {
      // El back deja la sesión abierta en la misma respuesta.
      await pedir("/api/v1/usuarios", {
        method: "POST",
        body: JSON.stringify({ email, contrasena, nombre, apellido }),
      });
      irA(2);
    } catch (e) {
      if (e instanceof ErrorAPI && e.estado === 409) {
        setAviso(
          <>
            Ya tenés una cuenta con ese email.{" "}
            <Link
              href="/entrar?volver=%2Fempezar"
              className="font-bold text-accion underline"
            >
              Entrar
            </Link>{" "}
            y seguí desde ahí.
          </>,
        );
      } else if (e instanceof ErrorAPI && e.estado === 422) {
        setError(e);
      } else {
        setAviso("No pudimos crear la cuenta. Probá de nuevo.");
      }
    } finally {
      setEnviando(false);
    }
  }

  async function crearPerfil() {
    const centavos = enCentavos(precio);
    if (centavos === null) return;

    setEnviando(true);
    setAviso(null);
    setError(null);
    try {
      const creado = await pedir<Profesional>("/api/v1/profesionales", {
        method: "POST",
        body: JSON.stringify({
          nombre,
          apellido,
          matricula,
          especialidad,
          bio: "",
          precioConsultaCentavos: centavos,
          modalidades,
          zona,
          obrasSociales,
          anticipacionMinimaMin: 120,
          horizonteDias: 60,
        }),
      });
      setPerfil(creado);
      irA(5);
    } catch (e) {
      if (e instanceof ErrorAPI && e.estado === 409) {
        setAviso("Ya hay un perfil con esa matrícula.");
      } else if (e instanceof ErrorAPI && e.estado === 422) {
        setError(e);
        // El 422 casi siempre es de la matrícula, que se cargó dos pasos atrás.
        // Dejar a la persona en este paso con un error que no puede corregir
        // acá sería un callejón.
        if (e.porCampo("matricula")) irA(2);
      } else {
        setAviso("No pudimos crear el perfil. Probá de nuevo.");
      }
    } finally {
      setEnviando(false);
    }
  }

  async function guardarBio() {
    if (!perfil) return;
    setEnviando(true);
    try {
      await pedir(`/api/v1/profesionales/${perfil.id}`, {
        method: "PUT",
        body: JSON.stringify(aPeticionProfesional(perfil, { bio })),
      });
      setPerfil({ ...perfil, bio });
      irA(6);
    } catch {
      setAviso("No pudimos guardar la descripción. Podés completarla más tarde.");
    } finally {
      setEnviando(false);
    }
  }

  async function guardarHorarios() {
    if (!perfil) return;
    setEnviando(true);
    try {
      await pedir(`/api/v1/profesionales/${perfil.id}/horarios`, {
        method: "PUT",
        body: JSON.stringify({ horarios }),
      });
      irA(7);
    } catch (e) {
      const detalle =
        e instanceof ErrorAPI && e.problema.errores?.length
          ? e.problema.errores.map((x) => x.mensaje).join(". ")
          : null;
      setAviso(detalle ? `No se guardó: ${detalle}.` : "No pudimos guardar los horarios.");
    } finally {
      setEnviando(false);
    }
  }

  const comun = { total, aviso, enviando };

  const borrador: Borrador = {
    nombre,
    apellido,
    matricula,
    // Recién a partir del paso que la pide: antes, el select todavía no se vio.
    especialidad: paso >= 2 ? especialidad : null,
    modalidades,
    zona,
    precio,
    obrasSociales,
    bio,
    horarios,
  };

  // Qué parte de la vista previa está llenando cada paso. Es lo que ata el
  // campo con su resultado sin tener que explicarlo.
  const FOCO: Record<number, Foco> = {
    1: "nombre",
    2: "matricula",
    3: "atencion",
    4: "precio",
    5: "bio",
    6: "agenda",
  };

  /**
   * Los pasos a la izquierda, cómo va quedando el perfil a la derecha.
   *
   * En móvil se apila y la vista previa queda debajo del formulario: primero se
   * completa, después se ve el efecto. En escritorio quedan lado a lado y la
   * vista previa acompaña el scroll.
   */
  const conVistaPrevia = (contenido: React.ReactNode) => (
    <main className="mx-auto w-full max-w-6xl px-4 py-10 sm:px-6 sm:py-14">
      <div className="grid gap-10 lg:grid-cols-[minmax(0,32rem)_minmax(0,1fr)] lg:gap-16">
        {contenido}
        <VistaPreviaPerfil borrador={borrador} foco={FOCO[paso] ?? "nombre"} />
      </div>
    </main>
  );

  if (paso === 1) {
    return conVistaPrevia(
      <Paso
        {...comun}
        numero={1}
        titulo="Empecemos por tu cuenta"
        ayuda="Con la misma cuenta publicás tu agenda y, si querés, reservás turnos."
        avanzar={crearCuenta}
        puedeAvanzar={completo[1]}
        textoAvanzar="Crear cuenta"
      >
        <div className="grid gap-5 sm:grid-cols-2">
          <Campo nombre="nombre" etiqueta="Nombre" valor={nombre} onCambiar={setNombre} autoComplete="given-name" error={error} />
          <Campo nombre="apellido" etiqueta="Apellido" valor={apellido} onCambiar={setApellido} autoComplete="family-name" error={error} />
        </div>
        <Campo nombre="email" etiqueta="Email" tipo="email" valor={email} onCambiar={setEmail} autoComplete="email" error={error} />
        <Campo
          nombre="contrasena"
          etiqueta="Contraseña"
          tipo="password"
          valor={contrasena}
          onCambiar={setContrasena}
          autoComplete="new-password"
          ayuda="Al menos 8 caracteres"
          error={error}
        />
      </Paso>
    );
  }

  if (paso === 2) {
    return conVistaPrevia(
      <Paso
        {...comun}
        numero={numeroVisible(2)}
        titulo="Tu matrícula"
        ayuda="Se publica en tu perfil. Es lo que le permite a un paciente confirmar que sos quien decís ser."
        avanzar={() => irA(3)}
        puedeAvanzar={completo[2]}
        volver={tieneSesion ? undefined : () => irA(1)}
      >
        <Campo
          nombre="matricula"
          etiqueta="Número de matrícula"
          valor={matricula}
          onCambiar={setMatricula}
          ayuda="Como figura en tu credencial: MN 98234, M.P. 12345."
          error={error}
        />
        <div className="grid min-w-0 gap-1.5">
          <label htmlFor="especialidad" className="text-sm font-semibold">
            Especialidad
          </label>
          <select
            id="especialidad"
            value={especialidad}
            onChange={(e) => setEspecialidad(e.target.value as Especialidad)}
            className="h-11 w-full rounded-lg border border-borde bg-superficie px-3"
          >
            {Object.entries(ESPECIALIDADES).map(([valor, texto]) => (
              <option key={valor} value={valor}>
                {texto}
              </option>
            ))}
          </select>
        </div>
      </Paso>
    );
  }

  if (paso === 3) {
    return conVistaPrevia(
      <Paso
        {...comun}
        numero={numeroVisible(3)}
        titulo="Cómo atendés"
        avanzar={() => irA(4)}
        puedeAvanzar={completo[3]}
        volver={() => irA(2)}
      >
        <Chips
          etiqueta="Modalidades"
          opciones={MODALIDADES.map((m) => NOMBRE_MODALIDAD[m])}
          elegidas={modalidades.map((m) => NOMBRE_MODALIDAD[m] ?? m)}
          onCambiar={(elegidas) =>
            setModalidades(MODALIDADES.filter((m) => elegidas.includes(NOMBRE_MODALIDAD[m])))
          }
        />
        <Campo
          nombre="zona"
          etiqueta="Zona"
          valor={zona}
          onCambiar={setZona}
          ayuda="Dónde atendés: CABA, Rosario, Vicente López."
          error={error}
        />
      </Paso>
    );
  }

  if (paso === 4) {
    return conVistaPrevia(
      <Paso
        {...comun}
        numero={numeroVisible(4)}
        titulo="Cuánto cobrás"
        ayuda="Es lo primero que mira un paciente después de tu especialidad."
        avanzar={crearPerfil}
        puedeAvanzar={completo[4]}
        textoAvanzar="Crear mi perfil"
        volver={() => irA(3)}
      >
        <Campo
          nombre="precio"
          etiqueta="Precio de la consulta"
          valor={precio}
          onCambiar={setPrecio}
          // Se formatea al salir del campo y no en cada tecla: reformatear
          // mientras se escribe mueve el cursor de lugar.
          onSalir={() => {
            const c = enCentavos(precio);
            if (c !== null) setPrecio(formatearPrecio(c));
          }}
          ayuda="En pesos. Lo podés cambiar cuando quieras."
          error={error}
        />
        <Chips
          etiqueta="Obras sociales que aceptás"
          opciones={OBRAS_SOCIALES}
          elegidas={obrasSociales}
          onCambiar={setObrasSociales}
          libre
        />
      </Paso>
    );
  }

  if (paso === 5) {
    return conVistaPrevia(
      <Paso
        {...comun}
        numero={numeroVisible(5)}
        titulo="Contá cómo trabajás"
        ayuda="Es lo primero que lee un paciente. Podés escribirlo ahora o más tarde."
        avanzar={guardarBio}
        puedeAvanzar
        volver={() => irA(4)}
        masTarde={() => irA(6)}
      >
        <div className="grid min-w-0 gap-1.5">
          <label htmlFor="bio" className="text-sm font-semibold">
            Descripción
          </label>
          <textarea
            id="bio"
            value={bio}
            onChange={(e) => setBio(e.target.value)}
            rows={5}
            placeholder="Qué atendés, cómo son tus sesiones, en qué te especializás."
            className="w-full rounded-lg border border-borde p-3"
          />
        </div>
      </Paso>
    );
  }

  if (paso === 6) {
    return conVistaPrevia(
      <Paso
        {...comun}
        numero={numeroVisible(6)}
        titulo="Tu semana"
        ayuda="Los horarios que cargues acá son los que un paciente puede reservar. Sin esto, tu perfil se ve pero nadie puede pedirte turno."
        avanzar={guardarHorarios}
        puedeAvanzar={horarios.length > 0}
        textoAvanzar="Publicar mi agenda"
        volver={() => irA(5)}
        masTarde={() => router.push("/panel")}
      >
        <EditorSemana horarios={horarios} onCambiar={setHorarios} />
      </Paso>
    );
  }

  // Paso 7: el final. No es una felicitación, es el link a lo que consiguió.
  return (
    <main className="mx-auto w-full max-w-lg px-4 py-16 text-center sm:px-6 sm:py-24">
      <p className="text-sm font-semibold uppercase tracking-wide text-libre">
        Tu agenda está publicada
      </p>
      <h1 className="mt-3 text-3xl font-black tracking-tight">
        Ya te pueden reservar turnos
      </h1>
      <p className="mt-3 text-tinta-suave">
        Tu perfil aparece en la búsqueda con los horarios que cargaste. Cuando
        alguien reserve, lo vas a ver en tu agenda.
      </p>

      <div className="mt-10 flex flex-col items-center gap-3">
        <Link
          href="/panel"
          className="rounded-lg bg-accion px-6 py-3 font-bold text-white transition-colors hover:bg-accion-viva"
        >
          Ir a mi panel
        </Link>
        {perfil && (
          <Link
            href={`/perfiles/${perfil.slug}`}
            className="text-sm font-semibold text-tinta-suave underline hover:text-accion"
          >
            Ver cómo te ve un paciente
          </Link>
        )}
      </div>
    </main>
  );
}

// useSearchParams necesita un límite de Suspense: sin él, Next no puede
// prerenderizar nada de esta ruta.
export default function Empezar() {
  return (
    <Suspense>
      <Onboarding />
    </Suspense>
  );
}
