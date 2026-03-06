package main

import (
	"strings"
	"testing"
)

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

func TestGameSession_OnOwnerEliminated_SetsPendingActivationForChoiceSpecial(t *testing.T) {
	g := newTestGameSession()
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-owner-choice",
			Definition: specialDefinition{
				ID:          "owner_choice",
				Title:       "Owner Choice",
				Trigger:     "onOwnerEliminated",
				Implemented: true,
				ChoiceKind:  "category",
				Effect: specialEffect{
					Type:   "forceRevealCategoryForAll",
					Params: map[string]any{},
				},
			},
		},
	}

	g.applyElimination("p1")
	special := g.Players["p1"].Specials[0]

	if special.Used {
		t.Fatalf("choice special must stay unused until manual activation")
	}
	if !special.PendingActivation {
		t.Fatalf("choice special must be marked pendingActivation")
	}
	if !special.RevealedPublic {
		t.Fatalf("choice special should become revealed when pending")
	}
}

func TestGameSession_OnOwnerEliminated_EliminatedPlayerCanApplyPendingSpecial(t *testing.T) {
	g := newTestGameSession()
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-owner-apply",
			Definition: specialDefinition{
				ID:          "owner_apply",
				Title:       "Owner Apply",
				Trigger:     "onOwnerEliminated",
				Implemented: true,
				ChoiceKind:  "category",
				Effect: specialEffect{
					Type:   "forceRevealCategoryForAll",
					Params: map[string]any{},
				},
			},
		},
	}

	g.applyElimination("p1")
	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-owner-apply",
		"payload": map[string]any{
			"category": "profession",
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial for pending card failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("pending activation should change state")
	}

	special := g.Players["p1"].Specials[0]
	if !special.Used {
		t.Fatalf("special should become used after apply")
	}
	if special.PendingActivation {
		t.Fatalf("pendingActivation should be cleared after apply")
	}
	if g.RoundRules.ForcedCategory == "" {
		t.Fatalf("effect should be applied for eliminated owner card")
	}
}

func TestGameSession_OnOwnerEliminated_BunkerChoiceWithoutRevealedBunkerIsConsumed(t *testing.T) {
	g := newTestGameSession()
	for i := range g.World.Bunker {
		g.World.Bunker[i].IsRevealed = false
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-owner-bunker",
			Definition: specialDefinition{
				ID:          "owner_bunker",
				Title:       "Owner Bunker",
				Trigger:     "onOwnerEliminated",
				Implemented: true,
				ChoiceKind:  "bunker",
				Effect: specialEffect{
					Type:   "discardBunkerCard",
					Params: map[string]any{},
				},
			},
		},
	}

	g.applyElimination("p1")
	special := g.Players["p1"].Specials[0]

	if !special.Used {
		t.Fatalf("bunker-choice special should be consumed when no revealed bunker cards exist")
	}
	if special.PendingActivation {
		t.Fatalf("bunker-choice special should not stay pending without revealed bunker cards")
	}
}

func TestGameSession_ApplySpecial_ReplaceBunkerCard(t *testing.T) {
	g := newTestGameSession()
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "bunker-old",
			Title:      "Old",
			ImageID:    "bunker-old",
			IsRevealed: true,
		},
	}
	g.DeckPools = map[string][]assetCard{
		"bunker": {
			{ID: "bunker-new", Label: "New bunker card"},
		},
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-replace-bunker",
			Definition: specialDefinition{
				ID:          "replace_bunker",
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
		"specialInstanceId": "sp-replace-bunker",
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
	if g.World.Bunker[0].ID != "bunker-new" {
		t.Fatalf("unexpected bunker card id after replace: %q", g.World.Bunker[0].ID)
	}
	if g.World.Bunker[0].ImageID != "bunker-new" {
		t.Fatalf("unexpected bunker image after replace: %q", g.World.Bunker[0].ImageID)
	}
	if !g.Players["p1"].Specials[0].Used {
		t.Fatalf("special must be marked as used")
	}
}

func TestGameSession_ApplySpecial_DiscardBunkerCard_AllowsRandomTarget(t *testing.T) {
	g := newTestGameSession()
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "bunker-visible",
			Title:      "Visible bunker",
			ImageID:    "bunker-visible",
			IsRevealed: true,
		},
		{
			Kind:       "bunker",
			ID:         "bunker-hidden",
			Title:      "Hidden bunker",
			ImageID:    "bunker-hidden",
			IsRevealed: false,
		},
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-discard-bunker",
			Definition: specialDefinition{
				ID:          "discard_bunker",
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
		"specialInstanceId": "sp-discard-bunker",
		"payload":           map[string]any{},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}
	if g.World.Bunker[0].Title != "Карта бункера потеряна" {
		t.Fatalf("unexpected discard title: %q", g.World.Bunker[0].Title)
	}
	if g.World.Bunker[0].ImageID != "" {
		t.Fatalf("discarded bunker card must not have image")
	}
	if !g.Players["p1"].Specials[0].Used {
		t.Fatalf("special must be marked as used")
	}
}

