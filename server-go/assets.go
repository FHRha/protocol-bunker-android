package main

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

var cardFileExts = map[string]struct{}{
	".png":  {},
	".jpg":  {},
	".jpeg": {},
	".webp": {},
}

func loadAssetCatalog(assetsRoot string) (assetCatalog, error) {
	decksRoot := filepath.Join(assetsRoot, "decks")
	info, err := os.Stat(decksRoot)
	if err != nil {
		return assetCatalog{}, fmt.Errorf("assets decks path error: %w", err)
	}
	if !info.IsDir() {
		return assetCatalog{}, fmt.Errorf("assets decks path is not directory: %s", decksRoot)
	}

	catalog := assetCatalog{Decks: map[string][]assetCard{}}
	deckEntries, err := os.ReadDir(decksRoot)
	if err != nil {
		return assetCatalog{}, fmt.Errorf("failed reading decks: %w", err)
	}

	for _, deckEntry := range deckEntries {
		if !deckEntry.IsDir() {
			continue
		}
		deckName := deckEntry.Name()
		if strings.EqualFold(deckName, "TEMP") {
			continue
		}
		deckDir := filepath.Join(decksRoot, deckName)
		files, err := os.ReadDir(deckDir)
		if err != nil {
			continue
		}

		cards := make([]assetCard, 0, len(files))
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if _, ok := cardFileExts[ext]; !ok {
				continue
			}
			relative := filepath.ToSlash(filepath.Join("decks", deckName, file.Name()))
			label := strings.TrimSpace(strings.TrimSuffix(file.Name(), ext))
			if label == "" {
				label = file.Name()
			}
			cards = append(cards, assetCard{
				ID:    relative,
				Deck:  deckName,
				Label: label,
			})
		}
		sort.Slice(cards, func(i, j int) bool {
			return cards[i].ID < cards[j].ID
		})
		catalog.Decks[deckName] = cards
	}

	return catalog, nil
}

func newRand(seed int64) *rand.Rand {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return rand.New(rand.NewSource(seed))
}

func drawRandomCard(pool []assetCard, rnd *rand.Rand) (assetCard, []assetCard, bool) {
	if len(pool) == 0 {
		return assetCard{}, pool, false
	}
	index := rnd.Intn(len(pool))
	card := pool[index]
	last := len(pool) - 1
	pool[index] = pool[last]
	pool[last] = assetCard{}
	pool = pool[:last]
	return card, pool, true
}

func availableScenarios(enableDev bool) []scenarioMeta {
	scenarios := []scenarioMeta{defaultScenarioClassic()}
	if enableDev {
		scenarios = append(scenarios, scenarioMeta{
			ID:          scenarioDevTest,
			Name:        "Dev Test Scenario",
			Description: "Dev-песочница для проверки механик и UI.",
			DevOnly:     true,
		})
	}
	return scenarios
}
