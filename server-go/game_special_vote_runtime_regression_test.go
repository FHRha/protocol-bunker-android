package main

import "testing"

func TestGameSession_ApplySpecial_StealBunkerCardToExiled_RemovesFaceAndMarksUsed(t *testing.T) {
	g := newTestGameSession()
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "bunker-visible",
			Title:      "Visible bunker",
			ImageID:    "bunker-visible",
			ImgURL:     "/assets/decks/1x/ru/Bunker/bunker-visible.png",
			IsRevealed: true,
		},
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-steal-bunker",
			Definition: specialDefinition{
				ID:          "steal_bunker",
				Title:       "Steal bunker",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "bunker",
				Effect: specialEffect{
					Type: "stealBunkerCardToExiled",
				},
			},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-steal-bunker",
		"payload": map[string]any{
			"bunkerIndex": 0,
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}

	card := g.World.Bunker[0]
	if card.ImageID != "" {
		t.Fatalf("removed bunker card must not keep image id")
	}
	if card.ImgURL != "" {
		t.Fatalf("removed bunker card must not keep imgUrl")
	}
	if !card.IsRevealed {
		t.Fatalf("removed bunker card should stay revealed placeholder")
	}
	if !g.Players["p1"].Specials[0].Used {
		t.Fatalf("special must be marked used")
	}
}