func TestGameSession_ApplySpecial_StealBaggageAndGiveSpecial_FullFlow(t *testing.T) {
	g := newTestGameSession()
	baggageDeck := categoryKeyToDeck["baggage"]
	if baggageDeck == "" {
		t.Fatalf("baggage deck name missing")
	}

	g.Players["p1"].Hand = []handCard{}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-need-more",
			Definition: specialDefinition{
				ID:          "need_more",
				Title:       "Мне Нужнее",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "player",
				TargetScope: "any_alive",
				Effect: specialEffect{
					Type:   "stealBaggage_and_giveSpecial",
					Params: map[string]any{"giveSpecialCount": 1},
				},
				AssetID: "decks/Особые условия/Мне Нужнее.jpg",
			},
		},
	}

	g.Players["p2"].Hand = []handCard{
		{
			InstanceID: "p2-bag-a",
			CardID:     "decks/Багаж/Bag-A.jpg",
			Deck:       baggageDeck,
			Label:      "Bag A",
			Revealed:   true,
		},
		{
			InstanceID: "p2-bag-b",
			CardID:     "decks/Багаж/Bag-B.jpg",
			Deck:       baggageDeck,
			Label:      "Bag B",
			Revealed:   true,
		},
	}
	g.Players["p2"].Specials = nil

	g.specialPool = []specialDefinition{
		{
			ID:          "gift-special",
			Title:       "Подарок",
			Trigger:     "active",
			Implemented: true,
			Effect: specialEffect{
				Type: "none",
			},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-need-more",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"baggageCardId":  "p2-bag-b",
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}

	if len(g.Players["p1"].Specials) != 0 {
		t.Fatalf("used special should be removed from player special list")
	}
	if len(g.Players["p1"].SpecialCategoryProxyCards) != 1 {
		t.Fatalf("expected one proxy card in special category")
	}
	if g.Players["p1"].SpecialCategoryProxyCards[0].Label != "Bag B" {
		t.Fatalf("unexpected proxy label: %q", g.Players["p1"].SpecialCategoryProxyCards[0].Label)
	}
	if len(g.Players["p1"].Hand) != 1 || g.Players["p1"].Hand[0].Label != "Bag B" {
		t.Fatalf("stolen baggage must be moved to actor hand")
	}

	if len(g.Players["p2"].Specials) != 1 {
		t.Fatalf("target should receive one new special")
	}

	targetHasHiddenSpecialBaggage := false
	for _, card := range g.Players["p2"].Hand {
		if card.Label == "Мне Нужнее" && card.Deck == baggageDeck {
			targetHasHiddenSpecialBaggage = true
			if card.Revealed {
				t.Fatalf("special baggage card on target must stay hidden")
			}
			if card.PublicBackCategory != specialDeckCategoryName {
				t.Fatalf("special baggage card must keep special back category, got %q", card.PublicBackCategory)
			}
		}
		if card.InstanceID == "p2-bag-b" {
			t.Fatalf("selected baggage card must be removed from target hand")
		}
	}
	if !targetHasHiddenSpecialBaggage {
		t.Fatalf("target must get hidden baggage card with used special face")
	}

	targetPublic := g.buildPublicCategories(g.Players["p2"])
	var baggageSlot *publicCategorySlot
	for i := range targetPublic {
		if targetPublic[i].Category == categoryKeyToLabel["baggage"] {
			baggageSlot = &targetPublic[i]
			break
		}
	}
	if baggageSlot == nil || len(baggageSlot.Cards) == 0 {
		t.Fatalf("expected baggage slot with hidden cards")
	}
	foundSpecialBack := false
	for _, card := range baggageSlot.Cards {
		if card.Hidden && card.BackCategory == specialDeckCategoryName {
			foundSpecialBack = true
			break
		}
	}
	if !foundSpecialBack {
		t.Fatalf("expected hidden baggage proxy card with special back category")
	}
}

