package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

const specialPerCardInstanceID = "sp-per-card"

func TestGameSession_SpecialCards_PerCardExecutable(t *testing.T) {
	definitions, err := loadImplementedSpecialDefinitionsFromFile(filepath.Join("..", "scenarios", "classic", "SPECIAL_CONDITIONS.json"))
	if err != nil {
		t.Fatalf("failed to load implemented specials: %v", err)
	}
	if len(definitions) == 0 {
		t.Fatalf("implemented specials list is empty")
	}

	for _, definition := range definitions {
		definition := definition
		t.Run(perCardSubtestName(definition), func(t *testing.T) {
			g := newPerCardSpecialSession(definitions)
			assignSpecialUnderTest(g, definition)
			executePerCardFlow(t, g, definition)
		})
	}
}

func TestGameSession_SpecialCards_CategorySelection_AllOptionsAccepted(t *testing.T) {
	definitions, err := loadImplementedSpecialDefinitionsFromFile(filepath.Join("..", "scenarios", "classic", "SPECIAL_CONDITIONS.json"))
	if err != nil {
		t.Fatalf("failed to load implemented specials: %v", err)
	}

	for _, definition := range definitions {
		definition := definition
		effectType := strings.TrimSpace(definition.Effect.Type)
		switch effectType {
		case "swapRevealedWithNeighbor", "replaceRevealedCard", "discardRevealedAndDealHidden":
		default:
			continue
		}

		category := strings.TrimSpace(asString(definition.Effect.Params["category"]))
		if category == "" {
			continue
		}

		t.Run("all_options__"+perCardSubtestName(definition), func(t *testing.T) {
			base := newPerCardSpecialSession(definitions)
			targetOptions := gCategoryCardIDs(base, "p2", category)
			if len(targetOptions) <= 1 {
				t.Skipf("category %q has <= 1 selectable cards for target; nothing to matrix-test", category)
			}

			switch effectType {
			case "swapRevealedWithNeighbor":
				actorOptions := gCategoryCardIDs(base, "p1", category)
				if len(actorOptions) == 0 {
					t.Fatalf("actor has no cards in category=%q", category)
				}
				for _, actorCardID := range actorOptions {
					for _, targetCardID := range targetOptions {
						g := newPerCardSpecialSession(definitions)
						assignSpecialUnderTest(g, definition)
						g.Phase = scenarioPhaseReveal
						result := g.handleAction("p1", "applySpecial", map[string]any{
							"specialInstanceId": specialPerCardInstanceID,
							"payload": map[string]any{
								"targetPlayerId": "p2",
								"actorCardId":    actorCardID,
								"targetCardId":   targetCardID,
							},
						})
						if result.Error != "" {
							t.Fatalf("swap failed for actorCardId=%s targetCardId=%s: %s", actorCardID, targetCardID, result.Error)
						}
						if !result.StateChanged {
							t.Fatalf("swap should change state for actorCardId=%s targetCardId=%s", actorCardID, targetCardID)
						}
					}
				}
			case "replaceRevealedCard", "discardRevealedAndDealHidden":
				for _, targetCardID := range targetOptions {
					g := newPerCardSpecialSession(definitions)
					assignSpecialUnderTest(g, definition)
					g.Phase = scenarioPhaseReveal
					result := g.handleAction("p1", "applySpecial", map[string]any{
						"specialInstanceId": specialPerCardInstanceID,
						"payload": map[string]any{
							"targetPlayerId": "p2",
							"targetCardId":   targetCardID,
						},
					})
					if result.Error != "" {
						t.Fatalf("%s failed for targetCardId=%s: %s", effectType, targetCardID, result.Error)
					}
					if !result.StateChanged {
						t.Fatalf("%s should change state for targetCardId=%s", effectType, targetCardID)
					}
				}
			}
		})
	}
}

func perCardSubtestName(def specialDefinition) string {
	id := strings.TrimSpace(def.ID)
	if id == "" {
		id = strings.TrimSpace(def.Title)
	}
	id = strings.ReplaceAll(id, "\\", "/")
	id = strings.ReplaceAll(id, "/", "_")
	if id == "" {
		id = "unknown"
	}
	trigger := strings.TrimSpace(def.Trigger)
	if trigger == "" {
		trigger = "active"
	}
	effect := strings.TrimSpace(def.Effect.Type)
	if effect == "" {
		effect = "none"
	}
	return fmt.Sprintf("%s__%s__%s", trigger, effect, id)
}

