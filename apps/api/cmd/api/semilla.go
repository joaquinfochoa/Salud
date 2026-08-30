package main

import (
	"context"
	"fmt"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
	"github.com/joaquinfochoa/Salud/apps/api/internal/service"
)

// sembrar carga profesionales de prueba. Solo se llama en desarrollo: main lo
// invoca únicamente cuando APP_ENV=development.
//
// Escribe a través del servicio y no del repositorio a propósito. Un seed que
// escribe por abajo se saltea las reglas de todas las altas: agregar un quinto
// profesional homónimo de otro metía dos registros con el mismo slug y sin
// error. Y al recibir el servicio en vez de un *memory.Profesional, este
// paquete deja de ser otro lugar que sabe qué repositorio está en juego:
// cambiar de repositorio vuelve a ser una sola línea, la de main.
//
// Los datos vienen del prototipo React, en legacy/prototype/src/data. Los
// precios estaban en pesos y acá van en centavos.
// precio devuelve un puntero al monto. EntradaProfesional.PrecioConsulta es
// *int64 justamente para distinguir "no lo mandaron" de "vale cero", así que
// no hay forma de escribir el literal inline.
func precio(centavos int64) *int64 {
	return &centavos
}

func sembrar(ctx context.Context, auth *service.Autenticacion, svc *service.Profesional, agenda *service.Agenda) error {
	entradas := []domain.EntradaProfesional{
		{
			Nombre:         "Martín",
			Apellido:       "González",
			Matricula:      "MN 98.234",
			Especialidad:   string(domain.EspecialidadPsicologia),
			Bio:            "Psicólogo clínico con orientación cognitivo-conductual. Atiendo adultos y adolescentes con ansiedad, depresión y crisis vitales. Más de 8 años de experiencia.",
			PrecioConsulta: precio(1_200_000),
			Modalidades:    []string{"telemedicina", "presencial"},
			Zona:           "CABA",
			ObrasSociales:  []string{"OSDE", "Swiss Medical", "Galeno", "Medifé"},
		},
		{
			Nombre:         "Carolina",
			Apellido:       "Vega",
			Matricula:      "MN 112.087",
			Especialidad:   string(domain.EspecialidadPsicologia),
			Bio:            "Especializada en terapia sistémica y de parejas. Trabajo con adultos en procesos de cambio, duelos y conflictos relacionales.",
			PrecioConsulta: precio(1_400_000),
			Modalidades:    []string{"telemedicina"},
			Zona:           "GBA Norte",
			ObrasSociales:  []string{"OSDE", "OMINT", "Swiss Medical", "Sanitas"},
		},
		{
			Nombre:         "Pablo",
			Apellido:       "Moreno",
			Matricula:      "MN 45.321",
			Especialidad:   string(domain.EspecialidadKinesiologia),
			Bio:            "Kinesiólogo especializado en traumatología deportiva y rehabilitación postquirúrgica. Atiendo a domicilio y en consultorio.",
			PrecioConsulta: precio(950_000),
			Modalidades:    []string{"presencial", "domicilio"},
			Zona:           "CABA",
			ObrasSociales:  []string{"OSDE", "Galeno", "IOMA", "PAMI"},
		},
		{
			Nombre:         "Gabriela",
			Apellido:       "Ríos",
			Matricula:      "MN 67.890",
			Especialidad:   string(domain.EspecialidadOdontologia),
			Bio:            "Odontóloga general con especialización en estética dental. Trabajo con materiales de primera calidad en un consultorio moderno en Palermo.",
			PrecioConsulta: precio(1_500_000),
			Modalidades:    []string{"presencial"},
			Zona:           "CABA",
			ObrasSociales:  []string{"OSDE", "Swiss Medical", "Galeno", "Medifé", "OMINT"},
		},
	}

	// Cada profesional necesita su usuario antes que su perfil: el perfil ya no
	// puede existir sin dueño. El servicio pone la fecha de alta y resuelve
	// slug y matrícula, igual que en cualquier POST.
	for i, entrada := range entradas {
		u, _, err := auth.Registrar(ctx, domain.EntradaUsuario{
			Email:      emailSemilla(entrada.Nombre, entrada.Apellido),
			Contrasena: contrasenaSemilla,
			Nombre:     entrada.Nombre,
			Apellido:   entrada.Apellido,
		})
		if err != nil {
			return fmt.Errorf("seed: registrando a %s %s: %w", entrada.Nombre, entrada.Apellido, err)
		}

		p, err := svc.Crear(ctx, u.ID, entrada)
		if err != nil {
			return fmt.Errorf("seed: dando de alta a %s %s: %w", entrada.Nombre, entrada.Apellido, err)
		}

		if _, _, err := agenda.ReemplazarHorarios(ctx, u.ID, p.ID, horarioSemilla(entrada, i)); err != nil {
			return fmt.Errorf("seed: cargando el horario de %s %s: %w", entrada.Nombre, entrada.Apellido, err)
		}
	}
	return nil
}

