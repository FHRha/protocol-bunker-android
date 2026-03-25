package main

import (
	"fmt"
	"math/rand"
	"slices"
	"sort"
	"strings"
	"time"
)

const (
	scenarioPhaseReveal           = "reveal"
	scenarioPhaseRevealDiscussion = "reveal_discussion"
	scenarioPhaseVoting           = "voting"
	scenarioPhaseResolution       = "resolution"
	scenarioPhaseEnded            = "ended"
	timerKindResolutionAuto       = "resolution_auto"

	votePhaseVoting        = "voting"
	votePhaseSpecialWindow = "voteSpecialWindow"
	votePhaseResolve       = "voteResolve"
)

type deckSlotConfig struct {
	deck string
	slot string
}

var categoryToDeckSlot = buildCategoryToDeckSlot()

func buildCategoryToDeckSlot() map[string]deckSlotConfig {
	slots := make(map[string]deckSlotConfig, len(categoryOrder))

	if len(categoryOrder) > 0 && len(coreDecks) > 0 {
		slots[categoryOrder[0]] = deckSlotConfig{deck: coreDecks[0]}
	}
	if len(categoryOrder) > 1 && len(coreDecks) > 1 {
		slots[categoryOrder[1]] = deckSlotConfig{deck: coreDecks[1]}
	}
	if len(categoryOrder) > 2 && len(coreDecks) > 2 {
		slots[categoryOrder[2]] = deckSlotConfig{deck: coreDecks[2]}
	}
	if len(categoryOrder) > 3 && len(coreDecks) > 3 {
		slots[categoryOrder[3]] = deckSlotConfig{deck: coreDecks[3]}
	}
	if len(categoryOrder) > 4 {
		slots[categoryOrder[4]] = deckSlotConfig{deck: factsDeck, slot: "facts1"}
	}
	if len(categoryOrder) > 5 {
		slots[categoryOrder[5]] = deckSlotConfig{deck: factsDeck, slot: "facts2"}
	}
	if len(categoryOrder) > 6 && len(coreDecks) > 4 {
		slots[categoryOrder[6]] = deckSlotConfig{deck: coreDecks[4]}
	}

	return slots
}

func resolveCategoryDeckSlot(category string) (deckSlotConfig, bool) {
	if cfg, ok := categoryToDeckSlot[category]; ok {
		return cfg, true
	}

	normalized := normalizeSpecialKey(category)
	if normalized == "" {
		return deckSlotConfig{}, false
	}

	for key, cfg := range categoryToDeckSlot {
		if normalizeSpecialKey(key) == normalized {
			return cfg, true
		}
	}

	for categoryKey := range categoryKeyToDeck {
		if normalizeSpecialKey(categoryKey) != normalized &&
			normalizeSpecialKey(scenarioCategoryLabel(scenarioClassic, categoryKey)) != normalized &&
			normalizeSpecialKey(scenarioCategoryLabel(scenarioDevTest, categoryKey)) != normalized {
			continue
		}
		deck := categoryKeyToDeck[categoryKey]
		if deck == "" {
			continue
		}
		return deckSlotConfig{deck: deck, slot: categorySlotKey(categoryKey)}, true
	}

	return deckSlotConfig{}, false
}

func healthDeckName() string {
	if len(coreDecks) > 1 {
		return coreDecks[1]
	}
	return "health"
}

type handCard struct {
	InstanceID         string
	CardID             string
	Deck               string
	SlotKey            string
	Label              string
	Revealed           bool
	Missing            bool
	PublicBackCategory string
}

type voteRecord struct {
	TargetID   string
	Submitted  int64
	IsValid    bool
	ReasonText string
}

type gamePlayer struct {
	PlayerID                  string
	Name                      string
	Status                    string
	Hand                      []handCard
	Specials                  []specialConditionState
	SpecialCategoryProxyCards []publicCategoryCard
	IsBot                     bool

	BannedAgainst        map[string]bool
	ForcedSelfVoteNext   bool
	ForcedWastedVoteNext bool
}

type gameActionResult struct {
	StateChanged bool
	Events       []gameEvent
	Error        string
	ErrorKey     string
	ErrorVars    map[string]any
}

type gameSession struct {
	RoomCode string
	HostID   string
	Scenario string
	IsDev    bool
	Settings gameSettings
	Ruleset  gameRuleset

	Players map[string]*gamePlayer
	Order   []string

	Phase                 string
	Round                 int
	CurrentTurnID         string
	RevealedThisRnd       map[string]bool
	LastRevealerID        string
	VotePhase             string
	Votes                 map[string]voteRecord
	BaseVotes             map[string]voteRecord
	VoteResults           map[string]voteRecord
	VoteCandidates        map[string]bool
	VoteDisabled          map[string]string
	VoteWeights           map[string]int
	AutoSelfVoteVoters    map[string]bool
	AutoWastedVoters      map[string]bool
	RevoteDisallowByVoter map[string]map[string]bool
	DoubleAgainst         string
	TieBreakUsed          bool
	VotesRemaining        int
	LastEliminatedID      string
	ResolutionNote        string
	Winners               []string
	LastStageText         string
	LastStageKey          string
	LastStageVars         map[string]any
	TotalExiles           int
	EventCounter          int
	World                 worldStateView
	WorldEvent            *worldEventView
	PostGame              *postGameStateView
	FinalThreats          []string
	FirstHealthRevealerID string
	RoundRules            roundRulesPublic
	ResolutionNoteKey     string
	ResolutionNoteVars    map[string]any
	ActiveTimer           *gameTimerState
	DeckPools             map[string][]assetCard
	CardCounter           int
	DevBotCounter         int

	rng            *rand.Rand
	specialPool    []specialDefinition
	specialCatalog []specialDefinition
	specialCounter int
}

