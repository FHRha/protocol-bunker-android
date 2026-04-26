import { useEffect, useMemo, useState } from "react";
import type { GameSettings, ManualRulesConfig, RoomState } from "@bunker/shared";
import { useUiLocaleNamespace, useUiLocaleNamespacesActivation } from "../localization";
import { getLocalizedWorldTitle } from "../cardLocalization";
import RulesModal from "../components/RulesModal";
import { LobbyKickModal } from "../lobby/LobbyKickModal";
import { LobbyManualRulesCard } from "../lobby/LobbyManualRulesCard";
import { LobbyPlayersCard } from "../lobby/LobbyPlayersCard";
import { LobbyRulesCard } from "../lobby/LobbyRulesCard";
import { LobbySettingsCard } from "../lobby/LobbySettingsCard";
import {
  buildRevealPlan,
  clampInt,
  fitVotesByTotal,
  generateVotesByDefault,
  normalizeVotesByRound,
  parseVotesSchedule,
  sumVotes,
} from "../lobby/rulesMath";

interface LobbyPageProps {
  roomState: RoomState | null;
  playerId: string | null;
  isControl: boolean;
  showHints: boolean;
  wsInteractive: boolean;
  onStart: () => void;
  onUpdateSettings: (settings: GameSettings) => void;
  onUpdateRules: (payload: {
    mode: "auto" | "manual";
    presetPlayerCount?: number;
    manualConfig?: ManualRulesConfig;
  }) => void;
  onKickPlayer: (targetPlayerId: string, options?: { skipConfirm?: boolean }) => void;
  onTransferHost: (targetPlayerId: string) => void;
}

type LobbyPlayer = RoomState["players"][number];

const GITHUB_URL = "https://github.com/FHRha";

function hasSuspiciousPlayerNameChars(value: string): boolean {
  return /[\u0000-\u001f\u007f]/.test(value);
}

function getSafePlayerName(
  playerName: string | null | undefined,
  fallbackName: string
): string {
  const trimmed = typeof playerName === "string" ? playerName.trim() : "";
  if (!trimmed || hasSuspiciousPlayerNameChars(trimmed)) return fallbackName;
  return trimmed;
}

