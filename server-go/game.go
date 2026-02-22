package main

import (
	"fmt"
	"math/rand"
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

	for categoryKey, label := range categoryKeyToLabel {
		if normalizeSpecialKey(label) != normalized && normalizeSpecialKey(categoryKey) != normalized {
			continue
		}
		deck := categoryKeyToDeck[categoryKey]
		if deck == "" {
			continue
		}
		return deckSlotConfig{deck: deck, slot: categoryKeyToSlot[categoryKey]}, true
	}

	return deckSlotConfig{}, false
}

func healthDeckName() string {
	if len(coreDecks) > 1 {
		return coreDecks[1]
	}
	return categoryKeyToDeck["health"]
}

type handCard struct {
	InstanceID string
	CardID     string
	Deck       string
	SlotKey    string
	Label      string
	Revealed   bool
	Missing    bool
}

type voteRecord struct {
	TargetID   string
	Submitted  int64
	IsValid    bool
	ReasonText string
}

type gamePlayer struct {
	PlayerID string
	Name     string
	Status   string
	Hand     []handCard
	Specials []specialConditionState
	IsBot    bool

	BannedAgainst        map[string]bool
	ForcedWastedVoteNext bool
}

type gameActionResult struct {
	StateChanged bool
	Events       []gameEvent
	Error        string
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
	AutoWastedVoters      map[string]bool
	RevoteDisallow        map[string]bool
	DoubleAgainst         string
	TieBreakUsed          bool
	VotesRemaining        int
	LastEliminatedID      string
	ResolutionNote        string
	Winners               []string
	LastStageText         string
	TotalExiles           int
	EventCounter          int
	World                 worldStateView
	WorldEvent            *worldEventView
	PostGame              *postGameStateView
	FinalThreats          []string
	FirstHealthRevealerID string
	RoundRules            roundRulesPublic
	ActiveTimer           *gameTimerState
	DeckPools             map[string][]assetCard
	CardCounter           int
	DevBotCounter         int

	rng            *rand.Rand
	specialPool    []specialDefinition
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
		RoomCode:         roomCode,
		HostID:           hostID,
		Scenario:         scenarioID,
		IsDev:            scenarioID == scenarioDevTest,
		Settings:         settings,
		Ruleset:          ruleset,
		Players:          make(map[string]*gamePlayer, len(roomPlayers)),
		Order:            make([]string, 0, len(roomPlayers)),
		Phase:            scenarioPhaseReveal,
		Round:            1,
		RevealedThisRnd:  map[string]bool{},
		Votes:            map[string]voteRecord{},
		BaseVotes:        map[string]voteRecord{},
		VoteResults:      map[string]voteRecord{},
		VoteCandidates:   map[string]bool{},
		VoteDisabled:     map[string]string{},
		VoteWeights:      map[string]int{},
		AutoWastedVoters: map[string]bool{},
		RevoteDisallow:   map[string]bool{},
		FinalThreats:     []string{},
		rng:              rnd,
		DeckPools:        pools,
	}
	session.World = rollWorldFromPools(pools, rnd, len(roomPlayers))
	sourceSpecials := specialDefinitions
	if len(sourceSpecials) == 0 {
		sourceSpecials = implementedSpecialDefinitions
	}
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
				Label:      "Нет карты",
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
			PlayerID:      rp.ID,
			Name:          rp.Name,
			Status:        playerAlive,
			Hand:          hand,
			Specials:      nil,
			IsBot:         false,
			BannedAgainst: map[string]bool{},
		}
		session.Order = append(session.Order, rp.ID)
	}
	session.assignInitialSpecials()

	session.CurrentTurnID = session.firstAliveID()
	session.VotesRemaining = session.votesForRound(session.Round)
	session.WorldEvent = nil
	session.revealNextBunkerCard(session.Round)
	session.LastStageText = "Раунд 1. Раскрытие карт."
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
	g.AutoWastedVoters = map[string]bool{}
	g.RevoteDisallow = map[string]bool{}
	g.DoubleAgainst = ""
	g.TieBreakUsed = false
	for _, id := range g.aliveIDs() {
		g.VoteCandidates[id] = true
	}
	for _, p := range g.Players {
		if p == nil || p.Status != playerAlive {
			continue
		}
		if p.ForcedWastedVoteNext {
			g.markVoteWasted(p.PlayerID, "Голос потрачен.")
			p.ForcedWastedVoteNext = false
		}
	}
	g.ResolutionNote = ""
	g.RoundRules.NoTalkUntilVoting = false
	g.LastStageText = fmt.Sprintf("Раунд %d. Голосование.", g.Round)
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
		g.LastStageText = "Игра завершена."
	} else {
		g.LastStageText = fmt.Sprintf("Игра завершена. В бункер попали: %s.", strings.Join(winners, ","))
	}
	event := g.makeEvent("gameEnd", g.LastStageText)
	if reason != "" {
		event.Message = fmt.Sprintf("%s (%s)", event.Message, reason)
	}
	return gameActionResult{StateChanged: true, Events: []gameEvent{event}}
}

