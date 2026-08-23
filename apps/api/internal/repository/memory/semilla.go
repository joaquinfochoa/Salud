package memory

import (
	"context"
	"fmt"
	"time"

	"github.com/joaquinfochoa/Salud/apps/api/internal/domain"
)

// Sembrar carga profesionales de prueba. Solo se llama en desarrollo: main.go lo
// invoca únicamente cuando APP_ENV=development.
//
// Los datos vienen del prototipo React, en legacy/prototype/src/data. Los
// precios estaban en pesos y acá van en centavos.
func Sembrar(ctx context.Context, repo *Profesional) error {
	ahora := time.Now().UTC()

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

	for i, entrada := range entradas {
		p, err := domain.NuevoProfesional(entrada, ahora.Add(time.Duration(i)*time.Second))
		if err != nil {
			return fmt.Errorf("seed: profesional %d (%s %s): %w", i, entrada.Nombre, entrada.Apellido, err)
		}
		if err := repo.Crear(ctx, p); err != nil {
			return fmt.Errorf("seed: guardando %s %s: %w", entrada.Nombre, entrada.Apellido, err)
		}
	}
	return nil
}
