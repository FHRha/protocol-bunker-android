package main

import (
	"strings"
	"time"
)

const (
	devTestBunkerCapacity = 5
)

type devTestSession struct {
	core *gameSession
}

func newDevTestSession(
	roomCode, hostID string,
	settings gameSettings,
	ruleset gameRuleset,
	roomPlayers []*player,
	assets assetCatalog,
	specialDefinitions []specialDefinition,
	seed int64,
) *devTestSession {
	core := newGameSession(
		roomCode,
		hostID,
		scenarioDevTest,
		settings,
		ruleset,
		roomPlayers,
		assets,
		specialDefinitions,
		seed,
	)
	return &devTestSession{core: core}
}

func (d *devTestSession) setHostID(hostID string) {
	if d == nil || d.core == nil {
		return
	}
	d.core.HostID = hostID
}

func (d *devTestSession) playerStatus(playerID string) (string, bool) {
	if d == nil || d.core == nil {
		return "", false
	}
	return d.core.playerStatus(playerID)
}

func (d *devTestSession) currentTimer() *gameTimerState {
	if d == nil || d.core == nil {
		return nil
	}
	return d.core.currentTimer()
}

func (d *devTestSession) buildGameView(room *room, playerID string) gameView {
	if d == nil || d.core == nil {
		return gameView{}
	}
	return d.core.buildGameView(room, playerID)
}

func (d *devTestSession) handleTimerExpired(now int64) gameActionResult {
	if d == nil || d.core == nil {
		return gameActionResult{Error: "dev session is not initialized"}
	}
	result := d.core.handleTimerExpired(now)
	return d.postProcess(result)
}

func (d *devTestSession) handleAction(actorID, actionType string, payload map[string]any) gameActionResult {
	if d == nil || d.core == nil {
		return gameActionResult{Error: "dev session is not initialized"}
	}
	if payload == nil {
		payload = map[string]any{}
	}

	switch actionType {
	case "setBunkerOutcome":
		result := d.core.handleAction(actorID, actionType, payload)
		return d.postProcess(result)
	case "devKickPlayer":
		return gameActionResult{Error: "Неизвестное действие."}
	case "devSkipRound":
		result := d.devSkipRound(actorID)
		return d.postProcess(result)
	case "markLeftBunker":
		targetID, _ := payload["targetPlayerId"].(string)
		result := d.markLeftBunker(targetID)
		return d.postProcess(result)
	case "devAddPlayer", "devRemovePlayer", "revealWorldThreat":
		result := d.core.handleAction(actorID, actionType, payload)
		return d.postProcess(result)
	default:
		player := d.core.Players[actorID]
		if player == nil {
			return gameActionResult{Error: "Игрок не найден."}
		}
		if player.Status != playerAlive {
			return gameActionResult{Error: "Вы исключены из игры."}
		}
		if d.core.Phase == scenarioPhaseEnded {
			return gameActionResult{Error: "Игра уже завершена."}
		}
		if actionType == "finalizeVoting" && d.core.VotePhase != votePhaseSpecialWindow {
			return gameActionResult{Error: "Окно спецусловий ещё не открыто."}
		}
		result := d.core.handleAction(actorID, actionType, payload)
		return d.postProcess(result)
	}
}

func (d *devTestSession) markLeftBunker(targetID string) gameActionResult {
	targetID = asString(targetID)
	targetID = strings.TrimSpace(targetID)
	if targetID == "" {
		return gameActionResult{Error: "Игрок не найден."}
	}
	target := d.core.Players[targetID]
	if target == nil || target.Status == playerLeftBunker {
		return gameActionResult{Error: "Игрок не найден."}
	}

	target.Status = playerLeftBunker
	delete(d.core.RevealedThisRnd, targetID)
	d.removeFromVotingDev(targetID)
	if d.core.CurrentTurnID == targetID {
		next := d.core.nextUnrevealedAliveAfter(targetID)
		if next == "" {
			next = d.core.firstAliveID()
		}
		d.core.CurrentTurnID = next
	}
	return gameActionResult{StateChanged: true}
}

func (d *devTestSession) removeFromVotingDev(targetID string) {
	d.core.removeFromVoting(targetID)
	if d.core.Phase == scenarioPhaseVoting && d.core.VotePhase == votePhaseVoting {
		aliveCount := len(d.core.aliveIDs())
		if aliveCount > 0 && len(d.core.Votes) >= aliveCount {
			_ = d.core.enterVoteSpecialWindow()
		}
	}
}

func (d *devTestSession) devSkipRound(actorID string) gameActionResult {
	if actorID != d.core.HostID {
		return gameActionResult{Error: "Только хост может пропустить раунд."}
	}
	if d.core.Phase == scenarioPhaseVoting || d.core.Phase == scenarioPhaseResolution {
		return gameActionResult{Error: "Нельзя пропустить раунд во время голосования."}
	}
	if d.core.Phase == scenarioPhaseEnded {
		return gameActionResult{Error: "Игра уже завершена."}
	}

	for _, id := range d.core.aliveIDs() {
		d.core.RevealedThisRnd[id] = true
	}
	if d.core.Phase != scenarioPhaseRevealDiscussion {
		d.core.Phase = scenarioPhaseRevealDiscussion
	}
	return d.core.advanceAfterDiscussion()
}