func TestGameSession_ApplySpecial_ReplaceBunkerCard_UpdatesImgURL(t *testing.T) {
	g := newTestGameSession()
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "bunker-old",
			Title:      "Old",
			ImageID:    "bunker-old",
			ImgURL:     "/assets/decks/1x/ru/Bunker/bunker-old.png",
			IsRevealed: true,
		},
	}
	g.DeckPools = map[string][]assetCard{
		"bunker": {
			{ID: "decks/1x/ru/Bunker/bunker-new.png", Label: "New bunker card"},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-replace-bunker-img",
			Definition: specialDefinition{
				ID:          "replace_bunker_img",
				Title:       "Replace bunker",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "bunker",
				Effect: specialEffect{
					Type: "replaceBunkerCard",
				},
			},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-replace-bunker-img",
		"payload": map[string]any{
			"bunkerIndex": 0,
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	card := g.World.Bunker[0]
	if card.ImageID != "decks/1x/ru/Bunker/bunker-new.png" {
		t.Fatalf("unexpected image id: %q", card.ImageID)
	}
	if card.ImgURL != "/assets/decks/1x/ru/Bunker/bunker-new.png" {
		t.Fatalf("unexpected imgUrl: %q", card.ImgURL)
	}
}

func TestGameSession_ApplySpecial_ReplaceBunkerCard_PublicViewUsesNewFace(t *testing.T) {
	g := newTestGameSession()
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "bunker-old",
			Title:      "Old",
			ImageID:    "decks/1x/ru/Bunker/bunker-old.png",
			ImgURL:     "/assets/decks/1x/ru/Bunker/bunker-old.png",
			IsRevealed: true,
		},
	}
	g.DeckPools = map[string][]assetCard{
		"bunker": {
			{ID: "decks/1x/ru/Bunker/bunker-new.png", Label: "New bunker card"},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-replace-bunker-view",
			Definition: specialDefinition{
				ID:          "replace_bunker_view",
				Title:       "Replace bunker",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "bunker",
				Effect: specialEffect{
					Type: "replaceBunkerCard",
				},
			},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-replace-bunker-view",
		"payload": map[string]any{
			"bunkerIndex": 0,
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}

	view := g.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")
	if view.World == nil || len(view.World.Bunker) != 1 {
		t.Fatalf("expected one bunker card in public view")
	}
	card := view.World.Bunker[0]
	if card.ImageID != "decks/1x/ru/Bunker/bunker-new.png" {
		t.Fatalf("unexpected public image id: %q", card.ImageID)
	}
	if card.ImgURL != "/assets/decks/1x/ru/Bunker/bunker-new.png" {
		t.Fatalf("unexpected public imgUrl: %q", card.ImgURL)
	}
}

func TestGameSession_ApplySpecial_DiscardBunkerCard_ClearsFaceFields(t *testing.T) {
	g := newTestGameSession()
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "bunker-visible",
			Title:      "Visible bunker",
			ImageID:    "bunker-visible",
			ImgURL:     "/assets/decks/1x/ru/Bunker/bunker-visible.png",
			IsRevealed: true,
		},
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-discard-bunker-img",
			Definition: specialDefinition{
				ID:          "discard_bunker_img",
				Title:       "Discard bunker",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "bunker",
				Effect: specialEffect{
					Type: "discardBunkerCard",
				},
			},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-discard-bunker-img",
		"payload": map[string]any{
			"bunkerIndex": 0,
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	card := g.World.Bunker[0]
	if card.ImageID != "" {
		t.Fatalf("discarded bunker card must clear imageId")
	}
	if card.ImgURL != "" {
		t.Fatalf("discarded bunker card must clear imgUrl")
	}
}

func TestGameSession_ApplySpecial_DiscardBunkerCard_PublicViewUsesHiddenPlaceholder(t *testing.T) {
	g := newTestGameSession()
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "bunker-visible",
			Title:      "Visible bunker",
			ImageID:    "decks/1x/ru/Bunker/bunker-visible.png",
			ImgURL:     "/assets/decks/1x/ru/Bunker/bunker-visible.png",
			IsRevealed: true,
		},
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-discard-bunker-view",
			Definition: specialDefinition{
				ID:          "discard_bunker_view",
				Title:       "Discard bunker",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "bunker",
				Effect: specialEffect{
					Type: "discardBunkerCard",
				},
			},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-discard-bunker-view",
		"payload": map[string]any{
			"bunkerIndex": 0,
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}

	view := g.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")
	if view.World == nil || len(view.World.Bunker) != 1 {
		t.Fatalf("expected one bunker card in public view")
	}
	card := view.World.Bunker[0]
	if card.ImageID != "" {
		t.Fatalf("discarded public bunker card must clear imageId")
	}
	if card.ImgURL != "" {
		t.Fatalf("discarded public bunker card must clear imgUrl")
	}
	if !card.IsRevealed {
		t.Fatalf("discarded bunker placeholder should stay revealed")
	}
}

func TestGameSession_ApplySpecial_SetRoundRule_ExposesRoundRulesInView(t *testing.T) {
	g := newTestGameSession()
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-round-rule",
			Definition: specialDefinition{
				ID:          "round_rule",
				Title:       "Silence",
				Trigger:     "active",
				Implemented: true,
				Effect: specialEffect{
					Type:   "setRoundRule",
					Params: map[string]any{"noTalkUntilVoting": true},
				},
			},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-round-rule",
		"payload":           map[string]any{},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !g.RoundRules.NoTalkUntilVoting {
		t.Fatalf("expected no-talk rule to be enabled")
	}

	view := g.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")
	if view.Public.RoundRules == nil {
		t.Fatalf("expected round rules in public view")
	}
	if !view.Public.RoundRules.NoTalkUntilVoting {
		t.Fatalf("expected public round rules to expose no-talk flag")
	}
}

func TestGameSession_ForcedWastedVoteOnNextVoting_PublicViewShowsForcedSelfReasonCode(t *testing.T) {
	g := newTestGameSession()
	g.Players["p2"].Status = playerEliminated
	g.FirstHealthRevealerID = "p2"
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-secret-self",
			Definition: specialDefinition{
				ID:          "sp-secret-self",
				Title:       "Secret Self Vote",
				Implemented: true,
				Trigger:     "secret_onEliminate",
				Effect: specialEffect{
					Type: "forcedWastedVoteOnNextVoting",
					Params: map[string]any{
						"condition": "firstRevealedHealthEliminated",
					},
				},
			},
		},
	}

	g.handleSecretEliminationTriggers("p2")
	g.startVoting()

	view := g.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")
	if len(view.Public.VotesPublic) == 0 {
		t.Fatalf("expected public votes")
	}

	for _, vote := range view.Public.VotesPublic {
		if vote.VoterID != "p1" {
			continue
		}
		if vote.Status != "voted" {
			t.Fatalf("expected p1 vote status to be voted, got %q", vote.Status)
		}
		if vote.TargetID != "p1" {
			t.Fatalf("expected p1 to be auto-voted for self, got %q", vote.TargetID)
		}
		if vote.ReasonCode != "VOTE_FORCED_SELF" {
			t.Fatalf("expected forced self reason code, got %q", vote.ReasonCode)
		}
		return
	}

	t.Fatalf("p1 public vote not found")
}

