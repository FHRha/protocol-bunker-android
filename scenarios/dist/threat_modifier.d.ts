import type { WorldFacedCard } from "@bunker/shared";
export interface ThreatDeltaResult {
    delta: number;
    reasons: string[];
}
export declare const getThreatDeltaFromBunkerCards: (cards: Array<Pick<WorldFacedCard, "title" | "isRevealed">>) => ThreatDeltaResult;
//# sourceMappingURL=threat_modifier.d.ts.map