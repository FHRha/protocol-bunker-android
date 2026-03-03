import assert from "node:assert/strict";
import test from "node:test";
import { buildLinkSet } from "../dist/urlBuilder.js";

function getControlToken(url) {
  return new URL(url).searchParams.get("token");
}

test("buildLinkSet uses controlCompanionToken with highest priority", () => {
  const links = buildLinkSet({
    lanBase: "http://127.0.0.1:8080",
    roomCode: "ABCD",
    overlayViewToken: "view-token",
    controlCompanionToken: "companion-token",
    editToken: "edit-token",
    controlToken: "control-token",
    overlayControlToken: "legacy-token",
  });

  assert.equal(getControlToken(links.controlPanelUrl.lan), "companion-token");
});

test("buildLinkSet falls back to editToken then controlToken then legacy token", () => {
  const withEdit = buildLinkSet({
    lanBase: "http://127.0.0.1:8080",
    roomCode: "ABCD",
    overlayViewToken: "view-token",
    editToken: "edit-token",
    controlToken: "control-token",
    overlayControlToken: "legacy-token",
  });
  assert.equal(getControlToken(withEdit.controlPanelUrl.lan), "edit-token");

  const withControl = buildLinkSet({
    lanBase: "http://127.0.0.1:8080",
    roomCode: "ABCD",
    overlayViewToken: "view-token",
    controlToken: "control-token",
    overlayControlToken: "legacy-token",
  });
  assert.equal(getControlToken(withControl.controlPanelUrl.lan), "control-token");

  const withLegacy = buildLinkSet({
    lanBase: "http://127.0.0.1:8080",
    roomCode: "ABCD",
    overlayViewToken: "view-token",
    overlayControlToken: "legacy-token",
  });
  assert.equal(getControlToken(withLegacy.controlPanelUrl.lan), "legacy-token");
});

test("buildLinkSet throws when control token is missing", () => {
  assert.throws(
    () =>
      buildLinkSet({
        lanBase: "http://127.0.0.1:8080",
        roomCode: "ABCD",
        overlayViewToken: "view-token",
      }),
    /control\/edit token is required/
  );
});