export default function LobbyPage({
  roomState,
  playerId,
  isControl,
  showHints,
  wsInteractive,
  onStart,
  onUpdateSettings,
  onUpdateRules,
  onKickPlayer,
  onTransferHost,
}: LobbyPageProps) {
  useUiLocaleNamespacesActivation([
    "lobby",
    "common",
    "room-settings",
    "rules",
    "format",
    "maps",
    "dev",
    "reconnect",
    "misc",
    "game",
  ]);
  const lobbyTexts = useUiLocaleNamespace("lobby", {
    fallbacks: ["common", "room-settings", "rules", "format", "maps", "dev", "reconnect", "misc", "game"],
  });
  const lobbyLocale = useMemo(() => {
    const rawPresetOptions = lobbyTexts.getRaw("rulesPresetOptions");
    const rulesPresetOptions = Array.isArray(rawPresetOptions)
      ? rawPresetOptions.map((value) => Number(value)).filter((value) => Number.isFinite(value))
      : [];

    return {
      authorGithubAria: lobbyTexts.t("authorGithubAria"),
      authorGithubLabel: lobbyTexts.t("authorGithubLabel"),
      controlMarker: lobbyTexts.t("controlMarker"),
      devKickNoTargets: lobbyTexts.t("devKickNoTargets"),
      hostMarker: lobbyTexts.t("hostMarker"),
      hostOnlyHint: lobbyTexts.t("hostOnlyHint"),
      lobbyKickAgreeLabel: lobbyTexts.t("lobbyKickAgreeLabel"),
      lobbyKickButton: lobbyTexts.t("lobbyKickButton"),
      lobbyKickSelectPlaceholder: lobbyTexts.t("lobbyKickSelectPlaceholder"),
      lobbyKickTitle: lobbyTexts.t("lobbyKickTitle"),
      lobbyLoading: lobbyTexts.t("lobbyLoading"),
      lobbyTitle: lobbyTexts.t("lobbyTitle"),
      manualAdjust: lobbyTexts.t("manualAdjust"),
      manualBunkerSlotsLabel: lobbyTexts.t("manualBunkerSlotsLabel"),
      manualFillFromTemplate: lobbyTexts.t("manualFillFromTemplate"),
      manualGenerate: lobbyTexts.t("manualGenerate"),
      manualModeTitle: lobbyTexts.t("manualModeTitle"),
      manualRevealsPlanLabel: lobbyTexts.t("manualRevealsPlanLabel"),
      manualRevealsRecommended: lobbyTexts.t("manualRevealsRecommended"),
      manualRevealsRequiredLabel: lobbyTexts.t("manualRevealsRequiredLabel"),
      manualRevealsWarning: lobbyTexts.t("manualRevealsWarning"),
      manualRoundAdd: lobbyTexts.t("manualRoundAdd"),
      manualRoundRemove: lobbyTexts.t("manualRoundRemove"),
      manualVotesFormatHint: lobbyTexts.t("manualVotesFormatHint"),
      maxPlayersHint: lobbyTexts.t("maxPlayersHint"),
      modalCancel: lobbyTexts.t("modalCancel"),
      offlineMarker: lobbyTexts.t("offlineMarker"),
      playersTitle: lobbyTexts.t("playersTitle"),
      rulesButtonShort: lobbyTexts.t("rulesButtonShort"),
      rulesModeAuto: lobbyTexts.t("rulesModeAuto"),
      rulesModeLabel: lobbyTexts.t("rulesModeLabel"),
      rulesModeManual: lobbyTexts.t("rulesModeManual"),
      rulesNeedMinPlayers: lobbyTexts.t("rulesNeedMinPlayers"),
      rulesPresetLabel: lobbyTexts.t("rulesPresetLabel"),
      rulesPresetOptions,
      rulesTitle: lobbyTexts.t("rulesTitle"),
      settingsAutomationAuto: lobbyTexts.t("settingsAutomationAuto"),
      settingsAutomationManual: lobbyTexts.t("settingsAutomationManual"),
      settingsAutomationMode: lobbyTexts.t("settingsAutomationMode"),
      settingsAutomationModeHint: lobbyTexts.t("settingsAutomationModeHint"),
      settingsAutomationSemi: lobbyTexts.t("settingsAutomationSemi"),
      settingsContinueAnyone: lobbyTexts.t("settingsContinueAnyone"),
      settingsContinueHost: lobbyTexts.t("settingsContinueHost"),
      settingsContinuePermission: lobbyTexts.t("settingsContinuePermission"),
      settingsContinueRevealer: lobbyTexts.t("settingsContinueRevealer"),
      settingsContinueTipText: lobbyTexts.t("settingsContinueTipText"),
      settingsFinalThreatReveal: lobbyTexts.t("settingsFinalThreatReveal"),
      settingsForcedDisaster: lobbyTexts.t("settingsForcedDisaster"),
      settingsForcedDisasterRandom: lobbyTexts.t("settingsForcedDisasterRandom"),
      settingsMaxPlayers: lobbyTexts.t("settingsMaxPlayers"),
      settingsOff: lobbyTexts.t("settingsOff"),
      settingsOn: lobbyTexts.t("settingsOn"),
      settingsOtherBlock: lobbyTexts.t("settingsOtherBlock"),
      settingsPostVoteTimer: lobbyTexts.t("settingsPostVoteTimer"),
      settingsPreVoteTimer: lobbyTexts.t("settingsPreVoteTimer"),
      settingsRevealDiscussionTimer: lobbyTexts.t("settingsRevealDiscussionTimer"),
      settingsRevealTimeoutAction: lobbyTexts.t("settingsRevealTimeoutAction"),
      settingsRevealTimeoutRandom: lobbyTexts.t("settingsRevealTimeoutRandom"),
      settingsRevealTimeoutSkip: lobbyTexts.t("settingsRevealTimeoutSkip"),
      settingsSpecialAnytime: lobbyTexts.t("settingsSpecialAnytime"),
      settingsSpecialUsage: lobbyTexts.t("settingsSpecialUsage"),
      settingsSpecialVotingOnly: lobbyTexts.t("settingsSpecialVotingOnly"),
      settingsThreatAnyone: lobbyTexts.t("settingsThreatAnyone"),
      settingsThreatHost: lobbyTexts.t("settingsThreatHost"),
      settingsThreatTipText: lobbyTexts.t("settingsThreatTipText"),
      settingsTimersBlock: lobbyTexts.t("settingsTimersBlock"),
      settingsTitle: lobbyTexts.t("settingsTitle"),
      startButton: lobbyTexts.t("startButton"),
      transferHostButton: lobbyTexts.t("transferHostButton"),
      transferHostSelectPlaceholder: lobbyTexts.t("transferHostSelectPlaceholder"),
      transferHostTitle: lobbyTexts.t("transferHostTitle"),
      votesByRoundLabel: lobbyTexts.t("votesByRoundLabel"),
      wsActionDisabledHint: lobbyTexts.t("wsActionDisabledHint"),
      scenarioLabel: (name: string) => lobbyTexts.t("scenarioLabel", { name }),
      rulesPlayers: (count: number) => lobbyTexts.t("rulesPlayers", { count }),
      rulesSeats: (count: number) => lobbyTexts.t("rulesSeats", { count }),
      rulesVotes: (votes: number[]) => lobbyTexts.t("rulesVotes", { values: votes.join(" / ") }),
      rulesExiles: (count: number) => lobbyTexts.t("rulesExiles", { count }),
      playerExtra: (count: number) => lobbyTexts.t("playerExtra", { count }),
      playerFallback: (index: number) => lobbyTexts.t("playerFallback", { index }),
      manualVotesRequired: (count: number) => lobbyTexts.t("manualVotesRequired", { count }),
      manualVotesSumHint: (sum: number, required: number) =>
        lobbyTexts.t("manualVotesSumHint", { sum, required }),
    };
  }, [lobbyTexts]);

  const [draft, setDraft] = useState<GameSettings | null>(roomState?.settings ?? null);
  const [kickTargetId, setKickTargetId] = useState("");
  const [kickModalOpen, setKickModalOpen] = useState(false);
  const [kickAgree, setKickAgree] = useState(false);
  const [transferHostTargetId, setTransferHostTargetId] = useState("");
  const [manualTemplatePlayers, setManualTemplatePlayers] = useState(4);
  const [manualVotesInput, setManualVotesInput] = useState("0");
  const [rulesOpen, setRulesOpen] = useState(false);

  const canControl = Boolean(isControl);
  const controlsDisabled = !wsInteractive;

  useEffect(() => {
    if (!roomState) return;
    setDraft(roomState.settings);
  }, [roomState?.settings]);

  useEffect(() => {
    if (!roomState) {
      setTransferHostTargetId("");
      return;
    }
    const candidateIds = roomState.players
      .map((player) => player.playerId)
      .filter((nextPlayerId) => nextPlayerId !== roomState.hostId);
    setTransferHostTargetId((prev) => (candidateIds.includes(prev) ? prev : candidateIds[0] ?? ""));
  }, [roomState?.players, roomState?.hostId]);

  useEffect(() => {
    if (!roomState) {
      setKickTargetId("");
      return;
    }
    const candidateIds = roomState.players
      .map((player) => player.playerId)
      .filter((nextPlayerId) => nextPlayerId !== roomState.controlId);
    setKickTargetId((prev) => (candidateIds.includes(prev) ? prev : candidateIds[0] ?? ""));
    if (candidateIds.length === 0) {
      setKickAgree(false);
      setKickModalOpen(false);
    }
  }, [roomState?.players, roomState?.controlId]);

  useEffect(() => {
    if (!roomState) return;
    const nextTemplate =
      roomState.ruleset.manualConfig?.seedTemplatePlayers ??
      roomState.rulesPresetCount ??
      roomState.ruleset.playerCount;
    setManualTemplatePlayers(clampInt(nextTemplate, 4, 16));
  }, [roomState?.ruleset, roomState?.rulesPresetCount]);

  useEffect(() => {
    if (!roomState) {
      setManualVotesInput("0");
      return;
    }
    const sourceVotes =
      roomState.ruleset.rulesetMode === "manual" && roomState.ruleset.manualConfig
        ? roomState.ruleset.manualConfig.votesByRound
        : roomState.ruleset.votesPerRound;
    setManualVotesInput(normalizeVotesByRound(sourceVotes).join("/"));
  }, [roomState?.ruleset]);

  if (!roomState) {
    return (
      <section className="panel">
        <h2>{lobbyLocale.lobbyTitle}</h2>
        <p className="muted">{lobbyLocale.lobbyLoading}</p>
      </section>
    );
  }

  const settings = draft ?? roomState.settings;
  const disasterOptions = useMemo(
    () =>
      (roomState.disasterOptions ?? []).map((option) => ({
        ...option,
        title: getLocalizedWorldTitle(lobbyTexts.locale, option),
      })),
    [lobbyTexts.locale, roomState.disasterOptions]
  );
  const supportsForcedDisaster = disasterOptions.length > 0;
  const disasterTitleById = new Map(disasterOptions.map((option) => [option.id, option.title]));
  const normalizedForcedDisasterId =
    settings.forcedDisasterId === "random" || disasterTitleById.has(settings.forcedDisasterId)
      ? settings.forcedDisasterId
      : "random";
  const forcedDisasterTitle =
    normalizedForcedDisasterId === "random"
      ? lobbyLocale.settingsForcedDisasterRandom
      : disasterTitleById.get(normalizedForcedDisasterId) ?? normalizedForcedDisasterId;
  const isClassic = roomState.scenarioMeta.id === "classic";
  const ruleset = roomState.ruleset;
  const rulesMode: "auto" | "manual" = ruleset.rulesetMode === "auto" ? "auto" : "manual";
  const presetCount =
    roomState.rulesPresetCount ?? ruleset.manualConfig?.seedTemplatePlayers ?? ruleset.playerCount;
  const rulesModeText = rulesMode === "manual" ? lobbyLocale.rulesModeManual : lobbyLocale.rulesModeAuto;
  const votesSummary = lobbyLocale.rulesVotes(ruleset.votesPerRound);
  const votesSeparatorIndex = votesSummary.indexOf(":");
  const votesLabel =
    votesSeparatorIndex >= 0 ? votesSummary.slice(0, votesSeparatorIndex + 1) : lobbyLocale.votesByRoundLabel;
  const votesValue =
    votesSeparatorIndex >= 0
      ? votesSummary.slice(votesSeparatorIndex + 1).trim()
      : ruleset.votesPerRound.join(" / ");
  const manualConfig: ManualRulesConfig = ruleset.manualConfig
    ? {
        ...ruleset.manualConfig,
        votesByRound: normalizeVotesByRound(ruleset.manualConfig.votesByRound),
        targetReveals: clampInt(ruleset.manualConfig.targetReveals ?? 7, 5, 7),
      }
    : {
        bunkerSlots: ruleset.bunkerSeats,
        votesByRound: normalizeVotesByRound([...ruleset.votesPerRound]),
        targetReveals: 7,
        seedTemplatePlayers: clampInt(presetCount, 4, 16),
      };
  const requiredVotes = Math.max(0, roomState.players.length - manualConfig.bunkerSlots);
  const manualVotesSum = sumVotes(manualConfig.votesByRound);
  const manualVotesMismatch = manualVotesSum !== requiredVotes;
  const revealPlan = buildRevealPlan(manualConfig.votesByRound.length, manualConfig.targetReveals);
  const revealPlanText = revealPlan.join("/");
  const manualRevealNotRecommended = manualConfig.targetReveals !== 7;
  const canStart = !isClassic || roomState.players.length >= 4;
  const minPlayersLimit = Math.max(isClassic ? 4 : 2, roomState.players.length);
  const wsHint = controlsDisabled ? lobbyLocale.wsActionDisabledHint : null;
  const continueTipText = lobbyLocale.settingsContinueTipText;
  const threatTipText = lobbyLocale.settingsThreatTipText;
  const rulesButtonLabel = lobbyLocale.rulesButtonShort;

  const sendRulesUpdate = (payload: {
    mode: "auto" | "manual";
    presetPlayerCount?: number;
    manualConfig?: ManualRulesConfig;
  }) => {
    if (controlsDisabled) return;
    onUpdateRules(payload);
  };

  const updateManualConfig = (patch: Partial<ManualRulesConfig>) => {
    const merged: ManualRulesConfig = {
      ...manualConfig,
      ...patch,
      seedTemplatePlayers: clampInt(
        patch.seedTemplatePlayers ?? manualConfig.seedTemplatePlayers ?? manualTemplatePlayers,
        4,
        16
      ),
      bunkerSlots: clampInt(patch.bunkerSlots ?? manualConfig.bunkerSlots, 1, 16),
      votesByRound: normalizeVotesByRound(patch.votesByRound ?? manualConfig.votesByRound),
      targetReveals: clampInt(patch.targetReveals ?? manualConfig.targetReveals, 5, 7),
    };
    sendRulesUpdate({
      mode: "manual",
      presetPlayerCount: merged.seedTemplatePlayers,
      manualConfig: merged,
    });
  };

  const applyRulesMode = (mode: "auto" | "manual") => {
    if (mode === "auto") {
      sendRulesUpdate({ mode: "auto" });
      return;
    }
    sendRulesUpdate({
      mode: "manual",
      presetPlayerCount: clampInt(manualTemplatePlayers, 4, 16),
      manualConfig: {
        ...manualConfig,
        seedTemplatePlayers: clampInt(manualTemplatePlayers, 4, 16),
      },
    });
  };

  const fillManualFromTemplate = () => {
    sendRulesUpdate({
      mode: "manual",
      presetPlayerCount: clampInt(manualTemplatePlayers, 4, 16),
    });
  };

  const maxPlayersVisible = 8;
  const visiblePlayers = roomState.players.slice(0, maxPlayersVisible);
  const extraPlayers = roomState.players.length - visiblePlayers.length;
  const kickCandidates = roomState.players.filter((player) => player.playerId !== roomState.controlId);
  const transferHostCandidates = roomState.players.filter((player) => player.playerId !== roomState.hostId);
  const playerIndexById = new Map(roomState.players.map((player, index) => [player.playerId, index]));
  const getFallbackPlayerName = (player: LobbyPlayer): string =>
    lobbyLocale.playerFallback((playerIndexById.get(player.playerId) ?? 0) + 1);
  const getLobbyPlayerName = (player: LobbyPlayer): string =>
    getSafePlayerName(player.name, getFallbackPlayerName(player));

  const applySettings = (next: GameSettings) => {
    if (controlsDisabled) return;
    setDraft(next);
    onUpdateSettings(next);
  };

  const updateField = <K extends keyof GameSettings>(key: K, value: GameSettings[K]) => {
    if (!draft) return;
    applySettings({ ...draft, [key]: value });
  };

  const updateAutomationMode = (mode: GameSettings["automationMode"]) => {
    if (!draft) return;
    applySettings({ ...draft, automationMode: mode });
  };

  return (
    <div className={`lobby-page lobbyLayout ${canControl ? "lobby--host" : "lobby--player"}`}>
      <div className="lobbyGrid">
        <div className="lobbyLeftColumn">
          <LobbyPlayersCard
            title={lobbyLocale.playersTitle}
            visiblePlayers={visiblePlayers}
            hostId={roomState.hostId}
            controlId={roomState.controlId}
            hostMarker={lobbyLocale.hostMarker}
            controlMarker={lobbyLocale.controlMarker}
            offlineMarker={lobbyLocale.offlineMarker}
            extraPlayers={extraPlayers}
            playerExtraText={extraPlayers > 0 ? lobbyLocale.playerExtra(extraPlayers) : null}
            canControl={canControl}
            kickCandidatesCount={kickCandidates.length}
            kickTitle={lobbyLocale.lobbyKickTitle}
            kickButton={lobbyLocale.lobbyKickButton}
            controlsDisabled={controlsDisabled}
            transferHostTitle={lobbyLocale.transferHostTitle}
            transferHostTargetId={transferHostTargetId}
            transferHostCandidates={transferHostCandidates}
            transferHostSelectPlaceholder={lobbyLocale.transferHostSelectPlaceholder}
            transferHostButton={lobbyLocale.transferHostButton}
            startButton={lobbyLocale.startButton}
            canStart={canStart}
            isClassic={isClassic}
            rulesNeedMinPlayers={lobbyLocale.rulesNeedMinPlayers}
            hostOnlyHint={lobbyLocale.hostOnlyHint}
            wsHint={wsHint}
            getLobbyPlayerName={getLobbyPlayerName}
            onOpenKickModal={() => {
              if (controlsDisabled) return;
              setKickAgree(false);
              setKickModalOpen(true);
            }}
            onTransferHostTargetChange={setTransferHostTargetId}
            onTransferHost={onTransferHost}
            onStart={() => {
              if (controlsDisabled) return;
              onStart();
            }}
          />

          <LobbyRulesCard
            title={lobbyLocale.rulesTitle}
            scenarioText={lobbyLocale.scenarioLabel(roomState.scenarioMeta.name)}
            rulesModeLabel={lobbyLocale.rulesModeLabel}
            canControl={canControl}
            isClassic={isClassic}
            rulesMode={rulesMode}
            rulesModeText={rulesModeText}
            controlsDisabled={controlsDisabled}
            rulesModeAuto={lobbyLocale.rulesModeAuto}
            rulesModeManual={lobbyLocale.rulesModeManual}
            onApplyRulesMode={applyRulesMode}
            playersText={lobbyLocale.rulesPlayers(roomState.players.length)}
            seatsText={lobbyLocale.rulesSeats(ruleset.bunkerSeats)}
            exilesText={lobbyLocale.rulesExiles(ruleset.totalExiles)}
            votesLabel={votesLabel}
            votesValue={votesValue}
          />
          {canControl && isClassic && rulesMode === "manual" ? (
            <LobbyManualRulesCard
              text={{
                manualModeTitle: lobbyLocale.manualModeTitle,
                rulesPresetLabel: lobbyLocale.rulesPresetLabel,
                manualFillFromTemplate: lobbyLocale.manualFillFromTemplate,
                manualBunkerSlotsLabel: lobbyLocale.manualBunkerSlotsLabel,
                manualVotesRequired: lobbyLocale.manualVotesRequired,
                manualRevealsRequiredLabel: lobbyLocale.manualRevealsRequiredLabel,
                manualRevealsRecommended: lobbyLocale.manualRevealsRecommended,
                manualRevealsWarning: lobbyLocale.manualRevealsWarning,
                manualRevealsPlanLabel: lobbyLocale.manualRevealsPlanLabel,
                votesByRoundLabel: lobbyLocale.votesByRoundLabel,
                manualRoundAdd: lobbyLocale.manualRoundAdd,
                manualRoundRemove: lobbyLocale.manualRoundRemove,
                manualGenerate: lobbyLocale.manualGenerate,
                manualVotesFormatHint: lobbyLocale.manualVotesFormatHint,
                manualVotesSumHint: lobbyLocale.manualVotesSumHint,
                manualAdjust: lobbyLocale.manualAdjust,
              }}
              controlsDisabled={controlsDisabled}
              manualTemplatePlayers={manualTemplatePlayers}
              setManualTemplatePlayers={setManualTemplatePlayers}
              rulesPresetOptions={lobbyLocale.rulesPresetOptions}
              fillManualFromTemplate={fillManualFromTemplate}
              manualConfig={manualConfig}
              requiredVotes={requiredVotes}
              manualRevealNotRecommended={manualRevealNotRecommended}
              revealPlanText={revealPlanText}
              updateManualConfig={updateManualConfig}
              manualVotesInput={manualVotesInput}
              setManualVotesInput={setManualVotesInput}
              parseVotesSchedule={parseVotesSchedule}
              generateVotesByDefault={generateVotesByDefault}
              fitVotesByTotal={fitVotesByTotal}
              manualVotesMismatch={manualVotesMismatch}
              manualVotesSum={manualVotesSum}
              wsHint={wsHint}
            />
          ) : null}
        </div>

        <div className="lobbyRightColumn">
          <LobbySettingsCard
            text={{
              settingsTitle: lobbyLocale.settingsTitle,
              settingsTimersBlock: lobbyLocale.settingsTimersBlock,
              settingsRevealDiscussionTimer: lobbyLocale.settingsRevealDiscussionTimer,
              settingsPreVoteTimer: lobbyLocale.settingsPreVoteTimer,
              settingsPostVoteTimer: lobbyLocale.settingsPostVoteTimer,
              settingsOtherBlock: lobbyLocale.settingsOtherBlock,
              settingsAutomationMode: lobbyLocale.settingsAutomationMode,
              settingsAutomationModeHint: lobbyLocale.settingsAutomationModeHint,
              settingsAutomationAuto: lobbyLocale.settingsAutomationAuto,
              settingsAutomationSemi: lobbyLocale.settingsAutomationSemi,
              settingsAutomationManual: lobbyLocale.settingsAutomationManual,
              settingsContinuePermission: lobbyLocale.settingsContinuePermission,
              settingsContinueHost: lobbyLocale.settingsContinueHost,
              settingsContinueRevealer: lobbyLocale.settingsContinueRevealer,
              settingsContinueAnyone: lobbyLocale.settingsContinueAnyone,
              settingsRevealTimeoutAction: lobbyLocale.settingsRevealTimeoutAction,
              settingsRevealTimeoutRandom: lobbyLocale.settingsRevealTimeoutRandom,
              settingsRevealTimeoutSkip: lobbyLocale.settingsRevealTimeoutSkip,
              settingsSpecialUsage: lobbyLocale.settingsSpecialUsage,
              settingsSpecialAnytime: lobbyLocale.settingsSpecialAnytime,
              settingsSpecialVotingOnly: lobbyLocale.settingsSpecialVotingOnly,
              settingsFinalThreatReveal: lobbyLocale.settingsFinalThreatReveal,
              settingsThreatHost: lobbyLocale.settingsThreatHost,
              settingsThreatAnyone: lobbyLocale.settingsThreatAnyone,
              settingsForcedDisaster: lobbyLocale.settingsForcedDisaster,
              settingsForcedDisasterRandom: lobbyLocale.settingsForcedDisasterRandom,
              settingsMaxPlayers: lobbyLocale.settingsMaxPlayers,
              maxPlayersHint: lobbyLocale.maxPlayersHint,
              settingsOn: lobbyLocale.settingsOn,
              settingsOff: lobbyLocale.settingsOff,
            }}
            canControl={canControl}
            controlsDisabled={controlsDisabled}
            settings={settings}
            continueTipText={continueTipText}
            threatTipText={threatTipText}
            supportsForcedDisaster={supportsForcedDisaster}
            normalizedForcedDisasterId={normalizedForcedDisasterId}
            disasterOptions={disasterOptions}
            forcedDisasterTitle={forcedDisasterTitle}
            minPlayersLimit={minPlayersLimit}
            wsHint={wsHint}
            updateField={updateField}
            updateAutomationMode={updateAutomationMode}
          />
        </div>
      </div>
      <button className="ghost rulesButton lobbyRulesButton" onClick={() => setRulesOpen(true)}>
        {rulesButtonLabel}
      </button>
      <a
        className="homeWatermark lobbyWatermark"
        href={GITHUB_URL}
        target="_blank"
        rel="noopener noreferrer"
        aria-label={lobbyLocale.authorGithubAria}
      >
        {lobbyLocale.authorGithubLabel}
      </a>
      <LobbyKickModal
        open={kickModalOpen && canControl}
        title={lobbyLocale.lobbyKickTitle}
        cancelLabel={lobbyLocale.modalCancel}
        noTargetsLabel={lobbyLocale.devKickNoTargets}
        selectPlaceholder={lobbyLocale.lobbyKickSelectPlaceholder}
        agreeLabel={lobbyLocale.lobbyKickAgreeLabel}
        submitLabel={lobbyLocale.lobbyKickButton}
        kickCandidates={kickCandidates}
        kickTargetId={kickTargetId}
        setKickTargetId={setKickTargetId}
        kickAgree={kickAgree}
        setKickAgree={setKickAgree}
        controlsDisabled={controlsDisabled}
        getLobbyPlayerName={getLobbyPlayerName}
        onClose={() => {
          setKickModalOpen(false);
          setKickAgree(false);
        }}
        onConfirm={(targetPlayerId) => {
          if (controlsDisabled) return;
          onKickPlayer(targetPlayerId, { skipConfirm: true });
          setKickModalOpen(false);
          setKickAgree(false);
        }}
      />
      <RulesModal open={rulesOpen} onClose={() => setRulesOpen(false)} />
    </div>
  );
}
