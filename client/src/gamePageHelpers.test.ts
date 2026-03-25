import { describe, expect, it } from "vitest";
import { getLocalizedWorldCardLabel, localizeOptionalCardLabel } from "./gamePageHelpers";

describe("gamePageHelpers", () => {
  it("localizes dossier card labels through cards dictionary", () => {
    expect(
      localizeOptionalCardLabel(
        {
          deck: "bunker",
          imgUrl: "decks/1x/ru/Bunker/bunker.grechka.png",
        },
        "en"
      )
    ).toBe("Buckwheat");
  });

  it("returns fallback for empty dossier card", () => {
    expect(localizeOptionalCardLabel(null, "ru", "No card")).toBe("No card");
  });

  it("localizes world card labels from imgUrl before raw title fallback", () => {
    expect(
      getLocalizedWorldCardLabel(
        {
          imgUrl: "/assets/decks/1x/ru/Bunker/bunker.uchebnik.png",
          title: "bunker.uchebnik",
        },
        "en",
        "Unnamed"
      )
    ).toBe("Textbook");
  });

  it("falls back to title/description when localization key is unavailable", () => {
    expect(
      getLocalizedWorldCardLabel(
        {
          imgUrl: "/assets/custom/path/unknown.png",
          title: "Server title",
        },
        "en",
        "Unnamed"
      )
    ).toBe("Server title");
  });
});
