import { z } from "zod";
export * from "./targeting.js";
export * from "./targetingLocale.js";
export type RoomPhase = "lobby" | "game";
export type ScenarioPhase = "reveal" | "reveal_discussion" | "voting" | "resolution" | "ended";
export type VotePhase = "voting" | "voteSpecialWindow" | "voteResolve";
export type PlayerStatus = "alive" | "eliminated" | "left_bunker";
export type GameEventKind =
  | "roundStart"
  | "votingStart"
  | "elimination"
  | "gameEnd"
  | "info"
  | "playerDisconnected"
  | "playerReconnected"
  | "playerLeftBunker";
export type SpecialConditionTrigger =
  | "active"
  | "onVote"
  | "onOwnerEliminated"
  | "onRevealOrActive"
  | "secret_onEliminate";
export type GameTimerKind = "reveal_discussion" | "pre_vote" | "post_vote";
export type ContinuePermission = "host_only" | "revealer_only" | "anyone";
export type RevealTimeoutAction = "random_card" | "skip_player";
export type SpecialUsageMode = "anytime" | "only_during_voting";
export type FinalThreatReveal = "host" | "anyone";
export type AutomationMode = "auto" | "semi" | "manual";
export type CardLocale = "ru" | "en";
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
  imgUrl?: string;
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
  automationMode: AutomationMode;
  continuePermission: ContinuePermission;
  revealTimeoutAction: RevealTimeoutAction;
  revealsBeforeVoting: number;
  specialUsage: SpecialUsageMode;
  maxPlayers: number;
  finalThreatReveal: FinalThreatReveal;
  forcedDisasterId: string;
  cardLocale?: CardLocale;
}

export interface ScenarioMeta {
  id: string;
  name: string;
  description?: string;
  devOnly?: boolean;
}

export interface CardRef {
  id: string; // Relative path under /assets (e.g. decks/Профессия/card1.jpg)
  deck: string;
  instanceId?: string;
  labelShort?: string;
  imgUrl?: string;
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
  choiceKind?: "player" | "neighbor" | "category" | "bunker" | "special" | "none";
  pendingActivation?: boolean;
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
  instanceId?: string;
  hidden?: boolean;
  backCategory?: string;
}

