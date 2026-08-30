import type { HorarioSemanal, TurnoConPaciente } from "./api";

const ZONA = "America/Argentina/Buenos_Aires";

// El orden es el de Date.getDay(): el domingo es 0. Escribirlo empezando por
// el lunes es el error clásico de desplazamiento, y hay un test que lo cubre.
const DIAS: HorarioSemanal["diaSemana"][] = [
  "domingo",
  "lunes",
  "martes",
  "miercoles",
  "jueves",
  "viernes",
  "sabado",
];

/**
 * Los turnos que el horario nuevo dejaría afuera, y que `PUT /horarios`
 * cancelaría al guardar.
 *
 * ponytail: esto duplica en el front la regla que el back ya tiene, porque la
 * API no ofrece modo simulación — `PUT /horarios` informa cuántos canceló
 * DESPUÉS de cancelarlos. El arreglo real es un `PUT /horarios?simular=true`
 * que devuelva los afectados sin escribir; se hace cuando alguien reporte que
 * el número previsto y el real no coinciden.
 *
 * Las dos mitigaciones importan más que el cálculo: la pantalla LISTA los
 * turnos en vez de contarlos, así que aunque esto se equivoque la persona ve
 * turnos concretos y juzga; y después de guardar se informa el número real que
 * devolvió la API.
 */
export function turnosHuerfanos(
  turnos: TurnoConPaciente[],
  horarios: HorarioSemanal[],
): TurnoConPaciente[] {
  return turnos.filter((turno) => {
    // Un turno ya cancelado no se puede cancelar de nuevo.
    if (turno.estado !== "reservado") return false;

    // ponytail: getDay() devuelve el día en la zona del proceso. Esta pantalla
    // corre en el browser del profesional, que está en Argentina, así que
    // coincide. Si algún día corre en un servidor en UTC, un turno de las 21:00
    // se cuenta como del día siguiente; el arreglo es derivar el día de
    // turno.inicio.slice(0, 10), que ya viene con el offset aplicado.
    const dia = DIAS[new Date(turno.inicio).getDay()];

    // El turno tiene que entrar ENTERO en algún bloque de ese día: el back
    // cancela el que se pasa del final aunque haya empezado adentro. Y alcanza
    // con que lo contenga alguno, porque un día puede tener mañana y tarde.
    return !horarios.some(
      (h) => h.diaSemana === dia && reloj(turno.inicio) >= h.desde && reloj(turno.fin) <= h.hasta,
    );
  });
}

/** `"09:50"`, para comparar contra las horas de reloj del horario. */
function reloj(iso: string): string {
  return new Intl.DateTimeFormat("es-AR", {
    hour: "2-digit",
    minute: "2-digit",
    hour12: false,
    timeZone: ZONA,
  }).format(new Date(iso));
}
