package main

import "testing"

func TestGameSession_CurrentThreatModifier_ByBunkerTitle(t *testing.T) {
	g := newTestGameSession()
	g.World.Counts.Threats = 6
	g.World.Threats = make([]worldFacedCardView, 7)
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "decks/Бункер/Вместе на 10 лет.jpg",
			Title:      "Вместе на 10 лет",
			IsRevealed: true,
		},
	}

	modifier := g.currentThreatModifier()
	if modifier.Delta != 1 {
		t.Fatalf("expected delta=1, got %d", modifier.Delta)
	}
	if modifier.FinalCount != 7 {
		t.Fatalf("expected finalCount=7, got %d", modifier.FinalCount)
	}
}

func TestGameSession_CurrentThreatModifier_ByBunkerIDFallback(t *testing.T) {
	g := newTestGameSession()
	g.World.Counts.Threats = 6
	g.World.Threats = make([]worldFacedCardView, 7)
	g.World.Bunker = []worldFacedCardView{
		{
			Kind:       "bunker",
			ID:         "decks/Бункер/Вместе на 10 лет.jpg",
			Title:      "Unknown bunker card",
			IsRevealed: true,
		},
	}

	modifier := g.currentThreatModifier()
	if modifier.Delta != 1 {
		t.Fatalf("expected delta=1 by id fallback, got %d", modifier.Delta)
	}
	if modifier.FinalCount != 7 {
		t.Fatalf("expected finalCount=7 by id fallback, got %d", modifier.FinalCount)
	}
}