export interface YouCategoryCard {
  instanceId: string;
  labelShort: string;
  deck?: string;
  revealed: boolean;
  imgUrl?: string;
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
  disasterOptions?: Array<{ id: string; title: string }>;
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

export type LocalizedVars = Record<string, string | number>;

export type VoteReasonCode =
  | "VOTE_BLOCKED"
  | "VOTE_FORCED_SELF"
  | "VOTE_SPENT"
  | "VOTE_TARGET_DISALLOWED"
  | "VOTE_TARGET_UNAVAILABLE"
  | "VOTE_BANNED_AGAINST_TARGET";

export interface ThreatModifierView {
  delta: number;
  reasons: string[];
  reasonCardIds?: string[];
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
  reasonKey?: string;
  reasonVars?: LocalizedVars;
  reasonCode?: VoteReasonCode;
  weight?: number;
  submittedAt?: number;
}

export interface GameEvent {
  id: string;
  kind: GameEventKind;
  message: string;
  messageKey?: string;
  messageVars?: LocalizedVars;
  createdAt: number;
}

export interface GameView {
  phase: ScenarioPhase;
  round: number;
  categoryOrder: string[];
  lastStageText?: string;
  lastStageTextKey?: string;
  lastStageTextVars?: LocalizedVars;
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
    yourVoteWeight?: number;
    votesRemainingInRound?: number;
    votesTotalThisRound?: number;
    revealLimit?: number;
    voting?: VotingView;
    votePhase?: VotePhase | null;
    votesPublic?: VotePublic[];
    votingProgress?: VotingProgress;
    disallowedVoteTargetIdsForYou?: string[];
    threatModifier?: ThreatModifierView;
    canOpenVotingModal?: boolean;
    canContinue?: boolean;
    activeTimer?: GameTimerState | null;
    voteModalOpen?: boolean;
    lastEliminated?: string;
    winners?: string[];
    resolutionNote?: string;
    resolutionNoteKey?: string;
    resolutionNoteVars?: LocalizedVars;
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
  deckId?: string;
  cardId?: string;
  locale?: string;
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

export type ScenarioAction =
  | { type: "revealCard"; payload: { cardId: string } }
  | { type: "vote"; payload: { targetPlayerId: string } }
  | { type: "finalizeVoting"; payload: {} }
  | {
      type: "applySpecial";
      payload: { specialInstanceId: string; payload?: Record<string, unknown> };
    }
  | { type: "revealWorldThreat"; payload: { index: number } }
  | { type: "setBunkerOutcome"; payload: { outcome: PostGameOutcome } }
  | { type: "devSkipRound"; payload: {} }
  | { type: "devKickPlayer"; payload: { targetPlayerId: string } }
  | { type: "markLeftBunker"; payload: { targetPlayerId: string } }
  | { type: "continueRound"; payload: {} }
  | { type: "devAddPlayer"; payload: { name?: string } }
  | { type: "devRemovePlayer"; payload: { targetPlayerId?: string } }
  | {
      type: "adminReplacePlayerCard";
      payload: {
        targetPlayerId: string;
        cardInstanceId: string;
        targetArea?: "hand" | "special";
        replacementMode: "random" | "specific";
        replacementCardId?: string;
      };
    }
  | {
      type: "adminSetWorldCardReveal";
      payload: {
        kind: "bunker" | "threat";
        index: number;
        revealed: boolean;
      };
    }
  | {
      type: "adminReplaceWorldCard";
      payload: {
        kind: "bunker" | "threat" | "disaster";
        index?: number;
        replacementMode: "random" | "specific";
        replacementCardId?: string;
      };
    }
  | {
      type: "adminSetWorldCount";
      payload: {
        kind: "bunker" | "threat";
        count: number;
      };
    }
  | {
      type: "adminApplySpecial";
      payload: {
        actorPlayerId: string;
        specialInstanceId?: string;
        specialId?: string;
        payload?: Record<string, unknown>;
      };
    };

export interface ScenarioActionResult {
  error?: string;
  errorKey?: string;
  errorVars?: LocalizedVars;
  stateChanged?: boolean;
}

export interface ControlSpecialCatalogEntry {
  id: string;
  title: string;
  text: string;
  implemented?: boolean;
  choiceKind?: "player" | "neighbor" | "category" | "bunker" | "special" | "none";
  targetScope?: SpecialTargetScope;
  allowSelfTarget?: boolean;
  effectType?: string;
  requires?: string[];
}

export interface ScenarioSession {
  getGameView(playerId: string): GameView;
  handleAction(playerId: string, action: ScenarioAction): ScenarioActionResult;
  getSpecialCatalog?(): ControlSpecialCatalogEntry[];
}

export interface ScenarioModule {
  meta: ScenarioMeta;
  createSession(ctx: ScenarioContext): ScenarioSession;
}

export const PlayerSummarySchema = z.object({
  playerId: z.string(),
  name: z.string(),
  connected: z.boolean(),
  disconnectedAt: z.number().int().nonnegative().optional(),
  totalAbsentMs: z.number().int().nonnegative().optional(),
  currentOfflineMs: z.number().int().nonnegative().optional(),
  kickRemainingMs: z.number().int().nonnegative().optional(),
  leftBunker: z.boolean().optional(),
});

export const ScenarioMetaSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().optional(),
  devOnly: z.boolean().optional(),
});

export const GameTimerStateSchema = z.object({
  kind: z.union([z.literal("reveal_discussion"), z.literal("pre_vote"), z.literal("post_vote")]),
  endsAt: z.number().int().nonnegative(),
});

export const ManualRulesConfigSchema = z.object({
  bunkerSlots: z.number().int().min(1).max(16),
  votesByRound: z.array(z.number().int().min(0).max(9)).min(1).max(64),
  targetReveals: z.number().int().min(5).max(7).default(7),
  seedTemplatePlayers: z.number().int().min(4).max(16).optional(),
});

