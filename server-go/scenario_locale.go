package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

var (
	scenarioLocaleOnce sync.Once
	scenarioLocaleData map[string]map[string]string
)

func scenarioLocaleRootCandidates() []string {
	wd, _ := os.Getwd()
	return []string{
		filepath.Join(wd, "locales", "scenario"),
		filepath.Join(wd, "..", "locales", "scenario"),
		filepath.Join(wd, "..", "..", "locales", "scenario"),
	}
}

func loadScenarioLocaleData() {
	scenarioLocaleData = map[string]map[string]string{}
	for _, scenarioID := range []string{scenarioClassic, scenarioDevTest} {
		scenarioLocaleData[scenarioID] = map[string]string{}
		for _, root := range scenarioLocaleRootCandidates() {
			filePath := filepath.Join(root, scenarioID, "ru.json")
			raw, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			parsed := map[string]string{}
			if err := json.Unmarshal(raw, &parsed); err != nil {
				continue
			}
			for key, value := range parsed {
				scenarioLocaleData[scenarioID][key] = value
			}
			break
		}
	}
}

func scenarioText(scenarioID, key, fallback string) string {
	scenarioLocaleOnce.Do(loadScenarioLocaleData)
	if dict := scenarioLocaleData[scenarioID]; dict != nil {
		if value := dict[key]; value != "" {
			return value
		}
	}
	if dict := scenarioLocaleData[scenarioClassic]; dict != nil {
		if value := dict[key]; value != "" {
			return value
		}
	}
	return fallback
}

func scenarioDeckLabel(scenarioID, deckID string) string {
	switch deckID {
	case "profession":
		return scenarioText(scenarioID, "deck.profession", deckID)
	case "health":
		return scenarioText(scenarioID, "deck.health", deckID)
	case "hobby":
		return scenarioText(scenarioID, "deck.hobby", deckID)
	case "baggage":
		return scenarioText(scenarioID, "deck.baggage", deckID)
	case "fact":
		return scenarioText(scenarioID, "deck.fact", deckID)
	case "biology":
		return scenarioText(scenarioID, "deck.biology", deckID)
	case "special":
		return scenarioText(scenarioID, "deck.special", deckID)
	case "bunker":
		return scenarioText(scenarioID, "deck.bunker", deckID)
	default:
		return deckID
	}
}

func scenarioCategoryLabel(scenarioID, categoryID string) string {
	switch categoryID {
	case "facts1":
		return scenarioText(scenarioID, "category.fact1", categoryID)
	case "facts2":
		return scenarioText(scenarioID, "category.fact2", categoryID)
	case "special":
		return scenarioText(scenarioID, "deck.special", categoryID)
	default:
		return scenarioDeckLabel(scenarioID, categoryID)
	}
}

func scenarioCardNoCardLabel(scenarioID string) string {
	return scenarioText(scenarioID, "card.noCard", "No card")
}

func scenarioSpecialNoneTitle(scenarioID string) string {
	return scenarioText(scenarioID, "special.none.title", "No available special condition")
}

func scenarioSpecialNoneText(scenarioID string) string {
	return scenarioText(scenarioID, "special.none.text", "Special deck is empty.")
}

func scenarioDevChoiceTitle(scenarioID string) string {
	return scenarioText(scenarioID, "dev.choice.title", "Test special condition")
}

func scenarioDevChoiceText(scenarioID string) string {
	return scenarioText(scenarioID, "dev.choice.text", "Choose any special condition for testing.")
}

func scenarioBotPrefix(scenarioID string) string {
	return scenarioText(scenarioID, "bot.prefix", "DEV")
}

func scenarioUnknownPlayerLabel(scenarioID string) string {
	return scenarioText(scenarioID, "fallback.unknownPlayer", "player")
}

func normalizeLocaleAlias(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	lower = strings.ReplaceAll(lower, "\u0451", "\u0435")
	replacer := strings.NewReplacer("-", " ", "_", " ", "/", " ", "\\", " ")
	lower = replacer.Replace(lower)
	return strings.Join(strings.Fields(lower), " ")
}

func resolveDeckIDByLabel(value string) string {
	normalized := normalizeLocaleAlias(value)
	switch normalized {
	case "profession", "professions", "профессия", "профа":
		return "profession"
	case "health", "здоровье", "hp":
		return "health"
	case "hobby", "hobbies", "хобби":
		return "hobby"
	case "baggage", "bag", "багаж":
		return "baggage"
	case "fact", "facts", "факты":
		return "fact"
	case "biology", "bio", "биология":
		return "biology"
	case "special", "specials", "special conditions", "особые условия":
		return "special"
	case "bunker", "бункер":
		return "bunker"
	case "disaster", "disasters", "катастрофа", "катастрофы":
		return "disaster"
	case "threat", "threats", "угроза", "угрозы":
		return "threat"
	case "back", "backs", "рубашки":
		return "back"
	default:
		return ""
	}
}

func resolveAssetDeckID(card assetCard) string {
	if byDeck := resolveDeckIDByLabel(card.Deck); byDeck != "" {
		return byDeck
	}
	raw := strings.TrimSpace(card.ID)
	if raw == "" {
		return ""
	}
	parts := strings.Split(strings.ReplaceAll(raw, "\\", "/"), "/")
	if len(parts) < 2 {
		return ""
	}
	return resolveDeckIDByLabel(parts[len(parts)-2])
}
