import { BACK_DECK_ID_ALIASES, BACK_FILE_BY_DECK_ID } from "../../locales/cards/backs";
import enCardsLocale from "../../locales/cards/en.json";
import ruCardsLocale from "../../locales/cards/ru.json";
import classicEnSpecialLocale from "../../locales/special_conditions/classic/en.json";
import classicRuSpecialLocale from "../../locales/special_conditions/classic/ru.json";
import devTestEnSpecialLocale from "../../locales/special_conditions/dev_test/en.json";
import devTestRuSpecialLocale from "../../locales/special_conditions/dev_test/ru.json";
import { ASSET_BASE } from "./config";

const imageCache = new Set<string>();
const CARD_ASSET_VARIANT = "1x";
const CARD_BACK_DECK_DIR = "Back";

type CardLocale = "ru" | "en";
type CardLocaleDictionary = {
  decks?: Record<string, string>;
  cards?: Record<string, string>;
};

const CARD_LOCALE_DICTIONARIES: Record<CardLocale, CardLocaleDictionary> = {
  ru: ruCardsLocale as CardLocaleDictionary,
  en: enCardsLocale as CardLocaleDictionary,
};

const CARD_ASSET_LOCALES: ReadonlySet<CardLocale> = new Set<CardLocale>(["ru"]);



type ScenarioId = "classic" | "dev_test";
type SpecialConditionLocaleEntry = {
  title?: string;
  text?: string;
};
type SpecialConditionLocaleDictionary = Record<string, SpecialConditionLocaleEntry>;

const SPECIAL_CONDITION_DICTIONARIES: Record<ScenarioId, Record<CardLocale, SpecialConditionLocaleDictionary>> = {
  classic: {
    ru: classicRuSpecialLocale as SpecialConditionLocaleDictionary,
    en: classicEnSpecialLocale as SpecialConditionLocaleDictionary,
  },
  dev_test: {
    ru: devTestRuSpecialLocale as SpecialConditionLocaleDictionary,
    en: devTestEnSpecialLocale as SpecialConditionLocaleDictionary,
  },
};

const normalizeScenarioId = (value?: string): ScenarioId => (value === "classic" ? "classic" : "dev_test");

const normalizeLookupValue = (value?: string): string =>
  String(value ?? "")
    .trim()
    .toLowerCase()
    .replace(/\.[a-z0-9]{2,5}$/i, "")
    .replace(/[_\s]+/g, "-")
    .replace(/[^a-z0-9а-яё-]+/gi, "")
    .replace(/-+/g, "-")
    .replace(/^-|-$/g, "");

const resolveSpecialLocaleKeyFromAssetPath = (assetPath?: string): string | undefined => {
  const cleaned = sanitizeAssetPath(assetPath);
  if (!cleaned) return undefined;
  const segments = cleaned.split("/").filter(Boolean);
  
  // Новый формат: decks/1x/ru/Special/special.bud-drugom.png
  if (segments.length >= 5 && segments[0] === "decks" && (segments[1] === "1x" || segments[1] === "2x")) {
    const deckDir = segments[3] ?? "";
    const fileName = segments[4] ?? "";
    if (deckDir.toLowerCase() === "special" && fileName) {
      return `Special/${fileName}`;
    }
  }
  
  // Старый формат: decks/Особые условия/БУДЬ ДРУГОМ.jpg
  if (segments.length >= 3 && segments[0] === "decks") {
    const deckDir = segments[1] ?? "";
    const fileName = segments[2] ?? "";
    if (deckDir.toLowerCase().includes("особ") || deckDir.toLowerCase() === "special") {
      const normalizedFileName = normalizeLookupValue(fileName);
      if (normalizedFileName) {
        return `Special/special.${normalizedFileName}.png`;
      }
    }
  }
  
  return undefined;
};

const buildSpecialConditionIndexes = () => {
  const indexes: Record<ScenarioId, Map<string, string>> = {
    classic: new Map<string, string>(),
    dev_test: new Map<string, string>(),
  };
  (["classic", "dev_test"] as const).forEach((scenarioId) => {
    (["ru", "en"] as const).forEach((locale) => {
      const dictionary = SPECIAL_CONDITION_DICTIONARIES[scenarioId][locale];
      Object.entries(dictionary).forEach(([key, entry]) => {
        const fileName = key.split("/").pop() ?? "";
        const stem = fileName.replace(/\.[a-z0-9]{2,5}$/i, "");
        const variants = [key, fileName, stem, entry.title];
        variants.forEach((variant) => {
          const normalized = normalizeLookupValue(variant);
          if (normalized && !indexes[scenarioId].has(normalized)) {
            indexes[scenarioId].set(normalized, key);
          }
        });
      });
    });
  });
  return indexes;
};

const SPECIAL_CONDITION_INDEXES = buildSpecialConditionIndexes();

