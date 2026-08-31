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
  await page.getByLabel("Celular").fill("11 5000-1234");
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
  await page.getByRole("link", { name: "Publicar mi agenda" }).first().click();
  await expect(page).toHaveURL("/empezar");

  // Paso 1: la cuenta.
  await page.getByLabel("Nombre").fill("Renata");
  await page.getByLabel("Apellido").fill("Kine");
  await page.getByLabel("Email").fill(`renata.${Date.now()}@ejemplo.com`);
  await page.getByLabel("Celular").fill("11 5000-4321");
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
  await page.getByLabel("Celular").fill("11 5000-1234");
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

/**
 * Cerrar sesión existía en la API desde la etapa de autenticación y no lo
 * llamaba nadie en el front: se entraba y no se salía nunca.
 *
 * El test comprueba las dos mitades: que la interfaz lo ofrezca, y que la
 * sesión quede cerrada del lado del servidor. Solo la primera dejaría pasar un
 * botón que borra la cookie y deja la sesión viva.
 */
test("se puede cerrar sesión, y queda cerrada", async ({ page }) => {
  await registrar(page, "Sale", "Delaapp");
  await expect(page).toHaveURL("/turnos");

  const encabezado = page.getByRole("navigation", { name: "Principal" });
  await expect(encabezado.getByRole("link", { name: "Mis turnos" })).toBeVisible();

  await encabezado.getByRole("button", { name: "Cerrar sesión" }).click();

  await expect(page).toHaveURL("/");
  await expect(encabezado.getByRole("link", { name: "Entrar" })).toBeVisible();

  const estado = await page.evaluate(async () => {
    const r = await fetch("http://localhost:8080/api/v1/usuarios/yo", {
      credentials: "include",
    });
    return r.status;
  });
  expect(estado).toBe(401);
});

/**
 * El profesional puede cancelar un turno puntual.
 *
 * El servicio ya lo permitía —verificarParte acepta al dueño del perfil— pero
 * la pantalla no lo ofrecía: la única salida era bloquear el rango entero y
 * cancelarle a todos los de esa franja.
 */
test("el profesional cancela un turno y el paciente lo ve cancelado", async ({ page }) => {
  // El apellido lleva la marca de tiempo, igual que el email. Sin eso, cada
  // corrida deja otra "Vera Cancela" en la misma agenda y el locator agarra la
  // de una corrida anterior, que ya está cancelada y no tiene botón.
  const apellido = `Cancela${Date.now()}`;

  // Se registra reservando, que es como llega un paciente de verdad.
  await page.goto("/perfiles/martin-gonzalez");
  await page.getByRole("link", { name: /^\d{2}:\d{2}$/ }).first().click();
  await page.getByLabel("Email").fill(`vera.${Date.now()}@ejemplo.com`);
  await page.getByLabel("Contraseña").fill("desarrollo123");
  await page.getByLabel("Nombre", { exact: true }).fill("Vera");
  await page.getByLabel("Apellido").fill(apellido);
  await page.getByLabel("Celular").fill("11 8500-1234");
  await page.getByLabel("Motivo de la consulta").fill("control");
  await page.getByRole("button", { name: /^Reservar \d{2}:\d{2}$/ }).click();
  await expect(page.locator("main").getByText("Turno reservado")).toBeVisible();

  // El profesional lo cancela, confirmando.
  const pro = await page.context().browser()!.newPage();
  await pro.goto("/entrar");
  await pro.getByLabel("Email").fill("martin.gonzalez@ejemplo.com");
  await pro.getByLabel("Contraseña").fill("desarrollo123");
  await pro.getByRole("button", { name: "Entrar" }).click();
  await pro.waitForURL("**/panel");
  await pro.goto("/panel/agenda");

  // La agenda abre en el primer día CON turnos, que no tiene por qué ser el de
  // este: el seed ya trae otros. Se navega al día del turno recién reservado.
  const dia = await page.evaluate(async () => {
    const r = await fetch("http://localhost:8080/api/v1/turnos", { credentials: "include" });
    const { datos } = await r.json();
    return new Intl.DateTimeFormat("es-AR", {
      weekday: "long",
      day: "numeric",
      month: "long",
      timeZone: "America/Argentina/Buenos_Aires",
    }).format(new Date(datos[0].inicio));
  });
  await pro.getByRole("button", { name: dia }).click();

  const fila = pro.locator("main li").filter({ hasText: apellido }).first();
  await fila.getByRole("button", { name: "Cancelar" }).click();
  await expect(pro.getByRole("heading", { name: /¿Cancelar el turno de Vera\?/ })).toBeVisible();
  await pro.getByRole("button", { name: "Sí, cancelar" }).click();
  await expect(
    pro.getByRole("alert").filter({ hasText: "Turno cancelado" }),
  ).toBeVisible();

  // Y el paciente lo ve del otro lado. Es la mitad que importa: cancelar sin
  // que la otra persona se entere no es cancelar.
  await page.goto("/turnos");
  await expect(page.getByText("Cancelado", { exact: false }).first()).toBeVisible();
  await pro.close();
});

/**
 * El área del paciente tiene sus propias secciones, y cambiar el email cambia
 * con qué se entra.
 *
 * El email es la identidad de login: si el cambio se guardara sin que se pueda
 * entrar con el nuevo, la persona quedaría afuera de su cuenta. Por eso el test
 * cierra sesión y vuelve a entrar en vez de conformarse con el 200.
 */
test("el paciente edita su cuenta y entra con el email nuevo", async ({ page }) => {
  await registrar(page, "Edita", "Sucuenta");
  await expect(page).toHaveURL("/turnos");

  const secciones = page.getByRole("navigation", { name: "Mi cuenta" });
  await expect(secciones.getByRole("link", { name: "Mis turnos" })).toBeVisible();
  await secciones.getByRole("link", { name: "Mi cuenta" }).click();
  await expect(page).toHaveURL("/cuenta");

  const nuevo = `edita.nuevo.${Date.now()}@ejemplo.com`;
  await page.getByLabel("Email").fill(nuevo);
  await page.getByLabel("Celular").fill("011 15 3333-2222");
  await page.getByRole("button", { name: "Guardar cambios" }).click();
  await expect(
    page.getByRole("alert").filter({ hasText: "entrás con tu email nuevo" }),
  ).toBeVisible();

  // El teléfono vuelve normalizado, no como se tipeó.
  await expect(page.getByLabel("Celular")).toHaveValue("+5491133332222");

  await page.getByRole("navigation", { name: "Principal" })
    .getByRole("button", { name: "Cerrar sesión" })
    .click();
  await expect(page).toHaveURL("/");

  await page.goto("/entrar");
  await page.getByLabel("Email").fill(nuevo);
  await page.getByLabel("Contraseña").fill("desarrollo123");
  await page.getByRole("button", { name: "Entrar" }).click();
  await expect(page).toHaveURL("/turnos");
});
