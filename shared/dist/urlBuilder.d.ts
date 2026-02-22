export interface UrlPair {
    lan: string;
    public?: string;
}
export declare const LINK_PATHS: {
    readonly app: "/";
    readonly spectator: "/spectate";
    readonly overlayView: "/overlay";
    readonly overlayAssets: "/overlay-assets";
    readonly overlayControl: "/overlay-control";
    readonly overlayControlState: "/overlay-control/state";
    readonly overlayControlSave: "/overlay-control/save";
    readonly overlayControlAction: "/overlay-control/action";
    readonly apiOverlayLinks: "/api/overlay-links";
};
export interface BuiltLinkSet {
    lanBase: string;
    publicBase?: string;
    appUrl: UrlPair;
    viewerUrl: UrlPair;
    overlayViewUrl: UrlPair;
    overlayDebugUrl: UrlPair;
    overlayControlUrl: UrlPair;
    overlayControlStateUrl: UrlPair;
    controlPanelUrl: UrlPair;
}
export interface BuildLinkSetInput {
    lanBase: string;
    publicBase?: string | null;
    roomCode: string;
    overlayViewToken: string;
    overlayControlToken: string;
}
export declare function normalizeBase(base: string): string;
export declare function buildLinkSet(input: BuildLinkSetInput): BuiltLinkSet;
//# sourceMappingURL=urlBuilder.d.ts.map