export const GameRulesetSchema = z.object({
  playerCount: z.number().int().min(4).max(16),
  votesPerRound: z.array(z.number().int().min(0).max(9)).min(1).max(64),
  totalExiles: z.number().int().min(0),
  bunkerSeats: z.number().int().min(1),
  rulesetMode: z.union([z.literal("auto"), z.literal("preset"), z.literal("manual")]),
  manualConfig: ManualRulesConfigSchema.optional(),
});

export const WorldCardSchema = z.object({
  kind: z.union([z.literal("bunker"), z.literal("disaster"), z.literal("threat")]),
  id: z.string(),
  title: z.string(),
  description: z.string(),
  text: z.string().optional(),
  imageId: z.string().optional(),
});

export const WorldFacedCardSchema = WorldCardSchema.extend({
  isRevealed: z.boolean(),
  revealedAtRound: z.number().int().nonnegative().optional(),
  revealedBy: z.string().optional(),
});

export const WorldState30Schema = z.object({
  disaster: WorldCardSchema,
  bunker: z.array(WorldFacedCardSchema),
  threats: z.array(WorldFacedCardSchema),
  counts: z.object({
    bunker: z.number().int().nonnegative(),
    threats: z.number().int().nonnegative(),
  }),
});

export const WorldEventSchema = z.object({
  type: z.literal("bunker_revealed"),
  index: z.number().int().nonnegative(),
  round: z.number().int().nonnegative(),
});

export const PostGameStateSchema = z.object({
  isActive: z.boolean(),
  enteredAt: z.number().int().nonnegative(),
  outcome: z.union([z.literal("survived"), z.literal("failed")]).optional(),
  decidedBy: z.string().optional(),
  decidedAt: z.number().int().nonnegative().optional(),
});


export const GameSettingsSchema = z.object({
  enableRevealDiscussionTimer: z.boolean(),
  revealDiscussionSeconds: z.number().int().min(5).max(600),
  enablePreVoteDiscussionTimer: z.boolean(),
  preVoteDiscussionSeconds: z.number().int().min(5).max(600),
  enablePostVoteDiscussionTimer: z.boolean(),
  postVoteDiscussionSeconds: z.number().int().min(5).max(600),
  automationMode: z.union([z.literal("auto"), z.literal("semi"), z.literal("manual")]),
  continuePermission: z.union([
    z.literal("host_only"),
    z.literal("revealer_only"),
    z.literal("anyone"),
  ]),
  revealTimeoutAction: z.union([z.literal("random_card"), z.literal("skip_player")]),
  revealsBeforeVoting: z.number().int().min(1),
  specialUsage: z.union([z.literal("anytime"), z.literal("only_during_voting")]),
  maxPlayers: z.number().int().min(2),
  finalThreatReveal: z.union([z.literal("host"), z.literal("anyone")]),
  forcedDisasterId: z.string().max(256),
  cardLocale: z.union([z.literal("ru"), z.literal("en")]).optional(),
});

export const CardRefSchema = z.object({
  id: z.string(),
  deck: z.string(),
  instanceId: z.string().optional(),
  labelShort: z.string().optional(),
  imgUrl: z.string().optional(),
  secret: z.boolean().optional(),
  missing: z.boolean().optional(),
});

export const CardInHandSchema = CardRefSchema.extend({
  revealed: z.boolean(),
});

export const SpecialConditionEffectSchema = z.object({
  type: z.string(),
  params: z.record(z.any()).optional(),
});

export const SpecialConditionInstanceSchema = z.object({
  instanceId: z.string(),
  id: z.string(),
  title: z.string(),
  text: z.string(),
  trigger: z.union([
    z.literal("active"),
    z.literal("onVote"),
    z.literal("onOwnerEliminated"),
    z.literal("onRevealOrActive"),
    z.literal("secret_onEliminate"),
  ]),
  effect: SpecialConditionEffectSchema,
  implemented: z.boolean(),
  revealedPublic: z.boolean(),
  used: z.boolean(),
  imgUrl: z.string().optional(),
  needsChoice: z.boolean().optional(),
  choiceKind: z
    .union([
      z.literal("player"),
      z.literal("neighbor"),
      z.literal("category"),
      z.literal("bunker"),
      z.literal("special"),
      z.literal("none"),
    ])
    .optional(),
  pendingActivation: z.boolean().optional(),
  allowSelfTarget: z.boolean().optional(),
  targetScope: z
    .union([
      z.literal("neighbors"),
      z.literal("any_alive"),
      z.literal("self"),
      z.literal("any_including_self"),
    ])
    .optional(),
});

