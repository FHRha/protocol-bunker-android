import { describe, expect, it } from "vitest";
import { getThreatDeltaFromBunkerCards } from "../src/threat_modifier";

const makeCard = (id: string, title: string, isRevealed = true) => ({ id, title, isRevealed });

describe("Threat modifier from bunker cards", () => {
  it("returns zero delta when no matching bunker cards are revealed", () => {
    const result = getThreatDeltaFromBunkerCards([
      makeCard("decks/bunker/bunker.masterskaya.jpg", "Мастерская"),
      makeCard("decks/bunker/bunker.aptechki.jpg", "Аптечки"),
    ]);

    expect(result.delta).toBe(0);
    expect(result.reasons).toEqual([]);
  });

  it("adds +1 for 'ВМЕСТЕ НА 10 ЛЕТ'", () => {
    const result = getThreatDeltaFromBunkerCards([
      makeCard("decks/bunker/bunker.vmeste-na-10-let.jpg", "ВМЕСТЕ НА 10 ЛЕТ"),
    ]);

    expect(result.delta).toBe(1);
    expect(result.reasons).toEqual(["ВМЕСТЕ НА 10 ЛЕТ"]);
  });

  it("adds -1 for 'ЗАГАДОЧНЫЙ ЖУРНАЛ'", () => {
    const result = getThreatDeltaFromBunkerCards([
      makeCard("decks/bunker/bunker.zagadochnyy-zhurnal.jpg", "ЗАГАДОЧНЫЙ ЖУРНАЛ"),
    ]);

    expect(result.delta).toBe(-1);
    expect(result.reasons).toEqual(["ЗАГАДОЧНЫЙ ЖУРНАЛ"]);
  });

  it("sums both modifiers when both cards are present", () => {
    const result = getThreatDeltaFromBunkerCards([
      makeCard("decks/bunker/bunker.vmeste-na-10-let.jpg", "ВМЕСТЕ НА 10 ЛЕТ"),
      makeCard("decks/bunker/bunker.zagadochnyy-zhurnal.jpg", "ЗАГАДОЧНЫЙ ЖУРНАЛ"),
    ]);

    expect(result.delta).toBe(0);
    expect(result.reasons).toEqual(["ВМЕСТЕ НА 10 ЛЕТ", "ЗАГАДОЧНЫЙ ЖУРНАЛ"]);
  });

  it("ignores hidden cards", () => {
    const result = getThreatDeltaFromBunkerCards([
      makeCard("decks/bunker/bunker.vmeste-na-10-let.jpg", "ВМЕСТЕ НА 10 ЛЕТ", false),
      makeCard("decks/bunker/bunker.zagadochnyy-zhurnal.jpg", "ЗАГАДОЧНЫЙ ЖУРНАЛ", false),
    ]);

    expect(result.delta).toBe(0);
    expect(result.reasons).toEqual([]);
  });
});
