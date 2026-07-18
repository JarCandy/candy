import { readFile } from "node:fs/promises";
import oniguruma from "vscode-oniguruma";
import textmate from "vscode-textmate";

const { createOnigScanner, createOnigString, loadWASM } = oniguruma;
const { Registry, parseRawGrammar } = textmate;

const files = [
  "./package.json",
  "./language-configuration.json",
  "./syntaxes/candy.tmLanguage.json"
];

for (const file of files) {
  JSON.parse(await readFile(new URL(`../${file}`, import.meta.url), "utf8"));
}

const wasm = await readFile(
  new URL("../node_modules/vscode-oniguruma/release/onig.wasm", import.meta.url)
);
await loadWASM(wasm.buffer);

const grammarUrl = new URL("../syntaxes/candy.tmLanguage.json", import.meta.url);
const registry = new Registry({
  onigLib: Promise.resolve({
    createOnigScanner,
    createOnigString
  }),
  loadGrammar: async (scopeName) => {
    if (scopeName !== "source.candy") {
      return null;
    }
    return parseRawGrammar(await readFile(grammarUrl, "utf8"), grammarUrl.pathname);
  }
});

const grammar = await registry.loadGrammar("source.candy");
if (!grammar) {
  throw new Error("Candy TextMate grammar could not be loaded.");
}

const cases = [
  ["let lang = custom(\"go\")", "let", "keyword.declaration.variable.candy"],
  ["let lang = custom(\"go\")", "lang", "variable.other.readwrite.candy"],
  ["pub name: *[]*string = \"none\"", "[]", "storage.modifier.slice.candy"],
  ["pub name: *[]*string = \"none\"", "string", "storage.type.builtin.candy"],
  ["go::lib(\"uuid\").NewString()", "go", "entity.name.namespace.candy"],
  ["go::lib(\"uuid\").NewString()", "lib", "entity.name.function.candy"],
  ["go::lib(\"uuid\").NewString()", "NewString", "entity.name.function.candy"],
  ["#[db::sqlite::table(\"User\")]", "table", "entity.name.function.candy"],
  ["use (\"db\" as db,)", ",", "invalid.illegal.comma.candy"],
  ["custom(\"go\", true)", ",", "invalid.illegal.comma.candy"],
  ["let name = \"Candy\";", ";", "invalid.illegal.semicolon.candy"],
  ["// comment", "comment", "comment.line.double-slash.candy"]
];

for (const [line, fragment, expectedScope] of cases) {
  const start = line.indexOf(fragment);
  const token = grammar.tokenizeLine(line).tokens.find(
    (item) => item.startIndex <= start && item.endIndex >= start + fragment.length
  );
  if (!token?.scopes.includes(expectedScope)) {
    throw new Error(
      `Expected ${JSON.stringify(fragment)} to have ${expectedScope}; got ${token?.scopes.join(", ")}`
    );
  }
}

console.log("Candy TextMate grammar loaded and highlighted the syntax fixture.");