func newPerCardSpecialSession(definitions []specialDefinition) *gameSession {
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
		makePerCardAssetCatalog(len(players)),
		definitions,
		20260306,
	)
	g.specialPool = cloneSpecialDefinitions(definitions)
	g.specialCatalog = cloneSpecialDefinitions(definitions)
	revealAllCards(g)
	ensureRevealedBunkerCard(g)
	return g
}

func makePerCardAssetCatalog(playerCount int) assetCatalog {
	catalog := makeHandAssetCatalog(playerCount)

	for deckIndex, deckName := range coreDecks {
		for i := 0; i < 8; i++ {
			catalog.Decks[deckName] = append(catalog.Decks[deckName], assetCard{
				ID:    fmt.Sprintf("decks/extra/core_%d_%d.jpg", deckIndex, i),
				Deck:  deckName,
				Label: fmt.Sprintf("extra-core-%d-%d", deckIndex, i),
			})
		}
	}

	for i := 0; i < 16; i++ {
		catalog.Decks[factsDeck] = append(catalog.Decks[factsDeck], assetCard{
			ID:    fmt.Sprintf("decks/extra/facts_%d.jpg", i),
			Deck:  factsDeck,
			Label: fmt.Sprintf("extra-fact-%d", i),
		})
	}

	catalog.Decks["bunker"] = make([]assetCard, 0, 12)
	for i := 0; i < 12; i++ {
		catalog.Decks["bunker"] = append(catalog.Decks["bunker"], assetCard{
			ID:    fmt.Sprintf("decks/bunker/test_%d.jpg", i),
			Deck:  "bunker",
			Label: fmt.Sprintf("Bunker %d", i),
		})
	}

	catalog.Decks["disaster"] = []assetCard{
		{ID: "decks/disaster/test_0.jpg", Deck: "disaster", Label: "Disaster 0"},
		{ID: "decks/disaster/test_1.jpg", Deck: "disaster", Label: "Disaster 1"},
	}

	catalog.Decks["threat"] = make([]assetCard, 0, 8)
	for i := 0; i < 8; i++ {
		catalog.Decks["threat"] = append(catalog.Decks["threat"], assetCard{
			ID:    fmt.Sprintf("decks/threat/test_%d.jpg", i),
			Deck:  "threat",
			Label: fmt.Sprintf("Threat %d", i),
		})
	}

	return catalog
}

func revealAllCards(g *gameSession) {
	for _, playerID := range g.Order {
		player := g.Players[playerID]
		if player == nil {
			continue
		}
		for i := range player.Hand {
			player.Hand[i].Revealed = true
		}
	}
}

func ensureRevealedBunkerCard(g *gameSession) {
	if len(g.World.Bunker) == 0 {
		g.World.Bunker = []worldFacedCardView{
			{
				Kind:       "bunker",
				ID:         "bunker-fallback-0",
				Title:      "Bunker fallback",
				ImageID:    "decks/bunker/fallback.jpg",
				IsRevealed: true,
			},
		}
		return
	}
	g.World.Bunker[0].IsRevealed = true
	if strings.TrimSpace(g.World.Bunker[0].ID) == "" {
		g.World.Bunker[0].ID = "bunker-fallback-0"
	}
	if strings.TrimSpace(g.World.Bunker[0].Title) == "" {
		g.World.Bunker[0].Title = "Bunker fallback"
	}
}

func assignSpecialUnderTest(g *gameSession, definition specialDefinition) {
	g.Players["p1"].Specials = []specialConditionState{
		{
			InstanceID: specialPerCardInstanceID,
			Definition: copySpecialDefinition(definition),
		},
	}
}

