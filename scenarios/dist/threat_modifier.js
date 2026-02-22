const THREAT_MODIFIER_BY_CARD_TITLE = {
    "ВМЕСТЕ НА 10 ЛЕТ": 1,
    "ЗАГАДОЧНЫЙ ЖУРНАЛ": -1,
};
const normalizeTitle = (value) => value.trim().replace(/\s+/g, " ").toUpperCase();
export const getThreatDeltaFromBunkerCards = (cards) => {
    let delta = 0;
    const reasons = [];
    for (const card of cards) {
        if (!card.isRevealed)
            continue;
        const modifier = THREAT_MODIFIER_BY_CARD_TITLE[normalizeTitle(card.title)];
        if (!modifier)
            continue;
        delta += modifier;
        reasons.push(card.title);
    }
    return { delta, reasons };
};
