package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type worldCountRow struct {
	Min     int
	Max     int
	Bunker  int
	Threats int
}

var worldCountTable = []worldCountRow{
	{Min: 4, Max: 4, Bunker: 5, Threats: 3},
	{Min: 5, Max: 6, Bunker: 5, Threats: 4},
	{Min: 7, Max: 9, Bunker: 5, Threats: 5},
	{Min: 10, Max: 16, Bunker: 5, Threats: 6},
}

func worldCounts(playerCount int) (int, int) {
	for _, row := range worldCountTable {
		if playerCount >= row.Min && playerCount <= row.Max {
			return row.Bunker, row.Threats
		}
	}
	last := worldCountTable[len(worldCountTable)-1]
	return last.Bunker, last.Threats
}

func normalizeDeckKey(raw string) string {
	key := strings.TrimSpace(strings.ToLower(raw))
	key = strings.ReplaceAll(key, "ё", "е")
	key = strings.NewReplacer(" ", "", "_", "", "-", "").Replace(key)
	return key
}

func findDeckByKeyword(decks map[string][]assetCard, keyword string) string {
	keyword = normalizeDeckKey(keyword)
	for deckName := range decks {
		if strings.Contains(normalizeDeckKey(deckName), keyword) {
			return deckName
		}
	}
	return ""
}

func drawWorldMany(
	pools map[string][]assetCard,
	deckName string,
	kind string,
	count int,
	rnd *rand.Rand,
) []worldFacedCardView {
	out := make([]worldFacedCardView, 0, count)
	for i := 0; i < count; i++ {
		card, next, ok := drawRandomCard(pools[deckName], rnd)
		pools[deckName] = next
		if !ok {
			out = append(out, worldFacedCardView{
				Kind:        kind,
				ID:          fmt.Sprintf("%s_fallback_%d", kind, i+1),
				Title:       fmt.Sprintf("%s #%d", kind, i+1),
				Description: "",
				IsRevealed:  false,
			})
			continue
		}
		out = append(out, worldFacedCardView{
			Kind:        kind,
			ID:          card.ID,
			Title:       card.Label,
			Description: card.Label,
			ImageID:     card.ID,
			IsRevealed:  false,
		})
	}
	return out
}

func drawWorldDisaster(
	pools map[string][]assetCard,
	deckName string,
	rnd *rand.Rand,
) worldCardView {
	card, next, ok := drawRandomCard(pools[deckName], rnd)
	pools[deckName] = next
	if !ok {
		return worldCardView{
			Kind:        "disaster",
			ID:          "disaster_fallback",
			Title:       "Катастрофа",
			Description: "Катастрофа",
		}
	}
	return worldCardView{
		Kind:        "disaster",
		ID:          card.ID,
		Title:       card.Label,
		Description: card.Label,
		ImageID:     card.ID,
	}
}

func rollWorldFromPools(pools map[string][]assetCard, rnd *rand.Rand, playerCount int) worldStateView {
	bunkerCount, threatCount := worldCounts(playerCount)

	bunkerDeck := findDeckByKeyword(pools, "бункер")
	disasterDeck := findDeckByKeyword(pools, "катастроф")
	threatDeck := findDeckByKeyword(pools, "угроз")

	world := worldStateView{}
	world.Counts.Bunker = bunkerCount
	world.Counts.Threats = threatCount

	world.Disaster = drawWorldDisaster(pools, disasterDeck, rnd)
	world.Bunker = drawWorldMany(pools, bunkerDeck, "bunker", bunkerCount, rnd)
	world.Threats = drawWorldMany(pools, threatDeck, "threat", threatCount+1, rnd)
	return world
}

func maskWorldCard(card worldFacedCardView) worldFacedCardView {
	if card.IsRevealed {
		return card
	}
	masked := card
	masked.Title = ""
	masked.Description = ""
	masked.Text = ""
	masked.ImageID = ""
	masked.IsRevealed = false
	masked.RevealedAtRound = nil
	masked.RevealedBy = ""
	return masked
}