func executePerCardFlow(t *testing.T, g *gameSession, definition specialDefinition) {
	t.Helper()

	switch strings.TrimSpace(definition.Trigger) {
	case "", "active", "onRevealOrActive", "onVote":
		prepareManualSpecialFlow(g, definition)
		payload := makePayloadForSpecial(t, g, definition)
		result := g.handleAction("p1", "applySpecial", map[string]any{
			"specialInstanceId": specialPerCardInstanceID,
			"payload":           payload,
		})
		if result.Error != "" {
			t.Fatalf("applySpecial failed (trigger=%q effect=%q): %s", definition.Trigger, definition.Effect.Type, result.Error)
		}
		if !result.StateChanged {
			t.Fatalf("expected state change for effect=%q", definition.Effect.Type)
		}
		assertSpecialConsumedForManualFlow(t, g, definition)

	case "onOwnerEliminated":
		g.applyElimination("p1")
		special := g.findPlayerSpecialByInstance(g.Players["p1"], specialPerCardInstanceID)
		if special == nil {
			t.Fatalf("special not found after owner elimination")
		}
		if special.PendingActivation {
			payload := makePayloadForSpecial(t, g, definition)
			result := g.handleAction("p1", "applySpecial", map[string]any{
				"specialInstanceId": specialPerCardInstanceID,
				"payload":           payload,
			})
			if result.Error != "" {
				t.Fatalf("pending applySpecial failed (effect=%q): %s", definition.Effect.Type, result.Error)
			}
			if !result.StateChanged {
				t.Fatalf("expected state change for pending effect=%q", definition.Effect.Type)
			}
			special = g.findPlayerSpecialByInstance(g.Players["p1"], specialPerCardInstanceID)
			if special == nil {
				t.Fatalf("special disappeared after pending activation")
			}
		}
		if !special.Used {
			t.Fatalf("special should be used after onOwnerEliminated flow")
		}

	case "secret_onEliminate":
		eliminatedID := secretTriggerTargetPlayerID(t, g, definition)
		g.handleSecretEliminationTriggers(eliminatedID)
		special := g.findPlayerSpecialByInstance(g.Players["p1"], specialPerCardInstanceID)
		if special == nil {
			t.Fatalf("secret special not found after elimination")
		}
		if !special.Used {
			t.Fatalf("secret special should be used after trigger")
		}
		if !special.RevealedPublic {
			t.Fatalf("secret special should be revealed after trigger")
		}
		if !g.Players["p1"].ForcedSelfVoteNext {
			t.Fatalf("secret trigger must set forced self vote for next voting")
		}

	default:
		t.Fatalf("unsupported trigger in per-card test: %q", definition.Trigger)
	}
}

func prepareManualSpecialFlow(g *gameSession, definition specialDefinition) {
	switch strings.TrimSpace(definition.Effect.Type) {
	case "banVoteAgainst", "voteWeight", "disableVote", "doubleVotesAgainst_and_disableSelfVote", "forceRevote":
		prepareVotingSpecialWindow(g)
	default:
		g.Phase = scenarioPhaseReveal
		g.VotePhase = ""
	}
}

func prepareVotingSpecialWindow(g *gameSession) {
	g.startVoting()
	g.Votes = map[string]voteRecord{
		"p1": {TargetID: "p2", IsValid: true, Submitted: 1},
		"p2": {TargetID: "p1", IsValid: true, Submitted: 2},
		"p3": {TargetID: "p2", IsValid: true, Submitted: 3},
	}
	g.BaseVotes = copyVotes(g.Votes)
	g.VotePhase = votePhaseSpecialWindow
}

func makePayloadForSpecial(t *testing.T, g *gameSession, definition specialDefinition) map[string]any {
	t.Helper()

	switch strings.TrimSpace(definition.Effect.Type) {
	case "banVoteAgainst", "disableVote", "doubleVotesAgainst_and_disableSelfVote":
		return map[string]any{"targetPlayerId": "p2"}
	case "voteWeight", "forceRevote", "redealAllRevealed", "setRoundRule", "addFinalThreat":
		return map[string]any{}
	case "swapRevealedWithNeighbor":
		category := strings.TrimSpace(asString(definition.Effect.Params["category"]))
		actorCardID := firstCategoryCardInstanceID(t, g, "p1", category)
		targetCardID := firstCategoryCardInstanceID(t, g, "p2", category)
		return map[string]any{
			"targetPlayerId": "p2",
			"actorCardId":    actorCardID,
			"targetCardId":   targetCardID,
		}
	case "replaceRevealedCard", "discardRevealedAndDealHidden":
		category := strings.TrimSpace(asString(definition.Effect.Params["category"]))
		targetCardID := firstCategoryCardInstanceID(t, g, "p2", category)
		return map[string]any{
			"targetPlayerId": "p2",
			"targetCardId":   targetCardID,
		}
	case "replaceBunkerCard", "discardBunkerCard", "stealBunkerCardToExiled":
		return map[string]any{
			"bunkerIndex": firstRevealedBunkerIndex(t, g),
		}
	case "forceRevealCategoryForAll":
		category := strings.TrimSpace(asString(definition.Effect.Params["category"]))
		if category == "" {
			category = "profession"
		}
		return map[string]any{"category": category}
	case "stealBaggage_and_giveSpecial":
		return map[string]any{
			"targetPlayerId": "p2",
			"baggageCardId":  firstCategoryCardInstanceID(t, g, "p2", "baggage"),
		}
	default:
		t.Fatalf("unsupported effect in per-card payload builder: %q", definition.Effect.Type)
		return nil
	}
}