// horarioSemilla le da a cada profesional una semana cargada.
//
// Antes el seed los dejaba sin agenda a propósito —"cargarlos es parte de
// probar la API"— y era correcto mientras la API fuera su propio consumidor.
// Con un front, un profesional sin horarios es una tarjeta que dice "todavía no
// publicó horarios": el producto arranca vacío en cada reinicio, y el E2E no
// tiene ningún hueco que elegir.
//
// La modalidad sale de la primera que ofrece el profesional, porque el dominio
// rechaza un bloque con una modalidad que la persona no atiende.
func horarioSemilla(entrada domain.EntradaProfesional, i int) []domain.EntradaHorarioSemanal {
	modalidad := entrada.Modalidades[0]

	// Cada uno atiende días y horas distintas. Cuatro agendas idénticas hacen
	// que el listado se vea generado, que es justo lo que un seed no tiene que
	// parecer.
	semanas := [][]domain.EntradaHorarioSemanal{
		{
			{DiaSemana: "lunes", Desde: "09:00", Hasta: "13:00", DuracionMin: 50, Modalidad: modalidad},
			{DiaSemana: "miercoles", Desde: "14:00", Hasta: "18:00", DuracionMin: 50, Modalidad: modalidad},
			{DiaSemana: "viernes", Desde: "09:00", Hasta: "12:00", DuracionMin: 50, Modalidad: modalidad},
		},
		{
			{DiaSemana: "martes", Desde: "14:00", Hasta: "20:00", DuracionMin: 50, Modalidad: modalidad},
			{DiaSemana: "jueves", Desde: "14:00", Hasta: "20:00", DuracionMin: 50, Modalidad: modalidad},
		},
		{
			{DiaSemana: "lunes", Desde: "08:00", Hasta: "12:00", DuracionMin: 40, Modalidad: modalidad},
			{DiaSemana: "martes", Desde: "08:00", Hasta: "12:00", DuracionMin: 40, Modalidad: modalidad},
			{DiaSemana: "jueves", Desde: "08:00", Hasta: "12:00", DuracionMin: 40, Modalidad: modalidad},
			{DiaSemana: "sabado", Desde: "09:00", Hasta: "13:00", DuracionMin: 40, Modalidad: modalidad},
		},
		{
			{DiaSemana: "miercoles", Desde: "10:00", Hasta: "19:00", DuracionMin: 30, Modalidad: modalidad},
			{DiaSemana: "viernes", Desde: "10:00", Hasta: "19:00", DuracionMin: 30, Modalidad: modalidad},
		},
	}

	return semanas[i%len(semanas)]
}

// contrasenaSemilla es la misma para los cuatro profesionales de prueba. Es
// segura porque sembrar() solo corre con APP_ENV=development —main lo gatea en
// cfg.EsDesarrollo()— y en producción el binario arranca vacío.
const contrasenaSemilla = "desarrollo123"

// emailSemilla arma "martin.gonzalez@ejemplo.com" a partir del nombre.
// Reutiliza GenerarSlug, que ya saca acentos y normaliza: escribir los cuatro
// emails a mano sería una segunda lista que se desincroniza de la primera.
func emailSemilla(nombre, apellido string) string {
	return domain.GenerarSlug(nombre) + "." + domain.GenerarSlug(apellido) + "@ejemplo.com"
}
