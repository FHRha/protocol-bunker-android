import { type AssetCatalog, type WorldState30 } from "@bunker/shared";
export declare const getWorldCounts: (playerCount: number) => {
    bunker: number;
    threats: number;
};
export declare const rollWorldFromAssets: (assets: AssetCatalog, rng: () => number, playerCount: number) => WorldState30;
//# sourceMappingURL=world_deck.d.ts.map