const getSpecialConditionEntry = (
  scenarioId: string | undefined,
  requestedLocale: CardLocale | undefined,
  source: { imgUrl?: string; id?: string; labelShort?: string; title?: string }
): SpecialConditionLocaleEntry | undefined => {
  const normalizedScenario = normalizeScenarioId(scenarioId);
  const locale = normalizeCardLocale(requestedLocale);
  const dictionary = SPECIAL_CONDITION_DICTIONARIES[normalizedScenario][locale];
  const fallbackDictionary = SPECIAL_CONDITION_DICTIONARIES[normalizedScenario].ru;

  const directKey = resolveSpecialLocaleKeyFromAssetPath(source.imgUrl ?? source.id);
  if (directKey) {
    return dictionary[directKey] ?? fallbackDictionary[directKey];
  }

  const index = SPECIAL_CONDITION_INDEXES[normalizedScenario];
  const candidates = [source.labelShort, source.title, source.id];
  if (source.id?.toLowerCase().startsWith("dev-choice-")) {
    candidates.push("dev-choice");
  }
  for (const candidate of candidates) {
    const key = index.get(normalizeLookupValue(candidate));
    if (!key) continue;
    const entry = dictionary[key] ?? fallbackDictionary[key];
    if (entry) return entry;
  }

  return undefined;
};

const DECK_FOLDER_ALIASES: Record<string, string> = {
  ...BACK_DECK_ID_ALIASES,
  back: "back",
  Back: "back",
  Profession: "profession",
  Health: "health",
  Hobby: "hobby",
  Baggage: "baggage",
  Fact: "fact",
  Biology: "biology",
  Special: "special",
  Bunker: "bunker",
  Disaster: "disaster",
  Threat: "threat",
};

const normalizeCardLocale = (value?: string): CardLocale => (value === "en" ? "en" : "ru");

const resolveAssetLocale = (requestedLocale?: string): CardLocale => {
  const normalized = normalizeCardLocale(requestedLocale);
  return CARD_ASSET_LOCALES.has(normalized) ? normalized : "ru";
};

