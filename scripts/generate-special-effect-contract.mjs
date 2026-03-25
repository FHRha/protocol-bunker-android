#!/usr/bin/env node
import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const here = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(here, "..");
const inputPath = path.resolve(root, "scenarios", "classic", "SPECIAL_CONDITIONS.json");
const outputPath = path.resolve(root, "shared", "special_effect_contract.json");

const args = new Set(process.argv.slice(2));
const checkOnly = args.has("--check");

const normalize = (value) => String(value ?? "").trim().toLowerCase();
const hasAny = (value, parts) => parts.some((part) => part && value.includes(part));

const resolveChoiceKindFromTargeting = (rawTargeting) => {
  const targeting = normalize(rawTargeting);
  if (!targeting) return "none";
  if (hasAny(targeting, ["choose special", "special"])) return "special";
  if (hasAny(targeting, ["bunker"])) return "bunker";
  if (hasAny(targeting, ["neighbor", "left", "right"])) return "neighbor";
  if (hasAny(targeting, ["category"])) return "category";
  if (hasAny(targeting, ["player", "target"])) return "player";
  return "none";
};

const resolveChoiceKind = (item) => {
  let choiceKind = resolveChoiceKindFromTargeting(item.uiTargeting);
  const effectLower = normalize(item?.effect?.type);
  if (choiceKind !== "none") return choiceKind;
  switch (effectLower) {
    case "banvoteagainst":
    case "disablevote":
    case "doublevotesagainst_and_disableselfvote":
    case "replacerevealedcard":
    case "discardrevealedanddealhidden":
    case "stealbaggage_and_givespecial":
      return "player";
    case "swaprevealedwithneighbor":
      return "neighbor";
    case "forcerevealcategoryforall":
      return "category";
    case "devchoosespecial":
      return "special";
    default:
      if (effectLower.includes("bunker")) return "bunker";
      return "none";
  }
};

const resolveTargetScopeFromTargeting = (rawTargeting, choiceKind) => {
  if (choiceKind === "category" || choiceKind === "none" || choiceKind === "special" || choiceKind === "bunker") {
    return "";
  }
  const targeting = normalize(rawTargeting);
  if (hasAny(targeting, ["not self", "not-self"])) return "any_alive";
  if (hasAny(targeting, ["neighbor", "left", "right"])) return "neighbors";
  if (hasAny(targeting, ["including self", "any_including_self"])) return "any_including_self";
  if (hasAny(targeting, ["only self", "self only", "only yourself"])) return "self";
  return "any_alive";
};

const resolveTargetScope = (item, choiceKind) => {
  let targetScope = resolveTargetScopeFromTargeting(item.uiTargeting, choiceKind);
  if (targetScope) return targetScope;
  const effectLower = normalize(item?.effect?.type);
  switch (effectLower) {
    case "swaprevealedwithneighbor":
      return "neighbors";
    case "banvoteagainst":
    case "disablevote":
    case "doublevotesagainst_and_disableselfvote":
    case "replacerevealedcard":
    case "discardrevealedanddealhidden":
    case "stealbaggage_and_givespecial":
      return "any_alive";
    default:
      return "";
  }
};

const raw = JSON.parse(fs.readFileSync(inputPath, "utf8"));
const implemented = raw.filter((item) => Boolean(item?.implemented));
if (implemented.length === 0) {
  throw new Error("No implemented specials found in SPECIAL_CONDITIONS.json");
}

const effectRules = new Map();
const categoryCardSelectionEffects = new Set();

for (const item of implemented) {
  const effectType = String(item?.effect?.type ?? "").trim();
  if (!effectType) {
    throw new Error(`Special "${item?.id ?? item?.title ?? "<unknown>"}" has empty effect.type`);
  }

  const choiceKind = resolveChoiceKind(item);
  const targetScope = resolveTargetScope(item, choiceKind);
  const currentRule = { choiceKind, targetScope };

  const existing = effectRules.get(effectType);
  if (existing) {
    if (existing.choiceKind !== currentRule.choiceKind || existing.targetScope !== currentRule.targetScope) {
      throw new Error(
        `Inconsistent rule for effect "${effectType}": existing=${JSON.stringify(existing)}, current=${JSON.stringify(currentRule)}`
      );
    }
  } else {
    effectRules.set(effectType, currentRule);
  }

  const category = String(item?.effect?.params?.category ?? "").trim();
  if (category && (choiceKind === "player" || choiceKind === "neighbor") && targetScope) {
    categoryCardSelectionEffects.add(effectType);
  }
}

const effectRulesObject = Object.fromEntries(
  [...effectRules.entries()].sort(([a], [b]) => a.localeCompare(b)).map(([effectType, rule]) => [effectType, rule])
);
const output = {
  effectRules: effectRulesObject,
  categoryCardSelectionEffects: [...categoryCardSelectionEffects].sort((a, b) => a.localeCompare(b)),
};
const rendered = `${JSON.stringify(output, null, 2)}\n`;

if (checkOnly) {
  const current = fs.existsSync(outputPath) ? fs.readFileSync(outputPath, "utf8") : "";
  if (current !== rendered) {
    console.error("special_effect_contract.json is out of date. Run:");
    console.error("  node ./scripts/generate-special-effect-contract.mjs");
    process.exit(1);
  }
  console.log("special_effect_contract.json is up to date");
  process.exit(0);
}

fs.mkdirSync(path.dirname(outputPath), { recursive: true });
fs.writeFileSync(outputPath, rendered, "utf8");
console.log(`Generated ${path.relative(root, outputPath)}`);