func newGameSession(roomCode, hostID, scenarioID string, settings gameSettings, ruleset gameRuleset, roomPlayers []*player, assets assetCatalog, specialDefinitions []specialDefinition, seed int64) *gameSession {
	rnd := newRand(seed)
	pools := make(map[string][]assetCard, len(assets.Decks))
	for deck, cards := range assets.Decks {
		copyCards := make([]assetCard, 0, len(cards))
		copyCards = append(copyCards, cards...)
		pools[deck] = copyCards
	}

	session := &gameSession{
		RoomCode:              roomCode,
		HostID:                hostID,
		Scenario:              scenarioID,
		IsDev:                 scenarioID == scenarioDevTest,
		Settings:              settings,
		Ruleset:               ruleset,
		Players:               make(map[string]*gamePlayer, len(roomPlayers)),
		Order:                 make([]string, 0, len(roomPlayers)),
		Phase:                 scenarioPhaseReveal,
		Round:                 1,
		RevealedThisRnd:       map[string]bool{},
		Votes:                 map[string]voteRecord{},
		BaseVotes:             map[string]voteRecord{},
		VoteResults:           map[string]voteRecord{},
		VoteCandidates:        map[string]bool{},
		VoteDisabled:          map[string]string{},
		VoteWeights:           map[string]int{},
		AutoSelfVoteVoters:    map[string]bool{},
		AutoWastedVoters:      map[string]bool{},
		RevoteDisallowByVoter: map[string]map[string]bool{},
		FinalThreats:          []string{},
		rng:                   rnd,
		DeckPools:             pools,
	}
	session.World = rollWorldFromPools(pools, rnd, len(roomPlayers), settings.ForcedDisasterID)
	sourceSpecials := specialDefinitions
	if len(sourceSpecials) == 0 {
		sourceSpecials = implementedSpecialDefinitions
	}
	sourceSpecials = localizeSpecialDefinitions(sourceSpecials, scenarioID, settings.CardLocale)
	session.specialCatalog = cloneSpecialDefinitions(sourceSpecials)
	session.specialPool = make([]specialDefinition, 0, len(sourceSpecials))
	for _, def := range sourceSpecials {
		session.specialPool = append(session.specialPool, def)
	}

	nextInstanceID := func(playerID string) string {
		session.CardCounter++
		return fmt.Sprintf("%s-%d", playerID, session.CardCounter)
	}

	drawFor := func(deckName, slot string, playerID string) handCard {
		pool := pools[deckName]
		card, next, ok := drawRandomCard(pool, rnd)
		pools[deckName] = next
		if !ok {
			return handCard{
				InstanceID: nextInstanceID(playerID),
				CardID:     "",
				Deck:       deckName,
				SlotKey:    slot,
				Label:      scenarioCardNoCardLabel(session.Scenario),
				Revealed:   false,
				Missing:    true,
			}
		}
		return handCard{
			InstanceID: nextInstanceID(playerID),
			CardID:     card.ID,
			Deck:       card.Deck,
			SlotKey:    slot,
			Label:      card.Label,
			Revealed:   false,
		}
	}

	for _, rp := range roomPlayers {
		hand := make([]handCard, 0, len(coreDecks)+len(factsSlotOrder))
		for _, deck := range coreDecks {
			hand = append(hand, drawFor(deck, "", rp.ID))
		}
		hand = append(hand, drawFor(factsDeck, "facts1", rp.ID))
		hand = append(hand, drawFor(factsDeck, "facts2", rp.ID))
		session.Players[rp.ID] = &gamePlayer{
			PlayerID:                  rp.ID,
			Name:                      rp.Name,
			Status:                    playerAlive,
			Hand:                      hand,
			Specials:                  nil,
			SpecialCategoryProxyCards: nil,
			IsBot:                     false,
			BannedAgainst:             map[string]bool{},
		}
		session.Order = append(session.Order, rp.ID)
	}
	session.assignInitialSpecials()

	session.CurrentTurnID = session.firstAliveID()
	session.VotesRemaining = session.votesForRound(session.Round)
	session.WorldEvent = nil
	session.revealNextBunkerCard(session.Round)
	session.setLastStage(scenarioText(session.Scenario, "event.roundStart", "Round 1. Reveal phase."), session.scenarioTextKey("event.roundStart", "classic.auto.121"), map[string]any{
		"round": session.Round,
		"v1":    session.Round,
	})
	session.scheduleRevealTimeoutIfNeeded()
	return session
}

func (g *gameSession) votesForRound(round int) int {
	index := round - 1
	if index < 0 || index >= len(g.Ruleset.VotesPerRnd) {
		return 0
	}
	return maxInt(0, g.Ruleset.VotesPerRnd[index])
}

func (g *gameSession) aliveIDs() []string {
	alive := make([]string, 0, len(g.Order))
	for _, id := range g.Order {
		player := g.Players[id]
		if player == nil {
			continue
		}
		if player.Status == playerAlive {
			alive = append(alive, id)
		}
	}
	return alive
}

func (g *gameSession) firstAliveID() string {
	for _, id := range g.Order {
		player := g.Players[id]
		if player != nil && player.Status == playerAlive {
			return id
		}
	}
	return ""
}

func (g *gameSession) nextUnrevealedAliveAfter(currentID string) string {
	if len(g.Order) == 0 {
		return ""
	}
	start := 0
	for i, id := range g.Order {
		if id == currentID {
			start = i
			break
		}
	}
	for offset := 1; offset <= len(g.Order); offset++ {
		index := (start + offset) % len(g.Order)
		id := g.Order[index]
		player := g.Players[id]
		if player == nil || player.Status != playerAlive {
			continue
		}
		if g.RevealedThisRnd[id] {
			continue
		}
		return id
	}
	return ""
}

func (g *gameSession) allAliveRevealed() bool {
	for _, id := range g.Order {
		player := g.Players[id]
		if player == nil || player.Status != playerAlive {
			continue
		}
		if !g.RevealedThisRnd[id] {
			return false
		}
	}
	return true
}

func (g *gameSession) startVoting() {
	g.clearActiveTimer()
	g.Phase = scenarioPhaseVoting
	g.VotePhase = votePhaseVoting
	g.Votes = map[string]voteRecord{}
	g.BaseVotes = map[string]voteRecord{}
	g.VoteResults = map[string]voteRecord{}
	g.VoteCandidates = map[string]bool{}
	g.VoteDisabled = map[string]string{}
	g.VoteWeights = map[string]int{}
	g.AutoSelfVoteVoters = map[string]bool{}
	g.AutoWastedVoters = map[string]bool{}
	g.RevoteDisallowByVoter = map[string]map[string]bool{}
	g.DoubleAgainst = ""
	g.TieBreakUsed = false
	for _, id := range g.aliveIDs() {
		g.VoteCandidates[id] = true
	}
	for _, p := range g.Players {
		if p == nil || p.Status != playerAlive {
			continue
		}
		if p.ForcedSelfVoteNext {
			g.markVoteSelf(p.PlayerID)
			p.ForcedSelfVoteNext = false
			p.ForcedWastedVoteNext = false
			continue
		}
		if p.ForcedWastedVoteNext {
			g.markVoteWasted(p.PlayerID, scenarioText(g.Scenario, "vote.spent", "Vote spent."))
			p.ForcedWastedVoteNext = false
		}
	}
	g.setResolutionNote("", "", nil)
	g.RoundRules.NoTalkUntilVoting = false
	g.setLastStage(scenarioText(g.Scenario, "event.votingStart", fmt.Sprintf("Round %d. Voting.", g.Round)), g.scenarioTextKey("event.votingStart", "classic.auto.119"), map[string]any{
		"round": g.Round,
		"v1":    g.Round,
	})
}

func (g *gameSession) finishGame(reason string) gameActionResult {
	if g.Phase == scenarioPhaseEnded && g.PostGame != nil && g.PostGame.IsActive {
		return gameActionResult{}
	}
	g.clearActiveTimer()
	g.ensureAllBunkerCardsRevealed()
	g.Phase = scenarioPhaseEnded
	g.VotePhase = ""
	g.CurrentTurnID = ""
	g.PostGame = &postGameStateView{
		IsActive:  true,
		EnteredAt: time.Now().UnixMilli(),
	}
	winners := make([]string, 0, len(g.Order))
	for _, id := range g.Order {
		player := g.Players[id]
		if player != nil && player.Status == playerAlive {
			winners = append(winners, player.Name)
		}
	}
	g.Winners = winners
	if len(winners) == 0 {
		g.setLastStage(scenarioText(g.Scenario, "classic.auto.117", "Game finished."), g.scenarioTextKey("event.gameEnd", "classic.auto.117"), map[string]any{
			"winners": "",
			"v1":      "",
			"v2":      "",
		})
	} else {
		g.setLastStage(scenarioText(g.Scenario, "classic.auto.117", fmt.Sprintf("Game finished. Winners: %s.", strings.Join(winners, ", "))), g.scenarioTextKey("event.gameEnd", "classic.auto.117"), map[string]any{
			"winners": strings.Join(winners, ", "),
			"v1":      strings.Join(winners, ", "),
			"v2":      "",
		})
	}
	event := g.makeEventLocalized("gameEnd", g.LastStageText, g.LastStageKey, g.LastStageVars)
	if reason != "" {
		event.Message = fmt.Sprintf("%s (%s)", event.Message, reason)
	}
	return gameActionResult{StateChanged: true, Events: []gameEvent{event}}
}

