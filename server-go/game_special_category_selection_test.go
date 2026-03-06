package main

import "testing"

func newCategorySelectionSession() *gameSession {
	settings := defaultSettings()
	ruleset := buildAutoRuleset(minClassicPlayers)
	players := []*player{
		{ID: "p1", Name: "Host"},
		{ID: "p2", Name: "P2"},
		{ID: "p3", Name: "P3"},
		{ID: "p4", Name: "P4"},
	}
	g := newGameSession(
		"TEST",
		"p1",
		scenarioClassic,
		settings,
		ruleset,
		players,
		makeHandAssetCatalog(len(players)),
		nil,
		777,
	)
	g.Phase = scenarioPhaseReveal
	return g
}

func findHandCardBySlot(t *testing.T, g *gameSession, playerID, slotKey string) *handCard {
	t.Helper()
	player := g.Players[playerID]
	if player == nil {
		t.Fatalf("player not found: %s", playerID)
	}
	for i := range player.Hand {
		if player.Hand[i].SlotKey == slotKey {
			return &player.Hand[i]
		}
	}
	t.Fatalf("slot %s not found for %s", slotKey, playerID)
	return nil
}

func TestGameSession_SpecialSwapUsesSelectedFactsCards(t *testing.T) {
	g := newCategorySelectionSession()
	def := specialDefinition{
		ID:          "test_swap_selected_fact",
		Title:       "Swap Facts",
		Implemented: true,
		ChoiceKind:  "neighbor",
		TargetScope: "neighbors",
		Effect: specialEffect{
			Type: "swapRevealedWithNeighbor",
			Params: map[string]any{
				"category": "facts",
			},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{{InstanceID: "sp-swap", Definition: def}}

	p1Fact1 := findHandCardBySlot(t, g, "p1", "facts1")
	p1Fact2 := findHandCardBySlot(t, g, "p1", "facts2")
	p2Fact1 := findHandCardBySlot(t, g, "p2", "facts1")
	p2Fact2 := findHandCardBySlot(t, g, "p2", "facts2")
	p1Fact1.Revealed = true
	p1Fact2.Revealed = true
	p2Fact1.Revealed = true
	p2Fact2.Revealed = true

	p1Fact1Before := p1Fact1.CardID
	p1Fact2Before := p1Fact2.CardID
	p2Fact1Before := p2Fact1.CardID
	p2Fact2Before := p2Fact2.CardID

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-swap",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"actorCardId":    p1Fact2.InstanceID,
			"targetCardId":   p2Fact2.InstanceID,
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}

	if p1Fact2.CardID != p2Fact2Before || p2Fact2.CardID != p1Fact2Before {
		t.Fatalf("selected cards were not swapped")
	}
	if p1Fact1.CardID != p1Fact1Before || p2Fact1.CardID != p2Fact1Before {
		t.Fatalf("non-selected cards should stay unchanged")
	}
}

func TestGameSession_SpecialReplaceUsesSelectedCategoryCard(t *testing.T) {
	g := newCategorySelectionSession()
	g.DeckPools[factsDeck] = append(g.DeckPools[factsDeck], assetCard{
		ID:    "decks/Факты/replace_extra.jpg",
		Deck:  factsDeck,
		Label: "Факт REPLACE",
	})
	def := specialDefinition{
		ID:          "test_replace_selected_fact",
		Title:       "Replace Fact",
		Implemented: true,
		ChoiceKind:  "player",
		TargetScope: "any_alive",
		Effect: specialEffect{
			Type: "replaceRevealedCard",
			Params: map[string]any{
				"category": "facts",
			},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{{InstanceID: "sp-replace", Definition: def}}

	p2Fact1 := findHandCardBySlot(t, g, "p2", "facts1")
	p2Fact2 := findHandCardBySlot(t, g, "p2", "facts2")
	p2Fact1.Revealed = true
	p2Fact2.Revealed = true
	p2Fact1Before := p2Fact1.CardID
	p2Fact2Before := p2Fact2.CardID

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-replace",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"targetCardId":   p2Fact2.InstanceID,
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}

	if p2Fact2.CardID == p2Fact2Before {
		t.Fatalf("selected card should be replaced")
	}
	if p2Fact1.CardID != p2Fact1Before {
		t.Fatalf("non-selected card should stay unchanged")
	}
}

func TestGameSession_SpecialDiscardUsesSelectedCategoryCard(t *testing.T) {
	g := newCategorySelectionSession()
	g.DeckPools[factsDeck] = append(g.DeckPools[factsDeck], assetCard{
		ID:    "decks/Факты/discard_extra.jpg",
		Deck:  factsDeck,
		Label: "Факт DISCARD",
	})
	def := specialDefinition{
		ID:          "test_discard_selected_fact",
		Title:       "Discard Fact",
		Implemented: true,
		ChoiceKind:  "player",
		TargetScope: "any_alive",
		Effect: specialEffect{
			Type: "discardRevealedAndDealHidden",
			Params: map[string]any{
				"category": "facts",
			},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{{InstanceID: "sp-discard", Definition: def}}

	p2Fact1 := findHandCardBySlot(t, g, "p2", "facts1")
	p2Fact2 := findHandCardBySlot(t, g, "p2", "facts2")
	p2Fact1.Revealed = true
	p2Fact2.Revealed = true
	p2Fact1Before := p2Fact1.CardID
	p2Fact2Before := p2Fact2.CardID

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-discard",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"targetCardId":   p2Fact2.InstanceID,
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}

	if p2Fact2.CardID == p2Fact2Before {
		t.Fatalf("selected card should be replaced")
	}
	if p2Fact2.Revealed {
		t.Fatalf("selected card should become hidden after discard effect")
	}
	if p2Fact1.CardID != p2Fact1Before {
		t.Fatalf("non-selected card should stay unchanged")
	}
}

func TestGameSession_SpecialSwapRejectsInvalidTargetCardID(t *testing.T) {
	g := newCategorySelectionSession()
	def := specialDefinition{
		ID:          "test_swap_invalid_target_card",
		Title:       "Swap Facts",
		Implemented: true,
		ChoiceKind:  "neighbor",
		TargetScope: "neighbors",
		Effect: specialEffect{
			Type: "swapRevealedWithNeighbor",
			Params: map[string]any{
				"category": "facts",
			},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{{InstanceID: "sp-swap-invalid", Definition: def}}

	p1Fact2 := findHandCardBySlot(t, g, "p1", "facts2")
	p1Fact2.Revealed = true
	findHandCardBySlot(t, g, "p2", "facts1").Revealed = true
	findHandCardBySlot(t, g, "p2", "facts2").Revealed = true

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-swap-invalid",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"actorCardId":    p1Fact2.InstanceID,
			"targetCardId":   p1Fact2.InstanceID, // карта не принадлежит цели
		},
	})
	if result.Error == "" {
		t.Fatalf("expected error for invalid targetCardId")
	}
}

func TestGameSession_SpecialReplaceRejectsHiddenSelectedCard(t *testing.T) {
	g := newCategorySelectionSession()
	g.DeckPools[factsDeck] = append(g.DeckPools[factsDeck], assetCard{
		ID:    "decks/Факты/replace_hidden_extra.jpg",
		Deck:  factsDeck,
		Label: "Факт REPLACE HIDDEN",
	})
	def := specialDefinition{
		ID:          "test_replace_hidden_selected_fact",
		Title:       "Replace Fact",
		Implemented: true,
		ChoiceKind:  "player",
		TargetScope: "any_alive",
		Effect: specialEffect{
			Type: "replaceRevealedCard",
			Params: map[string]any{
				"category": "facts",
			},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{{InstanceID: "sp-replace-hidden", Definition: def}}

	p2Fact2 := findHandCardBySlot(t, g, "p2", "facts2")
	p2Fact2.Revealed = false

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-replace-hidden",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"targetCardId":   p2Fact2.InstanceID,
		},
	})
	if result.Error == "" {
		t.Fatalf("expected error for hidden selected targetCardId")
	}
}

func TestGameSession_SpecialDiscardRejectsUnknownTargetCardID(t *testing.T) {
	g := newCategorySelectionSession()
	g.DeckPools[factsDeck] = append(g.DeckPools[factsDeck], assetCard{
		ID:    "decks/Факты/discard_invalid_extra.jpg",
		Deck:  factsDeck,
		Label: "Факт DISCARD INVALID",
	})
	def := specialDefinition{
		ID:          "test_discard_invalid_selected_fact",
		Title:       "Discard Fact",
		Implemented: true,
		ChoiceKind:  "player",
		TargetScope: "any_alive",
		Effect: specialEffect{
			Type: "discardRevealedAndDealHidden",
			Params: map[string]any{
				"category": "facts",
			},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{{InstanceID: "sp-discard-invalid", Definition: def}}

	findHandCardBySlot(t, g, "p2", "facts1").Revealed = true
	findHandCardBySlot(t, g, "p2", "facts2").Revealed = true

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-discard-invalid",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"targetCardId":   "unknown-card-id",
		},
	})
	if result.Error == "" {
		t.Fatalf("expected error for unknown targetCardId")
	}
}