export const PublicCategoryCardSchema = z.object({
  labelShort: z.string(),
  imgUrl: z.string().optional(),
  instanceId: z.string().optional(),
  hidden: z.boolean().optional(),
  backCategory: z.string().optional(),
});

export const YouCategoryCardSchema = z.object({
  instanceId: z.string(),
  labelShort: z.string(),
  deck: z.string().optional(),
  revealed: z.boolean(),
  imgUrl: z.string().optional(),
});

export const PublicCategorySlotSchema = z.object({
  category: z.string(),
  status: z.union([z.literal("hidden"), z.literal("revealed")]),
  cards: z.array(PublicCategoryCardSchema),
});

export const YouCategorySlotSchema = z.object({
  category: z.string(),
  cards: z.array(YouCategoryCardSchema),
});

export const DisasterOptionSchema = z.object({
  id: z.string().min(1),
  title: z.string().min(1),
});

export const PublicPlayerViewSchema = z.object({
  playerId: z.string(),
  name: z.string(),
  status: z.union([z.literal("alive"), z.literal("eliminated"), z.literal("left_bunker")]),
  connected: z.boolean(),
  disconnectedAt: z.number().int().nonnegative().optional(),
  totalAbsentMs: z.number().int().nonnegative().optional(),
  currentOfflineMs: z.number().int().nonnegative().optional(),
  kickRemainingMs: z.number().int().nonnegative().optional(),
  leftBunker: z.boolean().optional(),
  revealedCards: z.array(CardRefSchema),
  revealedCount: z.number().int().nonnegative(),
  totalCards: z.number().int().nonnegative(),
  specialRevealed: z.boolean(),
  categories: z.array(PublicCategorySlotSchema),
});

export const RoomStateSchema = z.object({
  roomCode: z.string(),
  players: z.array(PlayerSummarySchema),
  hostId: z.string(),
  controlId: z.string(),
  phase: z.union([z.literal("lobby"), z.literal("game")]),
  scenarioMeta: ScenarioMetaSchema,
  settings: GameSettingsSchema,
  ruleset: GameRulesetSchema,
  rulesOverriddenByHost: z.boolean(),
  rulesPresetCount: z.number().int().min(4).max(16).optional(),
  world: WorldState30Schema.optional(),
  isDev: z.boolean().optional(),
  disasterOptions: z.array(DisasterOptionSchema).optional(),
});

export const VotingViewSchema = z.object({
  hasVoted: z.boolean(),
});

export const VotePhaseSchema = z.union([
  z.literal("voting"),
  z.literal("voteSpecialWindow"),
  z.literal("voteResolve"),
]);

export const VotingProgressSchema = z.object({
  voted: z.number().int().nonnegative(),
  total: z.number().int().nonnegative(),
});

export const LocalizedVarsSchema = z.record(z.union([z.string(), z.number()]));

export const ThreatModifierViewSchema = z.object({
  delta: z.number().int(),
  reasons: z.array(z.string()),
  reasonCardIds: z.array(z.string()).optional(),
  baseCount: z.number().int().nonnegative(),
  finalCount: z.number().int().nonnegative(),
});

export const VotePublicSchema = z.object({
  voterId: z.string(),
  voterName: z.string(),
  targetId: z.string().optional(),
  targetName: z.string().optional(),
  status: z.union([z.literal("voted"), z.literal("not_voted"), z.literal("invalid")]),
  reason: z.string().optional(),
  reasonKey: z.string().optional(),
  reasonVars: LocalizedVarsSchema.optional(),
  reasonCode: z
    .union([
      z.literal("VOTE_BLOCKED"),
      z.literal("VOTE_FORCED_SELF"),
      z.literal("VOTE_SPENT"),
      z.literal("VOTE_TARGET_DISALLOWED"),
      z.literal("VOTE_TARGET_UNAVAILABLE"),
      z.literal("VOTE_BANNED_AGAINST_TARGET"),
    ])
    .optional(),
  weight: z.number().optional(),
  submittedAt: z.number().int().nonnegative().optional(),
});