func (g *gameSession) startNextRoundOrEnd() gameActionResult {
	if g.shouldEnd() {
		return g.finishGame("условия завершения достигнуты")
	}
	nextRound := g.Round + 1
	if nextRound > len(g.Ruleset.VotesPerRnd) {
		return g.finishGame("закончились раунды")
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
	g.AutoWastedVoters = map[string]bool{}
	g.RevoteDisallow = map[string]bool{}
	g.DoubleAgainst = ""
	g.TieBreakUsed = false
	g.RevealedThisRnd = map[string]bool{}
	g.CurrentTurnID = g.firstAliveID()
	g.VotesRemaining = g.votesForRound(nextRound)
	g.LastRevealerID = ""
	g.ResolutionNote = ""
	g.RoundRules = roundRulesPublic{}
	g.LastStageText = fmt.Sprintf("Раунд %d. Раскрытие карт.", nextRound)
	g.scheduleRevealTimeoutIfNeeded()

	event := g.makeEvent("roundStart", g.LastStageText)
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
		return gameActionResult{Error: "Игрок не найден."}
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
		return g.applySpecial(actorID, specialInstanceID, rawPayload)
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
		return gameActionResult{Error: "Неизвестное действие."}
	}
}

func (g *gameSession) revealCard(actorID, cardID string) gameActionResult {
	player := g.Players[actorID]
	if player == nil {
		return gameActionResult{Error: "Игрок не найден."}
	}
	if g.Phase == scenarioPhaseEnded {
		if g.PostGame == nil || !g.PostGame.IsActive || g.PostGame.Outcome != "" {
			return gameActionResult{Error: "Игра уже завершена."}
		}
		if player.Status == playerLeftBunker {
			return gameActionResult{Error: "Вы покинули игру."}
		}
		for i := range player.Hand {
			card := &player.Hand[i]
			if card.InstanceID != cardID {
				continue
			}
			if card.Revealed {
				return gameActionResult{Error: "Карта уже раскрыта."}
			}
			card.Revealed = true
			return gameActionResult{StateChanged: true}
		}
		return gameActionResult{Error: "Карта не найдена."}
	}
	if g.Phase != scenarioPhaseReveal {
		return gameActionResult{Error: "Сейчас нельзя раскрывать карту."}
	}
	if actorID != g.CurrentTurnID {
		return gameActionResult{Error: "Сейчас ход другого игрока."}
	}
	if player.Status != playerAlive {
		return gameActionResult{Error: "Игрок не может выполнить действие."}
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
		return gameActionResult{Error: "Карта не найдена."}
	}
	if targetCard.Revealed {
		return gameActionResult{Error: "Карта уже раскрыта."}
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
			return gameActionResult{Error: fmt.Sprintf("В этом раунде нужно раскрыть карту категории \"%s\".", forced)}
		}
	}

	targetCard.Revealed = true
	if targetCard.Deck == healthDeckName() && g.FirstHealthRevealerID == "" {
		g.FirstHealthRevealerID = actorID
	}
	g.RevealedThisRnd[actorID] = true
	g.LastRevealerID = actorID
	g.LastStageText = fmt.Sprintf("%s раскрыл карту.", player.Name)
	enterResult := g.enterRevealDiscussion()
	events := []gameEvent{g.makeEvent("info", g.LastStageText)}
	events = append(events, enterResult.Events...)
	return gameActionResult{StateChanged: true, Events: events}
}

