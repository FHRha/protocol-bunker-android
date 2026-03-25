package main

import (
	"slices"
	"testing"
)

func TestGameSession_ApplySpecial_RedealAllRevealed_PreservesRevealedPool(t *testing.T) {
	g := newTestGameSession()
	deckName := categoryKeyToDeck["profession"]
	if deckName == "" {
		t.Fatalf("profession deck name missing")
	}

	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-redeal-profession",
			Definition: specialDefinition{
				ID:          "redeal_profession",
				Title:       "Redeal Profession",
				Trigger:     "active",
				Implemented: true,
				Effect: specialEffect{
					Type: "redealAllRevealed",
					Params: map[string]any{
						"category": "profession",
					},
				},
			},
		},
	}

	before := make([]string, 0, len(g.Order))
	for _, playerID := range g.Order {
		player := g.Players[playerID]
		cards := g.getCardsByCategoryKey(player, "profession", false)
		if len(cards) == 0 {
			t.Fatalf("no profession card for %s", playerID)
		}
		cards[0].Revealed = true
		before = append(before, cards[0].CardID)
	}
	slices.Sort(before)

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-redeal-profession",
		"payload":           map[string]any{},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}
	if !g.Players["p1"].Specials[0].Used {
		t.Fatalf("special must be marked used")
	}

	after := make([]string, 0, len(g.Order))
	for _, playerID := range g.Order {
		player := g.Players[playerID]
		cards := g.getCardsByCategoryKey(player, "profession", false)
		if len(cards) == 0 {
			t.Fatalf("no profession card for %s after redeal", playerID)
		}
		if !cards[0].Revealed {
			t.Fatalf("revealed card should remain revealed for %s", playerID)
		}
		after = append(after, cards[0].CardID)
	}
	slices.Sort(after)

	if !slices.Equal(before, after) {
		t.Fatalf("redeal should preserve revealed card multiset: before=%v after=%v", before, after)
	}
}

func TestGameSession_ApplySpecial_AddFinalThreat_UsesParamOrSpecialID(t *testing.T) {
	g := newTestGameSession()
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-final-threat-explicit",
			Definition: specialDefinition{
				ID:          "special_explicit",
				Title:       "Explicit Threat",
				Trigger:     "active",
				Implemented: true,
				Effect: specialEffect{
					Type: "addFinalThreat",
					Params: map[string]any{
						"threatKey": "panic_mode",
					},
				},
			},
		},
		{
			InstanceID: "sp-final-threat-fallback",
			Definition: specialDefinition{
				ID:          "fallback_special_id",
				Title:       "Fallback Threat",
				Trigger:     "active",
				Implemented: true,
				Effect: specialEffect{
					Type:   "addFinalThreat",
					Params: map[string]any{},
				},
			},
		},
	}

	resultExplicit := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-final-threat-explicit",
		"payload":           map[string]any{},
	})
	if resultExplicit.Error != "" {
		t.Fatalf("explicit addFinalThreat failed: %s", resultExplicit.Error)
	}

	resultFallback := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-final-threat-fallback",
		"payload":           map[string]any{},
	})
	if resultFallback.Error != "" {
		t.Fatalf("fallback addFinalThreat failed: %s", resultFallback.Error)
	}

	if len(g.FinalThreats) != 2 {
		t.Fatalf("expected 2 final threats, got %d", len(g.FinalThreats))
	}
	if g.FinalThreats[0] != "panic_mode" {
		t.Fatalf("expected explicit threat key, got %q", g.FinalThreats[0])
	}
	if g.FinalThreats[1] != "fallback_special_id" {
		t.Fatalf("expected fallback to special id, got %q", g.FinalThreats[1])
	}
	if !g.Players["p1"].Specials[0].Used || !g.Players["p1"].Specials[1].Used {
		t.Fatalf("both specials must be marked used")
	}
}