func TestGameSession_ApplySpecial_StealBaggageAndGiveSpecial_PublicViewParity(t *testing.T) {
	g := newTestGameSession()
	baggageDeck := categoryKeyToDeck["baggage"]
	if baggageDeck == "" {
		t.Fatalf("baggage deck name missing")
	}

	g.Players["p1"].Hand = nil
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-need-more-view",
			Definition: specialDefinition{
				ID:          "need_more_view",
				Title:       "Need More",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "player",
				TargetScope: "any_alive",
				Effect: specialEffect{
					Type:   "stealBaggage_and_giveSpecial",
					Params: map[string]any{"giveSpecialCount": 1},
				},
				AssetID: "decks/1x/ru/Special/special.need-more.png",
			},
		},
	}
	g.Players["p2"].Hand = []handCard{
		{
			InstanceID: "p2-bag-a",
			CardID:     "decks/1x/ru/Baggage/baggage.visible-a.png",
			Deck:       baggageDeck,
			Label:      "Bag A",
			Revealed:   true,
		},
	}
	g.specialPool = []specialDefinition{
		{
			ID:          "gift-special-view",
			Title:       "Gift",
			Trigger:     "active",
			Implemented: true,
			Effect:      specialEffect{Type: "none"},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-need-more-view",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"baggageCardId":  "p2-bag-a",
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}

	view := g.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")

	var p1SpecialSlot *publicCategorySlot
	var p2BaggageSlot *publicCategorySlot
	for i := range view.Public.Players {
		player := &view.Public.Players[i]
		if player.PlayerID == "p1" {
			for j := range player.Categories {
				if player.Categories[j].Category == specialDeckCategoryName {
					p1SpecialSlot = &player.Categories[j]
					break
				}
			}
		}
		if player.PlayerID == "p2" {
			for j := range player.Categories {
				if player.Categories[j].Category == canonicalCategoryKey("baggage") {
					p2BaggageSlot = &player.Categories[j]
					break
				}
			}
		}
	}

	if p1SpecialSlot == nil || len(p1SpecialSlot.Cards) == 0 {
		t.Fatalf("expected actor special slot with proxy card")
	}
	if !p1SpecialSlot.Cards[0].Revealed {
		t.Fatalf("expected actor proxy card to be publicly revealed")
	}
	if p1SpecialSlot.Cards[0].ImgURL == "" {
		t.Fatalf("expected actor proxy card to expose imgUrl")
	}
	if p1SpecialSlot.Cards[0].BackCategory != canonicalCategoryKey("baggage") {
		t.Fatalf("unexpected actor proxy back category: %q", p1SpecialSlot.Cards[0].BackCategory)
	}

	if p2BaggageSlot == nil || len(p2BaggageSlot.Cards) == 0 {
		t.Fatalf("expected target baggage slot")
	}
	foundHiddenSpecialBack := false
	for _, card := range p2BaggageSlot.Cards {
		if card.Hidden && card.BackCategory == specialDeckCategoryName {
			if card.ImgURL != "" {
				t.Fatalf("hidden target baggage replacement must not expose imgUrl")
			}
			foundHiddenSpecialBack = true
		}
	}
	if !foundHiddenSpecialBack {
		t.Fatalf("expected hidden target baggage replacement with special back")
	}
}

func TestGameSession_ApplySpecial_StealBaggageAndGiveSpecial_TargetYouViewGetsReplacementAndGift(t *testing.T) {
	g := newTestGameSession()
	baggageDeck := categoryKeyToDeck["baggage"]
	if baggageDeck == "" {
		t.Fatalf("baggage deck name missing")
	}
	initialTargetSpecials := len(g.Players["p2"].Specials)

	g.Players["p1"].Hand = nil
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-need-more-target-view",
			Definition: specialDefinition{
				ID:          "need_more_target_view",
				Title:       "Need More",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "player",
				TargetScope: "any_alive",
				Effect: specialEffect{
					Type:   "stealBaggage_and_giveSpecial",
					Params: map[string]any{"giveSpecialCount": 1},
				},
				AssetID: "decks/1x/ru/Special/special.need-more.png",
			},
		},
	}
	g.Players["p2"].Hand = []handCard{
		{
			InstanceID: "p2-bag-b",
			CardID:     "decks/1x/ru/Baggage/baggage.visible-b.png",
			Deck:       baggageDeck,
			Label:      "Bag B",
			Revealed:   false,
		},
	}
	g.specialPool = []specialDefinition{
		{
			ID:          "gift-special-target-view",
			Title:       "Gift",
			Trigger:     "active",
			Implemented: true,
			Effect:      specialEffect{Type: "none"},
			AssetID:     "decks/1x/ru/Special/special.gift.png",
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-need-more-target-view",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"baggageCardId":  "p2-bag-b",
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}

	view := g.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p2")
	if len(view.You.Specials) != initialTargetSpecials+1 {
		t.Fatalf("expected one more gifted special in target you-view, got %d", len(view.You.Specials))
	}
	last := view.You.Specials[len(view.You.Specials)-1]
	if last.ID != "gift-special-target-view" {
		t.Fatalf("unexpected gifted special id: %q", last.ID)
	}

	var baggageSlot *youCategorySlot
	for i := range view.You.Categories {
		if view.You.Categories[i].Category == canonicalCategoryKey("baggage") {
			baggageSlot = &view.You.Categories[i]
			break
		}
	}
	if baggageSlot == nil || len(baggageSlot.Cards) == 0 {
		t.Fatalf("expected baggage slot in target you-view")
	}
	card := baggageSlot.Cards[0]
	if card.Deck != baggageDeck {
		t.Fatalf("unexpected baggage deck: %q", card.Deck)
	}
	if card.Label != "Need More" {
		t.Fatalf("unexpected replacement label: %q", card.Label)
	}
	if card.ImgURL == "" {
		t.Fatalf("expected replacement special card to expose imgUrl in target you-view")
	}
}
