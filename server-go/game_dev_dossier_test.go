package main

import "testing"

func makeHandAssetCatalog(playerCount int) assetCatalog {
	decks := map[string][]assetCard{}

	for _, deckName := range coreDecks {
		cards := make([]assetCard, 0, playerCount)
		for i := 0; i < playerCount; i++ {
			cards = append(cards, assetCard{
				ID:    "decks/" + deckName + "/card_" + string(rune('a'+i)) + ".jpg",
				Deck:  deckName,
				Label: deckName + " " + string(rune('1'+i)),
			})
		}
		decks[deckName] = cards
	}

	factsCards := make([]assetCard, 0, playerCount*2)
	for i := 0; i < playerCount*2; i++ {
		factsCards = append(factsCards, assetCard{
			ID:    "decks/" + factsDeck + "/fact_" + string(rune('a'+i)) + ".jpg",
			Deck:  factsDeck,
			Label: "Факт " + string(rune('1'+i)),
		})
	}
	decks[factsDeck] = factsCards

	return assetCatalog{Decks: decks}
}

func TestGameView_DevScenario_DossierCategoriesContainCards(t *testing.T) {
	settings := defaultSettingsForScenario(scenarioDevTest)
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
		scenarioDevTest,
		settings,
		ruleset,
		players,
		makeHandAssetCatalog(len(players)),
		nil,
		1234,
	)

	room := &room{
		Players: map[string]*player{
			"p1": {ID: "p1", Connected: true},
			"p2": {ID: "p2", Connected: true},
			"p3": {ID: "p3", Connected: true},
			"p4": {ID: "p4", Connected: true},
		},
	}

	view := session.buildGameView(room, "p1")
	if len(view.You.Categories) != len(categoryOrder)-1 {
		t.Fatalf("unexpected categories count: got=%d want=%d", len(view.You.Categories), len(categoryOrder)-1)
	}
	if len(view.You.Specials) == 0 {
		t.Fatalf("expected at least one special condition in dossier")
	}

	for _, slot := range view.You.Categories {
		if len(slot.Cards) == 0 {
			t.Fatalf("category %q has no cards in dossier", slot.Category)
		}
		for _, card := range slot.Cards {
			if card.InstanceID == "" {
				t.Fatalf("category %q has card with empty instanceId", slot.Category)
			}
		}
	}
}

func TestLocalizeGameViewForLocale_LocalizesDossierSpecial(t *testing.T) {
	view := gameView{}
	view.You.Specials = []specialConditionInstanceView{
		{
			ID:      "dev-choice-p1",
			Title:   "Обмен Карт Хобби (DEV)",
			Text:    "Поменяйся открытыми картами категории «Хобби» с ближайшим игроком справа или слева.",
			ImgURL:  "/assets/Special/special.obmen-kart-hobbi.png",
			Trigger: "active",
			Effect:  specialConditionEffectView{Type: "swapRevealedWithNeighbor"},
		},
	}

	localized := localizeGameViewForLocale(view, "en", scenarioDevTest)
	if got := localized.You.Specials[0].Title; got != "Swap Hobby Cards (DEV)" {
		t.Fatalf("unexpected localized special title: %q", got)
	}
	if got := localized.You.Specials[0].Text; got != "Swap revealed “Hobby” cards with the nearest player on your left or right." {
		t.Fatalf("unexpected localized special text: %q", got)
	}
}

func TestLocalizeGameEventForLocale_FormatsDevChoiceNotification(t *testing.T) {
	event := gameEvent{
		Kind:       "info",
		Message:    "{{name}} selected for the DEV card: {{title}}.",
		MessageKey: "event.devChoice.selected",
		MessageVars: map[string]any{
			"name":    "Alex",
			"title":   "Обмен Карт Хобби",
			"assetId": "Special/special.obmen-kart-hobbi.png",
		},
	}

	localized := localizeGameEventForLocale(event, "en", scenarioDevTest)
	if got := localized.Message; got != "Alex selected for the DEV card: Swap Hobby Cards." {
		t.Fatalf("unexpected localized event message: %q", got)
	}
	if got := localized.MessageVars["title"]; got != "Swap Hobby Cards" {
		t.Fatalf("unexpected localized event title var: %q", got)
	}
}
