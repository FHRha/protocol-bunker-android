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
	deckDirs, err := resolveDeckDirectories(decksRoot)
	if err != nil {
		return assetCatalog{}, err
	}

	for _, deckDir := range deckDirs {
		files, err := os.ReadDir(deckDir.fullPath)
		if err != nil {
			continue
		}

		deckName := deckDir.deckID
		cards := make([]assetCard, 0, len(files))
		for _, file := range files {
			if file.IsDir() {
				continue
			}
			ext := strings.ToLower(filepath.Ext(file.Name()))
			if _, ok := cardFileExts[ext]; !ok {
				continue
			}
			relative := filepath.ToSlash(filepath.Join(deckDir.relativeRoot, file.Name()))
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

type assetDeckDir struct {
	deckID       string
	fullPath     string
	relativeRoot string
}

func resolveDeckDirectories(decksRoot string) ([]assetDeckDir, error) {
	layout1xRoot := filepath.Join(decksRoot, "1x")
	if info, err := os.Stat(layout1xRoot); err == nil && info.IsDir() {
		return resolveNestedDeckDirectories(decksRoot, layout1xRoot)
	}
	return resolveFlatDeckDirectories(decksRoot)
}

func resolveNestedDeckDirectories(decksRoot, layoutRoot string) ([]assetDeckDir, error) {
	localeEntries, err := os.ReadDir(layoutRoot)
	if err != nil {
		return nil, fmt.Errorf("failed reading nested decks layout: %w", err)
	}

	localeName := ""
	for _, preferred := range []string{"ru", "en"} {
		for _, entry := range localeEntries {
			if entry.IsDir() && strings.EqualFold(entry.Name(), preferred) {
				localeName = entry.Name()
				break
			}
		}
		if localeName != "" {
			break
		}
	}
	if localeName == "" {
		for _, entry := range localeEntries {
			if entry.IsDir() {
				localeName = entry.Name()
				break
			}
		}
	}
	if localeName == "" {
		return nil, fmt.Errorf("no locale directories found in %s", layoutRoot)
	}

	localeRoot := filepath.Join(layoutRoot, localeName)
	deckEntries, err := os.ReadDir(localeRoot)
	if err != nil {
		return nil, fmt.Errorf("failed reading locale deck directories: %w", err)
	}

	dirs := make([]assetDeckDir, 0, len(deckEntries))
	for _, deckEntry := range deckEntries {
		if !deckEntry.IsDir() || strings.EqualFold(deckEntry.Name(), "TEMP") {
			continue
		}
		deckID := resolveDeckIDByLabel(deckEntry.Name())
		if deckID == "" {
			continue
		}
		relativeRoot, err := filepath.Rel(filepath.Dir(decksRoot), filepath.Join(localeRoot, deckEntry.Name()))
		if err != nil {
			return nil, fmt.Errorf("failed building relative asset path for %s: %w", deckEntry.Name(), err)
		}
		dirs = append(dirs, assetDeckDir{
			deckID:       deckID,
			fullPath:     filepath.Join(localeRoot, deckEntry.Name()),
			relativeRoot: relativeRoot,
		})
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].deckID < dirs[j].deckID
	})
	return dirs, nil
}

func resolveFlatDeckDirectories(decksRoot string) ([]assetDeckDir, error) {
	deckEntries, err := os.ReadDir(decksRoot)
	if err != nil {
		return nil, fmt.Errorf("failed reading decks: %w", err)
	}

	dirs := make([]assetDeckDir, 0, len(deckEntries))
	for _, deckEntry := range deckEntries {
		if !deckEntry.IsDir() || strings.EqualFold(deckEntry.Name(), "TEMP") {
			continue
		}
		deckID := resolveDeckIDByLabel(deckEntry.Name())
		if deckID == "" {
			deckID = deckEntry.Name()
		}
		dirs = append(dirs, assetDeckDir{
			deckID:       deckID,
			fullPath:     filepath.Join(decksRoot, deckEntry.Name()),
			relativeRoot: filepath.Join("decks", deckEntry.Name()),
		})
	}

	sort.Slice(dirs, func(i, j int) bool {
		return dirs[i].deckID < dirs[j].deckID
	})
	return dirs, nil
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
			Description: "Dev sandbox for testing mechanics and UI.",
			DevOnly:     true,
		})
	}
	return scenarios
}
