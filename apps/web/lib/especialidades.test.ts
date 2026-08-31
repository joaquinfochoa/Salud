import { describe, expect, it } from "vitest";
import { ESPECIALIDADES } from "./especialidades";
import type { components } from "./contrato";

type Especialidad = components["schemas"]["Especialidad"];

/**
 * La tabla de nombres para mostrar tiene que cubrir el enum del contrato, ni
 * más ni menos.
 *
 * Son dos listas en dos lugares, y la de acá se escribe a mano porque los
 * valores del dominio no llevan acentos. Sin este test, agregar una
 * especialidad en Go la deja apareciendo como "alergia-e-inmunologia" en el
 * select del alta, y nadie se entera hasta que un profesional la ve.
 *
 * El tipo `Especialidad` sale de `pnpm contrato`, así que esta lista es la del
 * contrato de verdad, no una copia.
 */
const DEL_CONTRATO: Especialidad[] = [
  "alergia-e-inmunologia",
  "cardiologia",
  "clinica-medica",
  "dermatologia",
  "endocrinologia",
  "enfermeria",
  "fonoaudiologia",
  "gastroenterologia",
  "ginecologia",
  "kinesiologia",
  "neumonologia",
  "neurologia",
  "nutricion",
  "obstetricia",
  "odontologia",
  "oftalmologia",
  "otorrinolaringologia",
  "pediatria",
  "podologia",
  "psicologia",
  "psicopedagogia",
  "psiquiatria",
  "reumatologia",
  "terapia-ocupacional",
  "traumatologia",
  "urologia",
];

describe("ESPECIALIDADES", () => {
  it("tiene un nombre para cada especialidad del contrato", () => {
    const sinNombre = DEL_CONTRATO.filter((e) => !ESPECIALIDADES[e]);
    expect(sinNombre).toEqual([]);
  });

  // Al revés también: un nombre para una especialidad que ya no existe queda
  // como opción muerta en el select.
  it("no tiene nombres de más", () => {
    const deMas = Object.keys(ESPECIALIDADES).filter(
      (k) => !DEL_CONTRATO.includes(k as Especialidad),
    );
    expect(deMas).toEqual([]);
  });

  it("los nombres visibles están ordenados alfabéticamente", () => {
    const visibles = Object.values(ESPECIALIDADES);
    expect(visibles).toEqual([...visibles].sort((a, b) => a.localeCompare(b, "es")));
  });
});
