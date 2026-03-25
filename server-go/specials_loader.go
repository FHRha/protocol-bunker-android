package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type specialConditionLocaleEntry struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}

var (
	specialConditionLocaleMu    sync.RWMutex
	specialConditionLocaleCache = map[string]map[string]map[string]specialConditionLocaleEntry{}
)

func specialConditionLocaleRootCandidates(scenarioID string) []string {
	wd, _ := os.Getwd()
	normalizedScenario := strings.TrimSpace(scenarioID)
	if normalizedScenario == "" {
		normalizedScenario = scenarioClassic
	}
	return []string{
		filepath.Join(wd, "locales", "special_conditions", normalizedScenario),
		filepath.Join(wd, "..", "locales", "special_conditions", normalizedScenario),
		filepath.Join(wd, "..", "..", "locales", "special_conditions", normalizedScenario),
		filepath.Join(wd, "locales", "special_conditions", scenarioClassic),
		filepath.Join(wd, "..", "locales", "special_conditions", scenarioClassic),
		filepath.Join(wd, "..", "..", "locales", "special_conditions", scenarioClassic),
	}
}

func loadSpecialConditionLocaleData(scenarioID string) map[string]map[string]specialConditionLocaleEntry {
	data := map[string]map[string]specialConditionLocaleEntry{}
	for _, locale := range []string{"ru", "en"} {
		dict := map[string]specialConditionLocaleEntry{}
		for _, root := range specialConditionLocaleRootCandidates(scenarioID) {
			filePath := filepath.Join(root, locale+".json")
			raw, err := os.ReadFile(filePath)
			if err != nil {
				continue
			}
			parsed := map[string]specialConditionLocaleEntry{}
			if err := json.Unmarshal(raw, &parsed); err != nil {
				continue
			}
			for key, value := range parsed {
				if _, exists := dict[key]; !exists {
					dict[key] = value
				}
			}
			break
		}
		data[locale] = dict
	}
	return data
}

func normalizeSpecialConditionLocaleKey(value string) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return ""
	}
	raw = strings.ReplaceAll(raw, "\\", "/")
	raw = strings.TrimPrefix(raw, "/")
	raw = strings.TrimPrefix(raw, "assets/")
	raw = strings.TrimPrefix(raw, "decks/")
	if strings.HasPrefix(raw, "1x/") {
		parts := strings.SplitN(raw, "/", 3)
		if len(parts) == 3 && (parts[1] == "ru" || parts[1] == "en") {
			raw = parts[2]
		}
	}
	raw = strings.TrimPrefix(raw, "./")
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp"} {
		if strings.HasSuffix(strings.ToLower(raw), ext) {
			raw = raw[:len(raw)-len(ext)]
			break
		}
	}
	return raw
}

func specialConditionLocaleEntryFor(scenarioID, locale, key string) (specialConditionLocaleEntry, bool) {
	normalizedScenario := strings.TrimSpace(scenarioID)
	if normalizedScenario == "" {
		normalizedScenario = scenarioClassic
	}
	normalizedLocale := normalizeCardLocale(locale)
	specialConditionLocaleMu.RLock()
	scenarioData := specialConditionLocaleCache[normalizedScenario]
	specialConditionLocaleMu.RUnlock()
	if scenarioData == nil {
		specialConditionLocaleMu.Lock()
		scenarioData = specialConditionLocaleCache[normalizedScenario]
		if scenarioData == nil {
			scenarioData = loadSpecialConditionLocaleData(normalizedScenario)
			specialConditionLocaleCache[normalizedScenario] = scenarioData
		}
		specialConditionLocaleMu.Unlock()
	}
	if dict := scenarioData[normalizedLocale]; dict != nil {
		if entry, ok := dict[key]; ok {
			return entry, true
		}
	}
	if normalizedLocale != "ru" {
		if dict := scenarioData["ru"]; dict != nil {
			if entry, ok := dict[key]; ok {
				return entry, true
			}
		}
	}
	return specialConditionLocaleEntry{}, false
}

func localizeSpecialDefinitions(definitions []specialDefinition, scenarioID, locale string) []specialDefinition {
	out := cloneSpecialDefinitions(definitions)
	for i := range out {
		keys := []string{
			normalizeSpecialConditionLocaleKey(out[i].AssetID),
			normalizeSpecialConditionLocaleKey(out[i].ID),
			normalizeSpecialConditionLocaleKey(out[i].Title),
		}
		for _, key := range keys {
			if key == "" {
				continue
			}
			if entry, ok := specialConditionLocaleEntryFor(scenarioID, locale, key); ok {
				if entry.Title != "" {
					out[i].Title = entry.Title
				}
				if entry.Text != "" {
					out[i].Text = entry.Text
				}
				break
			}
		}
	}
	return out
}

type specialDefinitionRaw struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Text        string   `json:"text"`
	File        string   `json:"file"`
	Trigger     string   `json:"trigger"`
	Implemented bool     `json:"implemented"`
	Requires    []string `json:"requires"`
	UITargeting string   `json:"uiTargeting"`
	Effect      struct {
		Type   string         `json:"type"`
		Params map[string]any `json:"params"`
	} `json:"effect"`
}

func cloneSpecialDefinitions(definitions []specialDefinition) []specialDefinition {
	out := make([]specialDefinition, 0, len(definitions))
	for _, def := range definitions {
		copyDef := def
		if def.Effect.Params != nil {
			params := make(map[string]any, len(def.Effect.Params))
			for k, v := range def.Effect.Params {
				params[k] = v
			}
			copyDef.Effect.Params = params
		}
		if def.Requires != nil {
			copyDef.Requires = append([]string(nil), def.Requires...)
		}
		out = append(out, copyDef)
	}
	return out
}

