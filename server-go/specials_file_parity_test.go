package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImplementedSpecialsFromJsonResolveToExistingAssets(t *testing.T) {
	defs, err := loadImplementedSpecialDefinitionsFromFile(filepath.Join("..", "scenarios", "classic", "SPECIAL_CONDITIONS.json"))
	if err != nil {
		t.Fatalf("failed to load specials file: %v", err)
	}

	catalog, err := loadAssetCatalog(filepath.Join("..", "assets"))
	if err != nil {
		t.Fatalf("failed to load asset catalog: %v", err)
	}

	session := &gameSession{
		DeckPools: make(map[string][]assetCard, len(catalog.Decks)),
	}
	for deckName, cards := range catalog.Decks {
		copied := make([]assetCard, 0, len(cards))
		copied = append(copied, cards...)
		session.DeckPools[deckName] = copied
	}

	lookup := session.buildSpecialAssetLookup()
	assetsRoot := filepath.Join("..", "assets")

	missing := []string{}
	for _, def := range defs {
		resolved := session.resolveSpecialAssetID(def, lookup)
		if resolved == "" {
			missing = append(missing, def.Title+" (empty asset)")
			continue
		}
		rel := strings.TrimPrefix(strings.ReplaceAll(resolved, "\\", "/"), "decks/")
		fullPath := filepath.Join(assetsRoot, "decks", filepath.FromSlash(rel))
		if _, statErr := os.Stat(fullPath); statErr != nil {
			missing = append(missing, def.Title+" -> "+resolved)
		}
	}

	if len(missing) > 0 {
		t.Fatalf("some implemented specials have no matching asset: %v", missing)
	}
}