func TestGameSession_ApplySpecial_StealBaggageAndGiveSpecial_InvalidBaggageCardID(t *testing.T) {
	g := newTestGameSession()
	baggageDeck := categoryKeyToDeck["baggage"]
	if baggageDeck == "" {
		t.Fatalf("baggage deck name missing")
	}

	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-need-more",
			Definition: specialDefinition{
				ID:          "need_more",
				Title:       "Мне Нужнее",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "player",
				TargetScope: "any_alive",
				Effect: specialEffect{
					Type: "stealBaggage_and_giveSpecial",
				},
			},
		},
	}
	g.Players["p2"].Hand = []handCard{
		{
			InstanceID: "p2-bag-a",
			CardID:     "decks/Багаж/Bag-A.jpg",
			Deck:       baggageDeck,
			Label:      "Bag A",
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-need-more",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"baggageCardId":  "missing-card",
		},
	})
	if result.Error == "" {
		t.Fatalf("expected error for invalid baggageCardId")
	}
}

func TestGameSession_StealBaggageChoiceFallbackByEffect(t *testing.T) {
	g := newTestGameSession()
	baggageDeck := categoryKeyToDeck["baggage"]
	if baggageDeck == "" {
		t.Fatalf("baggage deck name missing")
	}

	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-steal-fallback",
			Definition: specialDefinition{
				ID:          "need_more_fallback",
				Title:       "Мне Нужнее",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "",
				TargetScope: "",
				Requires:    []string{"targetHasBaggage"},
				Effect: specialEffect{
					Type:   "stealBaggage_and_giveSpecial",
					Params: map[string]any{"giveSpecialCount": 1},
				},
			},
		},
	}
	g.Players["p2"].Hand = []handCard{
		{
			InstanceID: "p2-bag-a",
			CardID:     "decks/Багаж/Bag-A.jpg",
			Deck:       baggageDeck,
			Label:      "Bag A",
			Revealed:   true,
		},
	}
	g.specialPool = []specialDefinition{
		{
			ID:          "gift-special",
			Title:       "Подарок",
			Trigger:     "active",
			Implemented: true,
			Effect: specialEffect{
				Type: "none",
			},
		},
	}

	view := g.buildSpecialInstances(g.Players["p1"])
	if len(view) != 1 {
		t.Fatalf("expected one special view")
	}
	if view[0].ChoiceKind != "player" {
		t.Fatalf("expected fallback choiceKind=player, got %q", view[0].ChoiceKind)
	}
	if view[0].TargetScope != "any_alive" {
		t.Fatalf("expected fallback targetScope=any_alive, got %q", view[0].TargetScope)
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-steal-fallback",
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
}

func TestGameSession_DevInitialSpecials_OnlyDevChoiceCard(t *testing.T) {
	g := newTestGameSession()
	g.IsDev = true

	player := g.Players["p1"]
	player.Specials = nil
	g.assignInitialSpecialsForPlayer(player, nil)

	if len(player.Specials) != 1 {
		t.Fatalf("expected exactly one dev special, got %d", len(player.Specials))
	}
	if player.Specials[0].Definition.Effect.Type != devChoiceEffectType {
		t.Fatalf("expected dev choice effect, got %q", player.Specials[0].Definition.Effect.Type)
	}
	if !player.Specials[0].RevealedPublic {
		t.Fatalf("dev special must be revealed")
	}
}

