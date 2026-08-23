package domain

import (
	"time"

	// La imagen del contenedor es distroless y no trae /usr/share/zoneinfo,
	// así que sin esto LoadLocation falla y el servidor no arranca en
	// producción — aunque funcione perfecto en la máquina de desarrollo y en
	// el CI, que sí tienen la base de zonas del sistema.
	//
	// Se importa acá y no en main porque ZonaHoraria se inicializa al cargar
	// este paquete, y el orden de init entre paquetes hermanos no está
	// garantizado: importándolo donde se usa, la dependencia queda explícita.
	_ "time/tzdata"
)

// ZonaHoraria es la zona en la que se interpretan los horarios de la agenda.
//
// ponytail: zona fija. El producto es Argentina, que tiene un huso único. El
// día que haya profesionales fuera del país esto pasa a ser un campo de
// Profesional, y el cambio queda local porque el resto del modelo ya trata las
// horas como hora de reloj.
//
// Que la zona sea constante no es lo mismo que guardar UTC: las horas siguen
// siendo de reloj, lo único fijo es a qué reloj se refieren.
var ZonaHoraria = cargarZonaHoraria()

// InicioDelDia devuelve la medianoche del día al que pertenece ese instante,
// en la zona del sistema.
//
// Vive acá y no en el cálculo porque la usan los dos: el dominio para recorrer
// los días del rango y el servicio para normalizar las fechas de la consulta.
// Definirla dos veces es la clase de duplicación que se desincroniza justo
// cuando alguien toca la zona horaria.
func InicioDelDia(t time.Time) time.Time {
	local := t.In(ZonaHoraria)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, ZonaHoraria)
}

func cargarZonaHoraria() *time.Location {
	loc, err := time.LoadLocation("America/Argentina/Buenos_Aires")
	if err != nil {
		// Sin zona horaria la agenda no significa nada, así que es mejor no
		// arrancar que arrancar calculando mal.
		panic("no se pudo cargar la zona horaria: " + err.Error())
	}
	return loc
}
