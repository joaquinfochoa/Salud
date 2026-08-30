import { expect, type Page, test } from "@playwright/test";

const API = "http://localhost:8080";

/** `2026-09-05` del próximo sábado. */
function proximoSabado(): string {
  const d = new Date();
  d.setDate(d.getDate() + ((6 - d.getDay() + 7) % 7 || 7));
  return new Intl.DateTimeFormat("en-CA", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    timeZone: "America/Argentina/Buenos_Aires",
  }).format(d);
}

/**
 * Se registra POR LA PANTALLA, no con un fetch a la API.
 *
 * La primera versión de este helper creaba la cuenta con un fetch directo, y
 * eso tapó un agujero real: no existía ninguna pantalla de registro. El único
 * alta vivía dentro del formulario de reservar un turno, así que un profesional
 * que llegaba desde la landing caía en /entrar sin forma de crear la cuenta. El
 * test pasaba en verde mientras una persona no podía registrarse.
 *
 * Un helper que evita la interfaz no prueba la interfaz.
 */
async function registrar(page: Page, nombre: string, apellido: string, volver?: string) {
  const email = `${nombre.toLowerCase()}.${Date.now()}@ejemplo.com`;
  await page.goto(volver ? `/crear-cuenta?volver=${encodeURIComponent(volver)}` : "/crear-cuenta");
  await page.getByLabel("Nombre").fill(nombre);
  await page.getByLabel("Apellido").fill(apellido);
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Contraseña").fill("desarrollo123");
  await page.getByRole("button", { name: "Crear cuenta" }).click();
  return email;
}

/**
 * El circuito que justifica la etapa entera: un profesional carga su semana y
 * esos huecos aparecen en su perfil público.
 *
 * Si esto pasa, las dos mitades del marketplace están conectadas. Sin
 * profesionales cargando su agenda no hay nada que buscar, y hasta esta etapa
 * eso solo se podía hacer con curl.
 *
 * El profesional se crea acá y no se usa el del seed. La primera versión usaba
 * a Martín y pasaba una sola vez: la segunda corrida contra la misma API
 * agregaba un bloque que se solapaba con el de la anterior, la API lo rechazaba
 * y el test fallaba. Un test que solo pasa contra una base recién creada es una
 * trampa que explota más adelante.
 */
test("un profesional carga su semana y aparece en su perfil público", async ({ page }) => {
  // Se entra por donde entra un profesional de verdad: el CTA de la landing de
  // captación, sin cuenta previa.
  await page.goto("/profesionales");
  await page.getByRole("link", { name: "Crear mi perfil" }).first().click();
  await expect(page).toHaveURL("/crear-cuenta?volver=%2Fpanel%2Fperfil");

  await page.getByLabel("Nombre").fill("Renata");
  await page.getByLabel("Apellido").fill("Kine");
  await page.getByLabel("Email").fill(`renata.${Date.now()}@ejemplo.com`);
  await page.getByLabel("Contraseña").fill("desarrollo123");
  await page.getByRole("button", { name: "Crear cuenta" }).click();

  // Registrarse abre la sesión en la misma respuesta: no hay que entrar después.
  await expect(page).toHaveURL("/panel/perfil");

  await page.getByLabel("Matrícula").fill(`MP ${Date.now().toString().slice(-6)}`);
  await page.getByLabel("Especialidad").selectOption("kinesiologia");
  await page.getByLabel("Descripción").fill("Kinesióloga. Rehabilitación deportiva.");
  await page.getByLabel("Precio de la consulta").fill("11000");
  await page.getByLabel("Zona").fill("CABA");
  await page.getByRole("button", { name: "En consultorio" }).click();
  await page.getByRole("button", { name: "Crear mi perfil" }).click();

  await expect(page).toHaveURL("/panel");
  // Recién creada no tiene horarios, así que nadie puede reservarle.
  await expect(page.getByText("Todavía no cargaste tus horarios")).toBeVisible();

  await page.getByRole("link", { name: "Configurar" }).click();
  await expect(page).toHaveURL("/panel/horarios");

  await page.getByRole("button", { name: "Agregar bloque a Sábado" }).click();
  await page.getByLabel("Desde").last().fill("10:00");
  await page.getByLabel("Hasta").last().fill("12:00");
  await page.getByRole("button", { name: "Guardar" }).click();

  // Filtrado por texto: el anunciador de rutas de Next también es role="alert".
  await expect(
    page.getByRole("alert").filter({ hasText: "Horarios guardados" }),
  ).toBeVisible();

  // Y del otro lado del marketplace. Con ?dia= porque sin él el perfil abre en
  // el primer día con huecos.
  const slug = await page.evaluate(async (API) => {
    const yo = await (await fetch(`${API}/api/v1/usuarios/yo`, { credentials: "include" })).json();
    const p = await (
      await fetch(`${API}/api/v1/profesionales/${yo.perfilProfesionalId}`)
    ).json();
    return p.slug as string;
  }, API);

  await page.goto(`/perfiles/${slug}?dia=${proximoSabado()}`);
  await expect(page.getByRole("link", { name: "10:00" })).toBeVisible();
  await expect(page.getByRole("link", { name: "10:50" })).toBeVisible();
});

test("una cuenta nueva sin perfil profesional queda en sus turnos", async ({ page }) => {
  await registrar(page, "Paula", "Paciente");
  await expect(page).toHaveURL("/turnos");
});

/**
 * El camino que estaba roto: llegar a /entrar sin cuenta y poder salir de ahí.
 *
 * Antes la única salida decía "se crea sola cuando reservás tu primer turno",
 * que a un profesional lo mandaba a reservarse un turno con otro profesional.
 */
test("desde entrar se puede crear una cuenta, conservando a dónde iba", async ({ page }) => {
  await page.goto("/entrar?volver=%2Fpanel%2Fperfil");
  await page.getByRole("link", { name: "Crear una" }).click();
  await expect(page).toHaveURL("/crear-cuenta?volver=%2Fpanel%2Fperfil");

  // Y la vuelta, para quien sí tiene cuenta.
  await page.getByRole("link", { name: "Entrar" }).click();
  await expect(page).toHaveURL("/entrar?volver=%2Fpanel%2Fperfil");
});

test("registrarse con un email que ya existe ofrece entrar", async ({ page }) => {
  const email = await registrar(page, "Repetida", "Cuenta");
  await expect(page).toHaveURL("/turnos");

  await page.goto("/crear-cuenta");
  await page.getByLabel("Nombre").fill("Repetida");
  await page.getByLabel("Apellido").fill("Cuenta");
  await page.getByLabel("Email").fill(email);
  await page.getByLabel("Contraseña").fill("desarrollo123");
  await page.getByRole("button", { name: "Crear cuenta" }).click();

  await expect(
    page.getByRole("alert").filter({ hasText: "Ya tenés una cuenta" }),
  ).toBeVisible();
});

/**
 * El panel es privado, y esto lo prueba desde afuera: sin sesión se rebota a
 * entrar conservando a dónde iba, para volver ahí después.
 */
test("el panel no se abre sin sesión", async ({ page }) => {
  await page.goto("/panel");
  await expect(page).toHaveURL("/entrar?volver=%2Fpanel");
});
