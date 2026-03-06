import { describe, expect, it } from "vitest";
import type { AssetCatalog, ScenarioContext } from "@bunker/shared";
import { scenario as classicScenario, classicSpecialContractForTests } from "../src/classic";

const makeRng = (seed = 1) => {
  let state = seed >>> 0;
  return () => {
    state = (state * 1664525 + 1013904223) % 0xffffffff;
    return state / 0xffffffff;
  };
};

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
      enablePresenterMode: false,
      continuePermission: "host_only",
      revealTimeoutAction: "random_card",
      revealsBeforeVoting: 2,
      specialUsage: "anytime",
      maxPlayers: 12,
      finalThreatReveal: "host",
      forcedDisasterId: "",
    },
  }) as ScenarioContext;

const revealCategoryCard = (
  session: ReturnType<typeof classicScenario.createSession>,
  playerId: string,
  category: string
) => {
  const view = session.getGameView(playerId);
  const slot = view.you.categories.find((entry) => entry.category === category);
  expect(slot, `slot not found: ${category}`).toBeTruthy();
  const cardId = slot?.cards?.[0]?.instanceId;
  expect(cardId, `card instanceId not found: ${category}`).toBeTruthy();
  const result = session.handleAction(playerId, {
    type: "revealCard",
    payload: { cardId: String(cardId) },
  });
  expect(result.error).toBeUndefined();
  return String(cardId);
};

const continueRoundAsHost = (session: ReturnType<typeof classicScenario.createSession>) => {
  const result = session.handleAction("p1", { type: "continueRound", payload: {} });
  expect(result.error).toBeUndefined();
};

const getCardAssetIdByInstance = (
  session: ReturnType<typeof classicScenario.createSession>,
  playerId: string,
  instanceId: string
) => {
  const view = session.getGameView(playerId);
  const card = view.you.hand.find((entry) => entry.instanceId === instanceId);
  expect(card).toBeTruthy();
  return String(card?.id ?? "");
};

describe("classic specials category card selection", () => {
  it("swapRevealedWithNeighbor uses selected card ids", () => {
    const session = classicScenario.createSession(
      makeContext([
        { playerId: "p1", name: "Host" },
        { playerId: "p2", name: "P2" },
        { playerId: "p3", name: "P3" },
        { playerId: "p4", name: "P4" },
      ])
    );

    const p1Fact1 = revealCategoryCard(session, "p1", "Факт №1");
    continueRoundAsHost(session);
    const p2Fact1 = revealCategoryCard(session, "p2", "Факт №1");

    const p1Fact1Before = getCardAssetIdByInstance(session, "p1", p1Fact1);
    const p2Fact1Before = getCardAssetIdByInstance(session, "p2", p2Fact1);

    const special = classicSpecialContractForTests.find(
      (entry) => entry.effectType === "swapRevealedWithNeighbor" && entry.category === "facts"
    );
    expect(special).toBeTruthy();

    const result = session.handleAction("p1", {
      type: "adminApplySpecial",
      payload: {
        actorPlayerId: "p1",
        specialId: special?.id,
        payload: {
          targetPlayerId: "p2",
          actorCardId: p1Fact1,
          targetCardId: p2Fact1,
        },
      },
    });
    expect(result.error).toBeUndefined();

    const p1Fact1After = getCardAssetIdByInstance(session, "p1", p1Fact1);
    const p2Fact1After = getCardAssetIdByInstance(session, "p2", p2Fact1);
    expect(p1Fact1After).toBe(p2Fact1Before);
    expect(p2Fact1After).toBe(p1Fact1Before);
  });

  it("swapRevealedWithNeighbor rejects hidden selected card", () => {
    const session = classicScenario.createSession(
      makeContext([
        { playerId: "p1", name: "Host" },
        { playerId: "p2", name: "P2" },
        { playerId: "p3", name: "P3" },
        { playerId: "p4", name: "P4" },
      ])
    );

    const p1Fact1 = revealCategoryCard(session, "p1", "Факт №1");
    continueRoundAsHost(session);
    revealCategoryCard(session, "p2", "Факт №1");

    const p2View = session.getGameView("p2");
    const p2Fact2 =
      p2View.you.categories.find((entry) => entry.category === "Факт №2")?.cards?.[0]?.instanceId ?? "";
    expect(p2Fact2).not.toBe("");

    const special = classicSpecialContractForTests.find(
      (entry) => entry.effectType === "swapRevealedWithNeighbor" && entry.category === "facts"
    );
    expect(special).toBeTruthy();

    const result = session.handleAction("p1", {
      type: "adminApplySpecial",
      payload: {
        actorPlayerId: "p1",
        specialId: special?.id,
        payload: {
          targetPlayerId: "p2",
          actorCardId: p1Fact1,
          targetCardId: p2Fact2,
        },
      },
    });
    expect(result.error).toBeTruthy();
  });
});
