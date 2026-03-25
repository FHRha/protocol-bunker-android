import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import type { AssetCatalog, ScenarioContext } from "@bunker/shared";
import { scenario as classicScenario } from "../src/classic";

interface EffectRule {
  choiceKind: string;
  targetScope: string;
}

interface EffectContract {
  effectRules: Record<string, EffectRule>;
  categoryCardSelectionEffects: string[];
}

const here = path.dirname(fileURLToPath(import.meta.url));

const makeDeck = (name: string, count: number) =>
  Array.from({ length: count }, (_, idx) => ({
    id: `decks/${name}/${name}_${idx + 1}.jpg`,
    deck: name,
    labelShort: `${name} ${idx + 1}`,
  }));

const makeAssets = (): AssetCatalog => ({
  decks: {
    Профессия: makeDeck("Профессия", 20),
    Здоровье: makeDeck("Здоровье", 20),
    Хобби: makeDeck("Хобби", 20),
    Багаж: makeDeck("Багаж", 20),
    Факты: makeDeck("Факты", 40),
    Биология: makeDeck("Биология", 20),
  },
});

const makeRng = (seed = 1) => {
  let state = seed >>> 0;
  return () => {
    state = (state * 1664525 + 1013904223) % 0xffffffff;
    return state / 0xffffffff;
  };
};

const makeContext = (players: Array<{ playerId: string; name: string }>): ScenarioContext =>
  ({
    roomCode: "TEST",
    createdAt: Date.now(),
    rng: makeRng(123),
    assets: makeAssets(),
    players,
    hostId: players[0]?.playerId ?? "p1",
    settings: {
      enableRevealDiscussionTimer: false,
      revealDiscussionSeconds: 30,
      enablePreVoteDiscussionTimer: false,
      preVoteDiscussionSeconds: 30,
      enablePostVoteDiscussionTimer: false,
      postVoteDiscussionSeconds: 30,
      automationMode: "semi",
      continuePermission: "host_only",
      revealTimeoutAction: "random_card",
      revealsBeforeVoting: 2,
      specialUsage: "anytime",
      maxPlayers: 12,
      finalThreatReveal: "host",
      forcedDisasterId: "",
    },
  }) as ScenarioContext;

const specialDefinitionsPath = path.resolve(here, "../classic/SPECIAL_CONDITIONS.json");
const specialDefinitions = JSON.parse(fs.readFileSync(specialDefinitionsPath, "utf8")) as Array<{
  effect?: { type?: string; params?: { category?: string } };
}>;
const contractPath = path.resolve(here, "../../shared/special_effect_contract.json");
const contract = JSON.parse(fs.readFileSync(contractPath, "utf8")) as EffectContract;

describe("special effect contract parity (TS classic engine)", () => {
  it("matches effect -> choiceKind/targetScope contract", () => {
    const session = classicScenario.createSession(
      makeContext([
        { playerId: "p1", name: "Host" },
        { playerId: "p2", name: "P2" },
      ])
    );
    const catalog = session.getSpecialCatalog();
    expect(catalog.length).toBeGreaterThan(0);

    const requireCategory = new Set(contract.categoryCardSelectionEffects);
    const usedEffects = new Set<string>();

    for (const entry of catalog) {
      const rule = contract.effectRules[entry.effectType];
      expect(rule, `missing rule for effect=${entry.effectType} special=${entry.id}`).toBeTruthy();
      if (!rule) continue;
      usedEffects.add(entry.effectType);

      expect(entry.choiceKind, `choiceKind mismatch for ${entry.effectType}/${entry.id}`).toBe(rule.choiceKind);
      expect(entry.targetScope ?? "", `targetScope mismatch for ${entry.effectType}/${entry.id}`).toBe(rule.targetScope);

      if (requireCategory.has(entry.effectType)) {
        const definition = specialDefinitions.find((item) => item.effect?.type === entry.effectType);
        expect(definition?.effect?.params?.category, `missing params.category for ${entry.effectType}/${entry.id}`).toBeTruthy();
      }
    }

    const unusedRules = Object.keys(contract.effectRules).filter((effectType) => !usedEffects.has(effectType));
    expect(unusedRules).toEqual([]);
  });
});

