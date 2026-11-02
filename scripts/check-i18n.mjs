import fs from "node:fs";
import assert from "node:assert/strict";
import locales from "../contracts/locales.json" with { type: "json" };

for (const directory of ["backend/internal/i18n/messages"]) {
  const catalogs = locales.supported.map((locale) =>
    JSON.parse(fs.readFileSync(`${directory}/${locale}.json`, "utf8")),
  );
  const keys = Object.keys(catalogs[0]).sort();
  for (const [index, messages] of catalogs.entries()) {
    assert.deepEqual(
      Object.keys(messages).sort(),
      keys,
      `${directory}: language keys differ`,
    );
    for (const [key, value] of Object.entries(messages)) {
      assert.equal(typeof value, "string");
      assert.ok(value.length, `${locales.supported[index]}: empty ${key}`);
      const variables = (text) =>
        [...text.matchAll(/\{(\w+)(?:,|\})/g)].map((m) => m[1]).sort();
      assert.deepEqual(
        variables(value),
        variables(catalogs[0][key]),
        `${key}: interpolation differs`,
      );
    }
  }
}
const messages = JSON.parse(
  fs.readFileSync("backend/internal/i18n/messages/en.json", "utf8"),
);
for (const file of ["web/index.html", "web/login.html"]) {
  const source = fs.readFileSync(file, "utf8");
  for (const match of source.matchAll(/\{\{t ["`]([^"`]+)["`]\}\}/g))
    assert.ok(messages[`ui.${match[1]}`], `${file}: missing ${match[1]}`);
}
console.log(
  "Three-language catalogs, interpolation variables and workspace templates are consistent.",
);