export const GameEventSchema = z.object({
  id: z.string(),
  kind: z.union([
    z.literal("roundStart"),
    z.literal("votingStart"),
    z.literal("elimination"),
    z.literal("gameEnd"),
    z.literal("info"),
    z.literal("playerDisconnected"),
    z.literal("playerReconnected"),
    z.literal("playerLeftBunker"),
  ]),
  message: z.string(),
  messageKey: z.string().optional(),
  messageVars: LocalizedVarsSchema.optional(),
  createdAt: z.number().int().nonnegative(),
});

export const GameViewSchema = z.object({
  phase: z.union([
    z.literal("reveal"),
    z.literal("reveal_discussion"),
    z.literal("voting"),
    z.literal("resolution"),
    z.literal("ended"),
  ]),
  round: z.number().int().nonnegative(),
  categoryOrder: z.array(z.string()),
  lastStageText: z.string().optional(),
  lastStageTextKey: z.string().optional(),
  lastStageTextVars: LocalizedVarsSchema.optional(),
  ruleset: GameRulesetSchema,
  world: WorldState30Schema.optional(),
  worldEvent: WorldEventSchema.optional(),
  postGame: PostGameStateSchema.optional(),
  you: z.object({
    playerId: z.string(),
    name: z.string(),
    hand: z.array(CardInHandSchema),
    categories: z.array(YouCategorySlotSchema),
    specialConditions: z.array(SpecialConditionInstanceSchema),
  }),
  public: z.object({
    players: z.array(PublicPlayerViewSchema),
    revealedThisRound: z.array(z.string()),
    roundRevealedCount: z.number().int().nonnegative().optional(),
    roundTotalAlive: z.number().int().nonnegative().optional(),
    currentTurnPlayerId: z.string().nullable().optional(),
    yourVoteWeight: z.number().int().positive().optional(),
    votesRemainingInRound: z.number().int().min(0).optional(),
    votesTotalThisRound: z.number().int().min(0).optional(),
    revealLimit: z.number().int().min(1).optional(),
    voting: VotingViewSchema.optional(),
    votePhase: VotePhaseSchema.nullable().optional(),
    votesPublic: z.array(VotePublicSchema).optional(),
    votingProgress: VotingProgressSchema.optional(),
    disallowedVoteTargetIdsForYou: z.array(z.string()).optional(),
    threatModifier: ThreatModifierViewSchema.optional(),
    canOpenVotingModal: z.boolean().optional(),
    canContinue: z.boolean().optional(),
    activeTimer: GameTimerStateSchema.nullable().optional(),
    voteModalOpen: z.boolean().optional(),
    lastEliminated: z.string().optional(),
    winners: z.array(z.string()).optional(),
    resolutionNote: z.string().optional(),
    resolutionNoteKey: z.string().optional(),
    resolutionNoteVars: LocalizedVarsSchema.optional(),
    roundRules: z
      .object({
        noTalkUntilVoting: z.boolean().optional(),
        forcedRevealCategory: z.string().optional(),
      })
      .optional(),
  }),
});

export const ClientHelloSchema = z.object({
  name: z.string().min(1),
  roomCode: z.string().min(1).optional(),
  create: z.boolean().optional(),
  scenarioId: z.string().min(1).optional(),
  locale: z.union([z.literal("ru"), z.literal("en")]).optional(),
  playerToken: z.string().min(1).optional(),
  tabId: z.string().min(1).optional(),
  sessionId: z.string().min(1).optional(),
});

