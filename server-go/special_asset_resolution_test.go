package main

import (
	"testing"

	"golang.org/x/text/encoding/charmap"
)

func makeLegacyMojibake(value string) string {
	encoded, err := charmap.Windows1251.NewDecoder().Bytes([]byte(value))
	if err != nil {
		return value
	}
	return string(encoded)
}

func TestResolveSpecialAssetID_RepairsMojibakeSpecialPath(t *testing.T) {
	g := &gameSession{
		DeckPools: map[string][]assetCard{
			"Особые условия": {
				{
					ID:    "decks/Особые условия/Будь другом.jpg",
					Deck:  "Особые условия",
					Label: "Будь другом",
				},
			},
		},
	}
	lookup := g.buildSpecialAssetLookup()
	def := specialDefinition{
		ID:      makeLegacyMojibake("Особые условия/БУДЬ ДРУГОМ.jpg"),
		Title:   makeLegacyMojibake("Будь Другом"),
		AssetID: "decks/" + makeLegacyMojibake("Особые условия/БУДЬ ДРУГОМ.jpg"),
	}

	got := g.resolveSpecialAssetID(def, lookup)
	want := "decks/Особые условия/Будь другом.jpg"
	if got != want {
		t.Fatalf("unexpected asset id: got=%q want=%q", got, want)
	}
}

func TestResolveSpecialAssetID_FallsBackToNormalizedAssetID(t *testing.T) {
	g := &gameSession{}
	def := specialDefinition{
		AssetID: "/assets/decks/Особые условия/Громкий Голос.jpg",
	}

	got := g.resolveSpecialAssetID(def, nil)
	want := "decks/Особые условия/Громкий Голос.jpg"
	if got != want {
		t.Fatalf("unexpected normalized asset id: got=%q want=%q", got, want)
	}
}
