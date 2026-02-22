package main

import (
	"testing"
	"time"
)

func newTestGameSession() *gameSession {
	settings := defaultSettings()
	ruleset := buildAutoRuleset(minClassicPlayers)
	players := []*player{
		{ID: "p1", Name: "Host"},
		{ID: "p2", Name: "P2"},
		{ID: "p3", Name: "P3"},
		{ID: "p4", Name: "P4"},
	}
	return newGameSession(
		"TEST",
		"p1",
		scenarioClassic,
		settings,
		ruleset,
		players,
		assetCatalog{Decks: map[string][]assetCard{}},
		nil,
		123,
	)
}

func TestGameSession_PostGameRevealAllowedBeforeOutcome(t *testing.T) {
	g := newTestGameSession()
	g.Phase = scenarioPhaseEnded
	g.PostGame = &postGameStateView{
		IsActive:  true,
		EnteredAt: time.Now().UnixMilli(),
	}

	cardID := g.Players["p1"].Hand[0].InstanceID
	result := g.handleAction("p1", "revealCard", map[string]any{"cardId": cardID})
	if result.Error != "" {
		t.Fatalf("expected reveal to be allowed in post-game, got error: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected stateChanged=true")
	}
	if !g.Players["p1"].Hand[0].Revealed {
		t.Fatalf("card should be revealed")
	}
}

func TestGameSession_PostGameRevealBlockedAfterOutcome(t *testing.T) {
	g := newTestGameSession()
	g.Phase = scenarioPhaseEnded
	g.PostGame = &postGameStateView{
		IsActive:  true,
		EnteredAt: time.Now().UnixMilli(),
	}
	setResult := g.handleAction("p1", "setBunkerOutcome", map[string]any{"outcome": "survived"})
	if setResult.Error != "" {
		t.Fatalf("setBunkerOutcome failed: %s", setResult.Error)
	}

	cardID := g.Players["p1"].Hand[0].InstanceID
	result := g.handleAction("p1", "revealCard", map[string]any{"cardId": cardID})
	if result.Error == "" {
		t.Fatalf("expected reveal error after outcome")
	}
	if g.Players["p1"].Hand[0].Revealed {
		t.Fatalf("card must remain hidden after outcome")
	}
}

func TestGameSession_GameViewContainsWorldAndPostGame(t *testing.T) {
	g := newTestGameSession()
	g.Phase = scenarioPhaseEnded
	g.PostGame = &postGameStateView{
		IsActive:  true,
		EnteredAt: time.Now().UnixMilli(),
	}
	room := &room{
		Players: map[string]*player{
			"p1": {ID: "p1", Connected: true},
			"p2": {ID: "p2", Connected: true},
			"p3": {ID: "p3", Connected: true},
			"p4": {ID: "p4", Connected: true},
		},
	}

	view := g.buildGameView(room, "p1")
	if view.World == nil {
		t.Fatalf("world must be present in gameView")
	}
	if view.PostGame == nil || !view.PostGame.IsActive {
		t.Fatalf("postGame must be present and active in gameView")
	}
	if view.Public.ThreatModifier == nil {
		t.Fatalf("threatModifier must be present in gameView.public")
	}
}