func (g *gameSession) startNextRoundOrEnd() gameActionResult {
	if g.shouldEnd() {
		return g.finishGame("Bunker capacity reached.")
	}
	nextRound := g.Round + 1
	if nextRound > len(g.Ruleset.VotesPerRnd) {
		return g.finishGame("No more rounds available.")
	}

	g.Round = nextRound
	g.clearActiveTimer()
	g.WorldEvent = nil
	g.revealNextBunkerCard(nextRound)
	g.Phase = scenarioPhaseReveal
	g.VotePhase = ""
	g.Votes = map[string]voteRecord{}
	g.BaseVotes = map[string]voteRecord{}
	g.VoteResults = map[string]voteRecord{}
	g.VoteCandidates = map[string]bool{}
	g.VoteDisabled = map[string]string{}
	g.VoteWeights = map[string]int{}
	g.AutoSelfVoteVoters = map[string]bool{}
	g.AutoWastedVoters = map[string]bool{}
	g.RevoteDisallowByVoter = map[string]map[string]bool{}
	g.DoubleAgainst = ""
	g.TieBreakUsed = false
	g.RevealedThisRnd = map[string]bool{}
	g.CurrentTurnID = g.firstAliveID()
	g.VotesRemaining = g.votesForRound(nextRound)
	g.LastRevealerID = ""
	g.setResolutionNote("", "", nil)
	g.RoundRules = roundRulesPublic{}
	g.setLastStage(scenarioText(g.Scenario, "classic.auto.121", fmt.Sprintf("Round %d. Reveal phase.", nextRound)), g.scenarioTextKey("event.roundStart", "classic.auto.121"), map[string]any{
		"round": nextRound,
		"v1":    nextRound,
	})
	g.scheduleRevealTimeoutIfNeeded()

	event := g.makeEventLocalized("roundStart", g.LastStageText, g.LastStageKey, g.LastStageVars)
	return gameActionResult{StateChanged: true, Events: []gameEvent{event}}
}

func (g *gameSession) shouldEnd() bool {
	aliveCount := len(g.aliveIDs())
	if aliveCount <= g.Ruleset.BunkerSeats {
		return true
	}
	if g.TotalExiles >= g.Ruleset.TotalExiles {
		return true
	}
	return false
}

func (g *gameSession) handleAction(actorID, actionType string, payload map[string]any) gameActionResult {
	player := g.Players[actorID]
	if player == nil {
		return g.actionErrorLocalized("Player not found.", g.scenarioTextKey("error.player.notFound", "error.player.notFound"), nil)
	}

	switch actionType {
	case "revealCard":
		cardID, _ := payload["cardId"].(string)
		return g.revealCard(actorID, cardID)
	case "continueRound":
		return g.continueRound(actorID)
	case "vote":
		targetID, _ := payload["targetPlayerId"].(string)
		return g.vote(actorID, targetID)
	case "finalizeVoting":
		return g.finalizeVoting(actorID)
	case "applySpecial":
		specialInstanceID, _ := payload["specialInstanceId"].(string)
		rawPayload, _ := payload["payload"].(map[string]any)
		if rawPayload == nil {
			rawPayload = map[string]any{}
		}
		return g.applySpecialWithPending(actorID, specialInstanceID, rawPayload)
	case "revealWorldThreat":
		index := asInt(payload["index"], -1)
		return g.revealWorldThreat(actorID, index)
	case "setBunkerOutcome":
		outcome, _ := payload["outcome"].(string)
		return g.setBunkerOutcome(actorID, outcome)
	case "devSkipRound":
		return g.startNextRoundOrEnd()
	case "devKickPlayer":
		targetID, _ := payload["targetPlayerId"].(string)
		return g.kickPlayer(targetID)
	case "devAddPlayer":
		name, _ := payload["name"].(string)
		return g.devAddPlayer(name)
	case "devRemovePlayer":
		targetID, _ := payload["targetPlayerId"].(string)
		return g.devRemovePlayer(targetID)
	case "markLeftBunker":
		targetID, _ := payload["targetPlayerId"].(string)
		return g.markLeftBunker(targetID)
	default:
		return g.actionErrorLocalized("Unknown action.", g.scenarioTextKey("error.action.unknown", "error.action.unknown"), nil)
	}
}

