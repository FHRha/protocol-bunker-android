import { z } from "zod";
export * from "./targeting.js";
export type RoomPhase = "lobby" | "game";
export type ScenarioPhase = "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended";
export type VotePhase = "voting" | "voteSpecialWindow" | "voteResolve";
export type PlayerStatus = "alive" | "eliminated" | "left_bunker";
export type GameEventKind = "roundStart" | "votingStart" | "elimination" | "gameEnd" | "info" | "playerDisconnected" | "playerReconnected" | "playerLeftBunker";
export type SpecialConditionTrigger = "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
export type GameTimerKind = "reveal_discussion" | "pre_vote" | "post_vote";
export type ContinuePermission = "host_only" | "revealer_only" | "anyone";
export type RevealTimeoutAction = "random_card" | "skip_player";
export type SpecialUsageMode = "anytime" | "only_during_voting";
export type FinalThreatReveal = "host" | "anyone";
export type SpecialTargetScope = "neighbors" | "any_alive" | "self" | "any_including_self";
export type WorldCardKind = "bunker" | "disaster" | "threat";
export type PostGameOutcome = "survived" | "failed";
export type HostChangeReason = "disconnect_timeout" | "left_bunker" | "eliminated" | "manual";
export type RulesetMode = "auto" | "preset" | "manual";
export type Role = "VIEW" | "PLAYER" | "CONTROL";
export interface WorldCard {
    kind: WorldCardKind;
    id: string;
    title: string;
    description: string;
    text?: string;
    imageId?: string;
}
export interface WorldFacedCard extends WorldCard {
    isRevealed: boolean;
    revealedAtRound?: number;
    revealedBy?: string;
}
export interface WorldState30 {
    disaster: WorldCard;
    bunker: WorldFacedCard[];
    threats: WorldFacedCard[];
    counts: {
        bunker: number;
        threats: number;
    };
}
export interface WorldEvent {
    type: "bunker_revealed";
    index: number;
    round: number;
}
export interface PostGameState {
    isActive: boolean;
    enteredAt: number;
    outcome?: PostGameOutcome;
    decidedBy?: string;
    decidedAt?: number;
}
export interface OverlayTagView {
    label: string;
    revealed: boolean;
    value: string;
}
export interface OverlayCategoryView {
    key: string;
    label: string;
    revealed: boolean;
    value: string;
    imgUrl?: string;
}
export interface OverlayOverrideEnabled {
    topBunker?: boolean;
    topCatastrophe?: boolean;
    topThreats?: boolean;
    playerNames?: boolean;
    playerTraits?: boolean;
    playerCategories?: boolean;
}
export interface OverlayOverrideTop {
    bunkerLines?: string[];
    catastropheText?: string;
    threatsLines?: string[];
}
export interface OverlayOverridePlayerTraits {
    sex?: string;
    age?: string;
    orient?: string;
}
export interface OverlayOverridePlayerEnabled {
    name?: boolean;
    traits?: boolean;
    categories?: Record<string, boolean>;
}
export interface OverlayOverridePlayer {
    name?: string;
    traits?: OverlayOverridePlayerTraits;
    categories?: Record<string, string>;
    enabled?: OverlayOverridePlayerEnabled;
}
export interface OverlayExtraText {
    id: string;
    text: string;
    x: number;
    y: number;
    align?: "left" | "center" | "right";
    size?: number;
    color?: string;
    shadow?: boolean;
    visible?: boolean;
}
export interface OverlayOverrides {
    enabled?: OverlayOverrideEnabled;
    top?: OverlayOverrideTop;
    players?: Record<string, OverlayOverridePlayer>;
    extraTexts?: OverlayExtraText[];
}
export interface OverlayPlayerView {
    id: string;
    nickname: string;
    connected?: boolean;
    alive: boolean;
    tags: {
        sex: OverlayTagView;
        age: OverlayTagView;
        orientation: OverlayTagView;
    };
    categories: OverlayCategoryView[];
}
export interface OverlayTopCardItem {
    title: string;
    subtitle?: string;
    imageId?: string;
}
export interface OverlayState {
    roomId: string;
    playerCount: number;
    top: {
        bunker: {
            revealed: number;
            total: number;
            lines: string[];
            items?: OverlayTopCardItem[];
        };
        catastrophe: {
            text: string;
            title?: string;
            imageId?: string;
        };
        threats: {
            revealed: number;
            total: number;
            lines: string[];
            items?: OverlayTopCardItem[];
        };
    };
    players: OverlayPlayerView[];
    overrides?: OverlayOverrides;
}
export interface GameTimerState {
    kind: GameTimerKind;
    endsAt: number;
}
export interface ManualRulesConfig {
    bunkerSlots: number;
    votesByRound: number[];
    targetReveals: number;
    seedTemplatePlayers?: number;
}
export interface GameRuleset {
    playerCount: number;
    votesPerRound: number[];
    totalExiles: number;
    bunkerSeats: number;
    rulesetMode: RulesetMode;
    manualConfig?: ManualRulesConfig;
}
export interface GameSettings {
    enableRevealDiscussionTimer: boolean;
    revealDiscussionSeconds: number;
    enablePreVoteDiscussionTimer: boolean;
    preVoteDiscussionSeconds: number;
    enablePostVoteDiscussionTimer: boolean;
    postVoteDiscussionSeconds: number;
    enablePresenterMode: boolean;
    continuePermission: ContinuePermission;
    revealTimeoutAction: RevealTimeoutAction;
    revealsBeforeVoting: number;
    specialUsage: SpecialUsageMode;
    maxPlayers: number;
    finalThreatReveal: FinalThreatReveal;
}
export interface ScenarioMeta {
    id: string;
    name: string;
    description?: string;
    devOnly?: boolean;
}
export interface CardRef {
    id: string;
    deck: string;
    instanceId?: string;
    labelShort?: string;
    secret?: boolean;
    missing?: boolean;
}
export interface CardInHand extends CardRef {
    revealed: boolean;
}
export interface SpecialConditionEffect {
    type: string;
    params?: Record<string, unknown>;
}
export interface SpecialConditionInstance {
    instanceId: string;
    id: string;
    title: string;
    text: string;
    trigger: SpecialConditionTrigger;
    effect: SpecialConditionEffect;
    implemented: boolean;
    revealedPublic: boolean;
    used: boolean;
    imgUrl?: string;
    needsChoice?: boolean;
    choiceKind?: "player" | "neighbor" | "category" | "none";
    allowSelfTarget?: boolean;
    targetScope?: SpecialTargetScope;
}
export interface PublicSpecialConditionView {
    instanceId: string;
    title: string;
    imgUrl?: string;
}
export interface PublicCategoryCard {
    labelShort: string;
    imgUrl?: string;
}
export interface YouCategoryCard {
    instanceId: string;
    labelShort: string;
    revealed: boolean;
}
export interface PlayerSummary {
    playerId: string;
    name: string;
    connected: boolean;
    disconnectedAt?: number;
    totalAbsentMs?: number;
    currentOfflineMs?: number;
    kickRemainingMs?: number;
    leftBunker?: boolean;
}
export interface RoomState {
    roomCode: string;
    players: PlayerSummary[];
    hostId: string;
    controlId: string;
    phase: RoomPhase;
    scenarioMeta: ScenarioMeta;
    settings: GameSettings;
    ruleset: GameRuleset;
    rulesOverriddenByHost: boolean;
    rulesPresetCount?: number;
    world?: WorldState30;
    isDev?: boolean;
}
export interface StatePatchPayload {
    roomState?: Partial<RoomState>;
    gameView?: Partial<GameView>;
}
export interface PublicCategorySlot {
    category: string;
    status: "hidden" | "revealed";
    cards: PublicCategoryCard[];
}
export interface YouCategorySlot {
    category: string;
    cards: YouCategoryCard[];
}
export interface PublicPlayerView {
    playerId: string;
    name: string;
    status: PlayerStatus;
    connected: boolean;
    disconnectedAt?: number;
    totalAbsentMs?: number;
    currentOfflineMs?: number;
    kickRemainingMs?: number;
    leftBunker?: boolean;
    revealedCards: CardRef[];
    revealedCount: number;
    totalCards: number;
    specialRevealed: boolean;
    categories: PublicCategorySlot[];
}
export interface VotingView {
    hasVoted: boolean;
}
export interface VotingProgress {
    voted: number;
    total: number;
}
export interface ThreatModifierView {
    delta: number;
    reasons: string[];
    baseCount: number;
    finalCount: number;
}
export interface VotePublic {
    voterId: string;
    voterName: string;
    targetId?: string;
    targetName?: string;
    status: "voted" | "not_voted" | "invalid";
    reason?: string;
    submittedAt?: number;
}
export interface GameEvent {
    id: string;
    kind: GameEventKind;
    message: string;
    createdAt: number;
}
export interface GameView {
    phase: ScenarioPhase;
    round: number;
    categoryOrder: string[];
    lastStageText?: string;
    ruleset: GameRuleset;
    world?: WorldState30;
    worldEvent?: WorldEvent;
    postGame?: PostGameState;
    you: {
        playerId: string;
        name: string;
        hand: CardInHand[];
        categories: YouCategorySlot[];
        specialConditions: SpecialConditionInstance[];
    };
    public: {
        players: PublicPlayerView[];
        revealedThisRound: string[];
        roundRevealedCount?: number;
        roundTotalAlive?: number;
        currentTurnPlayerId?: string | null;
        votesRemainingInRound?: number;
        votesTotalThisRound?: number;
        revealLimit?: number;
        voting?: VotingView;
        votePhase?: VotePhase | null;
        votesPublic?: VotePublic[];
        votingProgress?: VotingProgress;
        threatModifier?: ThreatModifierView;
        canOpenVotingModal?: boolean;
        canContinue?: boolean;
        activeTimer?: GameTimerState | null;
        voteModalOpen?: boolean;
        lastEliminated?: string;
        winners?: string[];
        resolutionNote?: string;
        roundRules?: {
            noTalkUntilVoting?: boolean;
            forcedRevealCategory?: string;
        };
    };
}
export interface AssetCard {
    id: string;
    deck: string;
    labelShort: string;
}
export interface AssetCatalog {
    decks: Record<string, AssetCard[]>;
}
export interface ScenarioPlayer {
    playerId: string;
    name: string;
}
export interface ScenarioContext {
    roomCode: string;
    createdAt: number;
    rng: () => number;
    assets: AssetCatalog;
    players: ScenarioPlayer[];
    settings: GameSettings;
    hostId: string;
    ruleset?: GameRuleset;
    onStateChange?: () => void;
    onEvent?: (event: GameEvent) => void;
}
export type ScenarioAction = {
    type: "revealCard";
    payload: {
        cardId: string;
    };
} | {
    type: "vote";
    payload: {
        targetPlayerId: string;
    };
} | {
    type: "finalizeVoting";
    payload: {};
} | {
    type: "applySpecial";
    payload: {
        specialInstanceId: string;
        payload?: Record<string, unknown>;
    };
} | {
    type: "revealWorldThreat";
    payload: {
        index: number;
    };
} | {
    type: "setBunkerOutcome";
    payload: {
        outcome: PostGameOutcome;
    };
} | {
    type: "devSkipRound";
    payload: {};
} | {
    type: "devKickPlayer";
    payload: {
        targetPlayerId: string;
    };
} | {
    type: "markLeftBunker";
    payload: {
        targetPlayerId: string;
    };
} | {
    type: "continueRound";
    payload: {};
} | {
    type: "devAddPlayer";
    payload: {
        name?: string;
    };
} | {
    type: "devRemovePlayer";
    payload: {
        targetPlayerId?: string;
    };
};
export interface ScenarioActionResult {
    error?: string;
    stateChanged?: boolean;
}
export interface ScenarioSession {
    getGameView(playerId: string): GameView;
    handleAction(playerId: string, action: ScenarioAction): ScenarioActionResult;
}
export interface ScenarioModule {
    meta: ScenarioMeta;
    createSession(ctx: ScenarioContext): ScenarioSession;
}
export declare const PlayerSummarySchema: z.ZodObject<{
    playerId: z.ZodString;
    name: z.ZodString;
    connected: z.ZodBoolean;
    disconnectedAt: z.ZodOptional<z.ZodNumber>;
    totalAbsentMs: z.ZodOptional<z.ZodNumber>;
    currentOfflineMs: z.ZodOptional<z.ZodNumber>;
    kickRemainingMs: z.ZodOptional<z.ZodNumber>;
    leftBunker: z.ZodOptional<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    playerId: string;
    name: string;
    connected: boolean;
    disconnectedAt?: number | undefined;
    totalAbsentMs?: number | undefined;
    currentOfflineMs?: number | undefined;
    kickRemainingMs?: number | undefined;
    leftBunker?: boolean | undefined;
}, {
    playerId: string;
    name: string;
    connected: boolean;
    disconnectedAt?: number | undefined;
    totalAbsentMs?: number | undefined;
    currentOfflineMs?: number | undefined;
    kickRemainingMs?: number | undefined;
    leftBunker?: boolean | undefined;
}>;
export declare const ScenarioMetaSchema: z.ZodObject<{
    id: z.ZodString;
    name: z.ZodString;
    description: z.ZodOptional<z.ZodString>;
    devOnly: z.ZodOptional<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    name: string;
    id: string;
    description?: string | undefined;
    devOnly?: boolean | undefined;
}, {
    name: string;
    id: string;
    description?: string | undefined;
    devOnly?: boolean | undefined;
}>;
export declare const GameTimerStateSchema: z.ZodObject<{
    kind: z.ZodUnion<[z.ZodLiteral<"reveal_discussion">, z.ZodLiteral<"pre_vote">, z.ZodLiteral<"post_vote">]>;
    endsAt: z.ZodNumber;
}, "strip", z.ZodTypeAny, {
    kind: "reveal_discussion" | "pre_vote" | "post_vote";
    endsAt: number;
}, {
    kind: "reveal_discussion" | "pre_vote" | "post_vote";
    endsAt: number;
}>;
export declare const ManualRulesConfigSchema: z.ZodObject<{
    bunkerSlots: z.ZodNumber;
    votesByRound: z.ZodArray<z.ZodNumber, "many">;
    targetReveals: z.ZodDefault<z.ZodNumber>;
    seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    bunkerSlots: number;
    votesByRound: number[];
    targetReveals: number;
    seedTemplatePlayers?: number | undefined;
}, {
    bunkerSlots: number;
    votesByRound: number[];
    targetReveals?: number | undefined;
    seedTemplatePlayers?: number | undefined;
}>;
export declare const GameRulesetSchema: z.ZodObject<{
    playerCount: z.ZodNumber;
    votesPerRound: z.ZodArray<z.ZodNumber, "many">;
    totalExiles: z.ZodNumber;
    bunkerSeats: z.ZodNumber;
    rulesetMode: z.ZodUnion<[z.ZodLiteral<"auto">, z.ZodLiteral<"preset">, z.ZodLiteral<"manual">]>;
    manualConfig: z.ZodOptional<z.ZodObject<{
        bunkerSlots: z.ZodNumber;
        votesByRound: z.ZodArray<z.ZodNumber, "many">;
        targetReveals: z.ZodDefault<z.ZodNumber>;
        seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        bunkerSlots: number;
        votesByRound: number[];
        targetReveals: number;
        seedTemplatePlayers?: number | undefined;
    }, {
        bunkerSlots: number;
        votesByRound: number[];
        targetReveals?: number | undefined;
        seedTemplatePlayers?: number | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    playerCount: number;
    votesPerRound: number[];
    totalExiles: number;
    bunkerSeats: number;
    rulesetMode: "auto" | "preset" | "manual";
    manualConfig?: {
        bunkerSlots: number;
        votesByRound: number[];
        targetReveals: number;
        seedTemplatePlayers?: number | undefined;
    } | undefined;
}, {
    playerCount: number;
    votesPerRound: number[];
    totalExiles: number;
    bunkerSeats: number;
    rulesetMode: "auto" | "preset" | "manual";
    manualConfig?: {
        bunkerSlots: number;
        votesByRound: number[];
        targetReveals?: number | undefined;
        seedTemplatePlayers?: number | undefined;
    } | undefined;
}>;
export declare const WorldCardSchema: z.ZodObject<{
    kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
    id: z.ZodString;
    title: z.ZodString;
    description: z.ZodString;
    text: z.ZodOptional<z.ZodString>;
    imageId: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    id: string;
    description: string;
    kind: "bunker" | "disaster" | "threat";
    title: string;
    text?: string | undefined;
    imageId?: string | undefined;
}, {
    id: string;
    description: string;
    kind: "bunker" | "disaster" | "threat";
    title: string;
    text?: string | undefined;
    imageId?: string | undefined;
}>;
export declare const WorldFacedCardSchema: z.ZodObject<{
    kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
    id: z.ZodString;
    title: z.ZodString;
    description: z.ZodString;
    text: z.ZodOptional<z.ZodString>;
    imageId: z.ZodOptional<z.ZodString>;
} & {
    isRevealed: z.ZodBoolean;
    revealedAtRound: z.ZodOptional<z.ZodNumber>;
    revealedBy: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    id: string;
    description: string;
    kind: "bunker" | "disaster" | "threat";
    title: string;
    isRevealed: boolean;
    text?: string | undefined;
    imageId?: string | undefined;
    revealedAtRound?: number | undefined;
    revealedBy?: string | undefined;
}, {
    id: string;
    description: string;
    kind: "bunker" | "disaster" | "threat";
    title: string;
    isRevealed: boolean;
    text?: string | undefined;
    imageId?: string | undefined;
    revealedAtRound?: number | undefined;
    revealedBy?: string | undefined;
}>;
export declare const WorldState30Schema: z.ZodObject<{
    disaster: z.ZodObject<{
        kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
        id: z.ZodString;
        title: z.ZodString;
        description: z.ZodString;
        text: z.ZodOptional<z.ZodString>;
        imageId: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        text?: string | undefined;
        imageId?: string | undefined;
    }, {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        text?: string | undefined;
        imageId?: string | undefined;
    }>;
    bunker: z.ZodArray<z.ZodObject<{
        kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
        id: z.ZodString;
        title: z.ZodString;
        description: z.ZodString;
        text: z.ZodOptional<z.ZodString>;
        imageId: z.ZodOptional<z.ZodString>;
    } & {
        isRevealed: z.ZodBoolean;
        revealedAtRound: z.ZodOptional<z.ZodNumber>;
        revealedBy: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        isRevealed: boolean;
        text?: string | undefined;
        imageId?: string | undefined;
        revealedAtRound?: number | undefined;
        revealedBy?: string | undefined;
    }, {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        isRevealed: boolean;
        text?: string | undefined;
        imageId?: string | undefined;
        revealedAtRound?: number | undefined;
        revealedBy?: string | undefined;
    }>, "many">;
    threats: z.ZodArray<z.ZodObject<{
        kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
        id: z.ZodString;
        title: z.ZodString;
        description: z.ZodString;
        text: z.ZodOptional<z.ZodString>;
        imageId: z.ZodOptional<z.ZodString>;
    } & {
        isRevealed: z.ZodBoolean;
        revealedAtRound: z.ZodOptional<z.ZodNumber>;
        revealedBy: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        isRevealed: boolean;
        text?: string | undefined;
        imageId?: string | undefined;
        revealedAtRound?: number | undefined;
        revealedBy?: string | undefined;
    }, {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        isRevealed: boolean;
        text?: string | undefined;
        imageId?: string | undefined;
        revealedAtRound?: number | undefined;
        revealedBy?: string | undefined;
    }>, "many">;
    counts: z.ZodObject<{
        bunker: z.ZodNumber;
        threats: z.ZodNumber;
    }, "strip", z.ZodTypeAny, {
        bunker: number;
        threats: number;
    }, {
        bunker: number;
        threats: number;
    }>;
}, "strip", z.ZodTypeAny, {
    bunker: {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        isRevealed: boolean;
        text?: string | undefined;
        imageId?: string | undefined;
        revealedAtRound?: number | undefined;
        revealedBy?: string | undefined;
    }[];
    disaster: {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        text?: string | undefined;
        imageId?: string | undefined;
    };
    threats: {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        isRevealed: boolean;
        text?: string | undefined;
        imageId?: string | undefined;
        revealedAtRound?: number | undefined;
        revealedBy?: string | undefined;
    }[];
    counts: {
        bunker: number;
        threats: number;
    };
}, {
    bunker: {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        isRevealed: boolean;
        text?: string | undefined;
        imageId?: string | undefined;
        revealedAtRound?: number | undefined;
        revealedBy?: string | undefined;
    }[];
    disaster: {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        text?: string | undefined;
        imageId?: string | undefined;
    };
    threats: {
        id: string;
        description: string;
        kind: "bunker" | "disaster" | "threat";
        title: string;
        isRevealed: boolean;
        text?: string | undefined;
        imageId?: string | undefined;
        revealedAtRound?: number | undefined;
        revealedBy?: string | undefined;
    }[];
    counts: {
        bunker: number;
        threats: number;
    };
}>;
export declare const WorldEventSchema: z.ZodObject<{
    type: z.ZodLiteral<"bunker_revealed">;
    index: z.ZodNumber;
    round: z.ZodNumber;
}, "strip", z.ZodTypeAny, {
    type: "bunker_revealed";
    index: number;
    round: number;
}, {
    type: "bunker_revealed";
    index: number;
    round: number;
}>;
export declare const PostGameStateSchema: z.ZodObject<{
    isActive: z.ZodBoolean;
    enteredAt: z.ZodNumber;
    outcome: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"survived">, z.ZodLiteral<"failed">]>>;
    decidedBy: z.ZodOptional<z.ZodString>;
    decidedAt: z.ZodOptional<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    isActive: boolean;
    enteredAt: number;
    outcome?: "survived" | "failed" | undefined;
    decidedBy?: string | undefined;
    decidedAt?: number | undefined;
}, {
    isActive: boolean;
    enteredAt: number;
    outcome?: "survived" | "failed" | undefined;
    decidedBy?: string | undefined;
    decidedAt?: number | undefined;
}>;
export declare const OverlayTagViewSchema: z.ZodObject<{
    label: z.ZodString;
    revealed: z.ZodBoolean;
    value: z.ZodString;
}, "strip", z.ZodTypeAny, {
    revealed: boolean;
    value: string;
    label: string;
}, {
    revealed: boolean;
    value: string;
    label: string;
}>;
export declare const OverlayCategoryViewSchema: z.ZodObject<{
    key: z.ZodString;
    label: z.ZodString;
    revealed: z.ZodBoolean;
    value: z.ZodString;
    imgUrl: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    revealed: boolean;
    value: string;
    label: string;
    key: string;
    imgUrl?: string | undefined;
}, {
    revealed: boolean;
    value: string;
    label: string;
    key: string;
    imgUrl?: string | undefined;
}>;
export declare const OverlayOverrideEnabledSchema: z.ZodObject<{
    topBunker: z.ZodOptional<z.ZodBoolean>;
    topCatastrophe: z.ZodOptional<z.ZodBoolean>;
    topThreats: z.ZodOptional<z.ZodBoolean>;
    playerNames: z.ZodOptional<z.ZodBoolean>;
    playerTraits: z.ZodOptional<z.ZodBoolean>;
    playerCategories: z.ZodOptional<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    topBunker?: boolean | undefined;
    topCatastrophe?: boolean | undefined;
    topThreats?: boolean | undefined;
    playerNames?: boolean | undefined;
    playerTraits?: boolean | undefined;
    playerCategories?: boolean | undefined;
}, {
    topBunker?: boolean | undefined;
    topCatastrophe?: boolean | undefined;
    topThreats?: boolean | undefined;
    playerNames?: boolean | undefined;
    playerTraits?: boolean | undefined;
    playerCategories?: boolean | undefined;
}>;
export declare const OverlayOverrideTopSchema: z.ZodObject<{
    bunkerLines: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
    catastropheText: z.ZodOptional<z.ZodString>;
    threatsLines: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
}, "strip", z.ZodTypeAny, {
    bunkerLines?: string[] | undefined;
    catastropheText?: string | undefined;
    threatsLines?: string[] | undefined;
}, {
    bunkerLines?: string[] | undefined;
    catastropheText?: string | undefined;
    threatsLines?: string[] | undefined;
}>;
export declare const OverlayOverridePlayerTraitsSchema: z.ZodObject<{
    sex: z.ZodOptional<z.ZodString>;
    age: z.ZodOptional<z.ZodString>;
    orient: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    sex?: string | undefined;
    age?: string | undefined;
    orient?: string | undefined;
}, {
    sex?: string | undefined;
    age?: string | undefined;
    orient?: string | undefined;
}>;
export declare const OverlayOverridePlayerEnabledSchema: z.ZodObject<{
    name: z.ZodOptional<z.ZodBoolean>;
    traits: z.ZodOptional<z.ZodBoolean>;
    categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodBoolean>>;
}, "strip", z.ZodTypeAny, {
    name?: boolean | undefined;
    traits?: boolean | undefined;
    categories?: Record<string, boolean> | undefined;
}, {
    name?: boolean | undefined;
    traits?: boolean | undefined;
    categories?: Record<string, boolean> | undefined;
}>;
export declare const OverlayOverridePlayerSchema: z.ZodObject<{
    name: z.ZodOptional<z.ZodString>;
    traits: z.ZodOptional<z.ZodObject<{
        sex: z.ZodOptional<z.ZodString>;
        age: z.ZodOptional<z.ZodString>;
        orient: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        sex?: string | undefined;
        age?: string | undefined;
        orient?: string | undefined;
    }, {
        sex?: string | undefined;
        age?: string | undefined;
        orient?: string | undefined;
    }>>;
    categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
    enabled: z.ZodOptional<z.ZodObject<{
        name: z.ZodOptional<z.ZodBoolean>;
        traits: z.ZodOptional<z.ZodBoolean>;
        categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodBoolean>>;
    }, "strip", z.ZodTypeAny, {
        name?: boolean | undefined;
        traits?: boolean | undefined;
        categories?: Record<string, boolean> | undefined;
    }, {
        name?: boolean | undefined;
        traits?: boolean | undefined;
        categories?: Record<string, boolean> | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    name?: string | undefined;
    traits?: {
        sex?: string | undefined;
        age?: string | undefined;
        orient?: string | undefined;
    } | undefined;
    categories?: Record<string, string> | undefined;
    enabled?: {
        name?: boolean | undefined;
        traits?: boolean | undefined;
        categories?: Record<string, boolean> | undefined;
    } | undefined;
}, {
    name?: string | undefined;
    traits?: {
        sex?: string | undefined;
        age?: string | undefined;
        orient?: string | undefined;
    } | undefined;
    categories?: Record<string, string> | undefined;
    enabled?: {
        name?: boolean | undefined;
        traits?: boolean | undefined;
        categories?: Record<string, boolean> | undefined;
    } | undefined;
}>;
export declare const OverlayExtraTextSchema: z.ZodObject<{
    id: z.ZodString;
    text: z.ZodString;
    x: z.ZodNumber;
    y: z.ZodNumber;
    align: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"left">, z.ZodLiteral<"center">, z.ZodLiteral<"right">]>>;
    size: z.ZodOptional<z.ZodNumber>;
    color: z.ZodOptional<z.ZodString>;
    shadow: z.ZodOptional<z.ZodBoolean>;
    visible: z.ZodOptional<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    id: string;
    text: string;
    x: number;
    y: number;
    align?: "left" | "right" | "center" | undefined;
    size?: number | undefined;
    color?: string | undefined;
    shadow?: boolean | undefined;
    visible?: boolean | undefined;
}, {
    id: string;
    text: string;
    x: number;
    y: number;
    align?: "left" | "right" | "center" | undefined;
    size?: number | undefined;
    color?: string | undefined;
    shadow?: boolean | undefined;
    visible?: boolean | undefined;
}>;
export declare const OverlayOverridesSchema: z.ZodObject<{
    enabled: z.ZodOptional<z.ZodObject<{
        topBunker: z.ZodOptional<z.ZodBoolean>;
        topCatastrophe: z.ZodOptional<z.ZodBoolean>;
        topThreats: z.ZodOptional<z.ZodBoolean>;
        playerNames: z.ZodOptional<z.ZodBoolean>;
        playerTraits: z.ZodOptional<z.ZodBoolean>;
        playerCategories: z.ZodOptional<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        topBunker?: boolean | undefined;
        topCatastrophe?: boolean | undefined;
        topThreats?: boolean | undefined;
        playerNames?: boolean | undefined;
        playerTraits?: boolean | undefined;
        playerCategories?: boolean | undefined;
    }, {
        topBunker?: boolean | undefined;
        topCatastrophe?: boolean | undefined;
        topThreats?: boolean | undefined;
        playerNames?: boolean | undefined;
        playerTraits?: boolean | undefined;
        playerCategories?: boolean | undefined;
    }>>;
    top: z.ZodOptional<z.ZodObject<{
        bunkerLines: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
        catastropheText: z.ZodOptional<z.ZodString>;
        threatsLines: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
    }, "strip", z.ZodTypeAny, {
        bunkerLines?: string[] | undefined;
        catastropheText?: string | undefined;
        threatsLines?: string[] | undefined;
    }, {
        bunkerLines?: string[] | undefined;
        catastropheText?: string | undefined;
        threatsLines?: string[] | undefined;
    }>>;
    players: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodObject<{
        name: z.ZodOptional<z.ZodString>;
        traits: z.ZodOptional<z.ZodObject<{
            sex: z.ZodOptional<z.ZodString>;
            age: z.ZodOptional<z.ZodString>;
            orient: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            sex?: string | undefined;
            age?: string | undefined;
            orient?: string | undefined;
        }, {
            sex?: string | undefined;
            age?: string | undefined;
            orient?: string | undefined;
        }>>;
        categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
        enabled: z.ZodOptional<z.ZodObject<{
            name: z.ZodOptional<z.ZodBoolean>;
            traits: z.ZodOptional<z.ZodBoolean>;
            categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodBoolean>>;
        }, "strip", z.ZodTypeAny, {
            name?: boolean | undefined;
            traits?: boolean | undefined;
            categories?: Record<string, boolean> | undefined;
        }, {
            name?: boolean | undefined;
            traits?: boolean | undefined;
            categories?: Record<string, boolean> | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        name?: string | undefined;
        traits?: {
            sex?: string | undefined;
            age?: string | undefined;
            orient?: string | undefined;
        } | undefined;
        categories?: Record<string, string> | undefined;
        enabled?: {
            name?: boolean | undefined;
            traits?: boolean | undefined;
            categories?: Record<string, boolean> | undefined;
        } | undefined;
    }, {
        name?: string | undefined;
        traits?: {
            sex?: string | undefined;
            age?: string | undefined;
            orient?: string | undefined;
        } | undefined;
        categories?: Record<string, string> | undefined;
        enabled?: {
            name?: boolean | undefined;
            traits?: boolean | undefined;
            categories?: Record<string, boolean> | undefined;
        } | undefined;
    }>>>;
    extraTexts: z.ZodOptional<z.ZodArray<z.ZodObject<{
        id: z.ZodString;
        text: z.ZodString;
        x: z.ZodNumber;
        y: z.ZodNumber;
        align: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"left">, z.ZodLiteral<"center">, z.ZodLiteral<"right">]>>;
        size: z.ZodOptional<z.ZodNumber>;
        color: z.ZodOptional<z.ZodString>;
        shadow: z.ZodOptional<z.ZodBoolean>;
        visible: z.ZodOptional<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        id: string;
        text: string;
        x: number;
        y: number;
        align?: "left" | "right" | "center" | undefined;
        size?: number | undefined;
        color?: string | undefined;
        shadow?: boolean | undefined;
        visible?: boolean | undefined;
    }, {
        id: string;
        text: string;
        x: number;
        y: number;
        align?: "left" | "right" | "center" | undefined;
        size?: number | undefined;
        color?: string | undefined;
        shadow?: boolean | undefined;
        visible?: boolean | undefined;
    }>, "many">>;
}, "strip", z.ZodTypeAny, {
    enabled?: {
        topBunker?: boolean | undefined;
        topCatastrophe?: boolean | undefined;
        topThreats?: boolean | undefined;
        playerNames?: boolean | undefined;
        playerTraits?: boolean | undefined;
        playerCategories?: boolean | undefined;
    } | undefined;
    top?: {
        bunkerLines?: string[] | undefined;
        catastropheText?: string | undefined;
        threatsLines?: string[] | undefined;
    } | undefined;
    players?: Record<string, {
        name?: string | undefined;
        traits?: {
            sex?: string | undefined;
            age?: string | undefined;
            orient?: string | undefined;
        } | undefined;
        categories?: Record<string, string> | undefined;
        enabled?: {
            name?: boolean | undefined;
            traits?: boolean | undefined;
            categories?: Record<string, boolean> | undefined;
        } | undefined;
    }> | undefined;
    extraTexts?: {
        id: string;
        text: string;
        x: number;
        y: number;
        align?: "left" | "right" | "center" | undefined;
        size?: number | undefined;
        color?: string | undefined;
        shadow?: boolean | undefined;
        visible?: boolean | undefined;
    }[] | undefined;
}, {
    enabled?: {
        topBunker?: boolean | undefined;
        topCatastrophe?: boolean | undefined;
        topThreats?: boolean | undefined;
        playerNames?: boolean | undefined;
        playerTraits?: boolean | undefined;
        playerCategories?: boolean | undefined;
    } | undefined;
    top?: {
        bunkerLines?: string[] | undefined;
        catastropheText?: string | undefined;
        threatsLines?: string[] | undefined;
    } | undefined;
    players?: Record<string, {
        name?: string | undefined;
        traits?: {
            sex?: string | undefined;
            age?: string | undefined;
            orient?: string | undefined;
        } | undefined;
        categories?: Record<string, string> | undefined;
        enabled?: {
            name?: boolean | undefined;
            traits?: boolean | undefined;
            categories?: Record<string, boolean> | undefined;
        } | undefined;
    }> | undefined;
    extraTexts?: {
        id: string;
        text: string;
        x: number;
        y: number;
        align?: "left" | "right" | "center" | undefined;
        size?: number | undefined;
        color?: string | undefined;
        shadow?: boolean | undefined;
        visible?: boolean | undefined;
    }[] | undefined;
}>;
export declare const OverlayPlayerViewSchema: z.ZodObject<{
    id: z.ZodString;
    nickname: z.ZodString;
    connected: z.ZodOptional<z.ZodBoolean>;
    alive: z.ZodBoolean;
    tags: z.ZodObject<{
        sex: z.ZodObject<{
            label: z.ZodString;
            revealed: z.ZodBoolean;
            value: z.ZodString;
        }, "strip", z.ZodTypeAny, {
            revealed: boolean;
            value: string;
            label: string;
        }, {
            revealed: boolean;
            value: string;
            label: string;
        }>;
        age: z.ZodObject<{
            label: z.ZodString;
            revealed: z.ZodBoolean;
            value: z.ZodString;
        }, "strip", z.ZodTypeAny, {
            revealed: boolean;
            value: string;
            label: string;
        }, {
            revealed: boolean;
            value: string;
            label: string;
        }>;
        orientation: z.ZodObject<{
            label: z.ZodString;
            revealed: z.ZodBoolean;
            value: z.ZodString;
        }, "strip", z.ZodTypeAny, {
            revealed: boolean;
            value: string;
            label: string;
        }, {
            revealed: boolean;
            value: string;
            label: string;
        }>;
    }, "strip", z.ZodTypeAny, {
        sex: {
            revealed: boolean;
            value: string;
            label: string;
        };
        age: {
            revealed: boolean;
            value: string;
            label: string;
        };
        orientation: {
            revealed: boolean;
            value: string;
            label: string;
        };
    }, {
        sex: {
            revealed: boolean;
            value: string;
            label: string;
        };
        age: {
            revealed: boolean;
            value: string;
            label: string;
        };
        orientation: {
            revealed: boolean;
            value: string;
            label: string;
        };
    }>;
    categories: z.ZodArray<z.ZodObject<{
        key: z.ZodString;
        label: z.ZodString;
        revealed: z.ZodBoolean;
        value: z.ZodString;
        imgUrl: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        revealed: boolean;
        value: string;
        label: string;
        key: string;
        imgUrl?: string | undefined;
    }, {
        revealed: boolean;
        value: string;
        label: string;
        key: string;
        imgUrl?: string | undefined;
    }>, "many">;
}, "strip", z.ZodTypeAny, {
    alive: boolean;
    id: string;
    categories: {
        revealed: boolean;
        value: string;
        label: string;
        key: string;
        imgUrl?: string | undefined;
    }[];
    nickname: string;
    tags: {
        sex: {
            revealed: boolean;
            value: string;
            label: string;
        };
        age: {
            revealed: boolean;
            value: string;
            label: string;
        };
        orientation: {
            revealed: boolean;
            value: string;
            label: string;
        };
    };
    connected?: boolean | undefined;
}, {
    alive: boolean;
    id: string;
    categories: {
        revealed: boolean;
        value: string;
        label: string;
        key: string;
        imgUrl?: string | undefined;
    }[];
    nickname: string;
    tags: {
        sex: {
            revealed: boolean;
            value: string;
            label: string;
        };
        age: {
            revealed: boolean;
            value: string;
            label: string;
        };
        orientation: {
            revealed: boolean;
            value: string;
            label: string;
        };
    };
    connected?: boolean | undefined;
}>;
export declare const OverlayStateSchema: z.ZodObject<{
    roomId: z.ZodString;
    playerCount: z.ZodNumber;
    top: z.ZodObject<{
        bunker: z.ZodObject<{
            revealed: z.ZodNumber;
            total: z.ZodNumber;
            lines: z.ZodArray<z.ZodString, "many">;
            items: z.ZodOptional<z.ZodArray<z.ZodObject<{
                title: z.ZodString;
                subtitle: z.ZodOptional<z.ZodString>;
                imageId: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }, {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }>, "many">>;
        }, "strip", z.ZodTypeAny, {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        }, {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        }>;
        catastrophe: z.ZodObject<{
            text: z.ZodString;
            title: z.ZodOptional<z.ZodString>;
            imageId: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            text: string;
            title?: string | undefined;
            imageId?: string | undefined;
        }, {
            text: string;
            title?: string | undefined;
            imageId?: string | undefined;
        }>;
        threats: z.ZodObject<{
            revealed: z.ZodNumber;
            total: z.ZodNumber;
            lines: z.ZodArray<z.ZodString, "many">;
            items: z.ZodOptional<z.ZodArray<z.ZodObject<{
                title: z.ZodString;
                subtitle: z.ZodOptional<z.ZodString>;
                imageId: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }, {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }>, "many">>;
        }, "strip", z.ZodTypeAny, {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        }, {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        }>;
    }, "strip", z.ZodTypeAny, {
        bunker: {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        };
        threats: {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        };
        catastrophe: {
            text: string;
            title?: string | undefined;
            imageId?: string | undefined;
        };
    }, {
        bunker: {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        };
        threats: {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        };
        catastrophe: {
            text: string;
            title?: string | undefined;
            imageId?: string | undefined;
        };
    }>;
    players: z.ZodArray<z.ZodObject<{
        id: z.ZodString;
        nickname: z.ZodString;
        connected: z.ZodOptional<z.ZodBoolean>;
        alive: z.ZodBoolean;
        tags: z.ZodObject<{
            sex: z.ZodObject<{
                label: z.ZodString;
                revealed: z.ZodBoolean;
                value: z.ZodString;
            }, "strip", z.ZodTypeAny, {
                revealed: boolean;
                value: string;
                label: string;
            }, {
                revealed: boolean;
                value: string;
                label: string;
            }>;
            age: z.ZodObject<{
                label: z.ZodString;
                revealed: z.ZodBoolean;
                value: z.ZodString;
            }, "strip", z.ZodTypeAny, {
                revealed: boolean;
                value: string;
                label: string;
            }, {
                revealed: boolean;
                value: string;
                label: string;
            }>;
            orientation: z.ZodObject<{
                label: z.ZodString;
                revealed: z.ZodBoolean;
                value: z.ZodString;
            }, "strip", z.ZodTypeAny, {
                revealed: boolean;
                value: string;
                label: string;
            }, {
                revealed: boolean;
                value: string;
                label: string;
            }>;
        }, "strip", z.ZodTypeAny, {
            sex: {
                revealed: boolean;
                value: string;
                label: string;
            };
            age: {
                revealed: boolean;
                value: string;
                label: string;
            };
            orientation: {
                revealed: boolean;
                value: string;
                label: string;
            };
        }, {
            sex: {
                revealed: boolean;
                value: string;
                label: string;
            };
            age: {
                revealed: boolean;
                value: string;
                label: string;
            };
            orientation: {
                revealed: boolean;
                value: string;
                label: string;
            };
        }>;
        categories: z.ZodArray<z.ZodObject<{
            key: z.ZodString;
            label: z.ZodString;
            revealed: z.ZodBoolean;
            value: z.ZodString;
            imgUrl: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            revealed: boolean;
            value: string;
            label: string;
            key: string;
            imgUrl?: string | undefined;
        }, {
            revealed: boolean;
            value: string;
            label: string;
            key: string;
            imgUrl?: string | undefined;
        }>, "many">;
    }, "strip", z.ZodTypeAny, {
        alive: boolean;
        id: string;
        categories: {
            revealed: boolean;
            value: string;
            label: string;
            key: string;
            imgUrl?: string | undefined;
        }[];
        nickname: string;
        tags: {
            sex: {
                revealed: boolean;
                value: string;
                label: string;
            };
            age: {
                revealed: boolean;
                value: string;
                label: string;
            };
            orientation: {
                revealed: boolean;
                value: string;
                label: string;
            };
        };
        connected?: boolean | undefined;
    }, {
        alive: boolean;
        id: string;
        categories: {
            revealed: boolean;
            value: string;
            label: string;
            key: string;
            imgUrl?: string | undefined;
        }[];
        nickname: string;
        tags: {
            sex: {
                revealed: boolean;
                value: string;
                label: string;
            };
            age: {
                revealed: boolean;
                value: string;
                label: string;
            };
            orientation: {
                revealed: boolean;
                value: string;
                label: string;
            };
        };
        connected?: boolean | undefined;
    }>, "many">;
    overrides: z.ZodOptional<z.ZodObject<{
        enabled: z.ZodOptional<z.ZodObject<{
            topBunker: z.ZodOptional<z.ZodBoolean>;
            topCatastrophe: z.ZodOptional<z.ZodBoolean>;
            topThreats: z.ZodOptional<z.ZodBoolean>;
            playerNames: z.ZodOptional<z.ZodBoolean>;
            playerTraits: z.ZodOptional<z.ZodBoolean>;
            playerCategories: z.ZodOptional<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            topBunker?: boolean | undefined;
            topCatastrophe?: boolean | undefined;
            topThreats?: boolean | undefined;
            playerNames?: boolean | undefined;
            playerTraits?: boolean | undefined;
            playerCategories?: boolean | undefined;
        }, {
            topBunker?: boolean | undefined;
            topCatastrophe?: boolean | undefined;
            topThreats?: boolean | undefined;
            playerNames?: boolean | undefined;
            playerTraits?: boolean | undefined;
            playerCategories?: boolean | undefined;
        }>>;
        top: z.ZodOptional<z.ZodObject<{
            bunkerLines: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
            catastropheText: z.ZodOptional<z.ZodString>;
            threatsLines: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
        }, "strip", z.ZodTypeAny, {
            bunkerLines?: string[] | undefined;
            catastropheText?: string | undefined;
            threatsLines?: string[] | undefined;
        }, {
            bunkerLines?: string[] | undefined;
            catastropheText?: string | undefined;
            threatsLines?: string[] | undefined;
        }>>;
        players: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodObject<{
            name: z.ZodOptional<z.ZodString>;
            traits: z.ZodOptional<z.ZodObject<{
                sex: z.ZodOptional<z.ZodString>;
                age: z.ZodOptional<z.ZodString>;
                orient: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                sex?: string | undefined;
                age?: string | undefined;
                orient?: string | undefined;
            }, {
                sex?: string | undefined;
                age?: string | undefined;
                orient?: string | undefined;
            }>>;
            categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
            enabled: z.ZodOptional<z.ZodObject<{
                name: z.ZodOptional<z.ZodBoolean>;
                traits: z.ZodOptional<z.ZodBoolean>;
                categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodBoolean>>;
            }, "strip", z.ZodTypeAny, {
                name?: boolean | undefined;
                traits?: boolean | undefined;
                categories?: Record<string, boolean> | undefined;
            }, {
                name?: boolean | undefined;
                traits?: boolean | undefined;
                categories?: Record<string, boolean> | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            name?: string | undefined;
            traits?: {
                sex?: string | undefined;
                age?: string | undefined;
                orient?: string | undefined;
            } | undefined;
            categories?: Record<string, string> | undefined;
            enabled?: {
                name?: boolean | undefined;
                traits?: boolean | undefined;
                categories?: Record<string, boolean> | undefined;
            } | undefined;
        }, {
            name?: string | undefined;
            traits?: {
                sex?: string | undefined;
                age?: string | undefined;
                orient?: string | undefined;
            } | undefined;
            categories?: Record<string, string> | undefined;
            enabled?: {
                name?: boolean | undefined;
                traits?: boolean | undefined;
                categories?: Record<string, boolean> | undefined;
            } | undefined;
        }>>>;
        extraTexts: z.ZodOptional<z.ZodArray<z.ZodObject<{
            id: z.ZodString;
            text: z.ZodString;
            x: z.ZodNumber;
            y: z.ZodNumber;
            align: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"left">, z.ZodLiteral<"center">, z.ZodLiteral<"right">]>>;
            size: z.ZodOptional<z.ZodNumber>;
            color: z.ZodOptional<z.ZodString>;
            shadow: z.ZodOptional<z.ZodBoolean>;
            visible: z.ZodOptional<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            text: string;
            x: number;
            y: number;
            align?: "left" | "right" | "center" | undefined;
            size?: number | undefined;
            color?: string | undefined;
            shadow?: boolean | undefined;
            visible?: boolean | undefined;
        }, {
            id: string;
            text: string;
            x: number;
            y: number;
            align?: "left" | "right" | "center" | undefined;
            size?: number | undefined;
            color?: string | undefined;
            shadow?: boolean | undefined;
            visible?: boolean | undefined;
        }>, "many">>;
    }, "strip", z.ZodTypeAny, {
        enabled?: {
            topBunker?: boolean | undefined;
            topCatastrophe?: boolean | undefined;
            topThreats?: boolean | undefined;
            playerNames?: boolean | undefined;
            playerTraits?: boolean | undefined;
            playerCategories?: boolean | undefined;
        } | undefined;
        top?: {
            bunkerLines?: string[] | undefined;
            catastropheText?: string | undefined;
            threatsLines?: string[] | undefined;
        } | undefined;
        players?: Record<string, {
            name?: string | undefined;
            traits?: {
                sex?: string | undefined;
                age?: string | undefined;
                orient?: string | undefined;
            } | undefined;
            categories?: Record<string, string> | undefined;
            enabled?: {
                name?: boolean | undefined;
                traits?: boolean | undefined;
                categories?: Record<string, boolean> | undefined;
            } | undefined;
        }> | undefined;
        extraTexts?: {
            id: string;
            text: string;
            x: number;
            y: number;
            align?: "left" | "right" | "center" | undefined;
            size?: number | undefined;
            color?: string | undefined;
            shadow?: boolean | undefined;
            visible?: boolean | undefined;
        }[] | undefined;
    }, {
        enabled?: {
            topBunker?: boolean | undefined;
            topCatastrophe?: boolean | undefined;
            topThreats?: boolean | undefined;
            playerNames?: boolean | undefined;
            playerTraits?: boolean | undefined;
            playerCategories?: boolean | undefined;
        } | undefined;
        top?: {
            bunkerLines?: string[] | undefined;
            catastropheText?: string | undefined;
            threatsLines?: string[] | undefined;
        } | undefined;
        players?: Record<string, {
            name?: string | undefined;
            traits?: {
                sex?: string | undefined;
                age?: string | undefined;
                orient?: string | undefined;
            } | undefined;
            categories?: Record<string, string> | undefined;
            enabled?: {
                name?: boolean | undefined;
                traits?: boolean | undefined;
                categories?: Record<string, boolean> | undefined;
            } | undefined;
        }> | undefined;
        extraTexts?: {
            id: string;
            text: string;
            x: number;
            y: number;
            align?: "left" | "right" | "center" | undefined;
            size?: number | undefined;
            color?: string | undefined;
            shadow?: boolean | undefined;
            visible?: boolean | undefined;
        }[] | undefined;
    }>>;
}, "strip", z.ZodTypeAny, {
    playerCount: number;
    top: {
        bunker: {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        };
        threats: {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        };
        catastrophe: {
            text: string;
            title?: string | undefined;
            imageId?: string | undefined;
        };
    };
    players: {
        alive: boolean;
        id: string;
        categories: {
            revealed: boolean;
            value: string;
            label: string;
            key: string;
            imgUrl?: string | undefined;
        }[];
        nickname: string;
        tags: {
            sex: {
                revealed: boolean;
                value: string;
                label: string;
            };
            age: {
                revealed: boolean;
                value: string;
                label: string;
            };
            orientation: {
                revealed: boolean;
                value: string;
                label: string;
            };
        };
        connected?: boolean | undefined;
    }[];
    roomId: string;
    overrides?: {
        enabled?: {
            topBunker?: boolean | undefined;
            topCatastrophe?: boolean | undefined;
            topThreats?: boolean | undefined;
            playerNames?: boolean | undefined;
            playerTraits?: boolean | undefined;
            playerCategories?: boolean | undefined;
        } | undefined;
        top?: {
            bunkerLines?: string[] | undefined;
            catastropheText?: string | undefined;
            threatsLines?: string[] | undefined;
        } | undefined;
        players?: Record<string, {
            name?: string | undefined;
            traits?: {
                sex?: string | undefined;
                age?: string | undefined;
                orient?: string | undefined;
            } | undefined;
            categories?: Record<string, string> | undefined;
            enabled?: {
                name?: boolean | undefined;
                traits?: boolean | undefined;
                categories?: Record<string, boolean> | undefined;
            } | undefined;
        }> | undefined;
        extraTexts?: {
            id: string;
            text: string;
            x: number;
            y: number;
            align?: "left" | "right" | "center" | undefined;
            size?: number | undefined;
            color?: string | undefined;
            shadow?: boolean | undefined;
            visible?: boolean | undefined;
        }[] | undefined;
    } | undefined;
}, {
    playerCount: number;
    top: {
        bunker: {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        };
        threats: {
            revealed: number;
            total: number;
            lines: string[];
            items?: {
                title: string;
                imageId?: string | undefined;
                subtitle?: string | undefined;
            }[] | undefined;
        };
        catastrophe: {
            text: string;
            title?: string | undefined;
            imageId?: string | undefined;
        };
    };
    players: {
        alive: boolean;
        id: string;
        categories: {
            revealed: boolean;
            value: string;
            label: string;
            key: string;
            imgUrl?: string | undefined;
        }[];
        nickname: string;
        tags: {
            sex: {
                revealed: boolean;
                value: string;
                label: string;
            };
            age: {
                revealed: boolean;
                value: string;
                label: string;
            };
            orientation: {
                revealed: boolean;
                value: string;
                label: string;
            };
        };
        connected?: boolean | undefined;
    }[];
    roomId: string;
    overrides?: {
        enabled?: {
            topBunker?: boolean | undefined;
            topCatastrophe?: boolean | undefined;
            topThreats?: boolean | undefined;
            playerNames?: boolean | undefined;
            playerTraits?: boolean | undefined;
            playerCategories?: boolean | undefined;
        } | undefined;
        top?: {
            bunkerLines?: string[] | undefined;
            catastropheText?: string | undefined;
            threatsLines?: string[] | undefined;
        } | undefined;
        players?: Record<string, {
            name?: string | undefined;
            traits?: {
                sex?: string | undefined;
                age?: string | undefined;
                orient?: string | undefined;
            } | undefined;
            categories?: Record<string, string> | undefined;
            enabled?: {
                name?: boolean | undefined;
                traits?: boolean | undefined;
                categories?: Record<string, boolean> | undefined;
            } | undefined;
        }> | undefined;
        extraTexts?: {
            id: string;
            text: string;
            x: number;
            y: number;
            align?: "left" | "right" | "center" | undefined;
            size?: number | undefined;
            color?: string | undefined;
            shadow?: boolean | undefined;
            visible?: boolean | undefined;
        }[] | undefined;
    } | undefined;
}>;
export declare const GameSettingsSchema: z.ZodObject<{
    enableRevealDiscussionTimer: z.ZodBoolean;
    revealDiscussionSeconds: z.ZodNumber;
    enablePreVoteDiscussionTimer: z.ZodBoolean;
    preVoteDiscussionSeconds: z.ZodNumber;
    enablePostVoteDiscussionTimer: z.ZodBoolean;
    postVoteDiscussionSeconds: z.ZodNumber;
    enablePresenterMode: z.ZodBoolean;
    continuePermission: z.ZodUnion<[z.ZodLiteral<"host_only">, z.ZodLiteral<"revealer_only">, z.ZodLiteral<"anyone">]>;
    revealTimeoutAction: z.ZodUnion<[z.ZodLiteral<"random_card">, z.ZodLiteral<"skip_player">]>;
    revealsBeforeVoting: z.ZodNumber;
    specialUsage: z.ZodUnion<[z.ZodLiteral<"anytime">, z.ZodLiteral<"only_during_voting">]>;
    maxPlayers: z.ZodNumber;
    finalThreatReveal: z.ZodUnion<[z.ZodLiteral<"host">, z.ZodLiteral<"anyone">]>;
}, "strip", z.ZodTypeAny, {
    enableRevealDiscussionTimer: boolean;
    revealDiscussionSeconds: number;
    enablePreVoteDiscussionTimer: boolean;
    preVoteDiscussionSeconds: number;
    enablePostVoteDiscussionTimer: boolean;
    postVoteDiscussionSeconds: number;
    enablePresenterMode: boolean;
    continuePermission: "host_only" | "revealer_only" | "anyone";
    revealTimeoutAction: "random_card" | "skip_player";
    revealsBeforeVoting: number;
    specialUsage: "anytime" | "only_during_voting";
    maxPlayers: number;
    finalThreatReveal: "anyone" | "host";
}, {
    enableRevealDiscussionTimer: boolean;
    revealDiscussionSeconds: number;
    enablePreVoteDiscussionTimer: boolean;
    preVoteDiscussionSeconds: number;
    enablePostVoteDiscussionTimer: boolean;
    postVoteDiscussionSeconds: number;
    enablePresenterMode: boolean;
    continuePermission: "host_only" | "revealer_only" | "anyone";
    revealTimeoutAction: "random_card" | "skip_player";
    revealsBeforeVoting: number;
    specialUsage: "anytime" | "only_during_voting";
    maxPlayers: number;
    finalThreatReveal: "anyone" | "host";
}>;
export declare const CardRefSchema: z.ZodObject<{
    id: z.ZodString;
    deck: z.ZodString;
    instanceId: z.ZodOptional<z.ZodString>;
    labelShort: z.ZodOptional<z.ZodString>;
    secret: z.ZodOptional<z.ZodBoolean>;
    missing: z.ZodOptional<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    id: string;
    deck: string;
    instanceId?: string | undefined;
    labelShort?: string | undefined;
    secret?: boolean | undefined;
    missing?: boolean | undefined;
}, {
    id: string;
    deck: string;
    instanceId?: string | undefined;
    labelShort?: string | undefined;
    secret?: boolean | undefined;
    missing?: boolean | undefined;
}>;
export declare const CardInHandSchema: z.ZodObject<{
    id: z.ZodString;
    deck: z.ZodString;
    instanceId: z.ZodOptional<z.ZodString>;
    labelShort: z.ZodOptional<z.ZodString>;
    secret: z.ZodOptional<z.ZodBoolean>;
    missing: z.ZodOptional<z.ZodBoolean>;
} & {
    revealed: z.ZodBoolean;
}, "strip", z.ZodTypeAny, {
    revealed: boolean;
    id: string;
    deck: string;
    instanceId?: string | undefined;
    labelShort?: string | undefined;
    secret?: boolean | undefined;
    missing?: boolean | undefined;
}, {
    revealed: boolean;
    id: string;
    deck: string;
    instanceId?: string | undefined;
    labelShort?: string | undefined;
    secret?: boolean | undefined;
    missing?: boolean | undefined;
}>;
export declare const SpecialConditionEffectSchema: z.ZodObject<{
    type: z.ZodString;
    params: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodAny>>;
}, "strip", z.ZodTypeAny, {
    type: string;
    params?: Record<string, any> | undefined;
}, {
    type: string;
    params?: Record<string, any> | undefined;
}>;
export declare const SpecialConditionInstanceSchema: z.ZodObject<{
    instanceId: z.ZodString;
    id: z.ZodString;
    title: z.ZodString;
    text: z.ZodString;
    trigger: z.ZodUnion<[z.ZodLiteral<"active">, z.ZodLiteral<"onVote">, z.ZodLiteral<"onOwnerEliminated">, z.ZodLiteral<"onRevealOrActive">, z.ZodLiteral<"secret_onEliminate">]>;
    effect: z.ZodObject<{
        type: z.ZodString;
        params: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodAny>>;
    }, "strip", z.ZodTypeAny, {
        type: string;
        params?: Record<string, any> | undefined;
    }, {
        type: string;
        params?: Record<string, any> | undefined;
    }>;
    implemented: z.ZodBoolean;
    revealedPublic: z.ZodBoolean;
    used: z.ZodBoolean;
    imgUrl: z.ZodOptional<z.ZodString>;
    needsChoice: z.ZodOptional<z.ZodBoolean>;
    choiceKind: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"player">, z.ZodLiteral<"neighbor">, z.ZodLiteral<"category">, z.ZodLiteral<"none">]>>;
    allowSelfTarget: z.ZodOptional<z.ZodBoolean>;
    targetScope: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"neighbors">, z.ZodLiteral<"any_alive">, z.ZodLiteral<"self">, z.ZodLiteral<"any_including_self">]>>;
}, "strip", z.ZodTypeAny, {
    id: string;
    title: string;
    text: string;
    instanceId: string;
    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
    effect: {
        type: string;
        params?: Record<string, any> | undefined;
    };
    implemented: boolean;
    revealedPublic: boolean;
    used: boolean;
    imgUrl?: string | undefined;
    needsChoice?: boolean | undefined;
    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
    allowSelfTarget?: boolean | undefined;
    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
}, {
    id: string;
    title: string;
    text: string;
    instanceId: string;
    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
    effect: {
        type: string;
        params?: Record<string, any> | undefined;
    };
    implemented: boolean;
    revealedPublic: boolean;
    used: boolean;
    imgUrl?: string | undefined;
    needsChoice?: boolean | undefined;
    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
    allowSelfTarget?: boolean | undefined;
    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
}>;
export declare const PublicCategoryCardSchema: z.ZodObject<{
    labelShort: z.ZodString;
    imgUrl: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    labelShort: string;
    imgUrl?: string | undefined;
}, {
    labelShort: string;
    imgUrl?: string | undefined;
}>;
export declare const YouCategoryCardSchema: z.ZodObject<{
    instanceId: z.ZodString;
    labelShort: z.ZodString;
    revealed: z.ZodBoolean;
}, "strip", z.ZodTypeAny, {
    revealed: boolean;
    instanceId: string;
    labelShort: string;
}, {
    revealed: boolean;
    instanceId: string;
    labelShort: string;
}>;
export declare const PublicCategorySlotSchema: z.ZodObject<{
    category: z.ZodString;
    status: z.ZodUnion<[z.ZodLiteral<"hidden">, z.ZodLiteral<"revealed">]>;
    cards: z.ZodArray<z.ZodObject<{
        labelShort: z.ZodString;
        imgUrl: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        labelShort: string;
        imgUrl?: string | undefined;
    }, {
        labelShort: string;
        imgUrl?: string | undefined;
    }>, "many">;
}, "strip", z.ZodTypeAny, {
    category: string;
    status: "hidden" | "revealed";
    cards: {
        labelShort: string;
        imgUrl?: string | undefined;
    }[];
}, {
    category: string;
    status: "hidden" | "revealed";
    cards: {
        labelShort: string;
        imgUrl?: string | undefined;
    }[];
}>;
export declare const YouCategorySlotSchema: z.ZodObject<{
    category: z.ZodString;
    cards: z.ZodArray<z.ZodObject<{
        instanceId: z.ZodString;
        labelShort: z.ZodString;
        revealed: z.ZodBoolean;
    }, "strip", z.ZodTypeAny, {
        revealed: boolean;
        instanceId: string;
        labelShort: string;
    }, {
        revealed: boolean;
        instanceId: string;
        labelShort: string;
    }>, "many">;
}, "strip", z.ZodTypeAny, {
    category: string;
    cards: {
        revealed: boolean;
        instanceId: string;
        labelShort: string;
    }[];
}, {
    category: string;
    cards: {
        revealed: boolean;
        instanceId: string;
        labelShort: string;
    }[];
}>;
export declare const PublicPlayerViewSchema: z.ZodObject<{
    playerId: z.ZodString;
    name: z.ZodString;
    status: z.ZodUnion<[z.ZodLiteral<"alive">, z.ZodLiteral<"eliminated">, z.ZodLiteral<"left_bunker">]>;
    connected: z.ZodBoolean;
    disconnectedAt: z.ZodOptional<z.ZodNumber>;
    totalAbsentMs: z.ZodOptional<z.ZodNumber>;
    currentOfflineMs: z.ZodOptional<z.ZodNumber>;
    kickRemainingMs: z.ZodOptional<z.ZodNumber>;
    leftBunker: z.ZodOptional<z.ZodBoolean>;
    revealedCards: z.ZodArray<z.ZodObject<{
        id: z.ZodString;
        deck: z.ZodString;
        instanceId: z.ZodOptional<z.ZodString>;
        labelShort: z.ZodOptional<z.ZodString>;
        secret: z.ZodOptional<z.ZodBoolean>;
        missing: z.ZodOptional<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        id: string;
        deck: string;
        instanceId?: string | undefined;
        labelShort?: string | undefined;
        secret?: boolean | undefined;
        missing?: boolean | undefined;
    }, {
        id: string;
        deck: string;
        instanceId?: string | undefined;
        labelShort?: string | undefined;
        secret?: boolean | undefined;
        missing?: boolean | undefined;
    }>, "many">;
    revealedCount: z.ZodNumber;
    totalCards: z.ZodNumber;
    specialRevealed: z.ZodBoolean;
    categories: z.ZodArray<z.ZodObject<{
        category: z.ZodString;
        status: z.ZodUnion<[z.ZodLiteral<"hidden">, z.ZodLiteral<"revealed">]>;
        cards: z.ZodArray<z.ZodObject<{
            labelShort: z.ZodString;
            imgUrl: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            labelShort: string;
            imgUrl?: string | undefined;
        }, {
            labelShort: string;
            imgUrl?: string | undefined;
        }>, "many">;
    }, "strip", z.ZodTypeAny, {
        category: string;
        status: "hidden" | "revealed";
        cards: {
            labelShort: string;
            imgUrl?: string | undefined;
        }[];
    }, {
        category: string;
        status: "hidden" | "revealed";
        cards: {
            labelShort: string;
            imgUrl?: string | undefined;
        }[];
    }>, "many">;
}, "strip", z.ZodTypeAny, {
    playerId: string;
    name: string;
    connected: boolean;
    status: "alive" | "eliminated" | "left_bunker";
    categories: {
        category: string;
        status: "hidden" | "revealed";
        cards: {
            labelShort: string;
            imgUrl?: string | undefined;
        }[];
    }[];
    revealedCards: {
        id: string;
        deck: string;
        instanceId?: string | undefined;
        labelShort?: string | undefined;
        secret?: boolean | undefined;
        missing?: boolean | undefined;
    }[];
    revealedCount: number;
    totalCards: number;
    specialRevealed: boolean;
    disconnectedAt?: number | undefined;
    totalAbsentMs?: number | undefined;
    currentOfflineMs?: number | undefined;
    kickRemainingMs?: number | undefined;
    leftBunker?: boolean | undefined;
}, {
    playerId: string;
    name: string;
    connected: boolean;
    status: "alive" | "eliminated" | "left_bunker";
    categories: {
        category: string;
        status: "hidden" | "revealed";
        cards: {
            labelShort: string;
            imgUrl?: string | undefined;
        }[];
    }[];
    revealedCards: {
        id: string;
        deck: string;
        instanceId?: string | undefined;
        labelShort?: string | undefined;
        secret?: boolean | undefined;
        missing?: boolean | undefined;
    }[];
    revealedCount: number;
    totalCards: number;
    specialRevealed: boolean;
    disconnectedAt?: number | undefined;
    totalAbsentMs?: number | undefined;
    currentOfflineMs?: number | undefined;
    kickRemainingMs?: number | undefined;
    leftBunker?: boolean | undefined;
}>;
export declare const RoomStateSchema: z.ZodObject<{
    roomCode: z.ZodString;
    players: z.ZodArray<z.ZodObject<{
        playerId: z.ZodString;
        name: z.ZodString;
        connected: z.ZodBoolean;
        disconnectedAt: z.ZodOptional<z.ZodNumber>;
        totalAbsentMs: z.ZodOptional<z.ZodNumber>;
        currentOfflineMs: z.ZodOptional<z.ZodNumber>;
        kickRemainingMs: z.ZodOptional<z.ZodNumber>;
        leftBunker: z.ZodOptional<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        playerId: string;
        name: string;
        connected: boolean;
        disconnectedAt?: number | undefined;
        totalAbsentMs?: number | undefined;
        currentOfflineMs?: number | undefined;
        kickRemainingMs?: number | undefined;
        leftBunker?: boolean | undefined;
    }, {
        playerId: string;
        name: string;
        connected: boolean;
        disconnectedAt?: number | undefined;
        totalAbsentMs?: number | undefined;
        currentOfflineMs?: number | undefined;
        kickRemainingMs?: number | undefined;
        leftBunker?: boolean | undefined;
    }>, "many">;
    hostId: z.ZodString;
    controlId: z.ZodString;
    phase: z.ZodUnion<[z.ZodLiteral<"lobby">, z.ZodLiteral<"game">]>;
    scenarioMeta: z.ZodObject<{
        id: z.ZodString;
        name: z.ZodString;
        description: z.ZodOptional<z.ZodString>;
        devOnly: z.ZodOptional<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        name: string;
        id: string;
        description?: string | undefined;
        devOnly?: boolean | undefined;
    }, {
        name: string;
        id: string;
        description?: string | undefined;
        devOnly?: boolean | undefined;
    }>;
    settings: z.ZodObject<{
        enableRevealDiscussionTimer: z.ZodBoolean;
        revealDiscussionSeconds: z.ZodNumber;
        enablePreVoteDiscussionTimer: z.ZodBoolean;
        preVoteDiscussionSeconds: z.ZodNumber;
        enablePostVoteDiscussionTimer: z.ZodBoolean;
        postVoteDiscussionSeconds: z.ZodNumber;
        enablePresenterMode: z.ZodBoolean;
        continuePermission: z.ZodUnion<[z.ZodLiteral<"host_only">, z.ZodLiteral<"revealer_only">, z.ZodLiteral<"anyone">]>;
        revealTimeoutAction: z.ZodUnion<[z.ZodLiteral<"random_card">, z.ZodLiteral<"skip_player">]>;
        revealsBeforeVoting: z.ZodNumber;
        specialUsage: z.ZodUnion<[z.ZodLiteral<"anytime">, z.ZodLiteral<"only_during_voting">]>;
        maxPlayers: z.ZodNumber;
        finalThreatReveal: z.ZodUnion<[z.ZodLiteral<"host">, z.ZodLiteral<"anyone">]>;
    }, "strip", z.ZodTypeAny, {
        enableRevealDiscussionTimer: boolean;
        revealDiscussionSeconds: number;
        enablePreVoteDiscussionTimer: boolean;
        preVoteDiscussionSeconds: number;
        enablePostVoteDiscussionTimer: boolean;
        postVoteDiscussionSeconds: number;
        enablePresenterMode: boolean;
        continuePermission: "host_only" | "revealer_only" | "anyone";
        revealTimeoutAction: "random_card" | "skip_player";
        revealsBeforeVoting: number;
        specialUsage: "anytime" | "only_during_voting";
        maxPlayers: number;
        finalThreatReveal: "anyone" | "host";
    }, {
        enableRevealDiscussionTimer: boolean;
        revealDiscussionSeconds: number;
        enablePreVoteDiscussionTimer: boolean;
        preVoteDiscussionSeconds: number;
        enablePostVoteDiscussionTimer: boolean;
        postVoteDiscussionSeconds: number;
        enablePresenterMode: boolean;
        continuePermission: "host_only" | "revealer_only" | "anyone";
        revealTimeoutAction: "random_card" | "skip_player";
        revealsBeforeVoting: number;
        specialUsage: "anytime" | "only_during_voting";
        maxPlayers: number;
        finalThreatReveal: "anyone" | "host";
    }>;
    ruleset: z.ZodObject<{
        playerCount: z.ZodNumber;
        votesPerRound: z.ZodArray<z.ZodNumber, "many">;
        totalExiles: z.ZodNumber;
        bunkerSeats: z.ZodNumber;
        rulesetMode: z.ZodUnion<[z.ZodLiteral<"auto">, z.ZodLiteral<"preset">, z.ZodLiteral<"manual">]>;
        manualConfig: z.ZodOptional<z.ZodObject<{
            bunkerSlots: z.ZodNumber;
            votesByRound: z.ZodArray<z.ZodNumber, "many">;
            targetReveals: z.ZodDefault<z.ZodNumber>;
            seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        }, {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        playerCount: number;
        votesPerRound: number[];
        totalExiles: number;
        bunkerSeats: number;
        rulesetMode: "auto" | "preset" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
    }, {
        playerCount: number;
        votesPerRound: number[];
        totalExiles: number;
        bunkerSeats: number;
        rulesetMode: "auto" | "preset" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
    }>;
    rulesOverriddenByHost: z.ZodBoolean;
    rulesPresetCount: z.ZodOptional<z.ZodNumber>;
    world: z.ZodOptional<z.ZodObject<{
        disaster: z.ZodObject<{
            kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
            id: z.ZodString;
            title: z.ZodString;
            description: z.ZodString;
            text: z.ZodOptional<z.ZodString>;
            imageId: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        }, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        }>;
        bunker: z.ZodArray<z.ZodObject<{
            kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
            id: z.ZodString;
            title: z.ZodString;
            description: z.ZodString;
            text: z.ZodOptional<z.ZodString>;
            imageId: z.ZodOptional<z.ZodString>;
        } & {
            isRevealed: z.ZodBoolean;
            revealedAtRound: z.ZodOptional<z.ZodNumber>;
            revealedBy: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }>, "many">;
        threats: z.ZodArray<z.ZodObject<{
            kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
            id: z.ZodString;
            title: z.ZodString;
            description: z.ZodString;
            text: z.ZodOptional<z.ZodString>;
            imageId: z.ZodOptional<z.ZodString>;
        } & {
            isRevealed: z.ZodBoolean;
            revealedAtRound: z.ZodOptional<z.ZodNumber>;
            revealedBy: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }>, "many">;
        counts: z.ZodObject<{
            bunker: z.ZodNumber;
            threats: z.ZodNumber;
        }, "strip", z.ZodTypeAny, {
            bunker: number;
            threats: number;
        }, {
            bunker: number;
            threats: number;
        }>;
    }, "strip", z.ZodTypeAny, {
        bunker: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        disaster: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        };
        threats: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        counts: {
            bunker: number;
            threats: number;
        };
    }, {
        bunker: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        disaster: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        };
        threats: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        counts: {
            bunker: number;
            threats: number;
        };
    }>>;
    isDev: z.ZodOptional<z.ZodBoolean>;
}, "strip", z.ZodTypeAny, {
    players: {
        playerId: string;
        name: string;
        connected: boolean;
        disconnectedAt?: number | undefined;
        totalAbsentMs?: number | undefined;
        currentOfflineMs?: number | undefined;
        kickRemainingMs?: number | undefined;
        leftBunker?: boolean | undefined;
    }[];
    roomCode: string;
    hostId: string;
    controlId: string;
    phase: "lobby" | "game";
    scenarioMeta: {
        name: string;
        id: string;
        description?: string | undefined;
        devOnly?: boolean | undefined;
    };
    settings: {
        enableRevealDiscussionTimer: boolean;
        revealDiscussionSeconds: number;
        enablePreVoteDiscussionTimer: boolean;
        preVoteDiscussionSeconds: number;
        enablePostVoteDiscussionTimer: boolean;
        postVoteDiscussionSeconds: number;
        enablePresenterMode: boolean;
        continuePermission: "host_only" | "revealer_only" | "anyone";
        revealTimeoutAction: "random_card" | "skip_player";
        revealsBeforeVoting: number;
        specialUsage: "anytime" | "only_during_voting";
        maxPlayers: number;
        finalThreatReveal: "anyone" | "host";
    };
    ruleset: {
        playerCount: number;
        votesPerRound: number[];
        totalExiles: number;
        bunkerSeats: number;
        rulesetMode: "auto" | "preset" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
    };
    rulesOverriddenByHost: boolean;
    rulesPresetCount?: number | undefined;
    world?: {
        bunker: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        disaster: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        };
        threats: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        counts: {
            bunker: number;
            threats: number;
        };
    } | undefined;
    isDev?: boolean | undefined;
}, {
    players: {
        playerId: string;
        name: string;
        connected: boolean;
        disconnectedAt?: number | undefined;
        totalAbsentMs?: number | undefined;
        currentOfflineMs?: number | undefined;
        kickRemainingMs?: number | undefined;
        leftBunker?: boolean | undefined;
    }[];
    roomCode: string;
    hostId: string;
    controlId: string;
    phase: "lobby" | "game";
    scenarioMeta: {
        name: string;
        id: string;
        description?: string | undefined;
        devOnly?: boolean | undefined;
    };
    settings: {
        enableRevealDiscussionTimer: boolean;
        revealDiscussionSeconds: number;
        enablePreVoteDiscussionTimer: boolean;
        preVoteDiscussionSeconds: number;
        enablePostVoteDiscussionTimer: boolean;
        postVoteDiscussionSeconds: number;
        enablePresenterMode: boolean;
        continuePermission: "host_only" | "revealer_only" | "anyone";
        revealTimeoutAction: "random_card" | "skip_player";
        revealsBeforeVoting: number;
        specialUsage: "anytime" | "only_during_voting";
        maxPlayers: number;
        finalThreatReveal: "anyone" | "host";
    };
    ruleset: {
        playerCount: number;
        votesPerRound: number[];
        totalExiles: number;
        bunkerSeats: number;
        rulesetMode: "auto" | "preset" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
    };
    rulesOverriddenByHost: boolean;
    rulesPresetCount?: number | undefined;
    world?: {
        bunker: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        disaster: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        };
        threats: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        counts: {
            bunker: number;
            threats: number;
        };
    } | undefined;
    isDev?: boolean | undefined;
}>;
export declare const VotingViewSchema: z.ZodObject<{
    hasVoted: z.ZodBoolean;
}, "strip", z.ZodTypeAny, {
    hasVoted: boolean;
}, {
    hasVoted: boolean;
}>;
export declare const VotePhaseSchema: z.ZodUnion<[z.ZodLiteral<"voting">, z.ZodLiteral<"voteSpecialWindow">, z.ZodLiteral<"voteResolve">]>;
export declare const VotingProgressSchema: z.ZodObject<{
    voted: z.ZodNumber;
    total: z.ZodNumber;
}, "strip", z.ZodTypeAny, {
    voted: number;
    total: number;
}, {
    voted: number;
    total: number;
}>;
export declare const ThreatModifierViewSchema: z.ZodObject<{
    delta: z.ZodNumber;
    reasons: z.ZodArray<z.ZodString, "many">;
    baseCount: z.ZodNumber;
    finalCount: z.ZodNumber;
}, "strip", z.ZodTypeAny, {
    delta: number;
    reasons: string[];
    baseCount: number;
    finalCount: number;
}, {
    delta: number;
    reasons: string[];
    baseCount: number;
    finalCount: number;
}>;
export declare const VotePublicSchema: z.ZodObject<{
    voterId: z.ZodString;
    voterName: z.ZodString;
    targetId: z.ZodOptional<z.ZodString>;
    targetName: z.ZodOptional<z.ZodString>;
    status: z.ZodUnion<[z.ZodLiteral<"voted">, z.ZodLiteral<"not_voted">, z.ZodLiteral<"invalid">]>;
    reason: z.ZodOptional<z.ZodString>;
    submittedAt: z.ZodOptional<z.ZodNumber>;
}, "strip", z.ZodTypeAny, {
    status: "voted" | "not_voted" | "invalid";
    voterId: string;
    voterName: string;
    targetId?: string | undefined;
    targetName?: string | undefined;
    reason?: string | undefined;
    submittedAt?: number | undefined;
}, {
    status: "voted" | "not_voted" | "invalid";
    voterId: string;
    voterName: string;
    targetId?: string | undefined;
    targetName?: string | undefined;
    reason?: string | undefined;
    submittedAt?: number | undefined;
}>;
export declare const GameEventSchema: z.ZodObject<{
    id: z.ZodString;
    kind: z.ZodUnion<[z.ZodLiteral<"roundStart">, z.ZodLiteral<"votingStart">, z.ZodLiteral<"elimination">, z.ZodLiteral<"gameEnd">, z.ZodLiteral<"info">, z.ZodLiteral<"playerDisconnected">, z.ZodLiteral<"playerReconnected">, z.ZodLiteral<"playerLeftBunker">]>;
    message: z.ZodString;
    createdAt: z.ZodNumber;
}, "strip", z.ZodTypeAny, {
    message: string;
    id: string;
    kind: "roundStart" | "votingStart" | "elimination" | "gameEnd" | "info" | "playerDisconnected" | "playerReconnected" | "playerLeftBunker";
    createdAt: number;
}, {
    message: string;
    id: string;
    kind: "roundStart" | "votingStart" | "elimination" | "gameEnd" | "info" | "playerDisconnected" | "playerReconnected" | "playerLeftBunker";
    createdAt: number;
}>;
export declare const GameViewSchema: z.ZodObject<{
    phase: z.ZodUnion<[z.ZodLiteral<"reveal">, z.ZodLiteral<"reveal_discussion">, z.ZodLiteral<"voting">, z.ZodLiteral<"resolution">, z.ZodLiteral<"ended">]>;
    round: z.ZodNumber;
    categoryOrder: z.ZodArray<z.ZodString, "many">;
    lastStageText: z.ZodOptional<z.ZodString>;
    ruleset: z.ZodObject<{
        playerCount: z.ZodNumber;
        votesPerRound: z.ZodArray<z.ZodNumber, "many">;
        totalExiles: z.ZodNumber;
        bunkerSeats: z.ZodNumber;
        rulesetMode: z.ZodUnion<[z.ZodLiteral<"auto">, z.ZodLiteral<"preset">, z.ZodLiteral<"manual">]>;
        manualConfig: z.ZodOptional<z.ZodObject<{
            bunkerSlots: z.ZodNumber;
            votesByRound: z.ZodArray<z.ZodNumber, "many">;
            targetReveals: z.ZodDefault<z.ZodNumber>;
            seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        }, {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        playerCount: number;
        votesPerRound: number[];
        totalExiles: number;
        bunkerSeats: number;
        rulesetMode: "auto" | "preset" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
    }, {
        playerCount: number;
        votesPerRound: number[];
        totalExiles: number;
        bunkerSeats: number;
        rulesetMode: "auto" | "preset" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
    }>;
    world: z.ZodOptional<z.ZodObject<{
        disaster: z.ZodObject<{
            kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
            id: z.ZodString;
            title: z.ZodString;
            description: z.ZodString;
            text: z.ZodOptional<z.ZodString>;
            imageId: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        }, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        }>;
        bunker: z.ZodArray<z.ZodObject<{
            kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
            id: z.ZodString;
            title: z.ZodString;
            description: z.ZodString;
            text: z.ZodOptional<z.ZodString>;
            imageId: z.ZodOptional<z.ZodString>;
        } & {
            isRevealed: z.ZodBoolean;
            revealedAtRound: z.ZodOptional<z.ZodNumber>;
            revealedBy: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }>, "many">;
        threats: z.ZodArray<z.ZodObject<{
            kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
            id: z.ZodString;
            title: z.ZodString;
            description: z.ZodString;
            text: z.ZodOptional<z.ZodString>;
            imageId: z.ZodOptional<z.ZodString>;
        } & {
            isRevealed: z.ZodBoolean;
            revealedAtRound: z.ZodOptional<z.ZodNumber>;
            revealedBy: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }, {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }>, "many">;
        counts: z.ZodObject<{
            bunker: z.ZodNumber;
            threats: z.ZodNumber;
        }, "strip", z.ZodTypeAny, {
            bunker: number;
            threats: number;
        }, {
            bunker: number;
            threats: number;
        }>;
    }, "strip", z.ZodTypeAny, {
        bunker: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        disaster: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        };
        threats: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        counts: {
            bunker: number;
            threats: number;
        };
    }, {
        bunker: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        disaster: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        };
        threats: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        counts: {
            bunker: number;
            threats: number;
        };
    }>>;
    worldEvent: z.ZodOptional<z.ZodObject<{
        type: z.ZodLiteral<"bunker_revealed">;
        index: z.ZodNumber;
        round: z.ZodNumber;
    }, "strip", z.ZodTypeAny, {
        type: "bunker_revealed";
        index: number;
        round: number;
    }, {
        type: "bunker_revealed";
        index: number;
        round: number;
    }>>;
    postGame: z.ZodOptional<z.ZodObject<{
        isActive: z.ZodBoolean;
        enteredAt: z.ZodNumber;
        outcome: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"survived">, z.ZodLiteral<"failed">]>>;
        decidedBy: z.ZodOptional<z.ZodString>;
        decidedAt: z.ZodOptional<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        isActive: boolean;
        enteredAt: number;
        outcome?: "survived" | "failed" | undefined;
        decidedBy?: string | undefined;
        decidedAt?: number | undefined;
    }, {
        isActive: boolean;
        enteredAt: number;
        outcome?: "survived" | "failed" | undefined;
        decidedBy?: string | undefined;
        decidedAt?: number | undefined;
    }>>;
    you: z.ZodObject<{
        playerId: z.ZodString;
        name: z.ZodString;
        hand: z.ZodArray<z.ZodObject<{
            id: z.ZodString;
            deck: z.ZodString;
            instanceId: z.ZodOptional<z.ZodString>;
            labelShort: z.ZodOptional<z.ZodString>;
            secret: z.ZodOptional<z.ZodBoolean>;
            missing: z.ZodOptional<z.ZodBoolean>;
        } & {
            revealed: z.ZodBoolean;
        }, "strip", z.ZodTypeAny, {
            revealed: boolean;
            id: string;
            deck: string;
            instanceId?: string | undefined;
            labelShort?: string | undefined;
            secret?: boolean | undefined;
            missing?: boolean | undefined;
        }, {
            revealed: boolean;
            id: string;
            deck: string;
            instanceId?: string | undefined;
            labelShort?: string | undefined;
            secret?: boolean | undefined;
            missing?: boolean | undefined;
        }>, "many">;
        categories: z.ZodArray<z.ZodObject<{
            category: z.ZodString;
            cards: z.ZodArray<z.ZodObject<{
                instanceId: z.ZodString;
                labelShort: z.ZodString;
                revealed: z.ZodBoolean;
            }, "strip", z.ZodTypeAny, {
                revealed: boolean;
                instanceId: string;
                labelShort: string;
            }, {
                revealed: boolean;
                instanceId: string;
                labelShort: string;
            }>, "many">;
        }, "strip", z.ZodTypeAny, {
            category: string;
            cards: {
                revealed: boolean;
                instanceId: string;
                labelShort: string;
            }[];
        }, {
            category: string;
            cards: {
                revealed: boolean;
                instanceId: string;
                labelShort: string;
            }[];
        }>, "many">;
        specialConditions: z.ZodArray<z.ZodObject<{
            instanceId: z.ZodString;
            id: z.ZodString;
            title: z.ZodString;
            text: z.ZodString;
            trigger: z.ZodUnion<[z.ZodLiteral<"active">, z.ZodLiteral<"onVote">, z.ZodLiteral<"onOwnerEliminated">, z.ZodLiteral<"onRevealOrActive">, z.ZodLiteral<"secret_onEliminate">]>;
            effect: z.ZodObject<{
                type: z.ZodString;
                params: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodAny>>;
            }, "strip", z.ZodTypeAny, {
                type: string;
                params?: Record<string, any> | undefined;
            }, {
                type: string;
                params?: Record<string, any> | undefined;
            }>;
            implemented: z.ZodBoolean;
            revealedPublic: z.ZodBoolean;
            used: z.ZodBoolean;
            imgUrl: z.ZodOptional<z.ZodString>;
            needsChoice: z.ZodOptional<z.ZodBoolean>;
            choiceKind: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"player">, z.ZodLiteral<"neighbor">, z.ZodLiteral<"category">, z.ZodLiteral<"none">]>>;
            allowSelfTarget: z.ZodOptional<z.ZodBoolean>;
            targetScope: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"neighbors">, z.ZodLiteral<"any_alive">, z.ZodLiteral<"self">, z.ZodLiteral<"any_including_self">]>>;
        }, "strip", z.ZodTypeAny, {
            id: string;
            title: string;
            text: string;
            instanceId: string;
            trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
            effect: {
                type: string;
                params?: Record<string, any> | undefined;
            };
            implemented: boolean;
            revealedPublic: boolean;
            used: boolean;
            imgUrl?: string | undefined;
            needsChoice?: boolean | undefined;
            choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
            allowSelfTarget?: boolean | undefined;
            targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
        }, {
            id: string;
            title: string;
            text: string;
            instanceId: string;
            trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
            effect: {
                type: string;
                params?: Record<string, any> | undefined;
            };
            implemented: boolean;
            revealedPublic: boolean;
            used: boolean;
            imgUrl?: string | undefined;
            needsChoice?: boolean | undefined;
            choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
            allowSelfTarget?: boolean | undefined;
            targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
        }>, "many">;
    }, "strip", z.ZodTypeAny, {
        playerId: string;
        name: string;
        categories: {
            category: string;
            cards: {
                revealed: boolean;
                instanceId: string;
                labelShort: string;
            }[];
        }[];
        hand: {
            revealed: boolean;
            id: string;
            deck: string;
            instanceId?: string | undefined;
            labelShort?: string | undefined;
            secret?: boolean | undefined;
            missing?: boolean | undefined;
        }[];
        specialConditions: {
            id: string;
            title: string;
            text: string;
            instanceId: string;
            trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
            effect: {
                type: string;
                params?: Record<string, any> | undefined;
            };
            implemented: boolean;
            revealedPublic: boolean;
            used: boolean;
            imgUrl?: string | undefined;
            needsChoice?: boolean | undefined;
            choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
            allowSelfTarget?: boolean | undefined;
            targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
        }[];
    }, {
        playerId: string;
        name: string;
        categories: {
            category: string;
            cards: {
                revealed: boolean;
                instanceId: string;
                labelShort: string;
            }[];
        }[];
        hand: {
            revealed: boolean;
            id: string;
            deck: string;
            instanceId?: string | undefined;
            labelShort?: string | undefined;
            secret?: boolean | undefined;
            missing?: boolean | undefined;
        }[];
        specialConditions: {
            id: string;
            title: string;
            text: string;
            instanceId: string;
            trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
            effect: {
                type: string;
                params?: Record<string, any> | undefined;
            };
            implemented: boolean;
            revealedPublic: boolean;
            used: boolean;
            imgUrl?: string | undefined;
            needsChoice?: boolean | undefined;
            choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
            allowSelfTarget?: boolean | undefined;
            targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
        }[];
    }>;
    public: z.ZodObject<{
        players: z.ZodArray<z.ZodObject<{
            playerId: z.ZodString;
            name: z.ZodString;
            status: z.ZodUnion<[z.ZodLiteral<"alive">, z.ZodLiteral<"eliminated">, z.ZodLiteral<"left_bunker">]>;
            connected: z.ZodBoolean;
            disconnectedAt: z.ZodOptional<z.ZodNumber>;
            totalAbsentMs: z.ZodOptional<z.ZodNumber>;
            currentOfflineMs: z.ZodOptional<z.ZodNumber>;
            kickRemainingMs: z.ZodOptional<z.ZodNumber>;
            leftBunker: z.ZodOptional<z.ZodBoolean>;
            revealedCards: z.ZodArray<z.ZodObject<{
                id: z.ZodString;
                deck: z.ZodString;
                instanceId: z.ZodOptional<z.ZodString>;
                labelShort: z.ZodOptional<z.ZodString>;
                secret: z.ZodOptional<z.ZodBoolean>;
                missing: z.ZodOptional<z.ZodBoolean>;
            }, "strip", z.ZodTypeAny, {
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }, {
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }>, "many">;
            revealedCount: z.ZodNumber;
            totalCards: z.ZodNumber;
            specialRevealed: z.ZodBoolean;
            categories: z.ZodArray<z.ZodObject<{
                category: z.ZodString;
                status: z.ZodUnion<[z.ZodLiteral<"hidden">, z.ZodLiteral<"revealed">]>;
                cards: z.ZodArray<z.ZodObject<{
                    labelShort: z.ZodString;
                    imgUrl: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }, {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }>, "many">;
            }, "strip", z.ZodTypeAny, {
                category: string;
                status: "hidden" | "revealed";
                cards: {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }[];
            }, {
                category: string;
                status: "hidden" | "revealed";
                cards: {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }[];
            }>, "many">;
        }, "strip", z.ZodTypeAny, {
            playerId: string;
            name: string;
            connected: boolean;
            status: "alive" | "eliminated" | "left_bunker";
            categories: {
                category: string;
                status: "hidden" | "revealed";
                cards: {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }[];
            }[];
            revealedCards: {
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            revealedCount: number;
            totalCards: number;
            specialRevealed: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }, {
            playerId: string;
            name: string;
            connected: boolean;
            status: "alive" | "eliminated" | "left_bunker";
            categories: {
                category: string;
                status: "hidden" | "revealed";
                cards: {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }[];
            }[];
            revealedCards: {
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            revealedCount: number;
            totalCards: number;
            specialRevealed: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }>, "many">;
        revealedThisRound: z.ZodArray<z.ZodString, "many">;
        roundRevealedCount: z.ZodOptional<z.ZodNumber>;
        roundTotalAlive: z.ZodOptional<z.ZodNumber>;
        currentTurnPlayerId: z.ZodOptional<z.ZodNullable<z.ZodString>>;
        votesRemainingInRound: z.ZodOptional<z.ZodNumber>;
        votesTotalThisRound: z.ZodOptional<z.ZodNumber>;
        revealLimit: z.ZodOptional<z.ZodNumber>;
        voting: z.ZodOptional<z.ZodObject<{
            hasVoted: z.ZodBoolean;
        }, "strip", z.ZodTypeAny, {
            hasVoted: boolean;
        }, {
            hasVoted: boolean;
        }>>;
        votePhase: z.ZodOptional<z.ZodNullable<z.ZodUnion<[z.ZodLiteral<"voting">, z.ZodLiteral<"voteSpecialWindow">, z.ZodLiteral<"voteResolve">]>>>;
        votesPublic: z.ZodOptional<z.ZodArray<z.ZodObject<{
            voterId: z.ZodString;
            voterName: z.ZodString;
            targetId: z.ZodOptional<z.ZodString>;
            targetName: z.ZodOptional<z.ZodString>;
            status: z.ZodUnion<[z.ZodLiteral<"voted">, z.ZodLiteral<"not_voted">, z.ZodLiteral<"invalid">]>;
            reason: z.ZodOptional<z.ZodString>;
            submittedAt: z.ZodOptional<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            status: "voted" | "not_voted" | "invalid";
            voterId: string;
            voterName: string;
            targetId?: string | undefined;
            targetName?: string | undefined;
            reason?: string | undefined;
            submittedAt?: number | undefined;
        }, {
            status: "voted" | "not_voted" | "invalid";
            voterId: string;
            voterName: string;
            targetId?: string | undefined;
            targetName?: string | undefined;
            reason?: string | undefined;
            submittedAt?: number | undefined;
        }>, "many">>;
        votingProgress: z.ZodOptional<z.ZodObject<{
            voted: z.ZodNumber;
            total: z.ZodNumber;
        }, "strip", z.ZodTypeAny, {
            voted: number;
            total: number;
        }, {
            voted: number;
            total: number;
        }>>;
        threatModifier: z.ZodOptional<z.ZodObject<{
            delta: z.ZodNumber;
            reasons: z.ZodArray<z.ZodString, "many">;
            baseCount: z.ZodNumber;
            finalCount: z.ZodNumber;
        }, "strip", z.ZodTypeAny, {
            delta: number;
            reasons: string[];
            baseCount: number;
            finalCount: number;
        }, {
            delta: number;
            reasons: string[];
            baseCount: number;
            finalCount: number;
        }>>;
        canOpenVotingModal: z.ZodOptional<z.ZodBoolean>;
        canContinue: z.ZodOptional<z.ZodBoolean>;
        activeTimer: z.ZodOptional<z.ZodNullable<z.ZodObject<{
            kind: z.ZodUnion<[z.ZodLiteral<"reveal_discussion">, z.ZodLiteral<"pre_vote">, z.ZodLiteral<"post_vote">]>;
            endsAt: z.ZodNumber;
        }, "strip", z.ZodTypeAny, {
            kind: "reveal_discussion" | "pre_vote" | "post_vote";
            endsAt: number;
        }, {
            kind: "reveal_discussion" | "pre_vote" | "post_vote";
            endsAt: number;
        }>>>;
        voteModalOpen: z.ZodOptional<z.ZodBoolean>;
        lastEliminated: z.ZodOptional<z.ZodString>;
        winners: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
        resolutionNote: z.ZodOptional<z.ZodString>;
        roundRules: z.ZodOptional<z.ZodObject<{
            noTalkUntilVoting: z.ZodOptional<z.ZodBoolean>;
            forcedRevealCategory: z.ZodOptional<z.ZodString>;
        }, "strip", z.ZodTypeAny, {
            noTalkUntilVoting?: boolean | undefined;
            forcedRevealCategory?: string | undefined;
        }, {
            noTalkUntilVoting?: boolean | undefined;
            forcedRevealCategory?: string | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        players: {
            playerId: string;
            name: string;
            connected: boolean;
            status: "alive" | "eliminated" | "left_bunker";
            categories: {
                category: string;
                status: "hidden" | "revealed";
                cards: {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }[];
            }[];
            revealedCards: {
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            revealedCount: number;
            totalCards: number;
            specialRevealed: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }[];
        revealedThisRound: string[];
        voting?: {
            hasVoted: boolean;
        } | undefined;
        roundRevealedCount?: number | undefined;
        roundTotalAlive?: number | undefined;
        currentTurnPlayerId?: string | null | undefined;
        votesRemainingInRound?: number | undefined;
        votesTotalThisRound?: number | undefined;
        revealLimit?: number | undefined;
        votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
        votesPublic?: {
            status: "voted" | "not_voted" | "invalid";
            voterId: string;
            voterName: string;
            targetId?: string | undefined;
            targetName?: string | undefined;
            reason?: string | undefined;
            submittedAt?: number | undefined;
        }[] | undefined;
        votingProgress?: {
            voted: number;
            total: number;
        } | undefined;
        threatModifier?: {
            delta: number;
            reasons: string[];
            baseCount: number;
            finalCount: number;
        } | undefined;
        canOpenVotingModal?: boolean | undefined;
        canContinue?: boolean | undefined;
        activeTimer?: {
            kind: "reveal_discussion" | "pre_vote" | "post_vote";
            endsAt: number;
        } | null | undefined;
        voteModalOpen?: boolean | undefined;
        lastEliminated?: string | undefined;
        winners?: string[] | undefined;
        resolutionNote?: string | undefined;
        roundRules?: {
            noTalkUntilVoting?: boolean | undefined;
            forcedRevealCategory?: string | undefined;
        } | undefined;
    }, {
        players: {
            playerId: string;
            name: string;
            connected: boolean;
            status: "alive" | "eliminated" | "left_bunker";
            categories: {
                category: string;
                status: "hidden" | "revealed";
                cards: {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }[];
            }[];
            revealedCards: {
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            revealedCount: number;
            totalCards: number;
            specialRevealed: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }[];
        revealedThisRound: string[];
        voting?: {
            hasVoted: boolean;
        } | undefined;
        roundRevealedCount?: number | undefined;
        roundTotalAlive?: number | undefined;
        currentTurnPlayerId?: string | null | undefined;
        votesRemainingInRound?: number | undefined;
        votesTotalThisRound?: number | undefined;
        revealLimit?: number | undefined;
        votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
        votesPublic?: {
            status: "voted" | "not_voted" | "invalid";
            voterId: string;
            voterName: string;
            targetId?: string | undefined;
            targetName?: string | undefined;
            reason?: string | undefined;
            submittedAt?: number | undefined;
        }[] | undefined;
        votingProgress?: {
            voted: number;
            total: number;
        } | undefined;
        threatModifier?: {
            delta: number;
            reasons: string[];
            baseCount: number;
            finalCount: number;
        } | undefined;
        canOpenVotingModal?: boolean | undefined;
        canContinue?: boolean | undefined;
        activeTimer?: {
            kind: "reveal_discussion" | "pre_vote" | "post_vote";
            endsAt: number;
        } | null | undefined;
        voteModalOpen?: boolean | undefined;
        lastEliminated?: string | undefined;
        winners?: string[] | undefined;
        resolutionNote?: string | undefined;
        roundRules?: {
            noTalkUntilVoting?: boolean | undefined;
            forcedRevealCategory?: string | undefined;
        } | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    round: number;
    phase: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended";
    ruleset: {
        playerCount: number;
        votesPerRound: number[];
        totalExiles: number;
        bunkerSeats: number;
        rulesetMode: "auto" | "preset" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
    };
    categoryOrder: string[];
    you: {
        playerId: string;
        name: string;
        categories: {
            category: string;
            cards: {
                revealed: boolean;
                instanceId: string;
                labelShort: string;
            }[];
        }[];
        hand: {
            revealed: boolean;
            id: string;
            deck: string;
            instanceId?: string | undefined;
            labelShort?: string | undefined;
            secret?: boolean | undefined;
            missing?: boolean | undefined;
        }[];
        specialConditions: {
            id: string;
            title: string;
            text: string;
            instanceId: string;
            trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
            effect: {
                type: string;
                params?: Record<string, any> | undefined;
            };
            implemented: boolean;
            revealedPublic: boolean;
            used: boolean;
            imgUrl?: string | undefined;
            needsChoice?: boolean | undefined;
            choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
            allowSelfTarget?: boolean | undefined;
            targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
        }[];
    };
    public: {
        players: {
            playerId: string;
            name: string;
            connected: boolean;
            status: "alive" | "eliminated" | "left_bunker";
            categories: {
                category: string;
                status: "hidden" | "revealed";
                cards: {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }[];
            }[];
            revealedCards: {
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            revealedCount: number;
            totalCards: number;
            specialRevealed: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }[];
        revealedThisRound: string[];
        voting?: {
            hasVoted: boolean;
        } | undefined;
        roundRevealedCount?: number | undefined;
        roundTotalAlive?: number | undefined;
        currentTurnPlayerId?: string | null | undefined;
        votesRemainingInRound?: number | undefined;
        votesTotalThisRound?: number | undefined;
        revealLimit?: number | undefined;
        votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
        votesPublic?: {
            status: "voted" | "not_voted" | "invalid";
            voterId: string;
            voterName: string;
            targetId?: string | undefined;
            targetName?: string | undefined;
            reason?: string | undefined;
            submittedAt?: number | undefined;
        }[] | undefined;
        votingProgress?: {
            voted: number;
            total: number;
        } | undefined;
        threatModifier?: {
            delta: number;
            reasons: string[];
            baseCount: number;
            finalCount: number;
        } | undefined;
        canOpenVotingModal?: boolean | undefined;
        canContinue?: boolean | undefined;
        activeTimer?: {
            kind: "reveal_discussion" | "pre_vote" | "post_vote";
            endsAt: number;
        } | null | undefined;
        voteModalOpen?: boolean | undefined;
        lastEliminated?: string | undefined;
        winners?: string[] | undefined;
        resolutionNote?: string | undefined;
        roundRules?: {
            noTalkUntilVoting?: boolean | undefined;
            forcedRevealCategory?: string | undefined;
        } | undefined;
    };
    world?: {
        bunker: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        disaster: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        };
        threats: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        counts: {
            bunker: number;
            threats: number;
        };
    } | undefined;
    lastStageText?: string | undefined;
    worldEvent?: {
        type: "bunker_revealed";
        index: number;
        round: number;
    } | undefined;
    postGame?: {
        isActive: boolean;
        enteredAt: number;
        outcome?: "survived" | "failed" | undefined;
        decidedBy?: string | undefined;
        decidedAt?: number | undefined;
    } | undefined;
}, {
    round: number;
    phase: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended";
    ruleset: {
        playerCount: number;
        votesPerRound: number[];
        totalExiles: number;
        bunkerSeats: number;
        rulesetMode: "auto" | "preset" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
    };
    categoryOrder: string[];
    you: {
        playerId: string;
        name: string;
        categories: {
            category: string;
            cards: {
                revealed: boolean;
                instanceId: string;
                labelShort: string;
            }[];
        }[];
        hand: {
            revealed: boolean;
            id: string;
            deck: string;
            instanceId?: string | undefined;
            labelShort?: string | undefined;
            secret?: boolean | undefined;
            missing?: boolean | undefined;
        }[];
        specialConditions: {
            id: string;
            title: string;
            text: string;
            instanceId: string;
            trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
            effect: {
                type: string;
                params?: Record<string, any> | undefined;
            };
            implemented: boolean;
            revealedPublic: boolean;
            used: boolean;
            imgUrl?: string | undefined;
            needsChoice?: boolean | undefined;
            choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
            allowSelfTarget?: boolean | undefined;
            targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
        }[];
    };
    public: {
        players: {
            playerId: string;
            name: string;
            connected: boolean;
            status: "alive" | "eliminated" | "left_bunker";
            categories: {
                category: string;
                status: "hidden" | "revealed";
                cards: {
                    labelShort: string;
                    imgUrl?: string | undefined;
                }[];
            }[];
            revealedCards: {
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            revealedCount: number;
            totalCards: number;
            specialRevealed: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }[];
        revealedThisRound: string[];
        voting?: {
            hasVoted: boolean;
        } | undefined;
        roundRevealedCount?: number | undefined;
        roundTotalAlive?: number | undefined;
        currentTurnPlayerId?: string | null | undefined;
        votesRemainingInRound?: number | undefined;
        votesTotalThisRound?: number | undefined;
        revealLimit?: number | undefined;
        votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
        votesPublic?: {
            status: "voted" | "not_voted" | "invalid";
            voterId: string;
            voterName: string;
            targetId?: string | undefined;
            targetName?: string | undefined;
            reason?: string | undefined;
            submittedAt?: number | undefined;
        }[] | undefined;
        votingProgress?: {
            voted: number;
            total: number;
        } | undefined;
        threatModifier?: {
            delta: number;
            reasons: string[];
            baseCount: number;
            finalCount: number;
        } | undefined;
        canOpenVotingModal?: boolean | undefined;
        canContinue?: boolean | undefined;
        activeTimer?: {
            kind: "reveal_discussion" | "pre_vote" | "post_vote";
            endsAt: number;
        } | null | undefined;
        voteModalOpen?: boolean | undefined;
        lastEliminated?: string | undefined;
        winners?: string[] | undefined;
        resolutionNote?: string | undefined;
        roundRules?: {
            noTalkUntilVoting?: boolean | undefined;
            forcedRevealCategory?: string | undefined;
        } | undefined;
    };
    world?: {
        bunker: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        disaster: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            text?: string | undefined;
            imageId?: string | undefined;
        };
        threats: {
            id: string;
            description: string;
            kind: "bunker" | "disaster" | "threat";
            title: string;
            isRevealed: boolean;
            text?: string | undefined;
            imageId?: string | undefined;
            revealedAtRound?: number | undefined;
            revealedBy?: string | undefined;
        }[];
        counts: {
            bunker: number;
            threats: number;
        };
    } | undefined;
    lastStageText?: string | undefined;
    worldEvent?: {
        type: "bunker_revealed";
        index: number;
        round: number;
    } | undefined;
    postGame?: {
        isActive: boolean;
        enteredAt: number;
        outcome?: "survived" | "failed" | undefined;
        decidedBy?: string | undefined;
        decidedAt?: number | undefined;
    } | undefined;
}>;
export declare const ClientHelloSchema: z.ZodObject<{
    name: z.ZodString;
    roomCode: z.ZodOptional<z.ZodString>;
    create: z.ZodOptional<z.ZodBoolean>;
    scenarioId: z.ZodOptional<z.ZodString>;
    playerToken: z.ZodOptional<z.ZodString>;
    tabId: z.ZodOptional<z.ZodString>;
    sessionId: z.ZodOptional<z.ZodString>;
}, "strip", z.ZodTypeAny, {
    name: string;
    roomCode?: string | undefined;
    create?: boolean | undefined;
    scenarioId?: string | undefined;
    playerToken?: string | undefined;
    tabId?: string | undefined;
    sessionId?: string | undefined;
}, {
    name: string;
    roomCode?: string | undefined;
    create?: boolean | undefined;
    scenarioId?: string | undefined;
    playerToken?: string | undefined;
    tabId?: string | undefined;
    sessionId?: string | undefined;
}>;
export declare const ClientMessageSchema: z.ZodDiscriminatedUnion<"type", [z.ZodObject<{
    type: z.ZodLiteral<"hello">;
    payload: z.ZodObject<{
        name: z.ZodString;
        roomCode: z.ZodOptional<z.ZodString>;
        create: z.ZodOptional<z.ZodBoolean>;
        scenarioId: z.ZodOptional<z.ZodString>;
        playerToken: z.ZodOptional<z.ZodString>;
        tabId: z.ZodOptional<z.ZodString>;
        sessionId: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        name: string;
        roomCode?: string | undefined;
        create?: boolean | undefined;
        scenarioId?: string | undefined;
        playerToken?: string | undefined;
        tabId?: string | undefined;
        sessionId?: string | undefined;
    }, {
        name: string;
        roomCode?: string | undefined;
        create?: boolean | undefined;
        scenarioId?: string | undefined;
        playerToken?: string | undefined;
        tabId?: string | undefined;
        sessionId?: string | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "hello";
    payload: {
        name: string;
        roomCode?: string | undefined;
        create?: boolean | undefined;
        scenarioId?: string | undefined;
        playerToken?: string | undefined;
        tabId?: string | undefined;
        sessionId?: string | undefined;
    };
}, {
    type: "hello";
    payload: {
        name: string;
        roomCode?: string | undefined;
        create?: boolean | undefined;
        scenarioId?: string | undefined;
        playerToken?: string | undefined;
        tabId?: string | undefined;
        sessionId?: string | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"resume">;
    payload: z.ZodObject<{
        roomCode: z.ZodString;
        sessionId: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        roomCode: string;
        sessionId: string;
    }, {
        roomCode: string;
        sessionId: string;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "resume";
    payload: {
        roomCode: string;
        sessionId: string;
    };
}, {
    type: "resume";
    payload: {
        roomCode: string;
        sessionId: string;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"startGame">;
    payload: z.ZodObject<{}, "strip", z.ZodTypeAny, {}, {}>;
}, "strip", z.ZodTypeAny, {
    type: "startGame";
    payload: {};
}, {
    type: "startGame";
    payload: {};
}>, z.ZodObject<{
    type: z.ZodLiteral<"ping">;
    payload: z.ZodObject<{}, "strip", z.ZodTypeAny, {}, {}>;
}, "strip", z.ZodTypeAny, {
    type: "ping";
    payload: {};
}, {
    type: "ping";
    payload: {};
}>, z.ZodObject<{
    type: z.ZodLiteral<"revealCard">;
    payload: z.ZodObject<{
        cardId: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        cardId: string;
    }, {
        cardId: string;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "revealCard";
    payload: {
        cardId: string;
    };
}, {
    type: "revealCard";
    payload: {
        cardId: string;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"vote">;
    payload: z.ZodObject<{
        targetPlayerId: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        targetPlayerId: string;
    }, {
        targetPlayerId: string;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "vote";
    payload: {
        targetPlayerId: string;
    };
}, {
    type: "vote";
    payload: {
        targetPlayerId: string;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"finalizeVoting">;
    payload: z.ZodObject<{}, "strip", z.ZodTypeAny, {}, {}>;
}, "strip", z.ZodTypeAny, {
    type: "finalizeVoting";
    payload: {};
}, {
    type: "finalizeVoting";
    payload: {};
}>, z.ZodObject<{
    type: z.ZodLiteral<"applySpecial">;
    payload: z.ZodObject<{
        specialInstanceId: z.ZodString;
        payload: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodAny>>;
    }, "strip", z.ZodTypeAny, {
        specialInstanceId: string;
        payload?: Record<string, any> | undefined;
    }, {
        specialInstanceId: string;
        payload?: Record<string, any> | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "applySpecial";
    payload: {
        specialInstanceId: string;
        payload?: Record<string, any> | undefined;
    };
}, {
    type: "applySpecial";
    payload: {
        specialInstanceId: string;
        payload?: Record<string, any> | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"revealWorldThreat">;
    payload: z.ZodObject<{
        index: z.ZodNumber;
    }, "strip", z.ZodTypeAny, {
        index: number;
    }, {
        index: number;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "revealWorldThreat";
    payload: {
        index: number;
    };
}, {
    type: "revealWorldThreat";
    payload: {
        index: number;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"setBunkerOutcome">;
    payload: z.ZodObject<{
        outcome: z.ZodUnion<[z.ZodLiteral<"survived">, z.ZodLiteral<"failed">]>;
    }, "strip", z.ZodTypeAny, {
        outcome: "survived" | "failed";
    }, {
        outcome: "survived" | "failed";
    }>;
}, "strip", z.ZodTypeAny, {
    type: "setBunkerOutcome";
    payload: {
        outcome: "survived" | "failed";
    };
}, {
    type: "setBunkerOutcome";
    payload: {
        outcome: "survived" | "failed";
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"devSkipRound">;
    payload: z.ZodObject<{}, "strip", z.ZodTypeAny, {}, {}>;
}, "strip", z.ZodTypeAny, {
    type: "devSkipRound";
    payload: {};
}, {
    type: "devSkipRound";
    payload: {};
}>, z.ZodObject<{
    type: z.ZodLiteral<"devKickPlayer">;
    payload: z.ZodObject<{
        targetPlayerId: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        targetPlayerId: string;
    }, {
        targetPlayerId: string;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "devKickPlayer";
    payload: {
        targetPlayerId: string;
    };
}, {
    type: "devKickPlayer";
    payload: {
        targetPlayerId: string;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"continueRound">;
    payload: z.ZodObject<{}, "strip", z.ZodTypeAny, {}, {}>;
}, "strip", z.ZodTypeAny, {
    type: "continueRound";
    payload: {};
}, {
    type: "continueRound";
    payload: {};
}>, z.ZodObject<{
    type: z.ZodLiteral<"kickFromLobby">;
    payload: z.ZodObject<{
        targetPlayerId: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        targetPlayerId: string;
    }, {
        targetPlayerId: string;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "kickFromLobby";
    payload: {
        targetPlayerId: string;
    };
}, {
    type: "kickFromLobby";
    payload: {
        targetPlayerId: string;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"updateSettings">;
    payload: z.ZodObject<{
        enableRevealDiscussionTimer: z.ZodBoolean;
        revealDiscussionSeconds: z.ZodNumber;
        enablePreVoteDiscussionTimer: z.ZodBoolean;
        preVoteDiscussionSeconds: z.ZodNumber;
        enablePostVoteDiscussionTimer: z.ZodBoolean;
        postVoteDiscussionSeconds: z.ZodNumber;
        enablePresenterMode: z.ZodBoolean;
        continuePermission: z.ZodUnion<[z.ZodLiteral<"host_only">, z.ZodLiteral<"revealer_only">, z.ZodLiteral<"anyone">]>;
        revealTimeoutAction: z.ZodUnion<[z.ZodLiteral<"random_card">, z.ZodLiteral<"skip_player">]>;
        revealsBeforeVoting: z.ZodNumber;
        specialUsage: z.ZodUnion<[z.ZodLiteral<"anytime">, z.ZodLiteral<"only_during_voting">]>;
        maxPlayers: z.ZodNumber;
        finalThreatReveal: z.ZodUnion<[z.ZodLiteral<"host">, z.ZodLiteral<"anyone">]>;
    }, "strip", z.ZodTypeAny, {
        enableRevealDiscussionTimer: boolean;
        revealDiscussionSeconds: number;
        enablePreVoteDiscussionTimer: boolean;
        preVoteDiscussionSeconds: number;
        enablePostVoteDiscussionTimer: boolean;
        postVoteDiscussionSeconds: number;
        enablePresenterMode: boolean;
        continuePermission: "host_only" | "revealer_only" | "anyone";
        revealTimeoutAction: "random_card" | "skip_player";
        revealsBeforeVoting: number;
        specialUsage: "anytime" | "only_during_voting";
        maxPlayers: number;
        finalThreatReveal: "anyone" | "host";
    }, {
        enableRevealDiscussionTimer: boolean;
        revealDiscussionSeconds: number;
        enablePreVoteDiscussionTimer: boolean;
        preVoteDiscussionSeconds: number;
        enablePostVoteDiscussionTimer: boolean;
        postVoteDiscussionSeconds: number;
        enablePresenterMode: boolean;
        continuePermission: "host_only" | "revealer_only" | "anyone";
        revealTimeoutAction: "random_card" | "skip_player";
        revealsBeforeVoting: number;
        specialUsage: "anytime" | "only_during_voting";
        maxPlayers: number;
        finalThreatReveal: "anyone" | "host";
    }>;
}, "strip", z.ZodTypeAny, {
    type: "updateSettings";
    payload: {
        enableRevealDiscussionTimer: boolean;
        revealDiscussionSeconds: number;
        enablePreVoteDiscussionTimer: boolean;
        preVoteDiscussionSeconds: number;
        enablePostVoteDiscussionTimer: boolean;
        postVoteDiscussionSeconds: number;
        enablePresenterMode: boolean;
        continuePermission: "host_only" | "revealer_only" | "anyone";
        revealTimeoutAction: "random_card" | "skip_player";
        revealsBeforeVoting: number;
        specialUsage: "anytime" | "only_during_voting";
        maxPlayers: number;
        finalThreatReveal: "anyone" | "host";
    };
}, {
    type: "updateSettings";
    payload: {
        enableRevealDiscussionTimer: boolean;
        revealDiscussionSeconds: number;
        enablePreVoteDiscussionTimer: boolean;
        preVoteDiscussionSeconds: number;
        enablePostVoteDiscussionTimer: boolean;
        postVoteDiscussionSeconds: number;
        enablePresenterMode: boolean;
        continuePermission: "host_only" | "revealer_only" | "anyone";
        revealTimeoutAction: "random_card" | "skip_player";
        revealsBeforeVoting: number;
        specialUsage: "anytime" | "only_during_voting";
        maxPlayers: number;
        finalThreatReveal: "anyone" | "host";
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"updateRules">;
    payload: z.ZodObject<{
        mode: z.ZodUnion<[z.ZodLiteral<"auto">, z.ZodLiteral<"manual">]>;
        presetPlayerCount: z.ZodOptional<z.ZodNumber>;
        manualConfig: z.ZodOptional<z.ZodObject<{
            bunkerSlots: z.ZodNumber;
            votesByRound: z.ZodArray<z.ZodNumber, "many">;
            targetReveals: z.ZodDefault<z.ZodNumber>;
            seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        }, {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        mode: "auto" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
        presetPlayerCount?: number | undefined;
    }, {
        mode: "auto" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
        presetPlayerCount?: number | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "updateRules";
    payload: {
        mode: "auto" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals: number;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
        presetPlayerCount?: number | undefined;
    };
}, {
    type: "updateRules";
    payload: {
        mode: "auto" | "manual";
        manualConfig?: {
            bunkerSlots: number;
            votesByRound: number[];
            targetReveals?: number | undefined;
            seedTemplatePlayers?: number | undefined;
        } | undefined;
        presetPlayerCount?: number | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"devAddPlayer">;
    payload: z.ZodObject<{
        name: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        name?: string | undefined;
    }, {
        name?: string | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "devAddPlayer";
    payload: {
        name?: string | undefined;
    };
}, {
    type: "devAddPlayer";
    payload: {
        name?: string | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"devRemovePlayer">;
    payload: z.ZodObject<{
        targetPlayerId: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        targetPlayerId?: string | undefined;
    }, {
        targetPlayerId?: string | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "devRemovePlayer";
    payload: {
        targetPlayerId?: string | undefined;
    };
}, {
    type: "devRemovePlayer";
    payload: {
        targetPlayerId?: string | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"requestHostTransfer">;
    payload: z.ZodObject<{}, "strip", z.ZodTypeAny, {}, {}>;
}, "strip", z.ZodTypeAny, {
    type: "requestHostTransfer";
    payload: {};
}, {
    type: "requestHostTransfer";
    payload: {};
}>, z.ZodObject<{
    type: z.ZodLiteral<"overlaySubscribe">;
    payload: z.ZodObject<{
        roomCode: z.ZodString;
        token: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        roomCode: string;
        token: string;
    }, {
        roomCode: string;
        token: string;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "overlaySubscribe";
    payload: {
        roomCode: string;
        token: string;
    };
}, {
    type: "overlaySubscribe";
    payload: {
        roomCode: string;
        token: string;
    };
}>]>;
export declare const ServerMessageSchema: z.ZodDiscriminatedUnion<"type", [z.ZodObject<{
    type: z.ZodLiteral<"roomState">;
    payload: z.ZodObject<{
        roomCode: z.ZodString;
        players: z.ZodArray<z.ZodObject<{
            playerId: z.ZodString;
            name: z.ZodString;
            connected: z.ZodBoolean;
            disconnectedAt: z.ZodOptional<z.ZodNumber>;
            totalAbsentMs: z.ZodOptional<z.ZodNumber>;
            currentOfflineMs: z.ZodOptional<z.ZodNumber>;
            kickRemainingMs: z.ZodOptional<z.ZodNumber>;
            leftBunker: z.ZodOptional<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            playerId: string;
            name: string;
            connected: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }, {
            playerId: string;
            name: string;
            connected: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }>, "many">;
        hostId: z.ZodString;
        controlId: z.ZodString;
        phase: z.ZodUnion<[z.ZodLiteral<"lobby">, z.ZodLiteral<"game">]>;
        scenarioMeta: z.ZodObject<{
            id: z.ZodString;
            name: z.ZodString;
            description: z.ZodOptional<z.ZodString>;
            devOnly: z.ZodOptional<z.ZodBoolean>;
        }, "strip", z.ZodTypeAny, {
            name: string;
            id: string;
            description?: string | undefined;
            devOnly?: boolean | undefined;
        }, {
            name: string;
            id: string;
            description?: string | undefined;
            devOnly?: boolean | undefined;
        }>;
        settings: z.ZodObject<{
            enableRevealDiscussionTimer: z.ZodBoolean;
            revealDiscussionSeconds: z.ZodNumber;
            enablePreVoteDiscussionTimer: z.ZodBoolean;
            preVoteDiscussionSeconds: z.ZodNumber;
            enablePostVoteDiscussionTimer: z.ZodBoolean;
            postVoteDiscussionSeconds: z.ZodNumber;
            enablePresenterMode: z.ZodBoolean;
            continuePermission: z.ZodUnion<[z.ZodLiteral<"host_only">, z.ZodLiteral<"revealer_only">, z.ZodLiteral<"anyone">]>;
            revealTimeoutAction: z.ZodUnion<[z.ZodLiteral<"random_card">, z.ZodLiteral<"skip_player">]>;
            revealsBeforeVoting: z.ZodNumber;
            specialUsage: z.ZodUnion<[z.ZodLiteral<"anytime">, z.ZodLiteral<"only_during_voting">]>;
            maxPlayers: z.ZodNumber;
            finalThreatReveal: z.ZodUnion<[z.ZodLiteral<"host">, z.ZodLiteral<"anyone">]>;
        }, "strip", z.ZodTypeAny, {
            enableRevealDiscussionTimer: boolean;
            revealDiscussionSeconds: number;
            enablePreVoteDiscussionTimer: boolean;
            preVoteDiscussionSeconds: number;
            enablePostVoteDiscussionTimer: boolean;
            postVoteDiscussionSeconds: number;
            enablePresenterMode: boolean;
            continuePermission: "host_only" | "revealer_only" | "anyone";
            revealTimeoutAction: "random_card" | "skip_player";
            revealsBeforeVoting: number;
            specialUsage: "anytime" | "only_during_voting";
            maxPlayers: number;
            finalThreatReveal: "anyone" | "host";
        }, {
            enableRevealDiscussionTimer: boolean;
            revealDiscussionSeconds: number;
            enablePreVoteDiscussionTimer: boolean;
            preVoteDiscussionSeconds: number;
            enablePostVoteDiscussionTimer: boolean;
            postVoteDiscussionSeconds: number;
            enablePresenterMode: boolean;
            continuePermission: "host_only" | "revealer_only" | "anyone";
            revealTimeoutAction: "random_card" | "skip_player";
            revealsBeforeVoting: number;
            specialUsage: "anytime" | "only_during_voting";
            maxPlayers: number;
            finalThreatReveal: "anyone" | "host";
        }>;
        ruleset: z.ZodObject<{
            playerCount: z.ZodNumber;
            votesPerRound: z.ZodArray<z.ZodNumber, "many">;
            totalExiles: z.ZodNumber;
            bunkerSeats: z.ZodNumber;
            rulesetMode: z.ZodUnion<[z.ZodLiteral<"auto">, z.ZodLiteral<"preset">, z.ZodLiteral<"manual">]>;
            manualConfig: z.ZodOptional<z.ZodObject<{
                bunkerSlots: z.ZodNumber;
                votesByRound: z.ZodArray<z.ZodNumber, "many">;
                targetReveals: z.ZodDefault<z.ZodNumber>;
                seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
            }, "strip", z.ZodTypeAny, {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals: number;
                seedTemplatePlayers?: number | undefined;
            }, {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals?: number | undefined;
                seedTemplatePlayers?: number | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals: number;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        }, {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals?: number | undefined;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        }>;
        rulesOverriddenByHost: z.ZodBoolean;
        rulesPresetCount: z.ZodOptional<z.ZodNumber>;
        world: z.ZodOptional<z.ZodObject<{
            disaster: z.ZodObject<{
                kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                id: z.ZodString;
                title: z.ZodString;
                description: z.ZodString;
                text: z.ZodOptional<z.ZodString>;
                imageId: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            }, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            }>;
            bunker: z.ZodArray<z.ZodObject<{
                kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                id: z.ZodString;
                title: z.ZodString;
                description: z.ZodString;
                text: z.ZodOptional<z.ZodString>;
                imageId: z.ZodOptional<z.ZodString>;
            } & {
                isRevealed: z.ZodBoolean;
                revealedAtRound: z.ZodOptional<z.ZodNumber>;
                revealedBy: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }>, "many">;
            threats: z.ZodArray<z.ZodObject<{
                kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                id: z.ZodString;
                title: z.ZodString;
                description: z.ZodString;
                text: z.ZodOptional<z.ZodString>;
                imageId: z.ZodOptional<z.ZodString>;
            } & {
                isRevealed: z.ZodBoolean;
                revealedAtRound: z.ZodOptional<z.ZodNumber>;
                revealedBy: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }>, "many">;
            counts: z.ZodObject<{
                bunker: z.ZodNumber;
                threats: z.ZodNumber;
            }, "strip", z.ZodTypeAny, {
                bunker: number;
                threats: number;
            }, {
                bunker: number;
                threats: number;
            }>;
        }, "strip", z.ZodTypeAny, {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        }, {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        }>>;
        isDev: z.ZodOptional<z.ZodBoolean>;
    }, "strip", z.ZodTypeAny, {
        players: {
            playerId: string;
            name: string;
            connected: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }[];
        roomCode: string;
        hostId: string;
        controlId: string;
        phase: "lobby" | "game";
        scenarioMeta: {
            name: string;
            id: string;
            description?: string | undefined;
            devOnly?: boolean | undefined;
        };
        settings: {
            enableRevealDiscussionTimer: boolean;
            revealDiscussionSeconds: number;
            enablePreVoteDiscussionTimer: boolean;
            preVoteDiscussionSeconds: number;
            enablePostVoteDiscussionTimer: boolean;
            postVoteDiscussionSeconds: number;
            enablePresenterMode: boolean;
            continuePermission: "host_only" | "revealer_only" | "anyone";
            revealTimeoutAction: "random_card" | "skip_player";
            revealsBeforeVoting: number;
            specialUsage: "anytime" | "only_during_voting";
            maxPlayers: number;
            finalThreatReveal: "anyone" | "host";
        };
        ruleset: {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals: number;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        };
        rulesOverriddenByHost: boolean;
        rulesPresetCount?: number | undefined;
        world?: {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        } | undefined;
        isDev?: boolean | undefined;
    }, {
        players: {
            playerId: string;
            name: string;
            connected: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }[];
        roomCode: string;
        hostId: string;
        controlId: string;
        phase: "lobby" | "game";
        scenarioMeta: {
            name: string;
            id: string;
            description?: string | undefined;
            devOnly?: boolean | undefined;
        };
        settings: {
            enableRevealDiscussionTimer: boolean;
            revealDiscussionSeconds: number;
            enablePreVoteDiscussionTimer: boolean;
            preVoteDiscussionSeconds: number;
            enablePostVoteDiscussionTimer: boolean;
            postVoteDiscussionSeconds: number;
            enablePresenterMode: boolean;
            continuePermission: "host_only" | "revealer_only" | "anyone";
            revealTimeoutAction: "random_card" | "skip_player";
            revealsBeforeVoting: number;
            specialUsage: "anytime" | "only_during_voting";
            maxPlayers: number;
            finalThreatReveal: "anyone" | "host";
        };
        ruleset: {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals?: number | undefined;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        };
        rulesOverriddenByHost: boolean;
        rulesPresetCount?: number | undefined;
        world?: {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        } | undefined;
        isDev?: boolean | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "roomState";
    payload: {
        players: {
            playerId: string;
            name: string;
            connected: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }[];
        roomCode: string;
        hostId: string;
        controlId: string;
        phase: "lobby" | "game";
        scenarioMeta: {
            name: string;
            id: string;
            description?: string | undefined;
            devOnly?: boolean | undefined;
        };
        settings: {
            enableRevealDiscussionTimer: boolean;
            revealDiscussionSeconds: number;
            enablePreVoteDiscussionTimer: boolean;
            preVoteDiscussionSeconds: number;
            enablePostVoteDiscussionTimer: boolean;
            postVoteDiscussionSeconds: number;
            enablePresenterMode: boolean;
            continuePermission: "host_only" | "revealer_only" | "anyone";
            revealTimeoutAction: "random_card" | "skip_player";
            revealsBeforeVoting: number;
            specialUsage: "anytime" | "only_during_voting";
            maxPlayers: number;
            finalThreatReveal: "anyone" | "host";
        };
        ruleset: {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals: number;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        };
        rulesOverriddenByHost: boolean;
        rulesPresetCount?: number | undefined;
        world?: {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        } | undefined;
        isDev?: boolean | undefined;
    };
}, {
    type: "roomState";
    payload: {
        players: {
            playerId: string;
            name: string;
            connected: boolean;
            disconnectedAt?: number | undefined;
            totalAbsentMs?: number | undefined;
            currentOfflineMs?: number | undefined;
            kickRemainingMs?: number | undefined;
            leftBunker?: boolean | undefined;
        }[];
        roomCode: string;
        hostId: string;
        controlId: string;
        phase: "lobby" | "game";
        scenarioMeta: {
            name: string;
            id: string;
            description?: string | undefined;
            devOnly?: boolean | undefined;
        };
        settings: {
            enableRevealDiscussionTimer: boolean;
            revealDiscussionSeconds: number;
            enablePreVoteDiscussionTimer: boolean;
            preVoteDiscussionSeconds: number;
            enablePostVoteDiscussionTimer: boolean;
            postVoteDiscussionSeconds: number;
            enablePresenterMode: boolean;
            continuePermission: "host_only" | "revealer_only" | "anyone";
            revealTimeoutAction: "random_card" | "skip_player";
            revealsBeforeVoting: number;
            specialUsage: "anytime" | "only_during_voting";
            maxPlayers: number;
            finalThreatReveal: "anyone" | "host";
        };
        ruleset: {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals?: number | undefined;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        };
        rulesOverriddenByHost: boolean;
        rulesPresetCount?: number | undefined;
        world?: {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        } | undefined;
        isDev?: boolean | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"gameView">;
    payload: z.ZodObject<{
        phase: z.ZodUnion<[z.ZodLiteral<"reveal">, z.ZodLiteral<"reveal_discussion">, z.ZodLiteral<"voting">, z.ZodLiteral<"resolution">, z.ZodLiteral<"ended">]>;
        round: z.ZodNumber;
        categoryOrder: z.ZodArray<z.ZodString, "many">;
        lastStageText: z.ZodOptional<z.ZodString>;
        ruleset: z.ZodObject<{
            playerCount: z.ZodNumber;
            votesPerRound: z.ZodArray<z.ZodNumber, "many">;
            totalExiles: z.ZodNumber;
            bunkerSeats: z.ZodNumber;
            rulesetMode: z.ZodUnion<[z.ZodLiteral<"auto">, z.ZodLiteral<"preset">, z.ZodLiteral<"manual">]>;
            manualConfig: z.ZodOptional<z.ZodObject<{
                bunkerSlots: z.ZodNumber;
                votesByRound: z.ZodArray<z.ZodNumber, "many">;
                targetReveals: z.ZodDefault<z.ZodNumber>;
                seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
            }, "strip", z.ZodTypeAny, {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals: number;
                seedTemplatePlayers?: number | undefined;
            }, {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals?: number | undefined;
                seedTemplatePlayers?: number | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals: number;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        }, {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals?: number | undefined;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        }>;
        world: z.ZodOptional<z.ZodObject<{
            disaster: z.ZodObject<{
                kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                id: z.ZodString;
                title: z.ZodString;
                description: z.ZodString;
                text: z.ZodOptional<z.ZodString>;
                imageId: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            }, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            }>;
            bunker: z.ZodArray<z.ZodObject<{
                kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                id: z.ZodString;
                title: z.ZodString;
                description: z.ZodString;
                text: z.ZodOptional<z.ZodString>;
                imageId: z.ZodOptional<z.ZodString>;
            } & {
                isRevealed: z.ZodBoolean;
                revealedAtRound: z.ZodOptional<z.ZodNumber>;
                revealedBy: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }>, "many">;
            threats: z.ZodArray<z.ZodObject<{
                kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                id: z.ZodString;
                title: z.ZodString;
                description: z.ZodString;
                text: z.ZodOptional<z.ZodString>;
                imageId: z.ZodOptional<z.ZodString>;
            } & {
                isRevealed: z.ZodBoolean;
                revealedAtRound: z.ZodOptional<z.ZodNumber>;
                revealedBy: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }, {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }>, "many">;
            counts: z.ZodObject<{
                bunker: z.ZodNumber;
                threats: z.ZodNumber;
            }, "strip", z.ZodTypeAny, {
                bunker: number;
                threats: number;
            }, {
                bunker: number;
                threats: number;
            }>;
        }, "strip", z.ZodTypeAny, {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        }, {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        }>>;
        worldEvent: z.ZodOptional<z.ZodObject<{
            type: z.ZodLiteral<"bunker_revealed">;
            index: z.ZodNumber;
            round: z.ZodNumber;
        }, "strip", z.ZodTypeAny, {
            type: "bunker_revealed";
            index: number;
            round: number;
        }, {
            type: "bunker_revealed";
            index: number;
            round: number;
        }>>;
        postGame: z.ZodOptional<z.ZodObject<{
            isActive: z.ZodBoolean;
            enteredAt: z.ZodNumber;
            outcome: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"survived">, z.ZodLiteral<"failed">]>>;
            decidedBy: z.ZodOptional<z.ZodString>;
            decidedAt: z.ZodOptional<z.ZodNumber>;
        }, "strip", z.ZodTypeAny, {
            isActive: boolean;
            enteredAt: number;
            outcome?: "survived" | "failed" | undefined;
            decidedBy?: string | undefined;
            decidedAt?: number | undefined;
        }, {
            isActive: boolean;
            enteredAt: number;
            outcome?: "survived" | "failed" | undefined;
            decidedBy?: string | undefined;
            decidedAt?: number | undefined;
        }>>;
        you: z.ZodObject<{
            playerId: z.ZodString;
            name: z.ZodString;
            hand: z.ZodArray<z.ZodObject<{
                id: z.ZodString;
                deck: z.ZodString;
                instanceId: z.ZodOptional<z.ZodString>;
                labelShort: z.ZodOptional<z.ZodString>;
                secret: z.ZodOptional<z.ZodBoolean>;
                missing: z.ZodOptional<z.ZodBoolean>;
            } & {
                revealed: z.ZodBoolean;
            }, "strip", z.ZodTypeAny, {
                revealed: boolean;
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }, {
                revealed: boolean;
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }>, "many">;
            categories: z.ZodArray<z.ZodObject<{
                category: z.ZodString;
                cards: z.ZodArray<z.ZodObject<{
                    instanceId: z.ZodString;
                    labelShort: z.ZodString;
                    revealed: z.ZodBoolean;
                }, "strip", z.ZodTypeAny, {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }, {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }>, "many">;
            }, "strip", z.ZodTypeAny, {
                category: string;
                cards: {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }[];
            }, {
                category: string;
                cards: {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }[];
            }>, "many">;
            specialConditions: z.ZodArray<z.ZodObject<{
                instanceId: z.ZodString;
                id: z.ZodString;
                title: z.ZodString;
                text: z.ZodString;
                trigger: z.ZodUnion<[z.ZodLiteral<"active">, z.ZodLiteral<"onVote">, z.ZodLiteral<"onOwnerEliminated">, z.ZodLiteral<"onRevealOrActive">, z.ZodLiteral<"secret_onEliminate">]>;
                effect: z.ZodObject<{
                    type: z.ZodString;
                    params: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodAny>>;
                }, "strip", z.ZodTypeAny, {
                    type: string;
                    params?: Record<string, any> | undefined;
                }, {
                    type: string;
                    params?: Record<string, any> | undefined;
                }>;
                implemented: z.ZodBoolean;
                revealedPublic: z.ZodBoolean;
                used: z.ZodBoolean;
                imgUrl: z.ZodOptional<z.ZodString>;
                needsChoice: z.ZodOptional<z.ZodBoolean>;
                choiceKind: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"player">, z.ZodLiteral<"neighbor">, z.ZodLiteral<"category">, z.ZodLiteral<"none">]>>;
                allowSelfTarget: z.ZodOptional<z.ZodBoolean>;
                targetScope: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"neighbors">, z.ZodLiteral<"any_alive">, z.ZodLiteral<"self">, z.ZodLiteral<"any_including_self">]>>;
            }, "strip", z.ZodTypeAny, {
                id: string;
                title: string;
                text: string;
                instanceId: string;
                trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                effect: {
                    type: string;
                    params?: Record<string, any> | undefined;
                };
                implemented: boolean;
                revealedPublic: boolean;
                used: boolean;
                imgUrl?: string | undefined;
                needsChoice?: boolean | undefined;
                choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                allowSelfTarget?: boolean | undefined;
                targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
            }, {
                id: string;
                title: string;
                text: string;
                instanceId: string;
                trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                effect: {
                    type: string;
                    params?: Record<string, any> | undefined;
                };
                implemented: boolean;
                revealedPublic: boolean;
                used: boolean;
                imgUrl?: string | undefined;
                needsChoice?: boolean | undefined;
                choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                allowSelfTarget?: boolean | undefined;
                targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
            }>, "many">;
        }, "strip", z.ZodTypeAny, {
            playerId: string;
            name: string;
            categories: {
                category: string;
                cards: {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }[];
            }[];
            hand: {
                revealed: boolean;
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            specialConditions: {
                id: string;
                title: string;
                text: string;
                instanceId: string;
                trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                effect: {
                    type: string;
                    params?: Record<string, any> | undefined;
                };
                implemented: boolean;
                revealedPublic: boolean;
                used: boolean;
                imgUrl?: string | undefined;
                needsChoice?: boolean | undefined;
                choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                allowSelfTarget?: boolean | undefined;
                targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
            }[];
        }, {
            playerId: string;
            name: string;
            categories: {
                category: string;
                cards: {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }[];
            }[];
            hand: {
                revealed: boolean;
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            specialConditions: {
                id: string;
                title: string;
                text: string;
                instanceId: string;
                trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                effect: {
                    type: string;
                    params?: Record<string, any> | undefined;
                };
                implemented: boolean;
                revealedPublic: boolean;
                used: boolean;
                imgUrl?: string | undefined;
                needsChoice?: boolean | undefined;
                choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                allowSelfTarget?: boolean | undefined;
                targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
            }[];
        }>;
        public: z.ZodObject<{
            players: z.ZodArray<z.ZodObject<{
                playerId: z.ZodString;
                name: z.ZodString;
                status: z.ZodUnion<[z.ZodLiteral<"alive">, z.ZodLiteral<"eliminated">, z.ZodLiteral<"left_bunker">]>;
                connected: z.ZodBoolean;
                disconnectedAt: z.ZodOptional<z.ZodNumber>;
                totalAbsentMs: z.ZodOptional<z.ZodNumber>;
                currentOfflineMs: z.ZodOptional<z.ZodNumber>;
                kickRemainingMs: z.ZodOptional<z.ZodNumber>;
                leftBunker: z.ZodOptional<z.ZodBoolean>;
                revealedCards: z.ZodArray<z.ZodObject<{
                    id: z.ZodString;
                    deck: z.ZodString;
                    instanceId: z.ZodOptional<z.ZodString>;
                    labelShort: z.ZodOptional<z.ZodString>;
                    secret: z.ZodOptional<z.ZodBoolean>;
                    missing: z.ZodOptional<z.ZodBoolean>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }, {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }>, "many">;
                revealedCount: z.ZodNumber;
                totalCards: z.ZodNumber;
                specialRevealed: z.ZodBoolean;
                categories: z.ZodArray<z.ZodObject<{
                    category: z.ZodString;
                    status: z.ZodUnion<[z.ZodLiteral<"hidden">, z.ZodLiteral<"revealed">]>;
                    cards: z.ZodArray<z.ZodObject<{
                        labelShort: z.ZodString;
                        imgUrl: z.ZodOptional<z.ZodString>;
                    }, "strip", z.ZodTypeAny, {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }, {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }>, "many">;
                }, "strip", z.ZodTypeAny, {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }, {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }>, "many">;
            }, "strip", z.ZodTypeAny, {
                playerId: string;
                name: string;
                connected: boolean;
                status: "alive" | "eliminated" | "left_bunker";
                categories: {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }[];
                revealedCards: {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                revealedCount: number;
                totalCards: number;
                specialRevealed: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }, {
                playerId: string;
                name: string;
                connected: boolean;
                status: "alive" | "eliminated" | "left_bunker";
                categories: {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }[];
                revealedCards: {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                revealedCount: number;
                totalCards: number;
                specialRevealed: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }>, "many">;
            revealedThisRound: z.ZodArray<z.ZodString, "many">;
            roundRevealedCount: z.ZodOptional<z.ZodNumber>;
            roundTotalAlive: z.ZodOptional<z.ZodNumber>;
            currentTurnPlayerId: z.ZodOptional<z.ZodNullable<z.ZodString>>;
            votesRemainingInRound: z.ZodOptional<z.ZodNumber>;
            votesTotalThisRound: z.ZodOptional<z.ZodNumber>;
            revealLimit: z.ZodOptional<z.ZodNumber>;
            voting: z.ZodOptional<z.ZodObject<{
                hasVoted: z.ZodBoolean;
            }, "strip", z.ZodTypeAny, {
                hasVoted: boolean;
            }, {
                hasVoted: boolean;
            }>>;
            votePhase: z.ZodOptional<z.ZodNullable<z.ZodUnion<[z.ZodLiteral<"voting">, z.ZodLiteral<"voteSpecialWindow">, z.ZodLiteral<"voteResolve">]>>>;
            votesPublic: z.ZodOptional<z.ZodArray<z.ZodObject<{
                voterId: z.ZodString;
                voterName: z.ZodString;
                targetId: z.ZodOptional<z.ZodString>;
                targetName: z.ZodOptional<z.ZodString>;
                status: z.ZodUnion<[z.ZodLiteral<"voted">, z.ZodLiteral<"not_voted">, z.ZodLiteral<"invalid">]>;
                reason: z.ZodOptional<z.ZodString>;
                submittedAt: z.ZodOptional<z.ZodNumber>;
            }, "strip", z.ZodTypeAny, {
                status: "voted" | "not_voted" | "invalid";
                voterId: string;
                voterName: string;
                targetId?: string | undefined;
                targetName?: string | undefined;
                reason?: string | undefined;
                submittedAt?: number | undefined;
            }, {
                status: "voted" | "not_voted" | "invalid";
                voterId: string;
                voterName: string;
                targetId?: string | undefined;
                targetName?: string | undefined;
                reason?: string | undefined;
                submittedAt?: number | undefined;
            }>, "many">>;
            votingProgress: z.ZodOptional<z.ZodObject<{
                voted: z.ZodNumber;
                total: z.ZodNumber;
            }, "strip", z.ZodTypeAny, {
                voted: number;
                total: number;
            }, {
                voted: number;
                total: number;
            }>>;
            threatModifier: z.ZodOptional<z.ZodObject<{
                delta: z.ZodNumber;
                reasons: z.ZodArray<z.ZodString, "many">;
                baseCount: z.ZodNumber;
                finalCount: z.ZodNumber;
            }, "strip", z.ZodTypeAny, {
                delta: number;
                reasons: string[];
                baseCount: number;
                finalCount: number;
            }, {
                delta: number;
                reasons: string[];
                baseCount: number;
                finalCount: number;
            }>>;
            canOpenVotingModal: z.ZodOptional<z.ZodBoolean>;
            canContinue: z.ZodOptional<z.ZodBoolean>;
            activeTimer: z.ZodOptional<z.ZodNullable<z.ZodObject<{
                kind: z.ZodUnion<[z.ZodLiteral<"reveal_discussion">, z.ZodLiteral<"pre_vote">, z.ZodLiteral<"post_vote">]>;
                endsAt: z.ZodNumber;
            }, "strip", z.ZodTypeAny, {
                kind: "reveal_discussion" | "pre_vote" | "post_vote";
                endsAt: number;
            }, {
                kind: "reveal_discussion" | "pre_vote" | "post_vote";
                endsAt: number;
            }>>>;
            voteModalOpen: z.ZodOptional<z.ZodBoolean>;
            lastEliminated: z.ZodOptional<z.ZodString>;
            winners: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
            resolutionNote: z.ZodOptional<z.ZodString>;
            roundRules: z.ZodOptional<z.ZodObject<{
                noTalkUntilVoting: z.ZodOptional<z.ZodBoolean>;
                forcedRevealCategory: z.ZodOptional<z.ZodString>;
            }, "strip", z.ZodTypeAny, {
                noTalkUntilVoting?: boolean | undefined;
                forcedRevealCategory?: string | undefined;
            }, {
                noTalkUntilVoting?: boolean | undefined;
                forcedRevealCategory?: string | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            players: {
                playerId: string;
                name: string;
                connected: boolean;
                status: "alive" | "eliminated" | "left_bunker";
                categories: {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }[];
                revealedCards: {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                revealedCount: number;
                totalCards: number;
                specialRevealed: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[];
            revealedThisRound: string[];
            voting?: {
                hasVoted: boolean;
            } | undefined;
            roundRevealedCount?: number | undefined;
            roundTotalAlive?: number | undefined;
            currentTurnPlayerId?: string | null | undefined;
            votesRemainingInRound?: number | undefined;
            votesTotalThisRound?: number | undefined;
            revealLimit?: number | undefined;
            votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
            votesPublic?: {
                status: "voted" | "not_voted" | "invalid";
                voterId: string;
                voterName: string;
                targetId?: string | undefined;
                targetName?: string | undefined;
                reason?: string | undefined;
                submittedAt?: number | undefined;
            }[] | undefined;
            votingProgress?: {
                voted: number;
                total: number;
            } | undefined;
            threatModifier?: {
                delta: number;
                reasons: string[];
                baseCount: number;
                finalCount: number;
            } | undefined;
            canOpenVotingModal?: boolean | undefined;
            canContinue?: boolean | undefined;
            activeTimer?: {
                kind: "reveal_discussion" | "pre_vote" | "post_vote";
                endsAt: number;
            } | null | undefined;
            voteModalOpen?: boolean | undefined;
            lastEliminated?: string | undefined;
            winners?: string[] | undefined;
            resolutionNote?: string | undefined;
            roundRules?: {
                noTalkUntilVoting?: boolean | undefined;
                forcedRevealCategory?: string | undefined;
            } | undefined;
        }, {
            players: {
                playerId: string;
                name: string;
                connected: boolean;
                status: "alive" | "eliminated" | "left_bunker";
                categories: {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }[];
                revealedCards: {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                revealedCount: number;
                totalCards: number;
                specialRevealed: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[];
            revealedThisRound: string[];
            voting?: {
                hasVoted: boolean;
            } | undefined;
            roundRevealedCount?: number | undefined;
            roundTotalAlive?: number | undefined;
            currentTurnPlayerId?: string | null | undefined;
            votesRemainingInRound?: number | undefined;
            votesTotalThisRound?: number | undefined;
            revealLimit?: number | undefined;
            votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
            votesPublic?: {
                status: "voted" | "not_voted" | "invalid";
                voterId: string;
                voterName: string;
                targetId?: string | undefined;
                targetName?: string | undefined;
                reason?: string | undefined;
                submittedAt?: number | undefined;
            }[] | undefined;
            votingProgress?: {
                voted: number;
                total: number;
            } | undefined;
            threatModifier?: {
                delta: number;
                reasons: string[];
                baseCount: number;
                finalCount: number;
            } | undefined;
            canOpenVotingModal?: boolean | undefined;
            canContinue?: boolean | undefined;
            activeTimer?: {
                kind: "reveal_discussion" | "pre_vote" | "post_vote";
                endsAt: number;
            } | null | undefined;
            voteModalOpen?: boolean | undefined;
            lastEliminated?: string | undefined;
            winners?: string[] | undefined;
            resolutionNote?: string | undefined;
            roundRules?: {
                noTalkUntilVoting?: boolean | undefined;
                forcedRevealCategory?: string | undefined;
            } | undefined;
        }>;
    }, "strip", z.ZodTypeAny, {
        round: number;
        phase: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended";
        ruleset: {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals: number;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        };
        categoryOrder: string[];
        you: {
            playerId: string;
            name: string;
            categories: {
                category: string;
                cards: {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }[];
            }[];
            hand: {
                revealed: boolean;
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            specialConditions: {
                id: string;
                title: string;
                text: string;
                instanceId: string;
                trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                effect: {
                    type: string;
                    params?: Record<string, any> | undefined;
                };
                implemented: boolean;
                revealedPublic: boolean;
                used: boolean;
                imgUrl?: string | undefined;
                needsChoice?: boolean | undefined;
                choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                allowSelfTarget?: boolean | undefined;
                targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
            }[];
        };
        public: {
            players: {
                playerId: string;
                name: string;
                connected: boolean;
                status: "alive" | "eliminated" | "left_bunker";
                categories: {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }[];
                revealedCards: {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                revealedCount: number;
                totalCards: number;
                specialRevealed: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[];
            revealedThisRound: string[];
            voting?: {
                hasVoted: boolean;
            } | undefined;
            roundRevealedCount?: number | undefined;
            roundTotalAlive?: number | undefined;
            currentTurnPlayerId?: string | null | undefined;
            votesRemainingInRound?: number | undefined;
            votesTotalThisRound?: number | undefined;
            revealLimit?: number | undefined;
            votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
            votesPublic?: {
                status: "voted" | "not_voted" | "invalid";
                voterId: string;
                voterName: string;
                targetId?: string | undefined;
                targetName?: string | undefined;
                reason?: string | undefined;
                submittedAt?: number | undefined;
            }[] | undefined;
            votingProgress?: {
                voted: number;
                total: number;
            } | undefined;
            threatModifier?: {
                delta: number;
                reasons: string[];
                baseCount: number;
                finalCount: number;
            } | undefined;
            canOpenVotingModal?: boolean | undefined;
            canContinue?: boolean | undefined;
            activeTimer?: {
                kind: "reveal_discussion" | "pre_vote" | "post_vote";
                endsAt: number;
            } | null | undefined;
            voteModalOpen?: boolean | undefined;
            lastEliminated?: string | undefined;
            winners?: string[] | undefined;
            resolutionNote?: string | undefined;
            roundRules?: {
                noTalkUntilVoting?: boolean | undefined;
                forcedRevealCategory?: string | undefined;
            } | undefined;
        };
        world?: {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        } | undefined;
        lastStageText?: string | undefined;
        worldEvent?: {
            type: "bunker_revealed";
            index: number;
            round: number;
        } | undefined;
        postGame?: {
            isActive: boolean;
            enteredAt: number;
            outcome?: "survived" | "failed" | undefined;
            decidedBy?: string | undefined;
            decidedAt?: number | undefined;
        } | undefined;
    }, {
        round: number;
        phase: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended";
        ruleset: {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals?: number | undefined;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        };
        categoryOrder: string[];
        you: {
            playerId: string;
            name: string;
            categories: {
                category: string;
                cards: {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }[];
            }[];
            hand: {
                revealed: boolean;
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            specialConditions: {
                id: string;
                title: string;
                text: string;
                instanceId: string;
                trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                effect: {
                    type: string;
                    params?: Record<string, any> | undefined;
                };
                implemented: boolean;
                revealedPublic: boolean;
                used: boolean;
                imgUrl?: string | undefined;
                needsChoice?: boolean | undefined;
                choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                allowSelfTarget?: boolean | undefined;
                targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
            }[];
        };
        public: {
            players: {
                playerId: string;
                name: string;
                connected: boolean;
                status: "alive" | "eliminated" | "left_bunker";
                categories: {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }[];
                revealedCards: {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                revealedCount: number;
                totalCards: number;
                specialRevealed: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[];
            revealedThisRound: string[];
            voting?: {
                hasVoted: boolean;
            } | undefined;
            roundRevealedCount?: number | undefined;
            roundTotalAlive?: number | undefined;
            currentTurnPlayerId?: string | null | undefined;
            votesRemainingInRound?: number | undefined;
            votesTotalThisRound?: number | undefined;
            revealLimit?: number | undefined;
            votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
            votesPublic?: {
                status: "voted" | "not_voted" | "invalid";
                voterId: string;
                voterName: string;
                targetId?: string | undefined;
                targetName?: string | undefined;
                reason?: string | undefined;
                submittedAt?: number | undefined;
            }[] | undefined;
            votingProgress?: {
                voted: number;
                total: number;
            } | undefined;
            threatModifier?: {
                delta: number;
                reasons: string[];
                baseCount: number;
                finalCount: number;
            } | undefined;
            canOpenVotingModal?: boolean | undefined;
            canContinue?: boolean | undefined;
            activeTimer?: {
                kind: "reveal_discussion" | "pre_vote" | "post_vote";
                endsAt: number;
            } | null | undefined;
            voteModalOpen?: boolean | undefined;
            lastEliminated?: string | undefined;
            winners?: string[] | undefined;
            resolutionNote?: string | undefined;
            roundRules?: {
                noTalkUntilVoting?: boolean | undefined;
                forcedRevealCategory?: string | undefined;
            } | undefined;
        };
        world?: {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        } | undefined;
        lastStageText?: string | undefined;
        worldEvent?: {
            type: "bunker_revealed";
            index: number;
            round: number;
        } | undefined;
        postGame?: {
            isActive: boolean;
            enteredAt: number;
            outcome?: "survived" | "failed" | undefined;
            decidedBy?: string | undefined;
            decidedAt?: number | undefined;
        } | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "gameView";
    payload: {
        round: number;
        phase: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended";
        ruleset: {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals: number;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        };
        categoryOrder: string[];
        you: {
            playerId: string;
            name: string;
            categories: {
                category: string;
                cards: {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }[];
            }[];
            hand: {
                revealed: boolean;
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            specialConditions: {
                id: string;
                title: string;
                text: string;
                instanceId: string;
                trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                effect: {
                    type: string;
                    params?: Record<string, any> | undefined;
                };
                implemented: boolean;
                revealedPublic: boolean;
                used: boolean;
                imgUrl?: string | undefined;
                needsChoice?: boolean | undefined;
                choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                allowSelfTarget?: boolean | undefined;
                targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
            }[];
        };
        public: {
            players: {
                playerId: string;
                name: string;
                connected: boolean;
                status: "alive" | "eliminated" | "left_bunker";
                categories: {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }[];
                revealedCards: {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                revealedCount: number;
                totalCards: number;
                specialRevealed: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[];
            revealedThisRound: string[];
            voting?: {
                hasVoted: boolean;
            } | undefined;
            roundRevealedCount?: number | undefined;
            roundTotalAlive?: number | undefined;
            currentTurnPlayerId?: string | null | undefined;
            votesRemainingInRound?: number | undefined;
            votesTotalThisRound?: number | undefined;
            revealLimit?: number | undefined;
            votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
            votesPublic?: {
                status: "voted" | "not_voted" | "invalid";
                voterId: string;
                voterName: string;
                targetId?: string | undefined;
                targetName?: string | undefined;
                reason?: string | undefined;
                submittedAt?: number | undefined;
            }[] | undefined;
            votingProgress?: {
                voted: number;
                total: number;
            } | undefined;
            threatModifier?: {
                delta: number;
                reasons: string[];
                baseCount: number;
                finalCount: number;
            } | undefined;
            canOpenVotingModal?: boolean | undefined;
            canContinue?: boolean | undefined;
            activeTimer?: {
                kind: "reveal_discussion" | "pre_vote" | "post_vote";
                endsAt: number;
            } | null | undefined;
            voteModalOpen?: boolean | undefined;
            lastEliminated?: string | undefined;
            winners?: string[] | undefined;
            resolutionNote?: string | undefined;
            roundRules?: {
                noTalkUntilVoting?: boolean | undefined;
                forcedRevealCategory?: string | undefined;
            } | undefined;
        };
        world?: {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        } | undefined;
        lastStageText?: string | undefined;
        worldEvent?: {
            type: "bunker_revealed";
            index: number;
            round: number;
        } | undefined;
        postGame?: {
            isActive: boolean;
            enteredAt: number;
            outcome?: "survived" | "failed" | undefined;
            decidedBy?: string | undefined;
            decidedAt?: number | undefined;
        } | undefined;
    };
}, {
    type: "gameView";
    payload: {
        round: number;
        phase: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended";
        ruleset: {
            playerCount: number;
            votesPerRound: number[];
            totalExiles: number;
            bunkerSeats: number;
            rulesetMode: "auto" | "preset" | "manual";
            manualConfig?: {
                bunkerSlots: number;
                votesByRound: number[];
                targetReveals?: number | undefined;
                seedTemplatePlayers?: number | undefined;
            } | undefined;
        };
        categoryOrder: string[];
        you: {
            playerId: string;
            name: string;
            categories: {
                category: string;
                cards: {
                    revealed: boolean;
                    instanceId: string;
                    labelShort: string;
                }[];
            }[];
            hand: {
                revealed: boolean;
                id: string;
                deck: string;
                instanceId?: string | undefined;
                labelShort?: string | undefined;
                secret?: boolean | undefined;
                missing?: boolean | undefined;
            }[];
            specialConditions: {
                id: string;
                title: string;
                text: string;
                instanceId: string;
                trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                effect: {
                    type: string;
                    params?: Record<string, any> | undefined;
                };
                implemented: boolean;
                revealedPublic: boolean;
                used: boolean;
                imgUrl?: string | undefined;
                needsChoice?: boolean | undefined;
                choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                allowSelfTarget?: boolean | undefined;
                targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
            }[];
        };
        public: {
            players: {
                playerId: string;
                name: string;
                connected: boolean;
                status: "alive" | "eliminated" | "left_bunker";
                categories: {
                    category: string;
                    status: "hidden" | "revealed";
                    cards: {
                        labelShort: string;
                        imgUrl?: string | undefined;
                    }[];
                }[];
                revealedCards: {
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                revealedCount: number;
                totalCards: number;
                specialRevealed: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[];
            revealedThisRound: string[];
            voting?: {
                hasVoted: boolean;
            } | undefined;
            roundRevealedCount?: number | undefined;
            roundTotalAlive?: number | undefined;
            currentTurnPlayerId?: string | null | undefined;
            votesRemainingInRound?: number | undefined;
            votesTotalThisRound?: number | undefined;
            revealLimit?: number | undefined;
            votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
            votesPublic?: {
                status: "voted" | "not_voted" | "invalid";
                voterId: string;
                voterName: string;
                targetId?: string | undefined;
                targetName?: string | undefined;
                reason?: string | undefined;
                submittedAt?: number | undefined;
            }[] | undefined;
            votingProgress?: {
                voted: number;
                total: number;
            } | undefined;
            threatModifier?: {
                delta: number;
                reasons: string[];
                baseCount: number;
                finalCount: number;
            } | undefined;
            canOpenVotingModal?: boolean | undefined;
            canContinue?: boolean | undefined;
            activeTimer?: {
                kind: "reveal_discussion" | "pre_vote" | "post_vote";
                endsAt: number;
            } | null | undefined;
            voteModalOpen?: boolean | undefined;
            lastEliminated?: string | undefined;
            winners?: string[] | undefined;
            resolutionNote?: string | undefined;
            roundRules?: {
                noTalkUntilVoting?: boolean | undefined;
                forcedRevealCategory?: string | undefined;
            } | undefined;
        };
        world?: {
            bunker: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            disaster: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                text?: string | undefined;
                imageId?: string | undefined;
            };
            threats: {
                id: string;
                description: string;
                kind: "bunker" | "disaster" | "threat";
                title: string;
                isRevealed: boolean;
                text?: string | undefined;
                imageId?: string | undefined;
                revealedAtRound?: number | undefined;
                revealedBy?: string | undefined;
            }[];
            counts: {
                bunker: number;
                threats: number;
            };
        } | undefined;
        lastStageText?: string | undefined;
        worldEvent?: {
            type: "bunker_revealed";
            index: number;
            round: number;
        } | undefined;
        postGame?: {
            isActive: boolean;
            enteredAt: number;
            outcome?: "survived" | "failed" | undefined;
            decidedBy?: string | undefined;
            decidedAt?: number | undefined;
        } | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"statePatch">;
    payload: z.ZodObject<{
        roomState: z.ZodOptional<z.ZodObject<{
            roomCode: z.ZodOptional<z.ZodString>;
            players: z.ZodOptional<z.ZodArray<z.ZodObject<{
                playerId: z.ZodString;
                name: z.ZodString;
                connected: z.ZodBoolean;
                disconnectedAt: z.ZodOptional<z.ZodNumber>;
                totalAbsentMs: z.ZodOptional<z.ZodNumber>;
                currentOfflineMs: z.ZodOptional<z.ZodNumber>;
                kickRemainingMs: z.ZodOptional<z.ZodNumber>;
                leftBunker: z.ZodOptional<z.ZodBoolean>;
            }, "strip", z.ZodTypeAny, {
                playerId: string;
                name: string;
                connected: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }, {
                playerId: string;
                name: string;
                connected: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }>, "many">>;
            hostId: z.ZodOptional<z.ZodString>;
            controlId: z.ZodOptional<z.ZodString>;
            phase: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"lobby">, z.ZodLiteral<"game">]>>;
            scenarioMeta: z.ZodOptional<z.ZodObject<{
                id: z.ZodString;
                name: z.ZodString;
                description: z.ZodOptional<z.ZodString>;
                devOnly: z.ZodOptional<z.ZodBoolean>;
            }, "strip", z.ZodTypeAny, {
                name: string;
                id: string;
                description?: string | undefined;
                devOnly?: boolean | undefined;
            }, {
                name: string;
                id: string;
                description?: string | undefined;
                devOnly?: boolean | undefined;
            }>>;
            settings: z.ZodOptional<z.ZodObject<{
                enableRevealDiscussionTimer: z.ZodBoolean;
                revealDiscussionSeconds: z.ZodNumber;
                enablePreVoteDiscussionTimer: z.ZodBoolean;
                preVoteDiscussionSeconds: z.ZodNumber;
                enablePostVoteDiscussionTimer: z.ZodBoolean;
                postVoteDiscussionSeconds: z.ZodNumber;
                enablePresenterMode: z.ZodBoolean;
                continuePermission: z.ZodUnion<[z.ZodLiteral<"host_only">, z.ZodLiteral<"revealer_only">, z.ZodLiteral<"anyone">]>;
                revealTimeoutAction: z.ZodUnion<[z.ZodLiteral<"random_card">, z.ZodLiteral<"skip_player">]>;
                revealsBeforeVoting: z.ZodNumber;
                specialUsage: z.ZodUnion<[z.ZodLiteral<"anytime">, z.ZodLiteral<"only_during_voting">]>;
                maxPlayers: z.ZodNumber;
                finalThreatReveal: z.ZodUnion<[z.ZodLiteral<"host">, z.ZodLiteral<"anyone">]>;
            }, "strip", z.ZodTypeAny, {
                enableRevealDiscussionTimer: boolean;
                revealDiscussionSeconds: number;
                enablePreVoteDiscussionTimer: boolean;
                preVoteDiscussionSeconds: number;
                enablePostVoteDiscussionTimer: boolean;
                postVoteDiscussionSeconds: number;
                enablePresenterMode: boolean;
                continuePermission: "host_only" | "revealer_only" | "anyone";
                revealTimeoutAction: "random_card" | "skip_player";
                revealsBeforeVoting: number;
                specialUsage: "anytime" | "only_during_voting";
                maxPlayers: number;
                finalThreatReveal: "anyone" | "host";
            }, {
                enableRevealDiscussionTimer: boolean;
                revealDiscussionSeconds: number;
                enablePreVoteDiscussionTimer: boolean;
                preVoteDiscussionSeconds: number;
                enablePostVoteDiscussionTimer: boolean;
                postVoteDiscussionSeconds: number;
                enablePresenterMode: boolean;
                continuePermission: "host_only" | "revealer_only" | "anyone";
                revealTimeoutAction: "random_card" | "skip_player";
                revealsBeforeVoting: number;
                specialUsage: "anytime" | "only_during_voting";
                maxPlayers: number;
                finalThreatReveal: "anyone" | "host";
            }>>;
            ruleset: z.ZodOptional<z.ZodObject<{
                playerCount: z.ZodNumber;
                votesPerRound: z.ZodArray<z.ZodNumber, "many">;
                totalExiles: z.ZodNumber;
                bunkerSeats: z.ZodNumber;
                rulesetMode: z.ZodUnion<[z.ZodLiteral<"auto">, z.ZodLiteral<"preset">, z.ZodLiteral<"manual">]>;
                manualConfig: z.ZodOptional<z.ZodObject<{
                    bunkerSlots: z.ZodNumber;
                    votesByRound: z.ZodArray<z.ZodNumber, "many">;
                    targetReveals: z.ZodDefault<z.ZodNumber>;
                    seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
                }, "strip", z.ZodTypeAny, {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                }, {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                }>>;
            }, "strip", z.ZodTypeAny, {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            }, {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            }>>;
            rulesOverriddenByHost: z.ZodOptional<z.ZodBoolean>;
            rulesPresetCount: z.ZodOptional<z.ZodOptional<z.ZodNumber>>;
            world: z.ZodOptional<z.ZodOptional<z.ZodObject<{
                disaster: z.ZodObject<{
                    kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                    id: z.ZodString;
                    title: z.ZodString;
                    description: z.ZodString;
                    text: z.ZodOptional<z.ZodString>;
                    imageId: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                }, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                }>;
                bunker: z.ZodArray<z.ZodObject<{
                    kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                    id: z.ZodString;
                    title: z.ZodString;
                    description: z.ZodString;
                    text: z.ZodOptional<z.ZodString>;
                    imageId: z.ZodOptional<z.ZodString>;
                } & {
                    isRevealed: z.ZodBoolean;
                    revealedAtRound: z.ZodOptional<z.ZodNumber>;
                    revealedBy: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }>, "many">;
                threats: z.ZodArray<z.ZodObject<{
                    kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                    id: z.ZodString;
                    title: z.ZodString;
                    description: z.ZodString;
                    text: z.ZodOptional<z.ZodString>;
                    imageId: z.ZodOptional<z.ZodString>;
                } & {
                    isRevealed: z.ZodBoolean;
                    revealedAtRound: z.ZodOptional<z.ZodNumber>;
                    revealedBy: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }>, "many">;
                counts: z.ZodObject<{
                    bunker: z.ZodNumber;
                    threats: z.ZodNumber;
                }, "strip", z.ZodTypeAny, {
                    bunker: number;
                    threats: number;
                }, {
                    bunker: number;
                    threats: number;
                }>;
            }, "strip", z.ZodTypeAny, {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            }, {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            }>>>;
            isDev: z.ZodOptional<z.ZodOptional<z.ZodBoolean>>;
        }, "strip", z.ZodTypeAny, {
            players?: {
                playerId: string;
                name: string;
                connected: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[] | undefined;
            roomCode?: string | undefined;
            hostId?: string | undefined;
            controlId?: string | undefined;
            phase?: "lobby" | "game" | undefined;
            scenarioMeta?: {
                name: string;
                id: string;
                description?: string | undefined;
                devOnly?: boolean | undefined;
            } | undefined;
            settings?: {
                enableRevealDiscussionTimer: boolean;
                revealDiscussionSeconds: number;
                enablePreVoteDiscussionTimer: boolean;
                preVoteDiscussionSeconds: number;
                enablePostVoteDiscussionTimer: boolean;
                postVoteDiscussionSeconds: number;
                enablePresenterMode: boolean;
                continuePermission: "host_only" | "revealer_only" | "anyone";
                revealTimeoutAction: "random_card" | "skip_player";
                revealsBeforeVoting: number;
                specialUsage: "anytime" | "only_during_voting";
                maxPlayers: number;
                finalThreatReveal: "anyone" | "host";
            } | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            rulesOverriddenByHost?: boolean | undefined;
            rulesPresetCount?: number | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            isDev?: boolean | undefined;
        }, {
            players?: {
                playerId: string;
                name: string;
                connected: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[] | undefined;
            roomCode?: string | undefined;
            hostId?: string | undefined;
            controlId?: string | undefined;
            phase?: "lobby" | "game" | undefined;
            scenarioMeta?: {
                name: string;
                id: string;
                description?: string | undefined;
                devOnly?: boolean | undefined;
            } | undefined;
            settings?: {
                enableRevealDiscussionTimer: boolean;
                revealDiscussionSeconds: number;
                enablePreVoteDiscussionTimer: boolean;
                preVoteDiscussionSeconds: number;
                enablePostVoteDiscussionTimer: boolean;
                postVoteDiscussionSeconds: number;
                enablePresenterMode: boolean;
                continuePermission: "host_only" | "revealer_only" | "anyone";
                revealTimeoutAction: "random_card" | "skip_player";
                revealsBeforeVoting: number;
                specialUsage: "anytime" | "only_during_voting";
                maxPlayers: number;
                finalThreatReveal: "anyone" | "host";
            } | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            rulesOverriddenByHost?: boolean | undefined;
            rulesPresetCount?: number | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            isDev?: boolean | undefined;
        }>>;
        gameView: z.ZodOptional<z.ZodObject<{
            phase: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"reveal">, z.ZodLiteral<"reveal_discussion">, z.ZodLiteral<"voting">, z.ZodLiteral<"resolution">, z.ZodLiteral<"ended">]>>;
            round: z.ZodOptional<z.ZodNumber>;
            categoryOrder: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
            lastStageText: z.ZodOptional<z.ZodOptional<z.ZodString>>;
            ruleset: z.ZodOptional<z.ZodObject<{
                playerCount: z.ZodNumber;
                votesPerRound: z.ZodArray<z.ZodNumber, "many">;
                totalExiles: z.ZodNumber;
                bunkerSeats: z.ZodNumber;
                rulesetMode: z.ZodUnion<[z.ZodLiteral<"auto">, z.ZodLiteral<"preset">, z.ZodLiteral<"manual">]>;
                manualConfig: z.ZodOptional<z.ZodObject<{
                    bunkerSlots: z.ZodNumber;
                    votesByRound: z.ZodArray<z.ZodNumber, "many">;
                    targetReveals: z.ZodDefault<z.ZodNumber>;
                    seedTemplatePlayers: z.ZodOptional<z.ZodNumber>;
                }, "strip", z.ZodTypeAny, {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                }, {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                }>>;
            }, "strip", z.ZodTypeAny, {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            }, {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            }>>;
            world: z.ZodOptional<z.ZodOptional<z.ZodObject<{
                disaster: z.ZodObject<{
                    kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                    id: z.ZodString;
                    title: z.ZodString;
                    description: z.ZodString;
                    text: z.ZodOptional<z.ZodString>;
                    imageId: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                }, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                }>;
                bunker: z.ZodArray<z.ZodObject<{
                    kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                    id: z.ZodString;
                    title: z.ZodString;
                    description: z.ZodString;
                    text: z.ZodOptional<z.ZodString>;
                    imageId: z.ZodOptional<z.ZodString>;
                } & {
                    isRevealed: z.ZodBoolean;
                    revealedAtRound: z.ZodOptional<z.ZodNumber>;
                    revealedBy: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }>, "many">;
                threats: z.ZodArray<z.ZodObject<{
                    kind: z.ZodUnion<[z.ZodLiteral<"bunker">, z.ZodLiteral<"disaster">, z.ZodLiteral<"threat">]>;
                    id: z.ZodString;
                    title: z.ZodString;
                    description: z.ZodString;
                    text: z.ZodOptional<z.ZodString>;
                    imageId: z.ZodOptional<z.ZodString>;
                } & {
                    isRevealed: z.ZodBoolean;
                    revealedAtRound: z.ZodOptional<z.ZodNumber>;
                    revealedBy: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }, {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }>, "many">;
                counts: z.ZodObject<{
                    bunker: z.ZodNumber;
                    threats: z.ZodNumber;
                }, "strip", z.ZodTypeAny, {
                    bunker: number;
                    threats: number;
                }, {
                    bunker: number;
                    threats: number;
                }>;
            }, "strip", z.ZodTypeAny, {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            }, {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            }>>>;
            worldEvent: z.ZodOptional<z.ZodOptional<z.ZodObject<{
                type: z.ZodLiteral<"bunker_revealed">;
                index: z.ZodNumber;
                round: z.ZodNumber;
            }, "strip", z.ZodTypeAny, {
                type: "bunker_revealed";
                index: number;
                round: number;
            }, {
                type: "bunker_revealed";
                index: number;
                round: number;
            }>>>;
            postGame: z.ZodOptional<z.ZodOptional<z.ZodObject<{
                isActive: z.ZodBoolean;
                enteredAt: z.ZodNumber;
                outcome: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"survived">, z.ZodLiteral<"failed">]>>;
                decidedBy: z.ZodOptional<z.ZodString>;
                decidedAt: z.ZodOptional<z.ZodNumber>;
            }, "strip", z.ZodTypeAny, {
                isActive: boolean;
                enteredAt: number;
                outcome?: "survived" | "failed" | undefined;
                decidedBy?: string | undefined;
                decidedAt?: number | undefined;
            }, {
                isActive: boolean;
                enteredAt: number;
                outcome?: "survived" | "failed" | undefined;
                decidedBy?: string | undefined;
                decidedAt?: number | undefined;
            }>>>;
            you: z.ZodOptional<z.ZodObject<{
                playerId: z.ZodString;
                name: z.ZodString;
                hand: z.ZodArray<z.ZodObject<{
                    id: z.ZodString;
                    deck: z.ZodString;
                    instanceId: z.ZodOptional<z.ZodString>;
                    labelShort: z.ZodOptional<z.ZodString>;
                    secret: z.ZodOptional<z.ZodBoolean>;
                    missing: z.ZodOptional<z.ZodBoolean>;
                } & {
                    revealed: z.ZodBoolean;
                }, "strip", z.ZodTypeAny, {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }, {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }>, "many">;
                categories: z.ZodArray<z.ZodObject<{
                    category: z.ZodString;
                    cards: z.ZodArray<z.ZodObject<{
                        instanceId: z.ZodString;
                        labelShort: z.ZodString;
                        revealed: z.ZodBoolean;
                    }, "strip", z.ZodTypeAny, {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }, {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }>, "many">;
                }, "strip", z.ZodTypeAny, {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }, {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }>, "many">;
                specialConditions: z.ZodArray<z.ZodObject<{
                    instanceId: z.ZodString;
                    id: z.ZodString;
                    title: z.ZodString;
                    text: z.ZodString;
                    trigger: z.ZodUnion<[z.ZodLiteral<"active">, z.ZodLiteral<"onVote">, z.ZodLiteral<"onOwnerEliminated">, z.ZodLiteral<"onRevealOrActive">, z.ZodLiteral<"secret_onEliminate">]>;
                    effect: z.ZodObject<{
                        type: z.ZodString;
                        params: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodAny>>;
                    }, "strip", z.ZodTypeAny, {
                        type: string;
                        params?: Record<string, any> | undefined;
                    }, {
                        type: string;
                        params?: Record<string, any> | undefined;
                    }>;
                    implemented: z.ZodBoolean;
                    revealedPublic: z.ZodBoolean;
                    used: z.ZodBoolean;
                    imgUrl: z.ZodOptional<z.ZodString>;
                    needsChoice: z.ZodOptional<z.ZodBoolean>;
                    choiceKind: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"player">, z.ZodLiteral<"neighbor">, z.ZodLiteral<"category">, z.ZodLiteral<"none">]>>;
                    allowSelfTarget: z.ZodOptional<z.ZodBoolean>;
                    targetScope: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"neighbors">, z.ZodLiteral<"any_alive">, z.ZodLiteral<"self">, z.ZodLiteral<"any_including_self">]>>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }, {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }>, "many">;
            }, "strip", z.ZodTypeAny, {
                playerId: string;
                name: string;
                categories: {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }[];
                hand: {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                specialConditions: {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }[];
            }, {
                playerId: string;
                name: string;
                categories: {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }[];
                hand: {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                specialConditions: {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }[];
            }>>;
            public: z.ZodOptional<z.ZodObject<{
                players: z.ZodArray<z.ZodObject<{
                    playerId: z.ZodString;
                    name: z.ZodString;
                    status: z.ZodUnion<[z.ZodLiteral<"alive">, z.ZodLiteral<"eliminated">, z.ZodLiteral<"left_bunker">]>;
                    connected: z.ZodBoolean;
                    disconnectedAt: z.ZodOptional<z.ZodNumber>;
                    totalAbsentMs: z.ZodOptional<z.ZodNumber>;
                    currentOfflineMs: z.ZodOptional<z.ZodNumber>;
                    kickRemainingMs: z.ZodOptional<z.ZodNumber>;
                    leftBunker: z.ZodOptional<z.ZodBoolean>;
                    revealedCards: z.ZodArray<z.ZodObject<{
                        id: z.ZodString;
                        deck: z.ZodString;
                        instanceId: z.ZodOptional<z.ZodString>;
                        labelShort: z.ZodOptional<z.ZodString>;
                        secret: z.ZodOptional<z.ZodBoolean>;
                        missing: z.ZodOptional<z.ZodBoolean>;
                    }, "strip", z.ZodTypeAny, {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }, {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }>, "many">;
                    revealedCount: z.ZodNumber;
                    totalCards: z.ZodNumber;
                    specialRevealed: z.ZodBoolean;
                    categories: z.ZodArray<z.ZodObject<{
                        category: z.ZodString;
                        status: z.ZodUnion<[z.ZodLiteral<"hidden">, z.ZodLiteral<"revealed">]>;
                        cards: z.ZodArray<z.ZodObject<{
                            labelShort: z.ZodString;
                            imgUrl: z.ZodOptional<z.ZodString>;
                        }, "strip", z.ZodTypeAny, {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }, {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }>, "many">;
                    }, "strip", z.ZodTypeAny, {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }, {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }>, "many">;
                }, "strip", z.ZodTypeAny, {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }, {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }>, "many">;
                revealedThisRound: z.ZodArray<z.ZodString, "many">;
                roundRevealedCount: z.ZodOptional<z.ZodNumber>;
                roundTotalAlive: z.ZodOptional<z.ZodNumber>;
                currentTurnPlayerId: z.ZodOptional<z.ZodNullable<z.ZodString>>;
                votesRemainingInRound: z.ZodOptional<z.ZodNumber>;
                votesTotalThisRound: z.ZodOptional<z.ZodNumber>;
                revealLimit: z.ZodOptional<z.ZodNumber>;
                voting: z.ZodOptional<z.ZodObject<{
                    hasVoted: z.ZodBoolean;
                }, "strip", z.ZodTypeAny, {
                    hasVoted: boolean;
                }, {
                    hasVoted: boolean;
                }>>;
                votePhase: z.ZodOptional<z.ZodNullable<z.ZodUnion<[z.ZodLiteral<"voting">, z.ZodLiteral<"voteSpecialWindow">, z.ZodLiteral<"voteResolve">]>>>;
                votesPublic: z.ZodOptional<z.ZodArray<z.ZodObject<{
                    voterId: z.ZodString;
                    voterName: z.ZodString;
                    targetId: z.ZodOptional<z.ZodString>;
                    targetName: z.ZodOptional<z.ZodString>;
                    status: z.ZodUnion<[z.ZodLiteral<"voted">, z.ZodLiteral<"not_voted">, z.ZodLiteral<"invalid">]>;
                    reason: z.ZodOptional<z.ZodString>;
                    submittedAt: z.ZodOptional<z.ZodNumber>;
                }, "strip", z.ZodTypeAny, {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }, {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }>, "many">>;
                votingProgress: z.ZodOptional<z.ZodObject<{
                    voted: z.ZodNumber;
                    total: z.ZodNumber;
                }, "strip", z.ZodTypeAny, {
                    voted: number;
                    total: number;
                }, {
                    voted: number;
                    total: number;
                }>>;
                threatModifier: z.ZodOptional<z.ZodObject<{
                    delta: z.ZodNumber;
                    reasons: z.ZodArray<z.ZodString, "many">;
                    baseCount: z.ZodNumber;
                    finalCount: z.ZodNumber;
                }, "strip", z.ZodTypeAny, {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                }, {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                }>>;
                canOpenVotingModal: z.ZodOptional<z.ZodBoolean>;
                canContinue: z.ZodOptional<z.ZodBoolean>;
                activeTimer: z.ZodOptional<z.ZodNullable<z.ZodObject<{
                    kind: z.ZodUnion<[z.ZodLiteral<"reveal_discussion">, z.ZodLiteral<"pre_vote">, z.ZodLiteral<"post_vote">]>;
                    endsAt: z.ZodNumber;
                }, "strip", z.ZodTypeAny, {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                }, {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                }>>>;
                voteModalOpen: z.ZodOptional<z.ZodBoolean>;
                lastEliminated: z.ZodOptional<z.ZodString>;
                winners: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
                resolutionNote: z.ZodOptional<z.ZodString>;
                roundRules: z.ZodOptional<z.ZodObject<{
                    noTalkUntilVoting: z.ZodOptional<z.ZodBoolean>;
                    forcedRevealCategory: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                }, {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                }>>;
            }, "strip", z.ZodTypeAny, {
                players: {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }[];
                revealedThisRound: string[];
                voting?: {
                    hasVoted: boolean;
                } | undefined;
                roundRevealedCount?: number | undefined;
                roundTotalAlive?: number | undefined;
                currentTurnPlayerId?: string | null | undefined;
                votesRemainingInRound?: number | undefined;
                votesTotalThisRound?: number | undefined;
                revealLimit?: number | undefined;
                votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
                votesPublic?: {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }[] | undefined;
                votingProgress?: {
                    voted: number;
                    total: number;
                } | undefined;
                threatModifier?: {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                } | undefined;
                canOpenVotingModal?: boolean | undefined;
                canContinue?: boolean | undefined;
                activeTimer?: {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                } | null | undefined;
                voteModalOpen?: boolean | undefined;
                lastEliminated?: string | undefined;
                winners?: string[] | undefined;
                resolutionNote?: string | undefined;
                roundRules?: {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                } | undefined;
            }, {
                players: {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }[];
                revealedThisRound: string[];
                voting?: {
                    hasVoted: boolean;
                } | undefined;
                roundRevealedCount?: number | undefined;
                roundTotalAlive?: number | undefined;
                currentTurnPlayerId?: string | null | undefined;
                votesRemainingInRound?: number | undefined;
                votesTotalThisRound?: number | undefined;
                revealLimit?: number | undefined;
                votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
                votesPublic?: {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }[] | undefined;
                votingProgress?: {
                    voted: number;
                    total: number;
                } | undefined;
                threatModifier?: {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                } | undefined;
                canOpenVotingModal?: boolean | undefined;
                canContinue?: boolean | undefined;
                activeTimer?: {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                } | null | undefined;
                voteModalOpen?: boolean | undefined;
                lastEliminated?: string | undefined;
                winners?: string[] | undefined;
                resolutionNote?: string | undefined;
                roundRules?: {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                } | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            round?: number | undefined;
            phase?: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended" | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            categoryOrder?: string[] | undefined;
            lastStageText?: string | undefined;
            worldEvent?: {
                type: "bunker_revealed";
                index: number;
                round: number;
            } | undefined;
            postGame?: {
                isActive: boolean;
                enteredAt: number;
                outcome?: "survived" | "failed" | undefined;
                decidedBy?: string | undefined;
                decidedAt?: number | undefined;
            } | undefined;
            you?: {
                playerId: string;
                name: string;
                categories: {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }[];
                hand: {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                specialConditions: {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }[];
            } | undefined;
            public?: {
                players: {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }[];
                revealedThisRound: string[];
                voting?: {
                    hasVoted: boolean;
                } | undefined;
                roundRevealedCount?: number | undefined;
                roundTotalAlive?: number | undefined;
                currentTurnPlayerId?: string | null | undefined;
                votesRemainingInRound?: number | undefined;
                votesTotalThisRound?: number | undefined;
                revealLimit?: number | undefined;
                votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
                votesPublic?: {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }[] | undefined;
                votingProgress?: {
                    voted: number;
                    total: number;
                } | undefined;
                threatModifier?: {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                } | undefined;
                canOpenVotingModal?: boolean | undefined;
                canContinue?: boolean | undefined;
                activeTimer?: {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                } | null | undefined;
                voteModalOpen?: boolean | undefined;
                lastEliminated?: string | undefined;
                winners?: string[] | undefined;
                resolutionNote?: string | undefined;
                roundRules?: {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                } | undefined;
            } | undefined;
        }, {
            round?: number | undefined;
            phase?: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended" | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            categoryOrder?: string[] | undefined;
            lastStageText?: string | undefined;
            worldEvent?: {
                type: "bunker_revealed";
                index: number;
                round: number;
            } | undefined;
            postGame?: {
                isActive: boolean;
                enteredAt: number;
                outcome?: "survived" | "failed" | undefined;
                decidedBy?: string | undefined;
                decidedAt?: number | undefined;
            } | undefined;
            you?: {
                playerId: string;
                name: string;
                categories: {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }[];
                hand: {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                specialConditions: {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }[];
            } | undefined;
            public?: {
                players: {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }[];
                revealedThisRound: string[];
                voting?: {
                    hasVoted: boolean;
                } | undefined;
                roundRevealedCount?: number | undefined;
                roundTotalAlive?: number | undefined;
                currentTurnPlayerId?: string | null | undefined;
                votesRemainingInRound?: number | undefined;
                votesTotalThisRound?: number | undefined;
                revealLimit?: number | undefined;
                votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
                votesPublic?: {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }[] | undefined;
                votingProgress?: {
                    voted: number;
                    total: number;
                } | undefined;
                threatModifier?: {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                } | undefined;
                canOpenVotingModal?: boolean | undefined;
                canContinue?: boolean | undefined;
                activeTimer?: {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                } | null | undefined;
                voteModalOpen?: boolean | undefined;
                lastEliminated?: string | undefined;
                winners?: string[] | undefined;
                resolutionNote?: string | undefined;
                roundRules?: {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                } | undefined;
            } | undefined;
        }>>;
    }, "strip", z.ZodTypeAny, {
        roomState?: {
            players?: {
                playerId: string;
                name: string;
                connected: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[] | undefined;
            roomCode?: string | undefined;
            hostId?: string | undefined;
            controlId?: string | undefined;
            phase?: "lobby" | "game" | undefined;
            scenarioMeta?: {
                name: string;
                id: string;
                description?: string | undefined;
                devOnly?: boolean | undefined;
            } | undefined;
            settings?: {
                enableRevealDiscussionTimer: boolean;
                revealDiscussionSeconds: number;
                enablePreVoteDiscussionTimer: boolean;
                preVoteDiscussionSeconds: number;
                enablePostVoteDiscussionTimer: boolean;
                postVoteDiscussionSeconds: number;
                enablePresenterMode: boolean;
                continuePermission: "host_only" | "revealer_only" | "anyone";
                revealTimeoutAction: "random_card" | "skip_player";
                revealsBeforeVoting: number;
                specialUsage: "anytime" | "only_during_voting";
                maxPlayers: number;
                finalThreatReveal: "anyone" | "host";
            } | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            rulesOverriddenByHost?: boolean | undefined;
            rulesPresetCount?: number | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            isDev?: boolean | undefined;
        } | undefined;
        gameView?: {
            round?: number | undefined;
            phase?: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended" | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            categoryOrder?: string[] | undefined;
            lastStageText?: string | undefined;
            worldEvent?: {
                type: "bunker_revealed";
                index: number;
                round: number;
            } | undefined;
            postGame?: {
                isActive: boolean;
                enteredAt: number;
                outcome?: "survived" | "failed" | undefined;
                decidedBy?: string | undefined;
                decidedAt?: number | undefined;
            } | undefined;
            you?: {
                playerId: string;
                name: string;
                categories: {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }[];
                hand: {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                specialConditions: {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }[];
            } | undefined;
            public?: {
                players: {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }[];
                revealedThisRound: string[];
                voting?: {
                    hasVoted: boolean;
                } | undefined;
                roundRevealedCount?: number | undefined;
                roundTotalAlive?: number | undefined;
                currentTurnPlayerId?: string | null | undefined;
                votesRemainingInRound?: number | undefined;
                votesTotalThisRound?: number | undefined;
                revealLimit?: number | undefined;
                votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
                votesPublic?: {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }[] | undefined;
                votingProgress?: {
                    voted: number;
                    total: number;
                } | undefined;
                threatModifier?: {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                } | undefined;
                canOpenVotingModal?: boolean | undefined;
                canContinue?: boolean | undefined;
                activeTimer?: {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                } | null | undefined;
                voteModalOpen?: boolean | undefined;
                lastEliminated?: string | undefined;
                winners?: string[] | undefined;
                resolutionNote?: string | undefined;
                roundRules?: {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                } | undefined;
            } | undefined;
        } | undefined;
    }, {
        roomState?: {
            players?: {
                playerId: string;
                name: string;
                connected: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[] | undefined;
            roomCode?: string | undefined;
            hostId?: string | undefined;
            controlId?: string | undefined;
            phase?: "lobby" | "game" | undefined;
            scenarioMeta?: {
                name: string;
                id: string;
                description?: string | undefined;
                devOnly?: boolean | undefined;
            } | undefined;
            settings?: {
                enableRevealDiscussionTimer: boolean;
                revealDiscussionSeconds: number;
                enablePreVoteDiscussionTimer: boolean;
                preVoteDiscussionSeconds: number;
                enablePostVoteDiscussionTimer: boolean;
                postVoteDiscussionSeconds: number;
                enablePresenterMode: boolean;
                continuePermission: "host_only" | "revealer_only" | "anyone";
                revealTimeoutAction: "random_card" | "skip_player";
                revealsBeforeVoting: number;
                specialUsage: "anytime" | "only_during_voting";
                maxPlayers: number;
                finalThreatReveal: "anyone" | "host";
            } | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            rulesOverriddenByHost?: boolean | undefined;
            rulesPresetCount?: number | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            isDev?: boolean | undefined;
        } | undefined;
        gameView?: {
            round?: number | undefined;
            phase?: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended" | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            categoryOrder?: string[] | undefined;
            lastStageText?: string | undefined;
            worldEvent?: {
                type: "bunker_revealed";
                index: number;
                round: number;
            } | undefined;
            postGame?: {
                isActive: boolean;
                enteredAt: number;
                outcome?: "survived" | "failed" | undefined;
                decidedBy?: string | undefined;
                decidedAt?: number | undefined;
            } | undefined;
            you?: {
                playerId: string;
                name: string;
                categories: {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }[];
                hand: {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                specialConditions: {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }[];
            } | undefined;
            public?: {
                players: {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }[];
                revealedThisRound: string[];
                voting?: {
                    hasVoted: boolean;
                } | undefined;
                roundRevealedCount?: number | undefined;
                roundTotalAlive?: number | undefined;
                currentTurnPlayerId?: string | null | undefined;
                votesRemainingInRound?: number | undefined;
                votesTotalThisRound?: number | undefined;
                revealLimit?: number | undefined;
                votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
                votesPublic?: {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }[] | undefined;
                votingProgress?: {
                    voted: number;
                    total: number;
                } | undefined;
                threatModifier?: {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                } | undefined;
                canOpenVotingModal?: boolean | undefined;
                canContinue?: boolean | undefined;
                activeTimer?: {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                } | null | undefined;
                voteModalOpen?: boolean | undefined;
                lastEliminated?: string | undefined;
                winners?: string[] | undefined;
                resolutionNote?: string | undefined;
                roundRules?: {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                } | undefined;
            } | undefined;
        } | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "statePatch";
    payload: {
        roomState?: {
            players?: {
                playerId: string;
                name: string;
                connected: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[] | undefined;
            roomCode?: string | undefined;
            hostId?: string | undefined;
            controlId?: string | undefined;
            phase?: "lobby" | "game" | undefined;
            scenarioMeta?: {
                name: string;
                id: string;
                description?: string | undefined;
                devOnly?: boolean | undefined;
            } | undefined;
            settings?: {
                enableRevealDiscussionTimer: boolean;
                revealDiscussionSeconds: number;
                enablePreVoteDiscussionTimer: boolean;
                preVoteDiscussionSeconds: number;
                enablePostVoteDiscussionTimer: boolean;
                postVoteDiscussionSeconds: number;
                enablePresenterMode: boolean;
                continuePermission: "host_only" | "revealer_only" | "anyone";
                revealTimeoutAction: "random_card" | "skip_player";
                revealsBeforeVoting: number;
                specialUsage: "anytime" | "only_during_voting";
                maxPlayers: number;
                finalThreatReveal: "anyone" | "host";
            } | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            rulesOverriddenByHost?: boolean | undefined;
            rulesPresetCount?: number | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            isDev?: boolean | undefined;
        } | undefined;
        gameView?: {
            round?: number | undefined;
            phase?: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended" | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals: number;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            categoryOrder?: string[] | undefined;
            lastStageText?: string | undefined;
            worldEvent?: {
                type: "bunker_revealed";
                index: number;
                round: number;
            } | undefined;
            postGame?: {
                isActive: boolean;
                enteredAt: number;
                outcome?: "survived" | "failed" | undefined;
                decidedBy?: string | undefined;
                decidedAt?: number | undefined;
            } | undefined;
            you?: {
                playerId: string;
                name: string;
                categories: {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }[];
                hand: {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                specialConditions: {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }[];
            } | undefined;
            public?: {
                players: {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }[];
                revealedThisRound: string[];
                voting?: {
                    hasVoted: boolean;
                } | undefined;
                roundRevealedCount?: number | undefined;
                roundTotalAlive?: number | undefined;
                currentTurnPlayerId?: string | null | undefined;
                votesRemainingInRound?: number | undefined;
                votesTotalThisRound?: number | undefined;
                revealLimit?: number | undefined;
                votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
                votesPublic?: {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }[] | undefined;
                votingProgress?: {
                    voted: number;
                    total: number;
                } | undefined;
                threatModifier?: {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                } | undefined;
                canOpenVotingModal?: boolean | undefined;
                canContinue?: boolean | undefined;
                activeTimer?: {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                } | null | undefined;
                voteModalOpen?: boolean | undefined;
                lastEliminated?: string | undefined;
                winners?: string[] | undefined;
                resolutionNote?: string | undefined;
                roundRules?: {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                } | undefined;
            } | undefined;
        } | undefined;
    };
}, {
    type: "statePatch";
    payload: {
        roomState?: {
            players?: {
                playerId: string;
                name: string;
                connected: boolean;
                disconnectedAt?: number | undefined;
                totalAbsentMs?: number | undefined;
                currentOfflineMs?: number | undefined;
                kickRemainingMs?: number | undefined;
                leftBunker?: boolean | undefined;
            }[] | undefined;
            roomCode?: string | undefined;
            hostId?: string | undefined;
            controlId?: string | undefined;
            phase?: "lobby" | "game" | undefined;
            scenarioMeta?: {
                name: string;
                id: string;
                description?: string | undefined;
                devOnly?: boolean | undefined;
            } | undefined;
            settings?: {
                enableRevealDiscussionTimer: boolean;
                revealDiscussionSeconds: number;
                enablePreVoteDiscussionTimer: boolean;
                preVoteDiscussionSeconds: number;
                enablePostVoteDiscussionTimer: boolean;
                postVoteDiscussionSeconds: number;
                enablePresenterMode: boolean;
                continuePermission: "host_only" | "revealer_only" | "anyone";
                revealTimeoutAction: "random_card" | "skip_player";
                revealsBeforeVoting: number;
                specialUsage: "anytime" | "only_during_voting";
                maxPlayers: number;
                finalThreatReveal: "anyone" | "host";
            } | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            rulesOverriddenByHost?: boolean | undefined;
            rulesPresetCount?: number | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            isDev?: boolean | undefined;
        } | undefined;
        gameView?: {
            round?: number | undefined;
            phase?: "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended" | undefined;
            ruleset?: {
                playerCount: number;
                votesPerRound: number[];
                totalExiles: number;
                bunkerSeats: number;
                rulesetMode: "auto" | "preset" | "manual";
                manualConfig?: {
                    bunkerSlots: number;
                    votesByRound: number[];
                    targetReveals?: number | undefined;
                    seedTemplatePlayers?: number | undefined;
                } | undefined;
            } | undefined;
            world?: {
                bunker: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                disaster: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    text?: string | undefined;
                    imageId?: string | undefined;
                };
                threats: {
                    id: string;
                    description: string;
                    kind: "bunker" | "disaster" | "threat";
                    title: string;
                    isRevealed: boolean;
                    text?: string | undefined;
                    imageId?: string | undefined;
                    revealedAtRound?: number | undefined;
                    revealedBy?: string | undefined;
                }[];
                counts: {
                    bunker: number;
                    threats: number;
                };
            } | undefined;
            categoryOrder?: string[] | undefined;
            lastStageText?: string | undefined;
            worldEvent?: {
                type: "bunker_revealed";
                index: number;
                round: number;
            } | undefined;
            postGame?: {
                isActive: boolean;
                enteredAt: number;
                outcome?: "survived" | "failed" | undefined;
                decidedBy?: string | undefined;
                decidedAt?: number | undefined;
            } | undefined;
            you?: {
                playerId: string;
                name: string;
                categories: {
                    category: string;
                    cards: {
                        revealed: boolean;
                        instanceId: string;
                        labelShort: string;
                    }[];
                }[];
                hand: {
                    revealed: boolean;
                    id: string;
                    deck: string;
                    instanceId?: string | undefined;
                    labelShort?: string | undefined;
                    secret?: boolean | undefined;
                    missing?: boolean | undefined;
                }[];
                specialConditions: {
                    id: string;
                    title: string;
                    text: string;
                    instanceId: string;
                    trigger: "active" | "onVote" | "onOwnerEliminated" | "onRevealOrActive" | "secret_onEliminate";
                    effect: {
                        type: string;
                        params?: Record<string, any> | undefined;
                    };
                    implemented: boolean;
                    revealedPublic: boolean;
                    used: boolean;
                    imgUrl?: string | undefined;
                    needsChoice?: boolean | undefined;
                    choiceKind?: "neighbor" | "player" | "category" | "none" | undefined;
                    allowSelfTarget?: boolean | undefined;
                    targetScope?: "neighbors" | "any_alive" | "self" | "any_including_self" | undefined;
                }[];
            } | undefined;
            public?: {
                players: {
                    playerId: string;
                    name: string;
                    connected: boolean;
                    status: "alive" | "eliminated" | "left_bunker";
                    categories: {
                        category: string;
                        status: "hidden" | "revealed";
                        cards: {
                            labelShort: string;
                            imgUrl?: string | undefined;
                        }[];
                    }[];
                    revealedCards: {
                        id: string;
                        deck: string;
                        instanceId?: string | undefined;
                        labelShort?: string | undefined;
                        secret?: boolean | undefined;
                        missing?: boolean | undefined;
                    }[];
                    revealedCount: number;
                    totalCards: number;
                    specialRevealed: boolean;
                    disconnectedAt?: number | undefined;
                    totalAbsentMs?: number | undefined;
                    currentOfflineMs?: number | undefined;
                    kickRemainingMs?: number | undefined;
                    leftBunker?: boolean | undefined;
                }[];
                revealedThisRound: string[];
                voting?: {
                    hasVoted: boolean;
                } | undefined;
                roundRevealedCount?: number | undefined;
                roundTotalAlive?: number | undefined;
                currentTurnPlayerId?: string | null | undefined;
                votesRemainingInRound?: number | undefined;
                votesTotalThisRound?: number | undefined;
                revealLimit?: number | undefined;
                votePhase?: "voting" | "voteSpecialWindow" | "voteResolve" | null | undefined;
                votesPublic?: {
                    status: "voted" | "not_voted" | "invalid";
                    voterId: string;
                    voterName: string;
                    targetId?: string | undefined;
                    targetName?: string | undefined;
                    reason?: string | undefined;
                    submittedAt?: number | undefined;
                }[] | undefined;
                votingProgress?: {
                    voted: number;
                    total: number;
                } | undefined;
                threatModifier?: {
                    delta: number;
                    reasons: string[];
                    baseCount: number;
                    finalCount: number;
                } | undefined;
                canOpenVotingModal?: boolean | undefined;
                canContinue?: boolean | undefined;
                activeTimer?: {
                    kind: "reveal_discussion" | "pre_vote" | "post_vote";
                    endsAt: number;
                } | null | undefined;
                voteModalOpen?: boolean | undefined;
                lastEliminated?: string | undefined;
                winners?: string[] | undefined;
                resolutionNote?: string | undefined;
                roundRules?: {
                    noTalkUntilVoting?: boolean | undefined;
                    forcedRevealCategory?: string | undefined;
                } | undefined;
            } | undefined;
        } | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"gameEvent">;
    payload: z.ZodObject<{
        id: z.ZodString;
        kind: z.ZodUnion<[z.ZodLiteral<"roundStart">, z.ZodLiteral<"votingStart">, z.ZodLiteral<"elimination">, z.ZodLiteral<"gameEnd">, z.ZodLiteral<"info">, z.ZodLiteral<"playerDisconnected">, z.ZodLiteral<"playerReconnected">, z.ZodLiteral<"playerLeftBunker">]>;
        message: z.ZodString;
        createdAt: z.ZodNumber;
    }, "strip", z.ZodTypeAny, {
        message: string;
        id: string;
        kind: "roundStart" | "votingStart" | "elimination" | "gameEnd" | "info" | "playerDisconnected" | "playerReconnected" | "playerLeftBunker";
        createdAt: number;
    }, {
        message: string;
        id: string;
        kind: "roundStart" | "votingStart" | "elimination" | "gameEnd" | "info" | "playerDisconnected" | "playerReconnected" | "playerLeftBunker";
        createdAt: number;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "gameEvent";
    payload: {
        message: string;
        id: string;
        kind: "roundStart" | "votingStart" | "elimination" | "gameEnd" | "info" | "playerDisconnected" | "playerReconnected" | "playerLeftBunker";
        createdAt: number;
    };
}, {
    type: "gameEvent";
    payload: {
        message: string;
        id: string;
        kind: "roundStart" | "votingStart" | "elimination" | "gameEnd" | "info" | "playerDisconnected" | "playerReconnected" | "playerLeftBunker";
        createdAt: number;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"error">;
    payload: z.ZodObject<{
        message: z.ZodString;
        code: z.ZodOptional<z.ZodString>;
        maxPlayers: z.ZodOptional<z.ZodNumber>;
    }, "strip", z.ZodTypeAny, {
        message: string;
        code?: string | undefined;
        maxPlayers?: number | undefined;
    }, {
        message: string;
        code?: string | undefined;
        maxPlayers?: number | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "error";
    payload: {
        message: string;
        code?: string | undefined;
        maxPlayers?: number | undefined;
    };
}, {
    type: "error";
    payload: {
        message: string;
        code?: string | undefined;
        maxPlayers?: number | undefined;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"helloAck">;
    payload: z.ZodObject<{
        playerId: z.ZodString;
        playerToken: z.ZodString;
    }, "strip", z.ZodTypeAny, {
        playerId: string;
        playerToken: string;
    }, {
        playerId: string;
        playerToken: string;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "helloAck";
    payload: {
        playerId: string;
        playerToken: string;
    };
}, {
    type: "helloAck";
    payload: {
        playerId: string;
        playerToken: string;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"hostChanged">;
    payload: z.ZodObject<{
        newHostId: z.ZodString;
        reason: z.ZodUnion<[z.ZodLiteral<"disconnect_timeout">, z.ZodLiteral<"left_bunker">, z.ZodLiteral<"eliminated">, z.ZodLiteral<"manual">]>;
    }, "strip", z.ZodTypeAny, {
        reason: "manual" | "eliminated" | "left_bunker" | "disconnect_timeout";
        newHostId: string;
    }, {
        reason: "manual" | "eliminated" | "left_bunker" | "disconnect_timeout";
        newHostId: string;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "hostChanged";
    payload: {
        reason: "manual" | "eliminated" | "left_bunker" | "disconnect_timeout";
        newHostId: string;
    };
}, {
    type: "hostChanged";
    payload: {
        reason: "manual" | "eliminated" | "left_bunker" | "disconnect_timeout";
        newHostId: string;
    };
}>, z.ZodObject<{
    type: z.ZodLiteral<"pong">;
    payload: z.ZodOptional<z.ZodObject<{}, "strip", z.ZodTypeAny, {}, {}>>;
}, "strip", z.ZodTypeAny, {
    type: "pong";
    payload?: {} | undefined;
}, {
    type: "pong";
    payload?: {} | undefined;
}>, z.ZodObject<{
    type: z.ZodLiteral<"overlayState">;
    payload: z.ZodObject<{
        ok: z.ZodBoolean;
        unauthorized: z.ZodOptional<z.ZodBoolean>;
        roomCode: z.ZodOptional<z.ZodString>;
        state: z.ZodOptional<z.ZodObject<{
            roomId: z.ZodString;
            playerCount: z.ZodNumber;
            top: z.ZodObject<{
                bunker: z.ZodObject<{
                    revealed: z.ZodNumber;
                    total: z.ZodNumber;
                    lines: z.ZodArray<z.ZodString, "many">;
                    items: z.ZodOptional<z.ZodArray<z.ZodObject<{
                        title: z.ZodString;
                        subtitle: z.ZodOptional<z.ZodString>;
                        imageId: z.ZodOptional<z.ZodString>;
                    }, "strip", z.ZodTypeAny, {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }, {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }>, "many">>;
                }, "strip", z.ZodTypeAny, {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                }, {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                }>;
                catastrophe: z.ZodObject<{
                    text: z.ZodString;
                    title: z.ZodOptional<z.ZodString>;
                    imageId: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                }, {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                }>;
                threats: z.ZodObject<{
                    revealed: z.ZodNumber;
                    total: z.ZodNumber;
                    lines: z.ZodArray<z.ZodString, "many">;
                    items: z.ZodOptional<z.ZodArray<z.ZodObject<{
                        title: z.ZodString;
                        subtitle: z.ZodOptional<z.ZodString>;
                        imageId: z.ZodOptional<z.ZodString>;
                    }, "strip", z.ZodTypeAny, {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }, {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }>, "many">>;
                }, "strip", z.ZodTypeAny, {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                }, {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                }>;
            }, "strip", z.ZodTypeAny, {
                bunker: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                threats: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                catastrophe: {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                };
            }, {
                bunker: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                threats: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                catastrophe: {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                };
            }>;
            players: z.ZodArray<z.ZodObject<{
                id: z.ZodString;
                nickname: z.ZodString;
                connected: z.ZodOptional<z.ZodBoolean>;
                alive: z.ZodBoolean;
                tags: z.ZodObject<{
                    sex: z.ZodObject<{
                        label: z.ZodString;
                        revealed: z.ZodBoolean;
                        value: z.ZodString;
                    }, "strip", z.ZodTypeAny, {
                        revealed: boolean;
                        value: string;
                        label: string;
                    }, {
                        revealed: boolean;
                        value: string;
                        label: string;
                    }>;
                    age: z.ZodObject<{
                        label: z.ZodString;
                        revealed: z.ZodBoolean;
                        value: z.ZodString;
                    }, "strip", z.ZodTypeAny, {
                        revealed: boolean;
                        value: string;
                        label: string;
                    }, {
                        revealed: boolean;
                        value: string;
                        label: string;
                    }>;
                    orientation: z.ZodObject<{
                        label: z.ZodString;
                        revealed: z.ZodBoolean;
                        value: z.ZodString;
                    }, "strip", z.ZodTypeAny, {
                        revealed: boolean;
                        value: string;
                        label: string;
                    }, {
                        revealed: boolean;
                        value: string;
                        label: string;
                    }>;
                }, "strip", z.ZodTypeAny, {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                }, {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                }>;
                categories: z.ZodArray<z.ZodObject<{
                    key: z.ZodString;
                    label: z.ZodString;
                    revealed: z.ZodBoolean;
                    value: z.ZodString;
                    imgUrl: z.ZodOptional<z.ZodString>;
                }, "strip", z.ZodTypeAny, {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }, {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }>, "many">;
            }, "strip", z.ZodTypeAny, {
                alive: boolean;
                id: string;
                categories: {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }[];
                nickname: string;
                tags: {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                };
                connected?: boolean | undefined;
            }, {
                alive: boolean;
                id: string;
                categories: {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }[];
                nickname: string;
                tags: {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                };
                connected?: boolean | undefined;
            }>, "many">;
            overrides: z.ZodOptional<z.ZodObject<{
                enabled: z.ZodOptional<z.ZodObject<{
                    topBunker: z.ZodOptional<z.ZodBoolean>;
                    topCatastrophe: z.ZodOptional<z.ZodBoolean>;
                    topThreats: z.ZodOptional<z.ZodBoolean>;
                    playerNames: z.ZodOptional<z.ZodBoolean>;
                    playerTraits: z.ZodOptional<z.ZodBoolean>;
                    playerCategories: z.ZodOptional<z.ZodBoolean>;
                }, "strip", z.ZodTypeAny, {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                }, {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                }>>;
                top: z.ZodOptional<z.ZodObject<{
                    bunkerLines: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
                    catastropheText: z.ZodOptional<z.ZodString>;
                    threatsLines: z.ZodOptional<z.ZodArray<z.ZodString, "many">>;
                }, "strip", z.ZodTypeAny, {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                }, {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                }>>;
                players: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodObject<{
                    name: z.ZodOptional<z.ZodString>;
                    traits: z.ZodOptional<z.ZodObject<{
                        sex: z.ZodOptional<z.ZodString>;
                        age: z.ZodOptional<z.ZodString>;
                        orient: z.ZodOptional<z.ZodString>;
                    }, "strip", z.ZodTypeAny, {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    }, {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    }>>;
                    categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodString>>;
                    enabled: z.ZodOptional<z.ZodObject<{
                        name: z.ZodOptional<z.ZodBoolean>;
                        traits: z.ZodOptional<z.ZodBoolean>;
                        categories: z.ZodOptional<z.ZodRecord<z.ZodString, z.ZodBoolean>>;
                    }, "strip", z.ZodTypeAny, {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    }, {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    }>>;
                }, "strip", z.ZodTypeAny, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }>>>;
                extraTexts: z.ZodOptional<z.ZodArray<z.ZodObject<{
                    id: z.ZodString;
                    text: z.ZodString;
                    x: z.ZodNumber;
                    y: z.ZodNumber;
                    align: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"left">, z.ZodLiteral<"center">, z.ZodLiteral<"right">]>>;
                    size: z.ZodOptional<z.ZodNumber>;
                    color: z.ZodOptional<z.ZodString>;
                    shadow: z.ZodOptional<z.ZodBoolean>;
                    visible: z.ZodOptional<z.ZodBoolean>;
                }, "strip", z.ZodTypeAny, {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }, {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }>, "many">>;
            }, "strip", z.ZodTypeAny, {
                enabled?: {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                } | undefined;
                top?: {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                } | undefined;
                players?: Record<string, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }> | undefined;
                extraTexts?: {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }[] | undefined;
            }, {
                enabled?: {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                } | undefined;
                top?: {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                } | undefined;
                players?: Record<string, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }> | undefined;
                extraTexts?: {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }[] | undefined;
            }>>;
        }, "strip", z.ZodTypeAny, {
            playerCount: number;
            top: {
                bunker: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                threats: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                catastrophe: {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                };
            };
            players: {
                alive: boolean;
                id: string;
                categories: {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }[];
                nickname: string;
                tags: {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                };
                connected?: boolean | undefined;
            }[];
            roomId: string;
            overrides?: {
                enabled?: {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                } | undefined;
                top?: {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                } | undefined;
                players?: Record<string, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }> | undefined;
                extraTexts?: {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }[] | undefined;
            } | undefined;
        }, {
            playerCount: number;
            top: {
                bunker: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                threats: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                catastrophe: {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                };
            };
            players: {
                alive: boolean;
                id: string;
                categories: {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }[];
                nickname: string;
                tags: {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                };
                connected?: boolean | undefined;
            }[];
            roomId: string;
            overrides?: {
                enabled?: {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                } | undefined;
                top?: {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                } | undefined;
                players?: Record<string, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }> | undefined;
                extraTexts?: {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }[] | undefined;
            } | undefined;
        }>>;
        presenter: z.ZodOptional<z.ZodAny>;
        presenterModeEnabled: z.ZodOptional<z.ZodBoolean>;
        role: z.ZodOptional<z.ZodUnion<[z.ZodLiteral<"VIEW">, z.ZodLiteral<"PLAYER">, z.ZodLiteral<"CONTROL">]>>;
        message: z.ZodOptional<z.ZodString>;
    }, "strip", z.ZodTypeAny, {
        ok: boolean;
        message?: string | undefined;
        roomCode?: string | undefined;
        unauthorized?: boolean | undefined;
        state?: {
            playerCount: number;
            top: {
                bunker: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                threats: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                catastrophe: {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                };
            };
            players: {
                alive: boolean;
                id: string;
                categories: {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }[];
                nickname: string;
                tags: {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                };
                connected?: boolean | undefined;
            }[];
            roomId: string;
            overrides?: {
                enabled?: {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                } | undefined;
                top?: {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                } | undefined;
                players?: Record<string, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }> | undefined;
                extraTexts?: {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }[] | undefined;
            } | undefined;
        } | undefined;
        presenter?: any;
        presenterModeEnabled?: boolean | undefined;
        role?: "VIEW" | "PLAYER" | "CONTROL" | undefined;
    }, {
        ok: boolean;
        message?: string | undefined;
        roomCode?: string | undefined;
        unauthorized?: boolean | undefined;
        state?: {
            playerCount: number;
            top: {
                bunker: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                threats: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                catastrophe: {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                };
            };
            players: {
                alive: boolean;
                id: string;
                categories: {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }[];
                nickname: string;
                tags: {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                };
                connected?: boolean | undefined;
            }[];
            roomId: string;
            overrides?: {
                enabled?: {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                } | undefined;
                top?: {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                } | undefined;
                players?: Record<string, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }> | undefined;
                extraTexts?: {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }[] | undefined;
            } | undefined;
        } | undefined;
        presenter?: any;
        presenterModeEnabled?: boolean | undefined;
        role?: "VIEW" | "PLAYER" | "CONTROL" | undefined;
    }>;
}, "strip", z.ZodTypeAny, {
    type: "overlayState";
    payload: {
        ok: boolean;
        message?: string | undefined;
        roomCode?: string | undefined;
        unauthorized?: boolean | undefined;
        state?: {
            playerCount: number;
            top: {
                bunker: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                threats: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                catastrophe: {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                };
            };
            players: {
                alive: boolean;
                id: string;
                categories: {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }[];
                nickname: string;
                tags: {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                };
                connected?: boolean | undefined;
            }[];
            roomId: string;
            overrides?: {
                enabled?: {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                } | undefined;
                top?: {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                } | undefined;
                players?: Record<string, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }> | undefined;
                extraTexts?: {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }[] | undefined;
            } | undefined;
        } | undefined;
        presenter?: any;
        presenterModeEnabled?: boolean | undefined;
        role?: "VIEW" | "PLAYER" | "CONTROL" | undefined;
    };
}, {
    type: "overlayState";
    payload: {
        ok: boolean;
        message?: string | undefined;
        roomCode?: string | undefined;
        unauthorized?: boolean | undefined;
        state?: {
            playerCount: number;
            top: {
                bunker: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                threats: {
                    revealed: number;
                    total: number;
                    lines: string[];
                    items?: {
                        title: string;
                        imageId?: string | undefined;
                        subtitle?: string | undefined;
                    }[] | undefined;
                };
                catastrophe: {
                    text: string;
                    title?: string | undefined;
                    imageId?: string | undefined;
                };
            };
            players: {
                alive: boolean;
                id: string;
                categories: {
                    revealed: boolean;
                    value: string;
                    label: string;
                    key: string;
                    imgUrl?: string | undefined;
                }[];
                nickname: string;
                tags: {
                    sex: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    age: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                    orientation: {
                        revealed: boolean;
                        value: string;
                        label: string;
                    };
                };
                connected?: boolean | undefined;
            }[];
            roomId: string;
            overrides?: {
                enabled?: {
                    topBunker?: boolean | undefined;
                    topCatastrophe?: boolean | undefined;
                    topThreats?: boolean | undefined;
                    playerNames?: boolean | undefined;
                    playerTraits?: boolean | undefined;
                    playerCategories?: boolean | undefined;
                } | undefined;
                top?: {
                    bunkerLines?: string[] | undefined;
                    catastropheText?: string | undefined;
                    threatsLines?: string[] | undefined;
                } | undefined;
                players?: Record<string, {
                    name?: string | undefined;
                    traits?: {
                        sex?: string | undefined;
                        age?: string | undefined;
                        orient?: string | undefined;
                    } | undefined;
                    categories?: Record<string, string> | undefined;
                    enabled?: {
                        name?: boolean | undefined;
                        traits?: boolean | undefined;
                        categories?: Record<string, boolean> | undefined;
                    } | undefined;
                }> | undefined;
                extraTexts?: {
                    id: string;
                    text: string;
                    x: number;
                    y: number;
                    align?: "left" | "right" | "center" | undefined;
                    size?: number | undefined;
                    color?: string | undefined;
                    shadow?: boolean | undefined;
                    visible?: boolean | undefined;
                }[] | undefined;
            } | undefined;
        } | undefined;
        presenter?: any;
        presenterModeEnabled?: boolean | undefined;
        role?: "VIEW" | "PLAYER" | "CONTROL" | undefined;
    };
}>]>;
export type ClientMessage = z.infer<typeof ClientMessageSchema>;
export type ServerMessage = z.infer<typeof ServerMessageSchema>;
export type ClientHelloPayload = z.infer<typeof ClientHelloSchema>;
export { formatLabelShort } from "./labelFormat.js";
export { getRulesetForPlayerCount, RULESET_PRESET_COUNTS, RULESET_TABLE } from "./ruleset.js";
export { buildLinkSet, normalizeBase, LINK_PATHS } from "./urlBuilder.js";
export type { BuildLinkSetInput, BuiltLinkSet, UrlPair } from "./urlBuilder.js";
//# sourceMappingURL=index.d.ts.map