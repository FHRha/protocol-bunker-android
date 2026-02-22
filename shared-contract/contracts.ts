export interface HealthResponse {
  status: "ok";
  service: "protocol-bunker-host";
  port: number;
  mode: "lan_only";
}

export interface ScenarioMeta {
  id: string;
  name: string;
  description?: string;
  devOnly?: boolean;
}

export type ScenariosResponse = ScenarioMeta[];

export interface OverlayLinksResponse {
  ok: true;
  roomCode: string;
  overlayViewToken: string;
  overlayControlToken: string;
  lanBase: string;
  publicBase: null;
  linkVisibility: "lan_only";
  buildProfile: "android-host-lan";
  links: {
    lanBase: string;
    publicBase: null;
    appUrl: { lan: string };
    viewerUrl: { lan: string };
    overlayViewUrl: { lan: string };
    overlayDebugUrl: { lan: string };
    overlayControlUrl: { lan: string };
    overlayControlStateUrl: { lan: string };
    controlPanelUrl: { lan: string };
  };
}

export interface OverlayControlCategory {
  key: string;
  label: string;
}

export interface OverlayControlPlayer {
  playerId: string;
  name: string;
  connected: boolean;
  alive: boolean;
  nickname: string;
  categories: OverlayControlCategory[];
}

export interface OverlayControlStateResponse {
  ok: true;
  role: "CONTROL";
  roomCode: string;
  categories: OverlayControlCategory[];
  players: OverlayControlPlayer[];
  overrides: Record<string, unknown>;
  overlayState: Record<string, unknown>;
  presenterModeEnabled: boolean;
  presenter: Record<string, unknown>;
}

export interface OverlayControlSaveResponse {
  ok: true;
  roomCode: string;
  overrides: Record<string, unknown>;
}

export interface OverlayControlActionResponse {
  ok: true;
  roomCode: string;
  role: "CONTROL";
  presenterModeEnabled: boolean;
  presenter: Record<string, unknown>;
}

export interface ApiErrorResponse {
  ok: false;
  message: string;
}