func (g *gameSession) continueRound(actorID string) gameActionResult {
	if g.Phase == scenarioPhaseEnded {
		return gameActionResult{Error: "Игра уже завершена."}
	}

	if !g.canContinue(actorID) {
		return gameActionResult{Error: "Недостаточно прав для продолжения."}
	}

	if g.Phase == scenarioPhaseRevealDiscussion {
		return g.advanceAfterDiscussion()
	}

	switch g.Phase {
	case scenarioPhaseResolution:
		if g.shouldEnd() {
			return g.finishGame("условия завершения достигнуты")
		}
		if g.VotesRemaining > 0 {
			g.startVoting()
			return gameActionResult{
				StateChanged: true,
				Events:       []gameEvent{g.makeEvent("votingStart", g.LastStageText)},
			}
		}
		return g.startNextRoundOrEnd()

	case scenarioPhaseVoting:
		if g.VotePhase == votePhaseSpecialWindow {
			return g.finalizeVoting(actorID)
		}
		return gameActionResult{Error: "Сначала завершите голосование."}

	default:
		return gameActionResult{Error: "Команда недоступна в текущей фазе."}
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
		return gameActionResult{Error: "Сейчас нет активного голосования."}
	}
	if actorID == targetID {
		return gameActionResult{Error: "Нельзя голосовать за себя."}
	}
	if !g.VoteCandidates[targetID] {
		return gameActionResult{Error: "Недопустимый кандидат."}
	}
	if g.RevoteDisallow[targetID] {
		return gameActionResult{Error: "Нельзя голосовать за этого кандидата."}
	}
	if reason, blocked := g.VoteDisabled[actorID]; blocked {
		if reason == "" {
			reason = "Ваш голос заблокирован."
		}
		return gameActionResult{Error: reason}
	}

	actor := g.Players[actorID]
	target := g.Players[targetID]
	if actor == nil || target == nil {
		return gameActionResult{Error: "Игрок не найден."}
	}
	if actor.Status != playerAlive {
		return gameActionResult{Error: "Вы исключены из игры."}
	}
	if target.Status != playerAlive {
		return gameActionResult{Error: "Кандидат не в игре."}
	}
	if target.BannedAgainst[actorID] {
		return gameActionResult{Error: "Вы не можете голосовать против этого игрока."}
	}
	if _, exists := g.Votes[actorID]; exists {
		return gameActionResult{Error: "Вы уже проголосовали."}
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
		return gameActionResult{Error: "voting is not active"}
	}
	if g.VotePhase != votePhaseSpecialWindow && g.VotePhase != votePhaseVoting {
		return gameActionResult{Error: "voting already finalized"}
	}
	if actorID != g.HostID {
		return gameActionResult{Error: "only CONTROL can finalize voting"}
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
			Events:       []gameEvent{g.makeEvent("info", g.LastStageText)},
		}
	}

	eliminatedID := ""
	if len(topCandidates) > 0 {
		eliminatedID = topCandidates[g.rng.Intn(len(topCandidates))]
	}

	events := make([]gameEvent, 0, 2)
	if eliminatedID != "" {
		g.applyElimination(eliminatedID)
		if g.LastEliminatedID != "" {
			name := g.playerName(g.LastEliminatedID)
			g.ResolutionNote = fmt.Sprintf("Voting result: %s eliminated.", name)
			events = append(events, g.makeEvent("elimination", g.ResolutionNote))
		}
	} else {
		g.ResolutionNote = "Tie: no one eliminated."
		events = append(events, g.makeEvent("info", g.ResolutionNote))
	}

	g.Phase = scenarioPhaseResolution
	g.VotePhase = votePhaseResolve
	g.LastStageText = g.ResolutionNote
	g.Votes = map[string]voteRecord{}

	if g.shouldEnd() {
		end := g.finishGame("after_voting")
		events = append(events, end.Events...)
	} else {
		g.schedulePhaseTimer(timerKindResolutionAuto, 2)
	}
	return gameActionResult{StateChanged: true, Events: events}
}

