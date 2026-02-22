import { z } from "zod";
export * from "./targeting.js";
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
export const OverlayTagViewSchema = z.object({
    label: z.string(),
    revealed: z.boolean(),
    value: z.string(),
});
export const OverlayCategoryViewSchema = z.object({
    key: z.string(),
    label: z.string(),
    revealed: z.boolean(),
    value: z.string(),
    imgUrl: z.string().optional(),
});
export const OverlayOverrideEnabledSchema = z.object({
    topBunker: z.boolean().optional(),
    topCatastrophe: z.boolean().optional(),
    topThreats: z.boolean().optional(),
    playerNames: z.boolean().optional(),
    playerTraits: z.boolean().optional(),
    playerCategories: z.boolean().optional(),
});
export const OverlayOverrideTopSchema = z.object({
    bunkerLines: z.array(z.string().max(120)).max(5).optional(),
    catastropheText: z.string().max(600).optional(),
    threatsLines: z.array(z.string().max(120)).max(6).optional(),
});
export const OverlayOverridePlayerTraitsSchema = z.object({
    sex: z.string().max(120).optional(),
    age: z.string().max(120).optional(),
    orient: z.string().max(120).optional(),
});
export const OverlayOverridePlayerEnabledSchema = z.object({
    name: z.boolean().optional(),
    traits: z.boolean().optional(),
    categories: z.record(z.string().max(40), z.boolean()).optional(),
});
export const OverlayOverridePlayerSchema = z.object({
    name: z.string().max(24).optional(),
    traits: OverlayOverridePlayerTraitsSchema.optional(),
    categories: z.record(z.string().max(120)).optional(),
    enabled: OverlayOverridePlayerEnabledSchema.optional(),
});
export const OverlayExtraTextSchema = z.object({
    id: z.string().min(1).max(64),
    text: z.string().max(120),
    x: z.number().min(0).max(1),
    y: z.number().min(0).max(1),
    align: z.union([z.literal("left"), z.literal("center"), z.literal("right")]).optional(),
    size: z.number().min(8).max(96).optional(),
    color: z.string().max(32).optional(),
    shadow: z.boolean().optional(),
    visible: z.boolean().optional(),
});
export const OverlayOverridesSchema = z.object({
    enabled: OverlayOverrideEnabledSchema.optional(),
    top: OverlayOverrideTopSchema.optional(),
    players: z.record(OverlayOverridePlayerSchema).optional(),
    extraTexts: z.array(OverlayExtraTextSchema).optional(),
});
export const OverlayPlayerViewSchema = z.object({
    id: z.string(),
    nickname: z.string(),
    connected: z.boolean().optional(),
    alive: z.boolean(),
    tags: z.object({
        sex: OverlayTagViewSchema,
        age: OverlayTagViewSchema,
        orientation: OverlayTagViewSchema,
    }),
    categories: z.array(OverlayCategoryViewSchema),
});
export const OverlayStateSchema = z.object({
    roomId: z.string(),
    playerCount: z.number().int().nonnegative(),
    top: z.object({
        bunker: z.object({
            revealed: z.number().int().nonnegative(),
            total: z.number().int().nonnegative(),
            lines: z.array(z.string()),
            items: z.array(z.object({
                title: z.string(),
                subtitle: z.string().optional(),
                imageId: z.string().optional(),
            })).optional(),
        }),
        catastrophe: z.object({
            text: z.string(),
            title: z.string().optional(),
            imageId: z.string().optional(),
        }),
        threats: z.object({
            revealed: z.number().int().nonnegative(),
            total: z.number().int().nonnegative(),
            lines: z.array(z.string()),
            items: z.array(z.object({
                title: z.string(),
                subtitle: z.string().optional(),
                imageId: z.string().optional(),
            })).optional(),
        }),
    }),
    players: z.array(OverlayPlayerViewSchema),
    overrides: OverlayOverridesSchema.optional(),
});
export const GameSettingsSchema = z.object({
    enableRevealDiscussionTimer: z.boolean(),
    revealDiscussionSeconds: z.number().int().min(5).max(600),
    enablePreVoteDiscussionTimer: z.boolean(),
    preVoteDiscussionSeconds: z.number().int().min(5).max(600),
    enablePostVoteDiscussionTimer: z.boolean(),
    postVoteDiscussionSeconds: z.number().int().min(5).max(600),
    enablePresenterMode: z.boolean(),
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
});
export const CardRefSchema = z.object({
    id: z.string(),
    deck: z.string(),
    instanceId: z.string().optional(),
    labelShort: z.string().optional(),
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
    choiceKind: z.union([z.literal("player"), z.literal("neighbor"), z.literal("category"), z.literal("none")]).optional(),
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
});
export const YouCategoryCardSchema = z.object({
    instanceId: z.string(),
    labelShort: z.string(),
    revealed: z.boolean(),
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
export const ThreatModifierViewSchema = z.object({
    delta: z.number().int(),
    reasons: z.array(z.string()),
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
        votesRemainingInRound: z.number().int().min(0).optional(),
        votesTotalThisRound: z.number().int().min(0).optional(),
        revealLimit: z.number().int().min(1).optional(),
        voting: VotingViewSchema.optional(),
        votePhase: VotePhaseSchema.nullable().optional(),
        votesPublic: z.array(VotePublicSchema).optional(),
        votingProgress: VotingProgressSchema.optional(),
        threatModifier: ThreatModifierViewSchema.optional(),
        canOpenVotingModal: z.boolean().optional(),
        canContinue: z.boolean().optional(),
        activeTimer: GameTimerStateSchema.nullable().optional(),
        voteModalOpen: z.boolean().optional(),
        lastEliminated: z.string().optional(),
        winners: z.array(z.string()).optional(),
        resolutionNote: z.string().optional(),
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
        payload: z.object({}),
    }),
    z.object({
        type: z.literal("overlaySubscribe"),
        payload: z.object({
            roomCode: z.string().min(1),
            token: z.string().min(1),
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
            maxPlayers: z.number().int().min(2).max(64).optional(),
        }),
    }),
    z.object({
        type: z.literal("helloAck"),
        payload: z.object({ playerId: z.string(), playerToken: z.string() }),
    }),
    z.object({
        type: z.literal("hostChanged"),
        payload: z.object({
            newHostId: z.string(),
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
    z.object({
        type: z.literal("overlayState"),
        payload: z.object({
            ok: z.boolean(),
            unauthorized: z.boolean().optional(),
            roomCode: z.string().optional(),
            state: OverlayStateSchema.optional(),
            presenter: z.any().optional(),
            presenterModeEnabled: z.boolean().optional(),
            role: z.union([z.literal("VIEW"), z.literal("PLAYER"), z.literal("CONTROL")]).optional(),
            message: z.string().optional(),
        }),
    }),
]);
export { formatLabelShort } from "./labelFormat.js";
export { getRulesetForPlayerCount, RULESET_PRESET_COUNTS, RULESET_TABLE } from "./ruleset.js";
export { buildLinkSet, normalizeBase, LINK_PATHS } from "./urlBuilder.js";
