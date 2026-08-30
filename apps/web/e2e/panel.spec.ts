import { expect, type Page, test } from "@playwright/test";

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
test("un profesional se da de alta paso a paso y queda reservable", async ({ page }) => {
  await page.goto("/profesionales");
  await page.getByRole("link", { name: "Crear mi perfil" }).first().click();
  await expect(page).toHaveURL("/empezar");

  // Paso 1: la cuenta.
  await page.getByLabel("Nombre").fill("Renata");
  await page.getByLabel("Apellido").fill("Kine");
  await page.getByLabel("Email").fill(`renata.${Date.now()}@ejemplo.com`);
  await page.getByLabel("Contraseña").fill("desarrollo123");
  await page.getByRole("button", { name: "Crear cuenta" }).click();

  // Paso 2: la matrícula.
  await expect(page).toHaveURL("/empezar?paso=2");
  await page.getByLabel("Número de matrícula").fill(`MP ${Date.now().toString().slice(-6)}`);
  await page.getByLabel("Especialidad").selectOption("kinesiologia");
  await page.getByRole("button", { name: "Continuar" }).click();

  // Paso 3: cómo atendés.
  await expect(page).toHaveURL("/empezar?paso=3");
  await page.getByRole("button", { name: "En consultorio" }).click();
  await page.getByLabel("Zona").fill("CABA");
  await page.getByRole("button", { name: "Continuar" }).click();

  // Paso 4: el precio. Acá se crea el perfil.
  await expect(page).toHaveURL("/empezar?paso=4");
  await page.getByLabel("Precio de la consulta").fill("11000");
  await page.getByRole("button", { name: "Crear mi perfil" }).click();

  // Paso 5: la bio es opcional y se puede dejar para después.
  await expect(page).toHaveURL("/empezar?paso=5");
  await page.getByRole("button", { name: "Completar más tarde" }).click();

  // Paso 6: la semana, que es lo que lo hace reservable.
  await expect(page).toHaveURL("/empezar?paso=6");
  await page.getByRole("button", { name: "Agregar bloque a Sábado" }).click();
  await page.getByLabel("Desde").last().fill("10:00");
  await page.getByLabel("Hasta").last().fill("12:00");
  await page.getByRole("button", { name: "Publicar mi agenda" }).click();

  // Y el final dice lo que consiguió, no "felicitaciones".
  await expect(page).toHaveURL("/empezar?paso=7");
  await expect(page.getByRole("heading", { name: "Ya te pueden reservar turnos" })).toBeVisible();

  // Del otro lado del marketplace: los huecos existen de verdad.
  // Se lee el href en vez de clickear y mirar la URL: page.url() puede leerse
  // antes de que la navegación termine, y el test falla por carrera y no por el
  // producto.
  const suPerfil = await page
    .getByRole("link", { name: "Ver cómo te ve un paciente" })
    .getAttribute("href");
  await page.goto(`${suPerfil}?dia=${proximoSabado()}`);
  await expect(page.getByRole("link", { name: "10:00" })).toBeVisible();
  await expect(page.getByRole("link", { name: "10:50" })).toBeVisible();
});

/** Los pasos obligatorios no se saltean, ni escribiendo la URL a mano. */
test("no se puede entrar a un paso salteando los anteriores", async ({ page }) => {
  await page.goto("/empezar?paso=4");
  await expect(page).toHaveURL("/empezar?paso=1");
  await expect(page.getByRole("heading", { name: "Empecemos por tu cuenta" })).toBeVisible();
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