func (g *gameSession) kickPlayer(targetID string) gameActionResult {
	player := g.Players[targetID]
	if player == nil {
		return gameActionResult{Error: "Игрок не найден."}
	}
	if player.Status != playerAlive {
		return gameActionResult{Error: "Игрок уже неактивен."}
	}
	g.applyElimination(targetID)
	g.ResolutionNote = fmt.Sprintf("Игрок %s исключён вручную.", player.Name)
	g.removeFromVoting(targetID)
	if g.CurrentTurnID == targetID {
		g.CurrentTurnID = g.nextUnrevealedAliveAfter(targetID)
	}
	if g.shouldEnd() {
		end := g.finishGame("ручное исключение")
		return gameActionResult{
			StateChanged: true,
			Events: append(
				[]gameEvent{g.makeEvent("elimination", g.ResolutionNote)},
				end.Events...,
			),
		}
	}
	return gameActionResult{
		StateChanged: true,
		Events:       []gameEvent{g.makeEvent("elimination", g.ResolutionNote)},
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
				Label:      "Нет карты",
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
		return gameActionResult{Error: "DEV-режим доступен только в сценарии dev_test."}
	}
	g.DevBotCounter++
	playerID := fmt.Sprintf("dev-%s-%d", strings.ToLower(g.RoomCode), g.DevBotCounter)
	for g.Players[playerID] != nil {
		g.DevBotCounter++
		playerID = fmt.Sprintf("dev-%s-%d", strings.ToLower(g.RoomCode), g.DevBotCounter)
	}

	fallbackName := fmt.Sprintf("DEV Игрок %d", g.DevBotCounter)
	displayName := sanitizeHumanText(strings.TrimSpace(name), fallbackName)
	newPlayer := &gamePlayer{
		PlayerID:      playerID,
		Name:          displayName,
		Status:        playerAlive,
		Hand:          g.drawInitialHandForPlayer(playerID),
		Specials:      nil,
		IsBot:         true,
		BannedAgainst: map[string]bool{},
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
		Events:       []gameEvent{g.makeEvent("info", fmt.Sprintf("DEV: добавлен игрок %s.", displayName))},
	}
}

func (g *gameSession) devRemovePlayer(targetID string) gameActionResult {
	if !g.IsDev {
		return gameActionResult{Error: "DEV-действие доступно только в dev_test."}
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
		return gameActionResult{Error: "Нет игрока для удаления."}
	}
	target := g.Players[removeID]
	if target == nil {
		return gameActionResult{Error: "Игрок не найден."}
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

	events := []gameEvent{g.makeEvent("info", fmt.Sprintf("Удалён игрок %s.", target.Name))}
	if g.shouldEnd() {
		end := g.finishGame("DEV: удаление игрока")
		events = append(events, end.Events...)
	}
	return gameActionResult{StateChanged: true, Events: events}
}

func (g *gameSession) markLeftBunker(targetID string) gameActionResult {
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		return gameActionResult{Error: "Нужно указать targetPlayerId."}
	}
	player := g.Players[targetID]
	if player == nil {
		return gameActionResult{Error: "Игрок не найден."}
	}
	if player.Status == playerLeftBunker {
		return gameActionResult{Error: "Игрок уже покинул бункер."}
	}

	wasAlive := player.Status == playerAlive
	player.Status = playerLeftBunker
	g.removeFromVoting(targetID)
	if g.CurrentTurnID == targetID {
		next := g.nextUnrevealedAliveAfter(targetID)
		if next == "" {
			next = g.firstAliveID()
		}
		g.CurrentTurnID = next
	}
	events := []gameEvent{g.makeEvent("playerLeftBunker", fmt.Sprintf("Игрок %s покинул бункер.", player.Name))}

	if wasAlive {
		g.TotalExiles++
		if g.VotesRemaining > 0 {
			g.VotesRemaining--
		}
	}

	if g.Phase == scenarioPhaseVoting && wasAlive {
		g.clearActiveTimer()
		g.Phase = scenarioPhaseResolution
		g.VotePhase = votePhaseResolve
		g.ResolutionNote = fmt.Sprintf("Голосование пропущено: %s покинул игру.", player.Name)
		g.LastStageText = g.ResolutionNote
		events = append(events, g.makeEvent("info", g.ResolutionNote))
		if g.shouldEnd() {
			end := g.finishGame("после выхода игрока")
			events = append(events, end.Events...)
		} else {
			g.schedulePhaseTimer(timerKindResolutionAuto, 2)
		}
		return gameActionResult{StateChanged: true, Events: events}
	}

	if g.shouldEnd() {
		end := g.finishGame("после выхода игрока")
		events = append(events, end.Events...)
	}
	return gameActionResult{StateChanged: true, Events: events}
}
func (g *gameSession) makeEvent(kind, message string) gameEvent {
	g.EventCounter++
	safeMessage := sanitizeHumanText(message, defaultEventMessage(kind))
	return gameEvent{
		ID:        fmt.Sprintf("%s-%d-%d", g.RoomCode, time.Now().UnixMilli(), g.EventCounter),
		Kind:      kind,
		Message:   safeMessage,
		CreatedAt: time.Now().UnixMilli(),
	}
}