func (g *gameSession) buildWorldView() *worldStateView {
	world := worldStateView{
		Disaster: g.World.Disaster,
		Bunker:   make([]worldFacedCardView, 0, len(g.World.Bunker)),
		Threats:  make([]worldFacedCardView, 0, len(g.World.Threats)),
	}
	world.Counts.Bunker = g.World.Counts.Bunker
	world.Counts.Threats = g.World.Counts.Threats

	for _, card := range g.World.Bunker {
		world.Bunker = append(world.Bunker, maskWorldCard(card))
	}
	for _, card := range g.World.Threats {
		world.Threats = append(world.Threats, maskWorldCard(card))
	}
	return &world
}

func (g *gameSession) revealNextBunkerCard(round int) {
	for i := range g.World.Bunker {
		if g.World.Bunker[i].IsRevealed {
			continue
		}
		g.World.Bunker[i].IsRevealed = true
		revealedRound := round
		g.World.Bunker[i].RevealedAtRound = &revealedRound
		g.WorldEvent = &worldEventView{
			Type:  "bunker_revealed",
			Index: i,
			Round: round,
		}
		return
	}
}

func (g *gameSession) ensureAllBunkerCardsRevealed() {
	revealRound := maxInt(1, len(g.Ruleset.VotesPerRnd))
	for i := range g.World.Bunker {
		if g.World.Bunker[i].IsRevealed {
			continue
		}
		g.World.Bunker[i].IsRevealed = true
		round := revealRound
		g.World.Bunker[i].RevealedAtRound = &round
	}
}

func normalizeThreatTitle(value string) string {
	upper := strings.ToUpper(strings.TrimSpace(value))
	return strings.Join(strings.Fields(upper), " ")
}

func (g *gameSession) currentThreatModifier() threatModifierView {
	baseCount := g.World.Counts.Threats
	delta := 0
	reasons := make([]string, 0, 2)

	// Parity hook from web scenarios (can be extended if more title rules appear).
	modifierByTitle := map[string]int{
		"ВМЕСТЕ НА 10 ЛЕТ":  1,
		"ЗАГАДОЧНЫЙ ЖУРНАЛ": -1,
	}
	for _, card := range g.World.Bunker {
		if !card.IsRevealed {
			continue
		}
		modifier := modifierByTitle[normalizeThreatTitle(card.Title)]
		if modifier == 0 {
			continue
		}
		delta += modifier
		reasons = append(reasons, card.Title)
	}

	finalCount := baseCount + delta
	if finalCount < 0 {
		finalCount = 0
	}
	if finalCount > len(g.World.Threats) {
		finalCount = len(g.World.Threats)
	}
	return threatModifierView{
		Delta:      delta,
		Reasons:    reasons,
		BaseCount:  baseCount,
		FinalCount: finalCount,
	}
}

func (g *gameSession) revealWorldThreat(actorID string, index int) gameActionResult {
	if g.Phase != scenarioPhaseEnded {
		return gameActionResult{Error: "Угрозы раскрываются в конце игры."}
	}
	modifier := g.currentThreatModifier()
	if index < 0 || index >= modifier.FinalCount {
		return gameActionResult{Error: "Некорректная карта угроз."}
	}
	if g.Settings.FinalThreatReveal == "host" && actorID != g.HostID {
		return gameActionResult{Error: "Только хост может открывать угрозы."}
	}

	target := &g.World.Threats[index]
	if target.IsRevealed {
		return gameActionResult{}
	}
	target.IsRevealed = true
	target.RevealedBy = actorID
	return gameActionResult{StateChanged: true}
}

func (g *gameSession) setBunkerOutcome(actorID, outcome string) gameActionResult {
	if g.Phase != scenarioPhaseEnded || g.PostGame == nil || !g.PostGame.IsActive {
		return gameActionResult{Error: "Игра ещё не завершена."}
	}
	if actorID != g.HostID {
		return gameActionResult{Error: "Только хост может выбрать исход бункера."}
	}
	if g.PostGame.Outcome != "" {
		return gameActionResult{Error: "Исход уже выбран."}
	}
	if outcome != "survived" && outcome != "failed" {
		return gameActionResult{Error: "Некорректный исход."}
	}

	g.PostGame.Outcome = outcome
	g.PostGame.DecidedBy = actorID
	g.PostGame.DecidedAt = time.Now().UnixMilli()
	if outcome == "survived" {
		g.LastStageText = "Финал: бункер выжил."
	} else {
		g.LastStageText = "Финал: бункер не выжил."
	}
	return gameActionResult{
		StateChanged: true,
		Events:       []gameEvent{g.makeEvent("info", g.LastStageText)},
	}
}