func firstCategoryCardInstanceID(t *testing.T, g *gameSession, playerID, categoryKey string) string {
	t.Helper()
	player := g.Players[playerID]
	if player == nil {
		t.Fatalf("player not found: %s", playerID)
	}
	cards := g.getCardsByCategoryKey(player, categoryKey, false)
	if len(cards) == 0 {
		t.Fatalf("no cards for player=%s category=%s", playerID, categoryKey)
	}
	return cards[0].InstanceID
}

func gCategoryCardIDs(g *gameSession, playerID, categoryKey string) []string {
	player := g.Players[playerID]
	if player == nil {
		return nil
	}
	cards := g.getCardsByCategoryKey(player, categoryKey, false)
	out := make([]string, 0, len(cards))
	for _, card := range cards {
		out = append(out, card.InstanceID)
	}
	return out
}

func firstRevealedBunkerIndex(t *testing.T, g *gameSession) int {
	t.Helper()
	for i, card := range g.World.Bunker {
		if card.IsRevealed {
			return i
		}
	}
	t.Fatalf("no revealed bunker cards available")
	return -1
}

func assertSpecialConsumedForManualFlow(t *testing.T, g *gameSession, definition specialDefinition) {
	t.Helper()
	if strings.TrimSpace(definition.Effect.Type) == "stealBaggage_and_giveSpecial" {
		if g.findPlayerSpecialByInstance(g.Players["p1"], specialPerCardInstanceID) != nil {
			t.Fatalf("stealBaggage_and_giveSpecial should remove used special from owner list")
		}
		return
	}
	special := g.findPlayerSpecialByInstance(g.Players["p1"], specialPerCardInstanceID)
	if special == nil {
		t.Fatalf("special disappeared after manual apply for effect=%q", definition.Effect.Type)
	}
	if !special.Used {
		t.Fatalf("special should be marked used for effect=%q", definition.Effect.Type)
	}
}

func secretTriggerTargetPlayerID(t *testing.T, g *gameSession, definition specialDefinition) string {
	t.Helper()
	condition := strings.TrimSpace(asString(definition.Effect.Params["condition"]))
	switch condition {
	case "leftNeighborEliminated":
		return "p4"
	case "rightNeighborEliminated":
		return "p2"
	case "youngestByRevealedAgeEliminated":
		setBiologyAgeForPlayer(t, g, "p1", 30)
		setBiologyAgeForPlayer(t, g, "p2", 20)
		setBiologyAgeForPlayer(t, g, "p3", 40)
		setBiologyAgeForPlayer(t, g, "p4", 50)
		return "p2"
	case "oldestByRevealedAgeEliminated":
		setBiologyAgeForPlayer(t, g, "p1", 30)
		setBiologyAgeForPlayer(t, g, "p2", 20)
		setBiologyAgeForPlayer(t, g, "p3", 40)
		setBiologyAgeForPlayer(t, g, "p4", 50)
		return "p4"
	case "firstRevealedHealthEliminated":
		g.FirstHealthRevealerID = "p3"
		return "p3"
	default:
		t.Fatalf("unsupported secret condition in per-card test: %q", condition)
		return ""
	}
}

func setBiologyAgeForPlayer(t *testing.T, g *gameSession, playerID string, age int) {
	t.Helper()
	player := g.Players[playerID]
	if player == nil {
		t.Fatalf("player not found: %s", playerID)
	}
	cards := g.getCardsByCategoryKey(player, "biology", false)
	if len(cards) == 0 {
		t.Fatalf("biology card not found for player: %s", playerID)
	}
	cards[0].Revealed = true
	cards[0].Label = fmt.Sprintf("Age %d", age)
}
