import type { SpecialTargetScope } from "./index.js";
export declare const computeTargetScope: (uiTargeting?: string, text?: string) => SpecialTargetScope | null;
export declare const computeNeighbors: (orderRing: string[], aliveSet: Set<string>, actorId: string) => {
    leftId?: string;
    rightId?: string;
};
export declare const getTargetCandidates: (scope: SpecialTargetScope, actorId: string, orderRing: string[], aliveSet: Set<string>) => string[];
//# sourceMappingURL=targeting.d.ts.map