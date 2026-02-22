package main

import "testing"

func findSpecialByEffect(effectType string) specialDefinition {
	for _, def := range implementedSpecialDefinitions {
		if def.Effect.Type == effectType {
			return def
		}
	}
	return specialDefinition{}
}

func TestGameSession_ApplySpecial_ForceRevealCategory(t *testing.T) {
	g := newTestGameSession()
	g.Phase = scenarioPhaseReveal
	g.CurrentTurnID = "p1"

	def := findSpecialByEffect("forceRevealCategoryForAll")
	if def.ID == "" {
		t.Fatalf("definition not found")
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-force",
			Definition: def,
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-force",
		"payload": map[string]any{
			"category": "profession",
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}
	if g.RoundRules.ForcedCategory != categoryKeyToLabel["profession"] {
		t.Fatalf("unexpected forced category: %q", g.RoundRules.ForcedCategory)
	}
	if !g.Players["p1"].Specials[0].Used {
		t.Fatalf("special must be marked as used")
	}
	if !g.Players["p1"].Specials[0].RevealedPublic {
		t.Fatalf("special must be publicly revealed after apply")
	}
}

func TestGameSession_ApplySpecial_DisableVoteInWindow(t *testing.T) {
	g := newTestGameSession()
	g.Phase = scenarioPhaseVoting
	g.startVoting()
	g.VotePhase = votePhaseSpecialWindow

	def := findSpecialByEffect("disableVote")
	if def.ID == "" {
		t.Fatalf("definition not found")
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-disable",
			Definition: def,
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-disable",
		"payload": map[string]any{
			"targetPlayerId": "p2",
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}
	if _, ok := g.VoteDisabled["p2"]; !ok {
		t.Fatalf("expected p2 vote to be disabled")
	}
	if record, ok := g.Votes["p2"]; !ok || record.IsValid {
		t.Fatalf("expected invalid auto vote for disabled player")
	}
}

func TestGameSession_Timer_RevealTimeoutRandomCard(t *testing.T) {
	g := newTestGameSession()
	g.Phase = scenarioPhaseReveal
	g.CurrentTurnID = "p1"
	g.Settings.EnableRevealDiscussionTimer = true
	g.Settings.RevealTimeoutAction = "random_card"
	g.ActiveTimer = &gameTimerState{
		Kind:   scenarioPhaseRevealDiscussion,
		EndsAt: 1,
	}

	result := g.handleTimerExpired(10)
	if result.Error != "" {
		t.Fatalf("timer handling failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}
	revealedCount := 0
	for _, card := range g.Players["p1"].Hand {
		if card.Revealed {
			revealedCount++
		}
	}
	if revealedCount == 0 {
		t.Fatalf("expected at least one revealed card after timeout random reveal")
	}
	if g.Phase != scenarioPhaseRevealDiscussion {
		t.Fatalf("expected reveal discussion phase, got %s", g.Phase)
	}
}

func TestGameSession_Timer_RevealTimeoutSkipPlayer(t *testing.T) {
	g := newTestGameSession()
	g.Phase = scenarioPhaseReveal
	g.CurrentTurnID = "p1"
	g.Settings.EnableRevealDiscussionTimer = true
	g.Settings.RevealTimeoutAction = "skip_player"
	g.ActiveTimer = &gameTimerState{
		Kind:   scenarioPhaseRevealDiscussion,
		EndsAt: 1,
	}

	result := g.handleTimerExpired(10)
	if result.Error != "" {
		t.Fatalf("timer handling failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}
	if !g.RevealedThisRnd["p1"] {
		t.Fatalf("expected p1 to be marked processed by timeout")
	}
	if g.CurrentTurnID == "p1" {
		t.Fatalf("expected turn to move to next player")
	}
}

