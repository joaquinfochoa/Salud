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
    await page.goto("/");
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
  await page.getByRole("button", { name: /^Reservar \d{2}:\d{2}$/ }).click();

  // Next tiene su propio anunciador de rutas con role="alert", así que hay que
  // desambiguar por texto. Se sigue afirmando el rol: el mensaje tiene que ser
  // anunciable por un lector de pantalla, no solo visible.
  await expect(
    page.getByRole("alert").filter({ hasText: /ya tenés una cuenta/i }),
  ).toBeVisible();
  await expect(page.getByRole("link", { name: "Entrar" })).toBeVisible();
});

/**
 * El criterio que hace que haber elegido Next signifique algo. Se verifica
 * pidiendo el HTML crudo, sin ejecutar JavaScript: si el nombre no está ahí, la
 * página se está renderizando en el cliente y el SEO no existe.
 */
test("las páginas públicas llegan completas desde el servidor", async ({ request }) => {
  const inicio = await (await request.get("/")).text();
  expect(inicio).toContain("González");

  const perfil = await (await request.get("/perfiles/martin-gonzalez")).text();
  expect(perfil).toContain("Martín González");
  expect(perfil).toContain("MN 98234");
  // El title lleva nombre, especialidad y zona: es lo que se ve en Google.
  expect(perfil).toMatch(/<title>Martín González — Psicología en CABA/);
  // Y los horarios, que son el contenido que hace útil la página.
  expect(perfil).toMatch(/>\d{2}:\d{2}</);
});
