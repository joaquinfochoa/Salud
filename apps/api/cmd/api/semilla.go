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
func sembrar(ctx context.Context, svc *service.Profesional) error {
	entradas := []domain.EntradaProfesional{
		{
			Nombre:         "Martín",
			Apellido:       "González",
			Matricula:      "MN 98.234",
			Especialidad:   string(domain.EspecialidadPsicologia),
			Bio:            "Psicólogo clínico con orientación cognitivo-conductual. Atiendo adultos y adolescentes con ansiedad, depresión y crisis vitales. Más de 8 años de experiencia.",
			PrecioConsulta: 1_200_000,
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
			PrecioConsulta: 1_400_000,
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
			PrecioConsulta: 950_000,
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
			PrecioConsulta: 1_500_000,
			Modalidades:    []string{"presencial"},
			Zona:           "CABA",
			ObrasSociales:  []string{"OSDE", "Swiss Medical", "Galeno", "Medifé", "OMINT"},
		},
	}

	// El servicio pone la fecha de alta y resuelve slug y matrícula, igual que
	// en cualquier POST.
	for _, entrada := range entradas {
		if _, err := svc.Crear(ctx, entrada); err != nil {
			return fmt.Errorf("seed: dando de alta a %s %s: %w", entrada.Nombre, entrada.Apellido, err)
		}
	}
	return nil
}
