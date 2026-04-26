import type { LocaleCode } from "./localization";
import cardsEn from "../../locales/cards/en.json";
import cardsRu from "../../locales/cards/ru.json";
import worldBunkerEn from "../../locales/world/bunker/en.json";
import worldBunkerRu from "../../locales/world/bunker/ru.json";
import worldDisastersEn from "../../locales/world/disasters/en.json";
import worldDisastersRu from "../../locales/world/disasters/ru.json";
import worldThreatsEn from "../../locales/world/threats/en.json";
import worldThreatsRu from "../../locales/world/threats/ru.json";

type CardsDictionary = {
  cards?: Record<string, string>;
};

type WorldBunkerDictionary = {
  subtitles?: Record<string, string>;
};

type WorldDisastersDictionary = {
  texts?: Record<string, string>;
};

type WorldThreatsDictionary = {
  subtitles?: Record<string, string>;
};

type LocalizableCard = {
  id?: string;
  deck?: string;
  imgUrl?: string;
  labelShort?: string;
  title?: string;
};

type LocalizableWorldCard = {
  kind?: string;
  id?: string;
  imageId?: string;
  imgUrl?: string;
  title?: string;
  description?: string;
};

const CARDS: Record<LocaleCode, CardsDictionary> = {
  ru: cardsRu as CardsDictionary,
  en: cardsEn as CardsDictionary,
};

const WORLD_BUNKER: Record<LocaleCode, WorldBunkerDictionary> = {
  ru: worldBunkerRu as WorldBunkerDictionary,
  en: worldBunkerEn as WorldBunkerDictionary,
};

const WORLD_DISASTERS: Record<LocaleCode, WorldDisastersDictionary> = {
  ru: worldDisastersRu as WorldDisastersDictionary,
  en: worldDisastersEn as WorldDisastersDictionary,
};

const WORLD_THREATS: Record<LocaleCode, WorldThreatsDictionary> = {
  ru: worldThreatsRu as WorldThreatsDictionary,
  en: worldThreatsEn as WorldThreatsDictionary,
};

const normalizeDeck = (deck?: string): string => {
  const raw = String(deck ?? "").trim().toLowerCase();
  if (!raw) return "";
  if (raw === "facts") return "fact";
  return raw;
};

const LOCALE_KEY_RE = /^[a-z0-9_-]+\.[a-z0-9_-]+$/i;

const isLocaleKey = (value?: string): boolean => LOCALE_KEY_RE.test(String(value ?? "").trim());

const cleanAssetRef = (value?: string): string => {
  const raw = String(value ?? "").trim();
  if (!raw) return "";
  return raw.replace(/\\/g, "/").replace(/[?#].*$/, "");
};

const getAssetSlug = (value?: string): string => {
  const clean = cleanAssetRef(value);
  if (!clean) return "";
  const lastSegment = clean.split("/").filter(Boolean).pop() ?? "";
  return lastSegment.replace(/\.[^.]+$/, "").trim().toLowerCase();
};

const stripDeckPrefix = (slug: string, deck: string): string => {
  if (!slug || !deck) return slug;
  const prefix = `${deck}.`;
  return slug.startsWith(prefix) ? slug.slice(prefix.length) : slug;
};

const getAssetDeck = (value?: string): string => {
  const clean = cleanAssetRef(value);
  if (!clean) return "";
  const localeMatch = clean.match(/\/decks\/1x\/(?:ru|en)\/([^/]+)\//i);
  if (localeMatch?.[1]) return normalizeDeck(localeMatch[1]);
  const segments = clean.split("/").filter(Boolean);
  if (segments.length < 2) return "";
  return normalizeDeck(segments[segments.length - 2]);
};

const buildCardKey = (card: LocalizableCard): string => {
  const explicitId = String(card.id ?? "").trim().toLowerCase();
  if (isLocaleKey(explicitId)) return explicitId;
  const labelId = String(card.labelShort ?? "").trim().toLowerCase();
  if (isLocaleKey(labelId)) return labelId;
  const deck = normalizeDeck(card.deck) || getAssetDeck(card.imgUrl);
  const slug = stripDeckPrefix(getAssetSlug(card.imgUrl) || explicitId, deck);
  if (!deck || !slug) return "";
  return `${deck}.${slug}`;
};

const getCardDictionaryValue = (locale: LocaleCode, key: string): string => {
  return CARDS[locale].cards?.[key] ?? "";
};

export const getLocalizedCardLabel = (locale: LocaleCode, card: LocalizableCard): string => {
  const key = buildCardKey(card);
  const localized = key ? getCardDictionaryValue(locale, key) : "";
  return localized || String(card.labelShort ?? card.title ?? "").trim();
};

const buildWorldKey = (card: LocalizableWorldCard): string => {
  const explicitId = String(card.id ?? "").trim().toLowerCase();
  if (isLocaleKey(explicitId)) return explicitId;
  const titleId = String(card.title ?? "").trim().toLowerCase();
  if (isLocaleKey(titleId)) return titleId;
  const kind = normalizeDeck(card.kind);
  const slug = stripDeckPrefix(getAssetSlug(card.imgUrl) || getAssetSlug(card.imageId) || getAssetSlug(card.id) || explicitId, kind);
  if (!kind || !slug) return "";
  return `${kind}.${slug}`;
};

export const getLocalizedWorldTitle = (locale: LocaleCode, card: LocalizableWorldCard): string => {
  const key = buildWorldKey(card);
  const localized = key ? getCardDictionaryValue(locale, key) : "";
  return localized || String(card.title ?? "").trim();
};

export const getLocalizedWorldDescription = (locale: LocaleCode, card: LocalizableWorldCard): string => {
  const key = buildWorldKey(card);
  if (!key) return String(card.description ?? "").trim();
  const kind = normalizeDeck(card.kind || key.split(".")[0]);
  if (kind === "bunker") {
    return WORLD_BUNKER[locale].subtitles?.[key] || String(card.description ?? "").trim();
  }
  if (kind === "threat") {
    return WORLD_THREATS[locale].subtitles?.[key] || String(card.description ?? "").trim();
  }
  if (kind === "disaster") {
    return WORLD_DISASTERS[locale].texts?.[key] || String(card.description ?? "").trim();
  }
  return String(card.description ?? "").trim();
};