func (g *gameSession) revealCard(actorID, cardID string) gameActionResult {
	player := g.Players[actorID]
	if player == nil {
		return gameActionResult{Error: "Player not found."}
	}
	if g.Phase == scenarioPhaseEnded {
		if g.PostGame == nil || !g.PostGame.IsActive || g.PostGame.Outcome != "" {
			return g.actionErrorLocalized("Game has already ended.", g.scenarioTextKey("error.game.alreadyEnded", "error.game.alreadyEnded"), nil)
		}
		if player.Status == playerLeftBunker {
			return g.actionErrorLocalized("Player already left the bunker.", g.scenarioTextKey("error.player.alreadyLeftBunker", "error.player.alreadyLeftBunker"), nil)
		}
		for i := range player.Hand {
			card := &player.Hand[i]
			if card.InstanceID != cardID {
				continue
			}
			if card.Revealed {
				return g.actionErrorLocalized("This card has already been revealed.", "error.card.alreadyRevealed", nil)
			}
			card.Revealed = true
			return gameActionResult{StateChanged: true}
		}
		return g.actionErrorLocalized("Card not found.", "error.card.notFound", nil)
	}
	if g.Phase != scenarioPhaseReveal {
		return g.actionErrorLocalized("You cannot reveal cards right now.", "error.reveal.notNow", nil)
	}
	if actorID != g.CurrentTurnID {
		return g.actionErrorLocalized("It is another player's turn right now.", "error.turn.otherPlayer", nil)
	}
	if player.Status != playerAlive {
		return g.actionErrorLocalized("You have been excluded from the game.", "error.player.excluded", nil)
	}
	var targetCard *handCard
	for i := range player.Hand {
		card := &player.Hand[i]
		if card.InstanceID == cardID {
			targetCard = card
			break
		}
	}
	if targetCard == nil {
		return g.actionErrorLocalized("Card not found.", "error.card.notFound", nil)
	}
	if targetCard.Revealed {
		return g.actionErrorLocalized("This card has already been revealed.", "error.card.alreadyRevealed", nil)
	}
	if g.RoundRules.ForcedCategory != "" {
		forced := g.RoundRules.ForcedCategory
		cfg, hasCfg := resolveCategoryDeckSlot(forced)
		hasForcedHidden := false
		if hasCfg {
			for _, card := range player.Hand {
				if card.Deck != cfg.deck {
					continue
				}
				if cfg.slot != "" && card.SlotKey != cfg.slot {
					continue
				}
				if !card.Revealed {
					hasForcedHidden = true
					break
				}
			}
		}
		matchesForced := true
		if hasCfg {
			matchesForced = targetCard.Deck == cfg.deck && (cfg.slot == "" || targetCard.SlotKey == cfg.slot)
		}
		if hasForcedHidden && !matchesForced {
			return g.actionErrorLocalized(
				fmt.Sprintf("This round you must reveal a card from category \"%s\".", forced),
				"error.reveal.forcedCategory",
				map[string]any{"category": forced},
			)
		}
	}
	targetCard.Revealed = true
	if targetCard.Deck == healthDeckName() && g.FirstHealthRevealerID == "" {
		g.FirstHealthRevealerID = actorID
	}
	g.RevealedThisRnd[actorID] = true
	g.LastRevealerID = actorID
	g.setLastStage(scenarioText(g.Scenario, "classic.auto.140", fmt.Sprintf("%s revealed a card.", player.Name)), g.scenarioTextKey("event.reveal.card", "classic.auto.140"), map[string]any{
		"name": player.Name,
		"v1":   player.Name,
	})
	enterResult := g.enterRevealDiscussion()
	events := []gameEvent{g.makeEventLocalized("info", g.LastStageText, g.LastStageKey, g.LastStageVars)}
	events = append(events, enterResult.Events...)
	return gameActionResult{StateChanged: true, Events: events}
}
func (g *gameSession) continueRound(actorID string) gameActionResult {
	if g.Phase == scenarioPhaseEnded {
		return g.actionErrorLocalized("Game has already ended.", g.scenarioTextKey("error.game.alreadyEnded", "error.game.alreadyEnded"), nil)
	}
	if !g.canContinue(actorID) {
		switch g.Settings.ContinuePermission {
		case "host_only":
			return g.actionErrorLocalized("Only the host can continue the turn.", "error.continue.hostOnly", nil)
		case "revealer_only":
			return g.actionErrorLocalized("Only the player who revealed the card can continue.", "error.continue.revealerOnly", nil)
		default:
			return g.actionErrorLocalized("Cannot continue the turn right now.", "error.continue.notNow", nil)
		}
	}
	if g.Phase == scenarioPhaseRevealDiscussion {
		return g.advanceAfterDiscussion()
	}
	switch g.Phase {
	case scenarioPhaseResolution:
		if g.shouldEnd() {
			return g.finishGame("after_resolution")
		}
		if g.VotesRemaining > 0 {
			g.startVoting()
			return gameActionResult{
				StateChanged: true,
				Events:       []gameEvent{g.makeEventLocalized("votingStart", g.LastStageText, g.LastStageKey, g.LastStageVars)},
			}
		}
		return g.startNextRoundOrEnd()
	case scenarioPhaseVoting:
		if g.VotePhase == votePhaseSpecialWindow {
			return g.finalizeVoting(actorID)
		}
		return g.actionErrorLocalized("Cannot continue the turn right now.", "error.continue.notNow", nil)
	default:
		return g.actionErrorLocalized("Cannot continue the turn right now.", "error.continue.notNow", nil)
	}
}
func (g *gameSession) canContinue(actorID string) bool {
	permission := g.Settings.ContinuePermission
	switch permission {
	case "anyone":
		return true
	case "host_only":
		return actorID == g.HostID
	case "revealer_only":
		if g.Phase == scenarioPhaseRevealDiscussion {
			return g.LastRevealerID == actorID
		}
		return actorID == g.HostID
	default:
		return actorID == g.HostID
	}
}

func (g *gameSession) vote(actorID, targetID string) gameActionResult {
	if g.Phase != scenarioPhaseVoting || g.VotePhase != votePhaseVoting {
		return g.actionErrorLocalized("There is no voting right now.", "error.voting.notNow", nil)
	}
	if actorID == targetID {
		return g.actionErrorLocalized("You cannot vote for yourself.", "error.vote.self", nil)
	}
	if !g.VoteCandidates[targetID] {
		return g.actionErrorLocalized("Invalid candidate.", "error.vote.invalidCandidate", nil)
	}
	if perVoter := g.RevoteDisallowByVoter[actorID]; perVoter != nil && perVoter[targetID] {
		return g.actionErrorLocalized("You cannot vote for this candidate.", "error.vote.disallowedCandidate", nil)
	}
	if reason, blocked := g.VoteDisabled[actorID]; blocked {
		if reason == "" {
			return g.actionErrorLocalized("Your vote is blocked.", "error.vote.blocked", nil)
		}
		return gameActionResult{Error: reason}
	}
	actor := g.Players[actorID]
	target := g.Players[targetID]
	if actor == nil || target == nil {
		return gameActionResult{Error: "Player not found."}
	}
	if actor.Status != playerAlive {
		return g.actionErrorLocalized("You have been excluded from the game.", "error.player.excluded", nil)
	}
	if target.Status != playerAlive {
		return g.actionErrorLocalized("The candidate is not in the game.", "error.vote.candidateNotAlive", nil)
	}
	if target.BannedAgainst[actorID] {
		return g.actionErrorLocalized("You cannot vote against this player.", "error.vote.cannotAgainst", nil)
	}
	if _, exists := g.Votes[actorID]; exists {
		return g.actionErrorLocalized("You have already voted.", "error.vote.alreadySubmitted", nil)
	}
	g.Votes[actorID] = voteRecord{
		TargetID:  targetID,
		Submitted: time.Now().UnixMilli(),
		IsValid:   true,
	}
	aliveCount := len(g.aliveIDs())
	if len(g.Votes) >= aliveCount {
		return g.enterVoteSpecialWindow()
	}
	return gameActionResult{StateChanged: true}
}
func (g *gameSession) finalizeVoting(actorID string) gameActionResult {
	if g.Phase != scenarioPhaseVoting {
		return g.actionErrorLocalized("Voting has not started.", "error.voting.notStarted", nil)
	}
	if g.VotePhase != votePhaseSpecialWindow && g.VotePhase != votePhaseVoting {
		return g.actionErrorLocalized("Voting has already been finalized.", g.scenarioTextKey("error.control.finalizeVoting", "error.control.finalizeVoting"), nil)
	}
	if actorID != g.HostID {
		return g.actionErrorLocalized("Only CONTROL can finalize voting.", g.scenarioTextKey("error.control.finalizeVoting", "error.control.finalizeVoting"), nil)
	}
	g.clearActiveTimer()
	source := g.currentVoteSource()
	if len(source) == 0 && len(g.BaseVotes) > 0 {
		source = g.BaseVotes
	}
	g.VoteResults = copyVotes(source)
	_, topCandidates := g.computeVoteTotals(source)
	if len(topCandidates) > 1 && !g.TieBreakUsed {
		g.startTieBreakRevote(topCandidates)
		return gameActionResult{
			StateChanged: true,
			Events:       []gameEvent{g.makeEventLocalized("info", g.LastStageText, g.LastStageKey, g.LastStageVars)},
		}
	}
	eliminatedID := ""
	if len(topCandidates) > 0 {
		eliminatedID = topCandidates[g.rng.Intn(len(topCandidates))]
	}
	events := make([]gameEvent, 0, 2)
	if eliminatedID != "" {
		triggerEvents := g.applyElimination(eliminatedID)
		if g.LastEliminatedID != "" {
			name := g.playerName(g.LastEliminatedID)
			g.setResolutionNote(scenarioText(g.Scenario, "classic.auto.122", fmt.Sprintf("Voting result: %s eliminated.", name)), g.scenarioTextKey("event.voting.eliminated", "classic.auto.122"), map[string]any{
				"name": name,
				"v1":   name,
			})
		}
		events = append(events, triggerEvents...)
	} else {
		g.setResolutionNote(scenarioText(g.Scenario, "event.tie.noEliminated", "Tie: no one eliminated."), g.scenarioTextKey("event.tie.noEliminated", "event.tie.noEliminated"), nil)
		events = append(events, g.makeEvent("info", g.ResolutionNote))
	}
	g.Phase = scenarioPhaseResolution
	g.VotePhase = votePhaseResolve
	g.setLastStage(g.ResolutionNote, g.ResolutionNoteKey, g.ResolutionNoteVars)
	g.Votes = map[string]voteRecord{}
	if g.shouldEnd() {
		end := g.finishGame("Bunker capacity reached.")
		events = append(events, end.Events...)
	} else {
		g.schedulePhaseTimer(timerKindResolutionAuto, 2)
	}
	return gameActionResult{StateChanged: true, Events: events}
}
func (g *gameSession) kickPlayer(targetID string) gameActionResult {
	player := g.Players[targetID]
	if player == nil {
		return g.actionErrorLocalized("Player not found.", g.scenarioTextKey("error.player.notFound", "error.player.notFound"), nil)
	}
	if player.Status != playerAlive {
		return g.actionErrorLocalized("Player is already excluded.", g.scenarioTextKey("error.player.excluded", "error.player.excluded"), nil)
	}
	triggerEvents := g.applyElimination(targetID)
	g.setResolutionNote(
		fmt.Sprintf("%s was eliminated.", player.Name),
		g.scenarioTextKey("event.elimination.resolution", "event.elimination.resolution"),
		map[string]any{"name": player.Name},
	)
	g.removeFromVoting(targetID)
	if g.CurrentTurnID == targetID {
		next := g.nextUnrevealedAliveAfter(targetID)
		if next == "" {
			next = g.firstAliveID()
		}
		g.CurrentTurnID = next
	}
	eliminationEvent := g.makeEventLocalized("elimination", g.ResolutionNote, g.ResolutionNoteKey, g.ResolutionNoteVars)
	if g.shouldEnd() {
		end := g.finishGame("Bunker capacity reached.")
		return gameActionResult{
			StateChanged: true,
			Events: append(
				append([]gameEvent{eliminationEvent}, triggerEvents...),
				end.Events...,
			),
		}
	}
	return gameActionResult{
		StateChanged: true,
		Events:       append([]gameEvent{eliminationEvent}, triggerEvents...),
	}
}

