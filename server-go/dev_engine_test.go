package main

import "testing"

func makeDevSessionForTest(playerCount int) *devTestSession {
	if playerCount < 2 {
		playerCount = 2
	}
	settings := defaultSettingsForScenario(scenarioDevTest)
	ruleset := buildAutoRuleset(maxInt(playerCount, minClassicPlayers))
	players := make([]*player, 0, playerCount)
	for i := 0; i < playerCount; i++ {
		id := "p" + string(rune('1'+i))
		name := "P" + string(rune('1'+i))
		players = append(players, &player{ID: id, Name: name})
	}
	return newDevTestSession(
		"ROOM",
		players[0].ID,
		settings,
		ruleset,
		players,
		makeHandAssetCatalog(playerCount),
		cloneSpecialDefinitions(implementedSpecialDefinitions),
		1234,
	)
}

func TestDevSession_SetBunkerOutcomeSupported(t *testing.T) {
	session := makeDevSessionForTest(4)
	session.core.finishGame("test")
	result := session.handleAction("p1", "setBunkerOutcome", map[string]any{"outcome": "survived"})
	if result.Error != "" {
		t.Fatalf("setBunkerOutcome failed: %s", result.Error)
	}
	if session.core.PostGame == nil || session.core.PostGame.Outcome != "survived" {
		t.Fatalf("expected post game outcome to be survived")
	}
}

func TestDevSession_DevSkipRoundHostOnly(t *testing.T) {
	session := makeDevSessionForTest(4)

	notHost := session.handleAction("p2", "devSkipRound", map[string]any{})
	if notHost.Error == "" {
		t.Fatalf("expected host-only error for devSkipRound")
	}

	session.core.Phase = scenarioPhaseVoting
	hostInVoting := session.handleAction("p1", "devSkipRound", map[string]any{})
	if hostInVoting.Error == "" {
		t.Fatalf("expected voting-phase error for devSkipRound")
	}
}

func TestDevSession_AutoRevealBotOnStateChange(t *testing.T) {
	session := makeDevSessionForTest(4)

	add := session.handleAction("p1", "devAddPlayer", map[string]any{"name": "Bot"})
	if add.Error != "" {
		t.Fatalf("devAddPlayer failed: %s", add.Error)
	}
	var botID string
	for _, id := range session.core.Order {
		if player := session.core.Players[id]; player != nil && player.IsBot {
			botID = id
		}
	}
	if botID == "" {
		t.Fatalf("bot player not created")
	}

	session.core.Phase = scenarioPhaseReveal
	session.core.CurrentTurnID = botID
	for _, id := range session.core.aliveIDs() {
		session.core.RevealedThisRnd[id] = false
	}
	// Trigger post-processing pipeline to run auto reveal.
	result := session.handleAction("p1", "devAddPlayer", map[string]any{"name": "Bot2"})
	if result.Error != "" {
		t.Fatalf("state-change trigger failed: %s", result.Error)
	}
	if !session.core.RevealedThisRnd[botID] {
		t.Fatalf("expected bot to auto-reveal on its turn")
	}
}

func TestDevSession_AutoVoteBots(t *testing.T) {
	session := makeDevSessionForTest(4)
	add := session.handleAction("p1", "devAddPlayer", map[string]any{"name": "Bot"})
	if add.Error != "" {
		t.Fatalf("devAddPlayer failed: %s", add.Error)
	}

	session.core.startVoting()
	if session.core.Phase != scenarioPhaseVoting || session.core.VotePhase != votePhaseVoting {
		t.Fatalf("voting did not start")
	}

	hostVote := session.handleAction("p1", "vote", map[string]any{"targetPlayerId": "p2"})
	if hostVote.Error != "" {
		t.Fatalf("host vote failed: %s", hostVote.Error)
	}

	botVotes := 0
	for voterID := range session.core.Votes {
		player := session.core.Players[voterID]
		if player != nil && player.IsBot {
			botVotes++
		}
	}
	if botVotes == 0 {
		t.Fatalf("expected at least one bot vote in auto-vote flow")
	}
	if session.core.VotePhase != votePhaseVoting {
		t.Fatalf("expected voting phase to remain active until all human votes, got=%s", session.core.VotePhase)
	}
}

func TestDevSession_EndConditionUsesDevCapacityAndElimination(t *testing.T) {
	session := makeDevSessionForTest(5)
	session.core.LastEliminatedID = "p5"
	for _, id := range session.core.Order {
		player := session.core.Players[id]
		if player == nil {
			continue
		}
		player.Status = playerAlive
	}
	session.core.Players["p5"].Status = playerEliminated

	result := session.postProcess(gameActionResult{StateChanged: true})
	if result.Error != "" {
		t.Fatalf("postProcess failed: %s", result.Error)
	}
	if session.core.Phase != scenarioPhaseEnded {
		t.Fatalf("expected ended phase by dev capacity rule, got=%s", session.core.Phase)
	}
}

func TestDevSession_FinalizeVotingOnlyInSpecialWindow(t *testing.T) {
	session := makeDevSessionForTest(4)
	session.core.startVoting()
	if session.core.VotePhase != votePhaseVoting {
		t.Fatalf("expected voting phase")
	}

	result := session.handleAction("p1", "finalizeVoting", map[string]any{})
	if result.Error == "" {
		t.Fatalf("expected finalizeVoting error outside voteSpecialWindow")
	}
}

func TestDevSession_BlocksRevealWhenEnded(t *testing.T) {
	session := makeDevSessionForTest(4)
	session.core.Phase = scenarioPhaseEnded
	cardID := session.core.Players["p1"].Hand[0].InstanceID

	result := session.handleAction("p1", "revealCard", map[string]any{"cardId": cardID})
	if result.Error == "" {
		t.Fatalf("expected revealCard error in ended phase for dev session")
	}
}

func TestDevSession_DevKickPlayerIsUnsupported(t *testing.T) {
	session := makeDevSessionForTest(4)
	result := session.handleAction("p1", "devKickPlayer", map[string]any{"targetPlayerId": "p2"})
	if result.Error == "" {
		t.Fatalf("expected unsupported action error for devKickPlayer")
	}
}

func TestDevSession_MarkLeftBunkerDoesNotAutoEndWithoutElimination(t *testing.T) {
	session := makeDevSessionForTest(4)
	result := session.handleAction("p1", "markLeftBunker", map[string]any{"targetPlayerId": "p2"})
	if result.Error != "" {
		t.Fatalf("markLeftBunker failed: %s", result.Error)
	}
	if session.core.Phase == scenarioPhaseEnded {
		t.Fatalf("dev session must not end only from markLeftBunker without elimination")
	}
}