func (g *gameSession) buildGameView(room *room, playerID string) gameView {
	view := gameView{
		Phase:      g.Phase,
		Round:      g.Round,
		Categories: append([]string(nil), categoryOrder...),
		LastStage:  sanitizeHumanText(g.LastStageText, defaultStageMessage(g.Round)),
		Ruleset:    g.Ruleset,
		World:      g.buildWorldView(),
		WorldEvent: g.WorldEvent,
		PostGame:   g.PostGame,
	}

	you := g.Players[playerID]
	if you == nil {
		you = &gamePlayer{PlayerID: playerID, Name: "Unknown", Status: playerEliminated}
	}
	view.You.PlayerID = you.PlayerID
	view.You.Name = you.Name
	view.You.Hand = make([]cardRef, 0, len(you.Hand))
	for _, card := range you.Hand {
		view.You.Hand = append(view.You.Hand, cardRef{
			ID:       card.CardID,
			Deck:     card.Deck,
			Instance: card.InstanceID,
			Label:    card.Label,
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
	view.Public.CanOpenVotingModal = g.VotePhase != ""
	view.Public.VoteModalOpen = g.VotePhase == votePhaseVoting
	view.Public.LastEliminated = g.LastEliminatedID
	view.Public.Winners = append([]string(nil), g.Winners...)
	view.Public.ResolutionNote = sanitizeHumanText(g.ResolutionNote, "Voting results updated.")
	if g.RoundRules.NoTalkUntilVoting || g.RoundRules.ForcedCategory != "" {
		rulesCopy := g.RoundRules
		if rulesCopy.ForcedCategory != "" {
			if key := resolveCategoryKey(rulesCopy.ForcedCategory); key != "" {
				rulesCopy.ForcedCategory = categoryKeyToLabel[key]
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

func (g *gameSession) playerHasVoted(playerID string) bool {
	_, exists := g.currentVoteSource()[playerID]
	return exists
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
				Submitted:  info.Submitted,
			})
			continue
		}
		out = append(out, votePublic{
			VoterID:   playerID,
			VoterName: sanitizeHumanText(player.Name, "Player"),
			Status:    "invalid",
			Reason:    sanitizeHumanText(info.ReasonText, "Vote rejected."),
			Submitted: info.Submitted,
		})
	}
	return out
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
				Missing:  card.Missing,
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
			cards := make([]publicCategoryCard, 0, len(player.Specials))
			for _, special := range player.Specials {
				if !g.IsDev && !special.RevealedPublic {
					continue
				}
				cards = append(cards, publicCategoryCard{
					Label:    special.Definition.Title,
					ImgURL:   specialImageURL(special.Definition),
					Revealed: special.RevealedPublic,
				})
			}
			status := "hidden"
			if len(cards) > 0 {
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
		for _, card := range player.Hand {
			if card.Deck != cfg.deck {
				continue
			}
			if cfg.slot != "" && card.SlotKey != cfg.slot {
				continue
			}
			if !g.IsDev && !card.Revealed {
				continue
			}
			imgURL := ""
			if card.CardID != "" {
				imgURL = "/assets/" + card.CardID
			}
			cards = append(cards, publicCategoryCard{
				Label:    card.Label,
				ImgURL:   imgURL,
				Revealed: card.Revealed,
			})
		}
		status := "hidden"
		if len(cards) > 0 {
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
			cards = append(cards, youCategoryCard{
				InstanceID: card.InstanceID,
				Label:      card.Label,
				Revealed:   card.Revealed,
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