func (d *devTestSession) postProcess(result gameActionResult) gameActionResult {
	if result.Error != "" || d.core == nil {
		return result
	}

	combined := result
	changed := result.StateChanged
	for {
		triggered := false
		if d.core.Phase == scenarioPhaseReveal {
			auto := d.autoRevealBots()
			combined = mergeActionResult(combined, auto)
			if auto.Error != "" {
				return combined
			}
			if auto.StateChanged {
				triggered = true
			}
		}
		if d.core.Phase == scenarioPhaseVoting && d.core.VotePhase == votePhaseVoting {
			auto := d.autoVoteBots()
			combined = mergeActionResult(combined, auto)
			if auto.Error != "" {
				return combined
			}
			if auto.StateChanged {
				triggered = true
			}
		}
		if !triggered {
			break
		}
		changed = true
	}

	if d.shouldEnd() {
		end := d.core.finishGame("dev_test_bunker_capacity")
		combined = mergeActionResult(combined, end)
		if end.StateChanged {
			changed = true
		}
	}

	combined.StateChanged = changed || combined.StateChanged
	return combined
}

func (d *devTestSession) shouldEnd() bool {
	if d.core == nil || d.core.Phase == scenarioPhaseEnded {
		return false
	}
	if d.core.LastEliminatedID == "" {
		return false
	}
	return len(d.core.aliveIDs()) <= devTestBunkerCapacity
}

func (d *devTestSession) autoRevealBots() gameActionResult {
	currentID := d.core.CurrentTurnID
	if currentID == "" {
		return gameActionResult{}
	}
	player := d.core.Players[currentID]
	if player == nil || player.Status != playerAlive || !player.IsBot {
		return gameActionResult{}
	}

	var cardID string
	if forced := d.core.RoundRules.ForcedCategory; forced != "" {
		if cfg, ok := resolveCategoryDeckSlot(forced); ok {
			for _, card := range player.Hand {
				if card.Revealed || card.Deck != cfg.deck {
					continue
				}
				if cfg.slot != "" && card.SlotKey != cfg.slot {
					continue
				}
				cardID = card.InstanceID
				break
			}
		}
	}
	if cardID == "" {
		for _, card := range player.Hand {
			if !card.Revealed {
				cardID = card.InstanceID
				break
			}
		}
	}
	if cardID == "" {
		return gameActionResult{}
	}
	return d.core.revealCard(currentID, cardID)
}

func (d *devTestSession) autoVoteBots() gameActionResult {
	if d.core.VotePhase != votePhaseVoting {
		return gameActionResult{}
	}
	aliveCount := len(d.core.aliveIDs())
	if aliveCount == 0 {
		return gameActionResult{}
	}

	changed := false
	for _, voterID := range d.core.Order {
		voter := d.core.Players[voterID]
		if voter == nil || voter.Status != playerAlive || !voter.IsBot {
			continue
		}
		if _, exists := d.core.Votes[voterID]; exists {
			continue
		}

		if reason, blocked := d.core.VoteDisabled[voterID]; blocked {
			d.core.markVoteWasted(voterID, reason)
			changed = true
			continue
		}

		viable := make([]string, 0, len(d.core.VoteCandidates))
		for candidateID := range d.core.VoteCandidates {
			if candidateID == voterID {
				continue
			}
			target := d.core.Players[candidateID]
			if target == nil || target.Status != playerAlive {
				continue
			}
			if target.BannedAgainst[voterID] {
				continue
			}
			if d.core.RevoteDisallow[candidateID] {
				continue
			}
			viable = append(viable, candidateID)
		}
		if len(viable) == 0 {
			d.core.markVoteWasted(voterID, "Нет доступных кандидатов.")
			changed = true
			continue
		}
		targetID := viable[d.core.rng.Intn(len(viable))]
		d.core.Votes[voterID] = voteRecord{
			TargetID:  targetID,
			Submitted: time.Now().UnixMilli(),
			IsValid:   true,
		}
		changed = true
	}

	if !changed {
		return gameActionResult{}
	}
	result := gameActionResult{StateChanged: true}
	if len(d.core.Votes) >= aliveCount && d.core.VotePhase == votePhaseVoting {
		window := d.core.enterVoteSpecialWindow()
		result = mergeActionResult(result, window)
	}
	return result
}

func mergeActionResult(base, next gameActionResult) gameActionResult {
	if base.Error == "" && next.Error != "" {
		base.Error = next.Error
	}
	base.StateChanged = base.StateChanged || next.StateChanged
	if len(next.Events) > 0 {
		base.Events = append(base.Events, next.Events...)
	}
	return base
}
