package main

import (
	"strings"
	"testing"
)

func makeImageAssetCatalog(playerCount int) assetCatalog {
	decks := map[string][]assetCard{}

	for _, deckName := range coreDecks {
		cards := make([]assetCard, 0, playerCount)
		for i := 0; i < playerCount; i++ {
			cards = append(cards, assetCard{
				ID:    "decks/1x/ru/" + deckName + "/" + deckName + ".card-" + string(rune('a'+i)) + ".png",
				Deck:  deckName,
				Label: deckName + " " + string(rune('1'+i)),
			})
		}
		decks[deckName] = cards
	}

	factsCards := make([]assetCard, 0, playerCount*2)
	for i := 0; i < playerCount*2; i++ {
		factsCards = append(factsCards, assetCard{
			ID:    "decks/1x/ru/" + factsDeck + "/" + factsDeck + ".card-" + string(rune('a'+i)) + ".png",
			Deck:  factsDeck,
			Label: "fact " + string(rune('1'+i)),
		})
	}
	decks[factsDeck] = factsCards

	decks["bunker"] = []assetCard{
		{ID: "decks/1x/ru/Bunker/bunker.card-a.png", Deck: "bunker", Label: "Bunker A"},
		{ID: "decks/1x/ru/Bunker/bunker.card-b.png", Deck: "bunker", Label: "Bunker B"},
		{ID: "decks/1x/ru/Bunker/bunker.card-c.png", Deck: "bunker", Label: "Bunker C"},
		{ID: "decks/1x/ru/Bunker/bunker.card-d.png", Deck: "bunker", Label: "Bunker D"},
		{ID: "decks/1x/ru/Bunker/bunker.card-e.png", Deck: "bunker", Label: "Bunker E"},
	}
	decks["disaster"] = []assetCard{
		{ID: "decks/1x/ru/Disaster/disaster.card-a.png", Deck: "disaster", Label: "Disaster A"},
	}
	decks["threat"] = []assetCard{
		{ID: "decks/1x/ru/Threat/threat.card-a.png", Deck: "threat", Label: "Threat A"},
		{ID: "decks/1x/ru/Threat/threat.card-b.png", Deck: "threat", Label: "Threat B"},
		{ID: "decks/1x/ru/Threat/threat.card-c.png", Deck: "threat", Label: "Threat C"},
		{ID: "decks/1x/ru/Threat/threat.card-d.png", Deck: "threat", Label: "Threat D"},
		{ID: "decks/1x/ru/Threat/threat.card-e.png", Deck: "threat", Label: "Threat E"},
		{ID: "decks/1x/ru/Threat/threat.card-f.png", Deck: "threat", Label: "Threat F"},
		{ID: "decks/1x/ru/Threat/threat.card-g.png", Deck: "threat", Label: "Threat G"},
	}

	return assetCatalog{Decks: decks}
}

func makeConnectedRoom(playerIDs ...string) *room {
	players := make(map[string]*player, len(playerIDs))
	for _, id := range playerIDs {
		players[id] = &player{ID: id, Connected: true}
	}
	return &room{Players: players}
}

func newImageTestGameSession(scenarioID string) *gameSession {
	settings := defaultSettingsForScenario(scenarioID)
	ruleset := buildAutoRuleset(minClassicPlayers)
	players := []*player{
		{ID: "p1", Name: "Host"},
		{ID: "p2", Name: "P2"},
		{ID: "p3", Name: "P3"},
		{ID: "p4", Name: "P4"},
	}
	return newGameSession(
		"ROOM",
		"p1",
		scenarioID,
		settings,
		ruleset,
		players,
		makeImageAssetCatalog(len(players)),
		nil,
		1234,
	)
}

func TestGameView_YouCategoriesIncludeImgURL(t *testing.T) {
	session := newImageTestGameSession(scenarioDevTest)
	view := session.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")

	if len(view.You.Categories) == 0 {
		t.Fatalf("expected categories in dossier")
	}

	for _, slot := range view.You.Categories {
		if len(slot.Cards) == 0 {
			t.Fatalf("category %q has no cards", slot.Category)
		}
		for _, card := range slot.Cards {
			if card.ImgURL == "" {
				t.Fatalf("category %q card %q is missing imgUrl", slot.Category, card.InstanceID)
			}
		}
	}
}

func TestGameView_PublicRevealedCardsIncludeImgURL(t *testing.T) {
	session := newImageTestGameSession(scenarioClassic)
	player := session.Players["p1"]
	if player == nil || len(player.Hand) == 0 {
		t.Fatalf("missing player hand")
	}
	player.Hand[0].Revealed = true

	view := session.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")
	if len(view.Public.Players) == 0 {
		t.Fatalf("expected public players")
	}

	var revealed *cardRef
	for _, publicPlayer := range view.Public.Players {
		if publicPlayer.PlayerID != "p1" {
			continue
		}
		if len(publicPlayer.RevealedCards) == 0 {
			t.Fatalf("expected revealed cards for p1")
		}
		revealed = &publicPlayer.RevealedCards[0]
		break
	}
	if revealed == nil {
		t.Fatalf("revealed card payload not found")
	}
	if revealed.ImgURL == "" {
		t.Fatalf("revealed card is missing imgUrl")
	}
}

func TestGameView_DevWorldCardsIncludeImgURL(t *testing.T) {
	session := newImageTestGameSession(scenarioDevTest)
	view := session.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")

	if view.World == nil {
		t.Fatalf("expected world payload")
	}
	if view.World.Disaster.ImgURL == "" {
		t.Fatalf("disaster is missing imgUrl")
	}
	if len(view.World.Bunker) == 0 || len(view.World.Threats) == 0 {
		t.Fatalf("expected bunker and threat cards")
	}
	for idx, card := range view.World.Bunker {
		if card.ImgURL == "" {
			t.Fatalf("bunker[%d] is missing imgUrl", idx)
		}
	}
	for idx, card := range view.World.Threats {
		if card.ImgURL == "" {
			t.Fatalf("threat[%d] is missing imgUrl", idx)
		}
	}
}

func TestGameView_AssetLocaleFallsBackToRuWhenOnlyRuDecksExist(t *testing.T) {
	settings := defaultSettingsForScenario(scenarioClassic)
	settings.CardLocale = "en"
	ruleset := buildAutoRuleset(minClassicPlayers)
	players := []*player{
		{ID: "p1", Name: "Host"},
		{ID: "p2", Name: "P2"},
		{ID: "p3", Name: "P3"},
		{ID: "p4", Name: "P4"},
	}
	session := newGameSession(
		"ROOM",
		"p1",
		scenarioClassic,
		settings,
		ruleset,
		players,
		makeImageAssetCatalog(len(players)),
		nil,
		4321,
	)

	view := session.buildGameView(makeConnectedRoom("p1", "p2", "p3", "p4"), "p1")
	if view.World == nil {
		t.Fatalf("expected world payload")
	}
	if got := view.World.Disaster.ImgURL; got == "" || !strings.Contains(got, "/decks/1x/ru/") {
		t.Fatalf("expected ru asset fallback for disaster imgUrl, got %q", got)
	}

	found := false
	for _, slot := range view.You.Categories {
		for _, card := range slot.Cards {
			if card.ImgURL == "" {
				continue
			}
			found = true
			if !strings.Contains(card.ImgURL, "/decks/1x/ru/") {
				t.Fatalf("expected ru asset fallback for category imgUrl, got %q", card.ImgURL)
			}
		}
	}
	if !found {
		t.Fatalf("expected at least one dossier card imgUrl")
	}
}