func containsAny(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(haystack, needle) {
			return true
		}
	}
	return false
}

func resolveChoiceKindFromTargeting(raw string) string {
	targeting := strings.ToLower(strings.TrimSpace(raw))
	if targeting == "" {
		return "none"
	}
	if containsAny(targeting, "choose special", "special", "особ") {
		return "special"
	}
	if containsAny(targeting, "bunker", "бункер") {
		return "bunker"
	}
	if containsAny(targeting, "neighbor", "сосед", "left", "right", "слева", "справа") {
		return "neighbor"
	}
	if containsAny(targeting, "category", "категор") {
		return "category"
	}
	if containsAny(targeting, "player", "игрок", "target", "цель") {
		return "player"
	}
	return "none"
}

func resolveTargetScopeFromTargeting(raw string, choiceKind string) string {
	if choiceKind == "category" || choiceKind == "none" || choiceKind == "special" || choiceKind == "bunker" {
		return ""
	}
	targeting := strings.ToLower(strings.TrimSpace(raw))
	if containsAny(targeting, "not self", "не себя", "кроме себя", "not-self") {
		return "any_alive"
	}
	if containsAny(targeting, "neighbor", "сосед", "left", "right", "слева", "справа") {
		return "neighbors"
	}
	if containsAny(targeting, "including self", "включая себя", "any_including_self") {
		return "any_including_self"
	}
	if containsAny(targeting, "only self", "self only", "only yourself", "только себя", "только себе", "у себя") {
		return "self"
	}
	return "any_alive"
}

func resolveSpecialsFile(cfg config) string {
	if trimmed := strings.TrimSpace(cfg.SpecialsFile); trimmed != "" {
		return trimmed
	}
	if trimmed := strings.TrimSpace(cfg.ScenariosSourceRoot); trimmed != "" {
		return filepath.Join(trimmed, "classic", "SPECIAL_CONDITIONS.json")
	}
	return ""
}

func loadImplementedSpecialDefinitionsFromFile(filePath string) ([]specialDefinition, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("specials file path is empty")
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed reading specials file: %w", err)
	}

	var raw []specialDefinitionRaw
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, fmt.Errorf("failed parsing specials json: %w", err)
	}

	out := make([]specialDefinition, 0, len(raw))
	for _, item := range raw {
		if !item.Implemented {
			continue
		}
		fileName := strings.TrimSpace(item.File)
		fileName = sanitizeHumanText(fileName, fileName)
		fileName = strings.ReplaceAll(fileName, "\\", "/")
		fileName = strings.TrimPrefix(fileName, "/")

		id := sanitizeHumanText(strings.TrimSpace(item.ID), strings.TrimSpace(item.ID))
		if id == "" {
			id = strings.TrimSpace(fileName)
		}
		if id == "" {
			id = sanitizeHumanText(strings.TrimSpace(item.Title), strings.TrimSpace(item.Title))
		}
		if id == "" {
			continue
		}
		choiceKind := resolveChoiceKindFromTargeting(item.UITargeting)
		effectLower := strings.ToLower(strings.TrimSpace(item.Effect.Type))
		if choiceKind == "none" {
			switch effectLower {
			case "banvoteagainst",
				"disablevote",
				"doublevotesagainst_and_disableselfvote",
				"replacerevealedcard",
				"discardrevealedanddealhidden",
				"stealbaggage_and_givespecial":
				choiceKind = "player"
			case "swaprevealedwithneighbor":
				choiceKind = "neighbor"
			case "forcerevealcategoryforall":
				choiceKind = "category"
			case "devchoosespecial":
				choiceKind = "special"
			default:
				if strings.Contains(effectLower, "bunker") {
					choiceKind = "bunker"
				}
			}
		}
		targetScope := resolveTargetScopeFromTargeting(item.UITargeting, choiceKind)
		if targetScope == "" {
			switch effectLower {
			case "swaprevealedwithneighbor":
				targetScope = "neighbors"
			case "banvoteagainst",
				"disablevote",
				"doublevotesagainst_and_disableselfvote",
				"replacerevealedcard",
				"discardrevealedanddealhidden",
				"stealbaggage_and_givespecial":
				targetScope = "any_alive"
			}
		}
		assetID := normalizeAssetIDPath(fileName)
		effectType := strings.TrimSpace(item.Effect.Type)
		if effectType == "" {
			effectType = "none"
		}
		params := item.Effect.Params
		if params == nil {
			params = map[string]any{}
		}
		trigger := strings.TrimSpace(item.Trigger)
		if trigger == "" {
			trigger = "active"
		}

		out = append(out, specialDefinition{
			ID:          id,
			Title:       sanitizeHumanText(strings.TrimSpace(item.Title), id),
			Text:        sanitizeHumanText(strings.TrimSpace(item.Text), ""),
			Trigger:     trigger,
			Effect:      specialEffect{Type: effectType, Params: params},
			Implemented: true,
			Requires:    append([]string(nil), item.Requires...),
			ChoiceKind:  choiceKind,
			TargetScope: targetScope,
			AllowSelf:   targetScope == "self" || targetScope == "any_including_self",
			AssetID:     assetID,
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("specials file has no implemented definitions")
	}
	return out, nil
}