func (g *gameSession) drawInitialHandForPlayer(playerID string) []handCard {
	drawFor := func(deckName, slot string) handCard {
		card, ok := g.drawCardFromDeck(deckName)
		if !ok {
			return handCard{
				InstanceID: g.nextCardInstanceID(playerID),
				CardID:     "",
				Deck:       deckName,
				SlotKey:    slot,
				Label:      scenarioCardNoCardLabel(g.Scenario),
				Revealed:   false,
				Missing:    true,
			}
		}
		return handCard{
			InstanceID: g.nextCardInstanceID(playerID),
			CardID:     card.ID,
			Deck:       card.Deck,
			SlotKey:    slot,
			Label:      card.Label,
			Revealed:   false,
		}
	}

	hand := make([]handCard, 0, len(coreDecks)+len(factsSlotOrder))
	for _, deck := range coreDecks {
		hand = append(hand, drawFor(deck, ""))
	}
	hand = append(hand, drawFor(factsDeck, "facts1"))
	hand = append(hand, drawFor(factsDeck, "facts2"))
	return hand
}

func (g *gameSession) devAddPlayer(name string) gameActionResult {
	if !g.IsDev {
		return gameActionResult{Error: "DEV actions are only available in dev_test."}
	}
	g.DevBotCounter++
	playerID := fmt.Sprintf("dev-%s-%d", strings.ToLower(g.RoomCode), g.DevBotCounter)
	for g.Players[playerID] != nil {
		g.DevBotCounter++
		playerID = fmt.Sprintf("dev-%s-%d", strings.ToLower(g.RoomCode), g.DevBotCounter)
	}

	fallbackName := fmt.Sprintf("%s %d", scenarioBotPrefix(g.Scenario), g.DevBotCounter)
	displayName := sanitizeHumanText(strings.TrimSpace(name), fallbackName)
	newPlayer := &gamePlayer{
		PlayerID:                  playerID,
		Name:                      displayName,
		Status:                    playerAlive,
		Hand:                      g.drawInitialHandForPlayer(playerID),
		Specials:                  nil,
		SpecialCategoryProxyCards: nil,
		IsBot:                     true,
		BannedAgainst:             map[string]bool{},
	}
	g.assignInitialSpecialsForPlayer(newPlayer, g.buildSpecialAssetLookup())
	g.Players[playerID] = newPlayer
	g.Order = append(g.Order, playerID)

	if g.Phase == scenarioPhaseVoting {
		g.VoteCandidates[playerID] = true
	}
	if g.CurrentTurnID == "" && g.Phase == scenarioPhaseReveal {
		g.CurrentTurnID = g.firstAliveID()
	}
	return gameActionResult{
		StateChanged: true,
		Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.playerAdded", fmt.Sprintf("DEV: Player %s added.", displayName)),
			"event.playerAdded",
			map[string]any{"name": displayName},
		)},
	}
}

func (g *gameSession) devRemovePlayer(targetID string) gameActionResult {
	if !g.IsDev {
		return gameActionResult{Error: "DEV actions are only available in dev_test."}
	}

	removeID := strings.TrimSpace(targetID)
	if removeID == "" {
		for i := len(g.Order) - 1; i >= 0; i-- {
			candidateID := g.Order[i]
			candidate := g.Players[candidateID]
			if candidate != nil && candidate.IsBot {
				removeID = candidateID
				break
			}
		}
	}
	if removeID == "" {
		return g.actionErrorLocalized("No player available for removal.", g.scenarioTextKey("error.removePlayer.none", ""), nil)
	}

	target := g.Players[removeID]
	if target == nil {
		return gameActionResult{Error: "Player not found."}
	}
	if removeID == g.HostID {
		return gameActionResult{Error: "Cannot remove the host."}
	}

	delete(g.Players, removeID)
	filteredOrder := make([]string, 0, len(g.Order))
	for _, id := range g.Order {
		if id != removeID {
			filteredOrder = append(filteredOrder, id)
		}
	}
	g.Order = filteredOrder
	g.removeFromVoting(removeID)

	if g.CurrentTurnID == removeID {
		next := g.nextUnrevealedAliveAfter(removeID)
		if next == "" {
			next = g.firstAliveID()
		}
		g.CurrentTurnID = next
	}

	events := []gameEvent{g.makeEventLocalized(
		"info",
		scenarioText(g.Scenario, "event.playerRemoved", fmt.Sprintf("DEV: Player %s removed.", target.Name)),
		"event.playerRemoved",
		map[string]any{"name": target.Name},
	)}
	if g.shouldEnd() {
		end := g.finishGame("Bunker capacity reached.")
		events = append(events, end.Events...)
	}
	return gameActionResult{StateChanged: true, Events: events}
}

