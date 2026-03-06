import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { classicSpecialContractForTests } from "../src/classic";

interface EffectRule {
  choiceKind: string;
  targetScope: string;
}

interface EffectContract {
  effectRules: Record<string, EffectRule>;
  categoryCardSelectionEffects: string[];
}

const here = path.dirname(fileURLToPath(import.meta.url));
const contractPath = path.resolve(here, "../../shared-contract/special_effect_contract.json");
const contract = JSON.parse(fs.readFileSync(contractPath, "utf8")) as EffectContract;

describe("special effect contract parity (TS classic engine)", () => {
  it("matches effect -> choiceKind/targetScope contract", () => {
    expect(classicSpecialContractForTests.length).toBeGreaterThan(0);
    const requireCategory = new Set(contract.categoryCardSelectionEffects);
    const usedEffects = new Set<string>();

    for (const entry of classicSpecialContractForTests) {
      const rule = contract.effectRules[entry.effectType];
      expect(rule, `missing rule for effect=${entry.effectType} special=${entry.id}`).toBeTruthy();
      if (!rule) continue;
      usedEffects.add(entry.effectType);

      expect(entry.choiceKind, `choiceKind mismatch for ${entry.effectType}/${entry.id}`).toBe(rule.choiceKind);
      expect(entry.targetScope, `targetScope mismatch for ${entry.effectType}/${entry.id}`).toBe(rule.targetScope);

      if (requireCategory.has(entry.effectType)) {
        expect(entry.category.trim().length, `missing params.category for ${entry.effectType}/${entry.id}`).toBeGreaterThan(0);
      }
    }

    const unusedRules = Object.keys(contract.effectRules).filter((effectType) => !usedEffects.has(effectType));
    expect(unusedRules).toEqual([]);
  });
});

