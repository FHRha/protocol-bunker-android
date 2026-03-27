package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeLocaleDeckCard(t *testing.T, assetsRoot, locale, deckDir, fileName string) {
	t.Helper()
	fullPath := filepath.Join(assetsRoot, "decks", "1x", locale, deckDir, fileName)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(fullPath, []byte("x"), 0o644); err != nil {
		t.Fatalf("write file failed: %v", err)
	}
}

func TestLoadAssetCatalogForLocale_PrefersRequestedLocale(t *testing.T) {
	assetsRoot := t.TempDir()
	writeLocaleDeckCard(t, assetsRoot, "ru", "Bunker", "bunker.ru-card.png")
	writeLocaleDeckCard(t, assetsRoot, "en", "Bunker", "bunker.en-card.png")

	enCatalog, err := loadAssetCatalogForLocale(assetsRoot, "en")
	if err != nil {
		t.Fatalf("load en catalog failed: %v", err)
	}
	enCards := enCatalog.Decks["bunker"]
	if len(enCards) != 1 {
		t.Fatalf("expected 1 en bunker card, got %d", len(enCards))
	}
	if !strings.Contains(enCards[0].ID, "decks/1x/en/Bunker/") {
		t.Fatalf("expected en card path, got %q", enCards[0].ID)
	}

	ruCatalog, err := loadAssetCatalogForLocale(assetsRoot, "ru")
	if err != nil {
		t.Fatalf("load ru catalog failed: %v", err)
	}
	ruCards := ruCatalog.Decks["bunker"]
	if len(ruCards) != 1 {
		t.Fatalf("expected 1 ru bunker card, got %d", len(ruCards))
	}
	if !strings.Contains(ruCards[0].ID, "decks/1x/ru/Bunker/") {
		t.Fatalf("expected ru card path, got %q", ruCards[0].ID)
	}
}

func TestLoadAssetCatalogForLocale_FallsBackWhenRequestedLocaleMissing(t *testing.T) {
	assetsRoot := t.TempDir()
	writeLocaleDeckCard(t, assetsRoot, "ru", "Bunker", "bunker.only-ru.png")

	catalog, err := loadAssetCatalogForLocale(assetsRoot, "en")
	if err != nil {
		t.Fatalf("load fallback catalog failed: %v", err)
	}
	cards := catalog.Decks["bunker"]
	if len(cards) != 1 {
		t.Fatalf("expected 1 bunker card, got %d", len(cards))
	}
	if !strings.Contains(cards[0].ID, "decks/1x/ru/Bunker/") {
		t.Fatalf("expected ru fallback path, got %q", cards[0].ID)
	}
}