func (g *gameSession) markLeftBunker(targetID string) gameActionResult {
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		return g.actionErrorLocalized("Player not found.", g.scenarioTextKey("error.player.notFound", "error.player.notFound"), nil)
	}
	player := g.Players[targetID]
	if player == nil {
		return g.actionErrorLocalized("Player not found.", g.scenarioTextKey("error.player.notFound", "error.player.notFound"), nil)
	}
	if player.Status == playerLeftBunker {
		return g.actionErrorLocalized("Player already left the bunker.", g.scenarioTextKey("error.player.alreadyLeftBunker", "error.player.alreadyLeftBunker"), nil)
	}

	player.Status = playerLeftBunker
	delete(g.RevealedThisRnd, targetID)
	g.removeFromVoting(targetID)
	if g.CurrentTurnID == targetID {
		next := g.nextUnrevealedAliveAfter(targetID)
		if next == "" {
			next = g.firstAliveID()
		}
		g.CurrentTurnID = next
	}

	events := []gameEvent{g.makeEventLocalized(
		"playerLeftBunker",
		scenarioText(g.Scenario, "event.player.leftBunker", fmt.Sprintf("%s left the bunker.", player.Name)),
		g.scenarioTextKey("event.player.leftBunker", "event.player.leftBunker"),
		map[string]any{"name": player.Name},
	)}

	if g.Phase == scenarioPhaseResolution {
		g.setResolutionNote(
			scenarioText(g.Scenario, "event.player.leftBunker", fmt.Sprintf("%s left the bunker.", player.Name)),
			g.scenarioTextKey("event.player.leftBunker", "event.player.leftBunker"),
			map[string]any{"name": player.Name},
		)
		g.setLastStage(g.ResolutionNote, g.ResolutionNoteKey, g.ResolutionNoteVars)
		events = append(events, g.makeEventLocalized("info", g.ResolutionNote, g.ResolutionNoteKey, g.ResolutionNoteVars))
		if g.shouldEnd() {
			end := g.finishGame("Bunker capacity reached.")
			events = append(events, end.Events...)
		} else {
			g.schedulePhaseTimer(timerKindResolutionAuto, 2)
		}
		return gameActionResult{StateChanged: true, Events: events}
	}

	if g.shouldEnd() {
		end := g.finishGame("Bunker capacity reached.")
		events = append(events, end.Events...)
	}
	return gameActionResult{StateChanged: true, Events: events}
}
func (g *gameSession) makeEvent(kind, message string) gameEvent {
	return g.makeEventLocalized(kind, message, "", nil)
}
func (g *gameSession) actionErrorLocalized(fallback, key string, vars map[string]any) gameActionResult {
	result := gameActionResult{Error: fallback}
	if key == "" {
		return result
	}
	if g.scenarioSupportsLocalizedScenarioText() {
		result.ErrorKey = key
		result.ErrorVars = cloneLocalizedVars(vars)
	}
	return result
}

func (g *gameSession) scenarioTextKey(devKey, classicKey string) string {
	switch g.Scenario {
	case scenarioClassic:
		return classicKey
	case scenarioDevTest:
		return devKey
	default:
		return ""
	}
}

func (g *gameSession) makeEventLocalized(kind, message, messageKey string, messageVars map[string]any) gameEvent {
	g.EventCounter++
	safeMessage := sanitizeHumanText(message, defaultEventMessage(kind))
	if !g.scenarioSupportsLocalizedScenarioText() {
		messageKey = ""
		messageVars = nil
	}
	return gameEvent{
		ID:          fmt.Sprintf("%s-%d-%d", g.RoomCode, time.Now().UnixMilli(), g.EventCounter),
		Kind:        kind,
		Message:     safeMessage,
		MessageKey:  messageKey,
		MessageVars: cloneLocalizedVars(messageVars),
		CreatedAt:   time.Now().UnixMilli(),
	}
}

func (g *gameSession) scenarioSupportsLocalizedScenarioText() bool {
	return g.Scenario == scenarioClassic || g.Scenario == scenarioDevTest
}

func cloneLocalizedVars(vars map[string]any) map[string]any {
	if len(vars) == 0 {
		return nil
	}
	out := make(map[string]any, len(vars))
	for key, value := range vars {
		out[key] = value
	}
	return out
}

func (g *gameSession) setLastStage(text, key string, vars map[string]any) {
	g.LastStageText = text
	if !g.scenarioSupportsLocalizedScenarioText() {
		g.LastStageKey = ""
		g.LastStageVars = nil
		return
	}
	g.LastStageKey = key
	g.LastStageVars = cloneLocalizedVars(vars)
}

func (g *gameSession) setResolutionNote(text, key string, vars map[string]any) {
	g.ResolutionNote = text
	if !g.scenarioSupportsLocalizedScenarioText() {
		g.ResolutionNoteKey = ""
		g.ResolutionNoteVars = nil
		return
	}
	g.ResolutionNoteKey = key
	g.ResolutionNoteVars = cloneLocalizedVars(vars)
}

func (g *gameSession) buildGameView(room *room, playerID string) gameView {
	view := gameView{
		Phase:         g.Phase,
		Round:         g.Round,
		Categories:    append([]string(nil), categoryOrder...),
		LastStage:     sanitizeHumanText(g.LastStageText, defaultStageMessage(g.Round)),
		LastStageKey:  g.LastStageKey,
		LastStageVars: cloneLocalizedVars(g.LastStageVars),
		Ruleset:       g.Ruleset,
		World:         g.buildWorldView(),
		WorldEvent:    g.WorldEvent,
		PostGame:      g.PostGame,
	}

	you := g.Players[playerID]
	if you == nil {
		you = &gamePlayer{PlayerID: playerID, Name: "Unknown", Status: playerEliminated}
	}
	view.You.PlayerID = you.PlayerID
	view.You.Name = you.Name
	view.You.Hand = make([]cardRef, 0, len(you.Hand))
	for _, card := range you.Hand {
		imgURL := ""
		if card.CardID != "" {
			imgURL = "/assets/" + card.CardID
		}
		view.You.Hand = append(view.You.Hand, cardRef{
			ID:       card.CardID,
			Deck:     card.Deck,
			Instance: card.InstanceID,
			Label:    card.Label,
			ImgURL:   imgURL,
			Missing:  card.Missing,
			Revealed: card.Revealed,
		})
	}
	view.You.Categories = g.buildYouCategories(you)
	view.You.Specials = g.buildSpecialInstances(you)

	view.Public.Players = g.buildPublicPlayers(room)
	view.Public.RevealedThisRound = g.revealedThisRoundIDs()
	view.Public.RoundRevealedCount = len(view.Public.RevealedThisRound)
	view.Public.RoundTotalAlive = len(g.aliveIDs())
	if g.CurrentTurnID != "" {
		current := g.CurrentTurnID
		view.Public.CurrentTurnPlayerID = &current
	}
	view.Public.VotesRemaining = g.VotesRemaining
	view.Public.VotesTotal = g.votesForRound(g.Round)
	view.Public.RevealLimit = len(g.aliveIDs())
	view.Public.YourVoteWeight = g.currentVoteWeightFor(playerID)
	view.Public.CanOpenVotingModal = g.VotePhase != ""
	view.Public.VoteModalOpen = g.VotePhase == votePhaseVoting
	view.Public.LastEliminated = g.LastEliminatedID
	view.Public.Winners = append([]string(nil), g.Winners...)
	view.Public.ResolutionNote = sanitizeHumanText(g.ResolutionNote, "Voting results updated.")
	view.Public.ResolutionNoteKey = g.ResolutionNoteKey
	view.Public.ResolutionNoteVars = cloneLocalizedVars(g.ResolutionNoteVars)
	if g.RoundRules.NoTalkUntilVoting || g.RoundRules.ForcedCategory != "" {
		rulesCopy := g.RoundRules
		if rulesCopy.ForcedCategory != "" {
			if key := resolveCategoryKey(rulesCopy.ForcedCategory); key != "" {
				rulesCopy.ForcedCategory = canonicalCategoryKey(key)
			} else {
				if looksLikeMojibake(rulesCopy.ForcedCategory) {
					rulesCopy.ForcedCategory = ""
				}
			}
		}
		view.Public.RoundRules = &rulesCopy
	}
	if g.ActiveTimer != nil {
		timerCopy := *g.ActiveTimer
		view.Public.ActiveTimer = &timerCopy
	}
	view.Public.CanContinue = g.canContinue(playerID)
	modifier := g.currentThreatModifier()
	view.Public.ThreatModifier = &modifier

	if g.VotePhase != "" {
		phase := g.VotePhase
		view.Public.VotePhase = &phase
		view.Public.DisallowedVoteTargetIDsForYou = g.disallowedVoteTargetsFor(playerID)
		view.Public.Voting = &votingView{HasVoted: g.playerHasVoted(playerID)}
		view.Public.VotingProgress = &votingProgress{
			Voted: len(g.currentVoteSource()),
			Total: len(g.aliveIDs()),
		}
		view.Public.VotesPublic = g.buildVotesPublic()
	}

	return view
}