const stripQueryAndHash = (value: string): string => value.replace(/[?#].*$/, "");

const sanitizeAssetPath = (value?: string): string | undefined => {
  if (!value) return undefined;
  let cleaned = String(value).trim();
  if (!cleaned) return undefined;
  if (cleaned.startsWith("http://") || cleaned.startsWith("https://")) {
    try {
      cleaned = new URL(cleaned).pathname;
    } catch {
      // keep raw value if URL parsing fails
    }
  }
  cleaned = cleaned.replace(/^\/+/, "");
  if (cleaned.startsWith("assets/")) {
    cleaned = cleaned.slice("assets/".length);
  }
  cleaned = stripQueryAndHash(cleaned);
  return cleaned || undefined;
};

const buildAssetUrl = (relativeAssetPath: string): string => `${ASSET_BASE}/${encodeURI(relativeAssetPath)}`;

const withLocalizedAssetPath = (relativeAssetPath: string, requestedLocale?: string): string => {
  const cleaned = sanitizeAssetPath(relativeAssetPath) ?? relativeAssetPath;
  const segments = cleaned.split("/").filter(Boolean);
  if (segments.length >= 4 && segments[0] === "decks" && (segments[1] === "1x" || segments[1] === "2x")) {
    const localizedSegments = [...segments];
    localizedSegments[2] = resolveAssetLocale(requestedLocale);
    return localizedSegments.join("/");
  }
  return cleaned;
};

const resolveDeckIdFromAssetPath = (assetPath?: string): string | undefined => {
  if (!assetPath) return undefined;
  const cleaned = sanitizeAssetPath(assetPath);
  if (!cleaned) return undefined;
  const segments = cleaned.split("/").filter(Boolean);
  let deckSegment = "";
  if (segments.length >= 4 && segments[0] === "decks" && (segments[1] === "1x" || segments[1] === "2x")) {
    deckSegment = segments[3] ?? "";
  } else if (segments.length >= 2 && segments[0] === "decks") {
    deckSegment = segments[1] ?? "";
  }
  if (!deckSegment) return undefined;
  return DECK_FOLDER_ALIASES[deckSegment] ?? DECK_FOLDER_ALIASES[deckSegment.toLowerCase()] ?? deckSegment.toLowerCase();
};

const resolveCardIdFromAssetPath = (assetPath?: string, fallbackDeck?: string): string | undefined => {
  if (!assetPath) return undefined;
  const cleaned = sanitizeAssetPath(assetPath);
  if (!cleaned) return undefined;
  const fileName = cleaned.split("/").filter(Boolean).pop();
  if (!fileName) return undefined;
  const stem = fileName.replace(/\.[a-z0-9]{2,5}$/i, "");
  const deckId = fallbackDeck ?? resolveDeckIdFromAssetPath(cleaned);
  if (deckId && stem.toLowerCase().startsWith(`${deckId}.`)) {
    return stem.slice(deckId.length + 1);
  }
  return stem;
};

const normalizeDeckId = (value?: string): string | undefined => {
  if (!value) return undefined;
  const normalized = String(value).trim();
  if (!normalized) return undefined;
  return BACK_DECK_ID_ALIASES[normalized.toLowerCase()] ?? normalized.toLowerCase();
};

const getCardDictionary = (requestedLocale?: string): CardLocaleDictionary => {
  const locale = normalizeCardLocale(requestedLocale);
  return CARD_LOCALE_DICTIONARIES[locale];
};

type CardLabelSource = {
  id?: string;
  deck?: string;
  labelShort?: string;
  imgUrl?: string;
};

export const resolveAssetIdFromImageUrl = (imgUrl?: string): string | undefined =>
  sanitizeAssetPath(imgUrl);

export const getCardFaceUrl = (imgUrl?: string, requestedLocale?: CardLocale): string | undefined => {
  if (!imgUrl) return undefined;
  if (imgUrl.startsWith("http")) return imgUrl;
  const cleaned = sanitizeAssetPath(imgUrl);
  if (!cleaned) return undefined;
  return buildAssetUrl(withLocalizedAssetPath(cleaned, requestedLocale));
};

export const getCardBackUrl = (category: string, cardLocale: CardLocale = "ru"): string | undefined => {
  const rawCategory = String(category ?? "").trim();
  const normalizedCategory = rawCategory.toLowerCase();
  const deckId = BACK_DECK_ID_ALIASES[normalizedCategory] ?? normalizedCategory;
  const backFileName = BACK_FILE_BY_DECK_ID[deckId];
  if (!backFileName) return undefined;
  const locale = resolveAssetLocale(cardLocale);
  return getCardFaceUrl(`decks/${CARD_ASSET_VARIANT}/${locale}/${CARD_BACK_DECK_DIR}/${backFileName}`, locale);
};

export const localizeCardLabel = (card: CardLabelSource, requestedLocale?: CardLocale): string => {
  const locale = normalizeCardLocale(requestedLocale);
  const localeDictionary = getCardDictionary(locale);
  const fallbackDictionary = getCardDictionary("ru");

  const assetPath = sanitizeAssetPath(card.imgUrl ?? card.id);
  const deckId = normalizeDeckId(card.deck) ?? resolveDeckIdFromAssetPath(assetPath);
  const cardId = resolveCardIdFromAssetPath(assetPath, deckId);
  
  if (deckId && cardId) {
    const key = `${deckId}.${cardId}`;
    const localized = localeDictionary.cards?.[key] ?? fallbackDictionary.cards?.[key];
    if (localized) return localized;
  }

  const rawLabel = String(card.labelShort ?? "").trim();
  if (rawLabel) {
    const direct = localeDictionary.cards?.[rawLabel] ?? fallbackDictionary.cards?.[rawLabel];
    if (direct) return direct;
    if (deckId) {
      const combinedKey = `${deckId}.${rawLabel}`;
      const combined = localeDictionary.cards?.[combinedKey] ?? fallbackDictionary.cards?.[combinedKey];
      if (combined) return combined;
    }
  }

  return rawLabel || (deckId
    ? localeDictionary.decks?.[deckId] ?? fallbackDictionary.decks?.[deckId] ?? deckId
    : "-");
};

export const localizeSpecialCondition = (
  scenarioId: string | undefined,
  special: { imgUrl?: string; id?: string; labelShort?: string; title?: string; text?: string },
  requestedLocale?: CardLocale
): { title: string; text: string } => {
  const entry = getSpecialConditionEntry(scenarioId, requestedLocale, special);
  return {
    title: String(entry?.title ?? special.title ?? special.labelShort ?? special.id ?? "-").trim() || "-",
    text: String(entry?.text ?? special.text ?? "").trim(),
  };
};

export const localizeSpecialOptionLabel = (
  scenarioId: string | undefined,
  option: { id?: string; title?: string; label?: string },
  requestedLocale?: CardLocale
): string => {
  const fromCards = localizeCardLabel({
    deck: "special",
    labelShort: option.label ?? option.title,
    id: option.id ? `special.${option.id}` : undefined,
  }, requestedLocale).trim();
  if (fromCards && fromCards !== option.label && fromCards !== option.title && !fromCards.startsWith('special.')) {
    return fromCards;
  }
  const entry = getSpecialConditionEntry(scenarioId, requestedLocale, { id: option.id, labelShort: option.label, title: option.title });
  return String(entry?.title ?? fromCards ?? option.label ?? option.title ?? option.id ?? "-").trim() || "-";
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
