import type { WorldFacedCard } from "@bunker/shared";

const normalizeTitle = (value: string) => value.trim().replace(/\s+/g, " ").toUpperCase();
const normalizeRef = (value: string) =>
  value
    .trim()
    .toLowerCase()
    .replace(/\\/g, "/")
    .replace(/^assets\//, "")
    .replace(/^decks\//, "")
    .replace(/\.[a-z0-9]{2,4}$/i, "")
    .replace(/ё/g, "е")
    .replace(/[^\p{L}\p{N}]+/gu, "");

type ThreatRule = { delta: number; title: string; idContains: string };

const THREAT_RULES: ThreatRule[] = [
  {
    delta: 1,
    title: normalizeTitle("ВМЕСТЕ НА 10 ЛЕТ"),
    idContains: normalizeRef("бункер/вместе на 10 лет"),
  },
  {
    delta: -1,
    title: normalizeTitle("ЗАГАДОЧНЫЙ ЖУРНАЛ"),
    idContains: normalizeRef("бункер/загадочный журнал"),
  },
];

const threatDeltaForCard = (card: Pick<WorldFacedCard, "title" | "id">): number => {
  const normalizedTitle = normalizeTitle(card.title ?? "");
  const normalizedId = normalizeRef(card.id ?? "");
  for (const rule of THREAT_RULES) {
    if (normalizedTitle === rule.title) return rule.delta;
    if (normalizedId && normalizedId.includes(rule.idContains)) return rule.delta;
  }
  return 0;
};

export interface ThreatDeltaResult {
  delta: number;
  reasons: string[];
}

export const getThreatDeltaFromBunkerCards = (
  cards: Array<Pick<WorldFacedCard, "title" | "id" | "isRevealed">>
): ThreatDeltaResult => {
  let delta = 0;
  const reasons: string[] = [];

  for (const card of cards) {
    if (!card.isRevealed) continue;
    const modifier = threatDeltaForCard(card);
    if (!modifier) continue;
    delta += modifier;
    reasons.push(card.title);
  }

  return { delta, reasons };
};
