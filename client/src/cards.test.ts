import { describe, expect, it } from "vitest";
import {
  getCardBackUrl,
  getCardFaceUrl,
  localizeCardLabel,
  localizeSpecialCondition,
  localizeSpecialOptionLabel,
  resolveAssetIdFromImageUrl,
} from "./cards";

describe("cards localization helpers", () => {
  it("localizes card labels from localized asset paths", () => {
    expect(
      localizeCardLabel(
        {
          deck: "bunker",
          imgUrl: "decks/1x/ru/Bunker/bunker.grechka.png",
        },
        "en"
      )
    ).toBe("Buckwheat");
  });

  it("localizes card labels from legacy asset ids", () => {
    expect(
      localizeCardLabel(
        {
          deck: "bunker",
          id: "decks/Bunker/bunker.uchebnik.png",
        },
        "en"
      )
    ).toBe("Textbook");
  });

  it("builds card back urls using requested locale assets", () => {
    expect(getCardBackUrl("bunker", "en")).toContain("/assets/decks/1x/en/Back/back.rubashka-bunker.png");
    expect(getCardBackUrl("facts2", "ru")).toContain("/assets/decks/1x/ru/Back/back.rubashka-fakty.png");
  });

  it("normalizes card face urls to requested locale path", () => {
    expect(getCardFaceUrl("decks/1x/ru/Bunker/bunker.grechka.png", "en")).toContain(
      "/assets/decks/1x/en/Bunker/bunker.grechka.png"
    );
    expect(getCardFaceUrl("decks/1x/ru/Bunker/bunker.grechka.png", "de")).toContain(
      "/assets/decks/1x/de/Bunker/bunker.grechka.png"
    );
    expect(getCardFaceUrl("decks/1x/ru/Bunker/bunker.grechka.png")).toContain(
      "/assets/decks/1x/ru/Bunker/bunker.grechka.png"
    );
    expect(resolveAssetIdFromImageUrl("/assets/decks/1x/ru/Bunker/bunker.grechka.png")).toBe(
      "decks/1x/ru/Bunker/bunker.grechka.png"
    );
  });

  it("localizes special conditions from new asset paths", () => {
    expect(
      localizeSpecialCondition(
        "classic",
        {
          imgUrl: "decks/1x/ru/Special/special.bud-drugom.png",
          title: "Будь другом",
          text: "",
        },
        "en"
      )
    ).toEqual({
      title: "Be a Friend",
      text: "The chosen player cannot vote against you until the end of the game.",
    });
  });

  it("localizes special picker labels from option id fallback", () => {
    expect(
      localizeSpecialOptionLabel(
        "dev_test",
        {
          id: "bud-drugom",
          title: "Будь другом",
        },
        "en"
      )
    ).toBe("Be a Friend");
  });
});
