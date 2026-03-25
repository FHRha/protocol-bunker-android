import { localizeCardLabel } from "./cards";

export function localizeOptionalCardLabel(
  card: { id?: string; deck?: string; labelShort?: string; imgUrl?: string } | null | undefined,
  cardLocale: "ru" | "en",
  fallback = "-"
): string {
  if (!card) return fallback;
  const localized = localizeCardLabel(card, cardLocale).trim();
  return localized || fallback;
}

export function getLocalizedWorldCardLabel(
  card: { imgUrl?: string; imageId?: string; title?: string; description?: string; kind?: string },
  cardLocale: "ru" | "en",
  unnamedCard = ""
): string {
  const imgUrl = card.imgUrl ?? card.imageId;
  if (imgUrl) {
    let deck: string | undefined;
    const cleaned = imgUrl.replace(/^\/assets\//, "");
    const segments = cleaned.split("/").filter(Boolean);
    if (segments.length >= 4 && segments[0] === "decks") {
      deck = segments[3];
    }
    const localized = localizeCardLabel({ imgUrl, deck }, cardLocale).trim();
    if (localized && localized !== "-" && localized !== imgUrl && !localized.startsWith("decks/")) {
      return localized;
    }
  }
  return card.title || card.description || unnamedCard;
}