export const ClientMessageSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("hello"),
    payload: ClientHelloSchema,
  }),
  z.object({
    type: z.literal("resume"),
    payload: z.object({
      roomCode: z.string().min(1),
      sessionId: z.string().min(1),
    }),
  }),
  z.object({
    type: z.literal("startGame"),
    payload: z.object({}),
  }),
  z.object({
    type: z.literal("ping"),
    payload: z.object({}),
  }),
  z.object({
    type: z.literal("revealCard"),
    payload: z.object({ cardId: z.string().min(1) }),
  }),
  z.object({
    type: z.literal("vote"),
    payload: z.object({ targetPlayerId: z.string().min(1) }),
  }),
  z.object({
    type: z.literal("finalizeVoting"),
    payload: z.object({}),
  }),
  z.object({
    type: z.literal("applySpecial"),
    payload: z.object({
      specialInstanceId: z.string().min(1),
      payload: z.record(z.any()).optional(),
    }),
  }),
  z.object({
    type: z.literal("revealWorldThreat"),
    payload: z.object({
      index: z.number().int().min(0),
    }),
  }),
  z.object({
    type: z.literal("setBunkerOutcome"),
    payload: z.object({
      outcome: z.union([z.literal("survived"), z.literal("failed")]),
    }),
  }),
  z.object({
    type: z.literal("devSkipRound"),
    payload: z.object({}),
  }),
  z.object({
    type: z.literal("devKickPlayer"),
    payload: z.object({
      targetPlayerId: z.string().min(1),
    }),
  }),
  z.object({
    type: z.literal("continueRound"),
    payload: z.object({}),
  }),
  z.object({
    type: z.literal("kickFromLobby"),
    payload: z.object({
      targetPlayerId: z.string().min(1),
    }),
  }),
  z.object({
    type: z.literal("updateSettings"),
    payload: GameSettingsSchema,
  }),
  z.object({
    type: z.literal("updateLocale"),
    payload: z.object({
      locale: z.union([z.literal("ru"), z.literal("en")]),
    }),
  }),
  z.object({
    type: z.literal("updateRules"),
    payload: z.object({
      mode: z.union([z.literal("auto"), z.literal("manual")]),
      presetPlayerCount: z.number().int().min(4).max(16).optional(),
      manualConfig: ManualRulesConfigSchema.optional(),
    }),
  }),
  z.object({
    type: z.literal("devAddPlayer"),
    payload: z.object({
      name: z.string().optional(),
    }),
  }),
  z.object({
    type: z.literal("devRemovePlayer"),
    payload: z.object({
      targetPlayerId: z.string().optional(),
    }),
  }),
  z.object({
    type: z.literal("requestHostTransfer"),
    payload: z.object({
      targetPlayerId: z.string().min(1).optional(),
    }),
  }),
]);

export const ServerMessageSchema = z.discriminatedUnion("type", [
  z.object({
    type: z.literal("roomState"),
    payload: RoomStateSchema,
  }),
  z.object({
    type: z.literal("gameView"),
    payload: GameViewSchema,
  }),
  z.object({
    type: z.literal("statePatch"),
    payload: z.object({
      roomState: RoomStateSchema.partial().optional(),
      gameView: GameViewSchema.partial().optional(),
    }),
  }),
  z.object({
    type: z.literal("gameEvent"),
    payload: GameEventSchema,
  }),
  z.object({
      type: z.literal("error"),
      payload: z.object({
        message: z.string(),
        code: z.string().optional(),
        errorKey: z.string().optional(),
        errorVars: LocalizedVarsSchema.optional(),
        maxPlayers: z.number().int().min(2).max(64).optional(),
      }),
    }),
  z.object({
    type: z.literal("helloAck"),
    payload: z.object({
      playerId: z.string(),
      playerToken: z.string(),
    }),
  }),
  z.object({
    type: z.literal("hostChanged"),
    payload: z.object({
      newHostId: z.string(),
      newControlId: z.string().optional(),
      reason: z.union([
        z.literal("disconnect_timeout"),
        z.literal("left_bunker"),
        z.literal("eliminated"),
        z.literal("manual"),
      ]),
    }),
  }),
  z.object({
    type: z.literal("pong"),
    payload: z.object({}).optional(),
  }),
]);

export type ClientMessage = z.infer<typeof ClientMessageSchema>;
export type ServerMessage = z.infer<typeof ServerMessageSchema>;
export type ClientHelloPayload = z.infer<typeof ClientHelloSchema>;

export { formatLabelShort } from "./labelFormat.js";
export { getRulesetForPlayerCount, RULESET_PRESET_COUNTS, RULESET_TABLE } from "./ruleset.js";
