import { expect, test } from "@playwright/test";

/**
 * El circuito completo, contra la API de Go real y sin mocks: el equivalente de
 * front a la convención que el back ya tiene escrita.
 *
 * Es exactamente el tipo de test que habría encontrado que CORS no existía.
 */
test("un visitante sin cuenta busca, reserva y ve su turno", async ({ page }) => {
  // Email único por corrida: el seed vive mientras viva el proceso de la API, y
  // dos corridas con el mismo email chocarían con un 409 de email en uso.
  const email = `e2e-${Date.now()}@ejemplo.com`;

  await test.step("la búsqueda muestra profesionales con sus horarios", async () => {
    await page.goto("/buscar");
    await expect(page.getByRole("heading", { name: /encontrá tu próximo turno/i })).toBeVisible();
    await expect(page.getByText("González").first()).toBeVisible();
  });

  await test.step("el perfil se abre y ofrece horarios", async () => {
    await page.getByRole("link", { name: /González/ }).first().click();
    await expect(page).toHaveURL(/\/perfiles\//);
    await expect(page.getByRole("heading", { name: /elegí un horario/i })).toBeVisible();
  });

  const horario = page.getByRole("link", { name: /^\d{2}:\d{2}$/ }).first();
  const textoHorario = (await horario.textContent())!.trim();

  await test.step("elegir un horario lleva a reservar con ese horario puesto", async () => {
    await horario.click();
    await expect(page).toHaveURL(/\/reservar\?inicio=/);
    await expect(page.getByRole("button", { name: `Reservar ${textoHorario}` })).toBeVisible();
  });

  await test.step("se crea la cuenta y el turno en el mismo paso", async () => {
    await page.getByLabel("Email").fill(email);
    await page.getByLabel("Contraseña").fill("unaclave8");
    await page.getByLabel("Nombre", { exact: true }).fill("Ana");
    await page.getByLabel("Apellido").fill("Prueba");
    await page.getByLabel("Celular").fill("11 6000-1234");
    await page.getByRole("button", { name: /^Reservar \d{2}:\d{2}$/ }).click();

    await expect(page.getByText(/turno reservado/i)).toBeVisible();
  });

  await test.step("el turno aparece en mis turnos", async () => {
    await page.getByRole("link", { name: /ver mis turnos/i }).click();
    await expect(page).toHaveURL(/\/turnos/);
    await expect(page.getByText(textoHorario).first()).toBeVisible();
  });
});

/**
 * El caso más probable de los dos que diseñamos: alguien que ya se registró y
 * no se acuerda. Tiene que ver un mensaje que lo diga y un camino para salir,
 * no un error genérico.
 */
test("reservar con un email ya registrado ofrece entrar", async ({ page }) => {
  await page.goto("/");
  await page.getByRole("link", { name: /González/ }).first().click();
  await page.getByRole("link", { name: /^\d{2}:\d{2}$/ }).first().click();

  // Uno de los cuatro del seed.
  await page.getByLabel("Email").fill("martin.gonzalez@ejemplo.com");
  await page.getByLabel("Contraseña").fill("unaclave8");
  await page.getByLabel("Nombre", { exact: true }).fill("Otro");
  await page.getByLabel("Apellido").fill("Nombre");
  await page.getByLabel("Celular").fill("11 6000-4321");
  await page.getByRole("button", { name: /^Reservar \d{2}:\d{2}$/ }).click();

  // Next tiene su propio anunciador de rutas con role="alert", así que hay que
  // desambiguar por texto. Se sigue afirmando el rol: el mensaje tiene que ser
  // anunciable por un lector de pantalla, no solo visible.
  const aviso = page.getByRole("alert").filter({ hasText: /ya tenés una cuenta/i });
  await expect(aviso).toBeVisible();
  // Acotado al aviso: desde que el encabezado tiene su propio "Entrar", buscarlo
  // en toda la página encuentra dos.
  await expect(aviso.getByRole("link", { name: "Entrar" })).toBeVisible();
});

/**
 * El criterio que hace que haber elegido Next signifique algo. Se verifica
 * pidiendo el HTML crudo, sin ejecutar JavaScript: si el nombre no está ahí, la
 * página se está renderizando en el cliente y el SEO no existe.
 */
test("las páginas públicas llegan completas desde el servidor", async ({ request }) => {
  // La landing: el mensaje Y contenido real. Es la página con más autoridad de
  // SEO, así que gastarla en marketing sin nada indexable sería tirarla.
  const inicio = await (await request.get("/")).text();
  expect(inicio).toContain("El horario que ves es el que reservás");
  expect(inicio).toContain("González");
  expect(inicio).toContain("Cómo funciona");
  // Y la disponibilidad de la portada son horarios de verdad, no una imagen.
  expect(inicio).toMatch(/>\d{2}:\d{2}</);

  const buscar = await (await request.get("/buscar")).text();
  expect(buscar).toContain("González");

  const perfil = await (await request.get("/perfiles/martin-gonzalez")).text();
  expect(perfil).toContain("Martín González");
  expect(perfil).toContain("MN 98234");
  // El title lleva nombre, especialidad y zona: es lo que se ve en Google.
  expect(perfil).toMatch(/<title>Martín González — Psicología en CABA/);
  // Y los horarios, que son el contenido que hace útil la página.
  expect(perfil).toMatch(/>\d{2}:\d{2}</);
});

/**
 * La tira de días son links y no estado de React, y esto es lo que lo prueba.
 *
 * Con el JavaScript apagado el perfil tiene que seguir dejando cambiar de día:
 * es la razón por la que la pantalla que más importa para la búsqueda orgánica
 * se renderiza entera en el servidor.
 */
test.describe("sin JavaScript", () => {
  test.use({ javaScriptEnabled: false });

  test("la agenda del perfil deja cambiar de día", async ({ page }) => {
    await page.goto("/perfiles/martin-gonzalez");

    const agenda = page.getByRole("list").filter({ has: page.getByRole("link", { name: /^\d{2}:\d{2}$/ }) });
    const primerDia = await page.getByRole("heading", { level: 3 }).first().textContent();

    // El segundo día con horarios de la tira: el primero ya está abierto.
    const dias = page.getByRole("link", { name: /(Lun|Mar|Mié|Jue|Vie|Sáb|Dom)\s*\d+/ });
    await dias.nth(1).click();

    await expect(page.getByRole("heading", { level: 3 }).first()).not.toHaveText(primerDia ?? "");
    await expect(agenda.getByRole("link").first()).toBeVisible();
  });
});