func TestGameSession_ResolveBunkerIndex_AllowsDevUnrevealedCards(t *testing.T) {
	g := newTestGameSession()
	g.IsDev = true
	if len(g.World.Bunker) == 0 {
		t.Fatalf("expected bunker cards")
	}
	for i := range g.World.Bunker {
		g.World.Bunker[i].IsRevealed = false
	}

	index, errText := g.resolveBunkerIndex(map[string]any{"bunkerIndex": 0}, false)
	if errText != "" {
		t.Fatalf("resolveBunkerIndex returned error in dev: %s", errText)
	}
	if index != 0 {
		t.Fatalf("unexpected index: %d", index)
	}
}

func TestGameSession_ApplySpecial_StealBaggageAndGiveSpecial_DevTreatsVisibleAsRevealed(t *testing.T) {
	g := newTestGameSession()
	g.IsDev = true
	baggageDeck := categoryKeyToDeck["baggage"]
	if baggageDeck == "" {
		t.Fatalf("baggage deck name missing")
	}

	g.Players["p1"].Hand = nil
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: "sp-need-more-dev",
			Definition: specialDefinition{
				ID:          "need_more_dev",
				Title:       "Мне нужнее",
				Trigger:     "active",
				Implemented: true,
				ChoiceKind:  "player",
				TargetScope: "any_alive",
				Effect: specialEffect{
					Type:   "stealBaggage_and_giveSpecial",
					Params: map[string]any{"giveSpecialCount": 1},
				},
			},
		},
	}
	g.Players["p2"].Hand = []handCard{
		{
			InstanceID: "p2-bag-hidden",
			CardID:     "decks/Багаж/Bag-Hidden.jpg",
			Deck:       baggageDeck,
			Label:      "Bag Hidden",
			Revealed:   false,
		},
	}
	g.specialPool = []specialDefinition{
		{
			ID:          "gift-special-dev",
			Title:       "Подарок",
			Trigger:     "active",
			Implemented: true,
			Effect: specialEffect{
				Type: "none",
			},
		},
	}

	result := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-need-more-dev",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"baggageCardId":  "p2-bag-hidden",
		},
	})
	if result.Error != "" {
		t.Fatalf("applySpecial failed: %s", result.Error)
	}
	if !result.StateChanged {
		t.Fatalf("expected state changed")
	}
	if len(g.Players["p1"].Hand) == 0 {
		t.Fatalf("expected stolen card in actor hand")
	}
	if !g.Players["p1"].Hand[0].Revealed {
		t.Fatalf("stolen card should be treated as revealed in dev")
	}
	if len(g.Players["p1"].SpecialCategoryProxyCards) == 0 {
		t.Fatalf("expected proxy card for stolen baggage")
	}
	if !g.Players["p1"].SpecialCategoryProxyCards[0].Revealed {
		t.Fatalf("proxy card should be revealed in dev")
	}
}