func (g *gameSession) currentVoteSource() map[string]voteRecord {
	if g.VotePhase == votePhaseResolve {
		if len(g.VoteResults) > 0 {
			return g.VoteResults
		}
		if len(g.BaseVotes) > 0 {
			return g.BaseVotes
		}
	}
	if g.VotePhase == votePhaseSpecialWindow && len(g.BaseVotes) > 0 {
		return g.BaseVotes
	}
	if g.VotePhase == votePhaseResolve {
		return g.VoteResults
	}
	return g.Votes
}

func (g *gameSession) disallowedVoteTargetsFor(playerID string) []string {
	perVoter := g.RevoteDisallowByVoter[playerID]
	if len(perVoter) == 0 {
		return nil
	}
	out := make([]string, 0, len(perVoter))
	for targetID := range perVoter {
		out = append(out, targetID)
	}
	sort.Strings(out)
	return out
}

func (g *gameSession) playerHasVoted(playerID string) bool {
	_, exists := g.currentVoteSource()[playerID]
	return exists
}

func (g *gameSession) currentVoteWeightFor(playerID string) int {
	if playerID == "" {
		return 1
	}
	if weight, ok := g.VoteWeights[playerID]; ok && weight > 0 {
		return weight
	}
	return 1
}

func (g *gameSession) revealedThisRoundIDs() []string {
	out := make([]string, 0, len(g.RevealedThisRnd))
	for _, id := range g.Order {
		if g.RevealedThisRnd[id] {
			out = append(out, id)
		}
	}
	return out
}

func (g *gameSession) buildVotesPublic() []votePublic {
	source := g.currentVoteSource()
	effective := g.buildEffectiveVotes(source)
	out := make([]votePublic, 0, len(g.Order))
	for _, playerID := range g.Order {
		player := g.Players[playerID]
		if player == nil {
			continue
		}
		info, ok := effective[playerID]
		if !ok || info.Status == "not_voted" {
			out = append(out, votePublic{
				VoterID:   playerID,
				VoterName: sanitizeHumanText(player.Name, "Player"),
				Status:    "not_voted",
			})
			continue
		}
		if info.Status == "voted" && info.TargetID != "" {
			targetName := "Unknown player"
			if target := g.Players[info.TargetID]; target != nil {
				targetName = sanitizeHumanText(target.Name, "Unknown player")
			}
			out = append(out, votePublic{
				VoterID:    playerID,
				VoterName:  sanitizeHumanText(player.Name, "Player"),
				TargetID:   info.TargetID,
				TargetName: sanitizeHumanText(targetName, "Unknown player"),
				Status:     "voted",
				ReasonCode: g.inferVoteReasonCode(playerID, info),
				ReasonKey:  g.inferVoteReasonKey(playerID, info),
				ReasonVars: g.inferVoteReasonVars(),
				Submitted:  info.Submitted,
			})
			continue
		}
		out = append(out, votePublic{
			VoterID:   playerID,
			VoterName: sanitizeHumanText(player.Name, "Player"),
			Status:    "invalid",
			Reason:    sanitizeHumanText(info.ReasonText, "Vote rejected."),
			ReasonKey: g.inferVoteReasonKey(playerID, info),
			ReasonCode: g.inferVoteReasonCode(
				playerID,
				info,
			),
			ReasonVars: g.inferVoteReasonVars(),
			Submitted:  info.Submitted,
		})
	}
	return out
}

func (g *gameSession) inferVoteReasonCode(playerID string, info effectiveVoteRecord) string {
	if info.Status == "blocked" {
		normalized := strings.ToLower(strings.TrimSpace(info.ReasonText))
		switch {
		case strings.Contains(normalized, "vote.blocked.byspecial"),
			strings.Contains(normalized, "blocked by special"),
			strings.Contains(normalized, "vote blocked"):
			return "VOTE_BLOCKED"
		case strings.Contains(normalized, "vote.spent.byspecial"),
			strings.Contains(normalized, "spent by special"),
			strings.Contains(normalized, "vote spent"),
			strings.Contains(normalized, "spent"):
			return "VOTE_SPENT"
		}
	}
	if info.Status == "voted" {
		if g.AutoSelfVoteVoters[playerID] && info.TargetID == playerID {
			return "VOTE_FORCED_SELF"
		}
		return ""
	}
	if g.AutoSelfVoteVoters[playerID] {
		return "VOTE_FORCED_SELF"
	}
	if g.AutoWastedVoters[playerID] {
		return "VOTE_SPENT"
	}
	if _, blocked := g.VoteDisabled[playerID]; blocked {
		normalized := strings.ToLower(strings.TrimSpace(info.ReasonText))
		switch {
		case strings.Contains(normalized, "vote.blocked.byspecial"),
			strings.Contains(normalized, "blocked by special"),
			strings.Contains(normalized, "vote blocked"):
			return "VOTE_BLOCKED"
		case strings.Contains(normalized, "vote.spent.byspecial"),
			strings.Contains(normalized, "spent by special"),
			strings.Contains(normalized, "vote spent"),
			strings.Contains(normalized, "spent"):
			return "VOTE_SPENT"
		}
	}
	if info.TargetID != "" {
		if disallowed := g.RevoteDisallowByVoter[playerID]; len(disallowed) > 0 && disallowed[info.TargetID] {
			return "VOTE_TARGET_DISALLOWED"
		}
		if g.DoubleAgainst != "" && info.TargetID == g.DoubleAgainst {
			return "VOTE_BANNED_AGAINST_TARGET"
		}
	}
	normalized := strings.ToLower(strings.TrimSpace(info.ReasonText))
	switch {
	case strings.Contains(normalized, "cannot vote against this player"),
		strings.Contains(normalized, "vote disallowed"),
		strings.Contains(normalized, "target disallowed"):
		return "VOTE_BANNED_AGAINST_TARGET"
	default:
		return ""
	}
}

func (g *gameSession) inferVoteReasonKey(playerID string, info effectiveVoteRecord) string {
	if !g.scenarioSupportsLocalizedScenarioText() {
		return ""
	}
	switch g.inferVoteReasonCode(playerID, info) {
	case "VOTE_BLOCKED":
		return "error.vote.blocked"
	case "VOTE_TARGET_DISALLOWED":
		return "error.vote.disallowedCandidate"
	case "VOTE_TARGET_UNAVAILABLE":
		return "error.vote.candidateNotAlive"
	case "VOTE_BANNED_AGAINST_TARGET":
		return "error.vote.cannotAgainst"
	default:
		return ""
	}
}

