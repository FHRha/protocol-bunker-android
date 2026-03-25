import test from "node:test";
import assert from "node:assert/strict";
import { computeNeighbors, computeTargetScope, getTargetCandidates, TARGETING_TERMS } from "../dist/index.js";

test("TARGETING_TERMS merges both locale dictionaries", () => {
  assert.ok(TARGETING_TERMS.anyPlayer.includes("any player"));
  assert.ok(TARGETING_TERMS.choose.includes("choose"));
  assert.ok(TARGETING_TERMS.neighbors.some((term) => term.includes("left")));
  assert.ok(TARGETING_TERMS.noTarget.includes("passive"));
});

test("computeTargetScope resolves english and russian targeting semantics", () => {
  assert.equal(computeTargetScope("not self", ""), "any_alive");
  assert.equal(computeTargetScope("", "choose the neighbor on the left"), "neighbors");
  assert.equal(computeTargetScope("", "можно себя"), "any_including_self");
  assert.equal(computeTargetScope("только себя", ""), "self");
  assert.equal(computeTargetScope("passive", ""), null);
});

test("computeNeighbors skips dead players in both directions", () => {
  const ring = ["p1", "p2", "p3", "p4", "p5"];
  const alive = new Set(["p1", "p3", "p5"]);
  assert.deepEqual(computeNeighbors(ring, alive, "p3"), { leftId: "p1", rightId: "p5" });
});

test("getTargetCandidates respects scope semantics", () => {
  const ring = ["p1", "p2", "p3", "p4"];
  const alive = new Set(["p1", "p2", "p4"]);

  assert.deepEqual(getTargetCandidates("self", "p1", ring, alive), ["p1"]);
  assert.deepEqual(getTargetCandidates("neighbors", "p1", ring, alive), ["p4", "p2"]);
  assert.deepEqual(getTargetCandidates("any_including_self", "p1", ring, alive), ["p1", "p2", "p4"]);
  assert.deepEqual(getTargetCandidates("any_alive", "p1", ring, alive), ["p2", "p4"]);
});
