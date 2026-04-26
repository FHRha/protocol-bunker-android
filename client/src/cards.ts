import { BACK_DECK_ID_ALIASES, BACK_FILE_BY_DECK_ID } from "../../locales/cards/backs";
import { ASSET_BASE } from "./config";

const imageCache = new Set<string>();
const LOCALE_PATH_RE = /(^|\/)decks\/1x\/(ru|en)(\/)/i;

function rewriteAssetPathLocale(path: string, cardLocale: CardLocale): string {
  return path.replace(LOCALE_PATH_RE, `$1decks/1x/${cardLocale}$3`);
}

export const getCardFaceUrl = (imgUrl?: string, cardLocale?: CardLocale): string | undefined => {
  if (!imgUrl) return undefined;
  if (imgUrl.startsWith("http")) return imgUrl;
  let cleaned = imgUrl.replace(/^\/+/, "");
  if (cleaned.startsWith("assets/")) {
    cleaned = cleaned.slice("assets/".length);
  }
  if (cardLocale) {
    cleaned = rewriteAssetPathLocale(cleaned, cardLocale);
  }
  return `${ASSET_BASE}/${encodeURI(cleaned)}`;
};

type CardLocale = "ru" | "en";

export const getCardBackUrl = (category: string, cardLocale: CardLocale = "ru"): string | undefined => {
  const rawCategory = String(category ?? "").trim();
  const normalizedCategory = rawCategory.toLowerCase();
  const deckId = BACK_DECK_ID_ALIASES[normalizedCategory] ?? normalizedCategory;
  const backFileName = BACK_FILE_BY_DECK_ID[deckId];
  if (!backFileName) return undefined;
  return getCardFaceUrl(`decks/1x/${cardLocale}/Back/${backFileName}`);
};

export const preloadImages = (urls: Array<string | undefined>) => {
  urls.forEach((url) => {
    if (!url || imageCache.has(url)) return;
    const img = new Image();
    img.src = url;
    imageCache.add(url);
  });
};

export const preloadCategoryBacks = (categories: string[], cardLocale: CardLocale = "ru") => {
  preloadImages(categories.map((category) => getCardBackUrl(category, cardLocale)));
};