func TestGameSession_DevChoice_ResolvesSpecialAssetForNeedMoreFlow(t *testing.T) {
	g := newTestGameSession()
	g.IsDev = true

	baggageDeck := categoryKeyToDeck["baggage"]
	if baggageDeck == "" {
		t.Fatalf("baggage deck name missing")
	}

	needMoreAssetID := "decks/Особые условия/Мне нужнее.jpg"
	g.DeckPools[specialDeckCategoryName] = []assetCard{
		{
			ID:    needMoreAssetID,
			Deck:  specialDeckCategoryName,
			Label: "Мне нужнее",
		},
	}

	g.specialCatalog = []specialDefinition{
		{
			ID:          "need_more",
			Title:       "Мне нужнее",
			Trigger:     "active",
			Implemented: true,
			ChoiceKind:  "player",
			TargetScope: "any_alive",
			Effect: specialEffect{
				Type:   "stealBaggage_and_giveSpecial",
				Params: map[string]any{"giveSpecialCount": 1},
			},
			AssetID: "",
		},
	}
	g.specialPool = []specialDefinition{
		{
			ID:          "gift-special",
			Title:       "Подарок",
			Trigger:     "active",
			Implemented: true,
			Effect: specialEffect{
				Type: "none",
			},
		},
	}

	devChoice, ok := g.buildDevChoiceDefinition("p1")
	if !ok {
		t.Fatalf("expected dev choice definition")
	}
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID:     "sp-dev-choice",
			Definition:     devChoice,
			RevealedPublic: true,
		},
	}
	g.Players["p2"].Hand = []handCard{
		{
			InstanceID: "p2-bag-a",
			CardID:     "decks/Багаж/Bag-A.jpg",
			Deck:       baggageDeck,
			Label:      "Bag A",
			Revealed:   true,
		},
	}

	chooseResult := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-dev-choice",
		"payload": map[string]any{
			"specialId": "need_more",
		},
	})
	if chooseResult.Error != "" {
		t.Fatalf("dev choose apply failed: %s", chooseResult.Error)
	}
	if !chooseResult.StateChanged {
		t.Fatalf("expected state change on dev choose")
	}

	if len(g.Players["p1"].Specials) != 1 {
		t.Fatalf("expected one special after dev choose")
	}
	if got := g.Players["p1"].Specials[0].Definition.AssetID; got != needMoreAssetID {
		t.Fatalf("dev-selected special asset must resolve, got %q", got)
	}

	applyResult := g.handleAction("p1", "applySpecial", map[string]any{
		"specialInstanceId": "sp-dev-choice",
		"payload": map[string]any{
			"targetPlayerId": "p2",
			"baggageCardId":  "p2-bag-a",
		},
	})
	if applyResult.Error != "" {
		t.Fatalf("need more apply failed: %s", applyResult.Error)
	}
	if !applyResult.StateChanged {
		t.Fatalf("expected state change on need more apply")
	}

	targetHasSpecialFaceInBaggage := false
	for _, card := range g.Players["p2"].Hand {
		if strings.Contains(card.Label, "Мне нужнее") && card.Deck == baggageDeck {
			targetHasSpecialFaceInBaggage = true
			if card.CardID != needMoreAssetID {
				t.Fatalf("expected special face card id %q, got %q", needMoreAssetID, card.CardID)
			}
			break
		}
	}
	if !targetHasSpecialFaceInBaggage {
		t.Fatalf("target must have baggage replacement with selected special face")
	}
}

func TestGameSession_SecretOnEliminate_AutoVotesSelfOnNextVoting(t *testing.T) {
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

	events := g.handleSecretEliminationTriggers("p2")
	if !g.Players["p1"].Specials[0].Used {
		t.Fatalf("secret special should be marked as used after trigger")
	}
	if !g.Players["p1"].Specials[0].RevealedPublic {
		t.Fatalf("secret special should become revealed after trigger")
	}
	if len(events) == 0 {
		t.Fatalf("expected reveal notification event for triggered secret special")
	}
	if !g.Players["p1"].ForcedSelfVoteNext {
		t.Fatalf("expected forced self vote flag on next voting")
	}

	g.startVoting()

	vote, ok := g.Votes["p1"]
	if !ok {
		t.Fatalf("expected auto vote record for p1")
	}
	if vote.TargetID != "p1" {
		t.Fatalf("expected auto vote target to be self, got %q", vote.TargetID)
	}
	if !vote.IsValid {
		t.Fatalf("expected self auto vote to be valid")
	}
	if _, disabled := g.VoteDisabled["p1"]; disabled {
		t.Fatalf("self auto vote must not disable voter")
	}
}

func TestGameSession_SecretOnEliminate_AutoSelfVoteOnlyForOneVoting(t *testing.T) {
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
	g.startTieBreakRevote([]string{"p1", "p3"})

	if _, ok := g.Votes["p1"]; ok {
		t.Fatalf("expected self auto vote to apply only for the first voting attempt, not revote")
	}
}