func (g *gameSession) inferVoteReasonVars() map[string]any {
	return nil
}

func (g *gameSession) buildPublicPlayers(room *room) []publicPlayerView {
	out := make([]publicPlayerView, 0, len(g.Order))
	now := time.Now().UnixMilli()
	for _, playerID := range g.Order {
		gamePlayer := g.Players[playerID]
		if gamePlayer == nil {
			continue
		}
		var roomPlayer *player
		if room != nil && room.Players != nil {
			roomPlayer = room.Players[playerID]
		}
		connected := roomPlayer != nil && roomPlayer.Connected
		public := publicPlayerView{
			PlayerID:        gamePlayer.PlayerID,
			Name:            gamePlayer.Name,
			Status:          gamePlayer.Status,
			Connected:       connected,
			LeftBunker:      gamePlayer.Status == playerLeftBunker,
			RevealedCards:   make([]cardRef, 0, 8),
			SpecialRevealed: (g.IsDev && len(gamePlayer.Specials) > 0) || g.playerHasRevealedSpecial(gamePlayer),
			Categories:      make([]publicCategorySlot, 0, len(categoryOrder)),
		}
		if roomPlayer != nil {
			if roomPlayer.DisconnectedAtMS != nil {
				disconnectedAt := *roomPlayer.DisconnectedAtMS
				public.DisconnectedAt = &disconnectedAt
			}
			totalAbsent := roomPlayer.TotalAbsentMS
			public.TotalAbsentMS = &totalAbsent
			currentOffline := int64(0)
			if !roomPlayer.Connected && roomPlayer.DisconnectedAtMS != nil {
				currentOffline = maxInt64(0, now-*roomPlayer.DisconnectedAtMS)
			}
			public.CurrentOffMS = &currentOffline
			remaining := maxInt64(0, disconnectGraceMS-(totalAbsent+currentOffline))
			public.KickRemainMS = &remaining
		}

		revealedCount := 0
		for _, card := range gamePlayer.Hand {
			if !card.Revealed {
				continue
			}
			revealedCount++
			public.RevealedCards = append(public.RevealedCards, cardRef{
				ID:       card.CardID,
				Deck:     card.Deck,
				Instance: card.InstanceID,
				Label:    card.Label,
				ImgURL: func() string {
					if card.CardID == "" {
						return ""
					}
					return "/assets/" + card.CardID
				}(),
				Missing: card.Missing,
			})
		}
		public.RevealedCount = revealedCount
		public.TotalCards = len(gamePlayer.Hand)
		public.Categories = g.buildPublicCategories(gamePlayer)
		out = append(out, public)
	}
	return out
}

func (g *gameSession) buildPublicCategories(player *gamePlayer) []publicCategorySlot {
	slots := make([]publicCategorySlot, 0, len(categoryOrder))
	for _, category := range categoryOrder {
		if len(categoryOrder) > 0 && category == categoryOrder[len(categoryOrder)-1] {
			cards := make([]publicCategoryCard, 0, len(player.Specials)+len(player.SpecialCategoryProxyCards))
			for _, special := range player.Specials {
				revealed := special.RevealedPublic || g.IsDev
				imgURL := ""
				if revealed {
					imgURL = specialImageURL(special.Definition)
				}
				cards = append(cards, publicCategoryCard{
					InstanceID:   special.InstanceID,
					Label:        special.Definition.Title,
					ImgURL:       imgURL,
					Revealed:     revealed,
					Hidden:       !revealed,
					BackCategory: specialDeckCategoryName,
				})
			}
			cards = append(cards, player.SpecialCategoryProxyCards...)
			status := "hidden"
			if slices.ContainsFunc(cards, func(card publicCategoryCard) bool { return !card.Hidden }) {
				status = "revealed"
			}
			slots = append(slots, publicCategorySlot{
				Category: category,
				Status:   status,
				Cards:    cards,
			})
			continue
		}
		cfg, ok := resolveCategoryDeckSlot(category)
		if !ok {
			slots = append(slots, publicCategorySlot{
				Category: category,
				Status:   "hidden",
				Cards:    nil,
			})
			continue
		}
		cards := make([]publicCategoryCard, 0, 2)
		hiddenCards := make([]publicCategoryCard, 0, 2)
		hasRevealed := false
		for _, card := range player.Hand {
			if card.Deck != cfg.deck {
				continue
			}
			if cfg.slot != "" && card.SlotKey != cfg.slot {
				continue
			}
			backCategory := category
			if trimmed := strings.TrimSpace(card.PublicBackCategory); trimmed != "" {
				backCategory = trimmed
			}
			if !g.IsDev && !card.Revealed {
				hiddenCards = append(hiddenCards, publicCategoryCard{
					InstanceID:   card.InstanceID,
					Label:        scenarioCardNoCardLabel(g.Scenario),
					Revealed:     false,
					Hidden:       true,
					BackCategory: backCategory,
				})
				continue
			}
			imgURL := ""
			if card.CardID != "" {
				imgURL = "/assets/" + card.CardID
			}
			hasRevealed = hasRevealed || card.Revealed || g.IsDev
			cards = append(cards, publicCategoryCard{
				InstanceID:   card.InstanceID,
				Label:        card.Label,
				ImgURL:       imgURL,
				Revealed:     card.Revealed || g.IsDev,
				BackCategory: backCategory,
			})
		}
		cards = append(cards, hiddenCards...)
		status := "hidden"
		if hasRevealed {
			status = "revealed"
		}
		slots = append(slots, publicCategorySlot{
			Category: category,
			Status:   status,
			Cards:    cards,
		})
	}
	return slots
}

func (g *gameSession) buildYouCategories(player *gamePlayer) []youCategorySlot {
	slots := make([]youCategorySlot, 0, len(categoryOrder)-1)
	for _, category := range categoryOrder {
		if len(categoryOrder) > 0 && category == categoryOrder[len(categoryOrder)-1] {
			continue
		}
		cfg, ok := resolveCategoryDeckSlot(category)
		if !ok {
			slots = append(slots, youCategorySlot{
				Category: category,
				Cards:    nil,
			})
			continue
		}
		cards := make([]youCategoryCard, 0, 2)
		for _, card := range player.Hand {
			if card.Deck != cfg.deck {
				continue
			}
			if cfg.slot != "" && card.SlotKey != cfg.slot {
				continue
			}
			imgURL := ""
			if card.CardID != "" {
				imgURL = "/assets/" + card.CardID
			}
			cards = append(cards, youCategoryCard{
				InstanceID: card.InstanceID,
				Label:      card.Label,
				Deck:       cfg.deck,
				Revealed:   card.Revealed,
				ImgURL:     imgURL,
			})
		}
		sort.Slice(cards, func(i, j int) bool {
			return cards[i].InstanceID < cards[j].InstanceID
		})
		slots = append(slots, youCategorySlot{
			Category: category,
			Cards:    cards,
		})
	}
	return slots
}

func (g *gameSession) setHostID(hostID string) {
	g.HostID = hostID
}

func (g *gameSession) playerStatus(playerID string) (string, bool) {
	player := g.Players[playerID]
	if player == nil {
		return "", false
	}
	return player.Status, true
}
