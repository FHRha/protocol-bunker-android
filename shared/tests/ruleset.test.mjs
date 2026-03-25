import test from "node:test";
import assert from "node:assert/strict";
import { RULESET_PRESET_COUNTS, RULESET_TABLE, formatLabelShort, getRulesetForPlayerCount } from "../dist/index.js";

test("RULESET_PRESET_COUNTS covers 4 through 16", () => {
  assert.deepEqual(RULESET_PRESET_COUNTS, [4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16]);
});

test("getRulesetForPlayerCount clamps and returns cloned preset data", () => {
  const low = getRulesetForPlayerCount(1);
  const high = getRulesetForPlayerCount(99);
  assert.equal(low.playerCount, 4);
  assert.equal(high.playerCount, 16);

  const preset = RULESET_TABLE[8];
  const rules = getRulesetForPlayerCount(8);
  assert.deepEqual(rules.votesPerRound, preset.votesPerRound);
  rules.votesPerRound[0] = 99;
  assert.notDeepEqual(rules.votesPerRound, preset.votesPerRound);
});

test("formatLabelShort normalizes file-like labels", () => {
  assert.equal(formatLabelShort("bunker.grechka.png"), "Bunker.grechka");
  assert.equal(formatLabelShort("  VERY_LONG_LABEL  "), "Very Long Label");
  assert.equal(formatLabelShort(""), "(без названия)");
});
