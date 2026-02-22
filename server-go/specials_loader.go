package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
	if choiceKind == "category" || choiceKind == "none" {
		return ""
	}
	targeting := strings.ToLower(strings.TrimSpace(raw))
	if containsAny(targeting, "neighbor", "сосед", "left", "right", "слева", "справа") {
		return "neighbors"
	}
	if containsAny(targeting, "including self", "включая себя", "any_including_self") {
		return "any_including_self"
	}
	if containsAny(targeting, "self", "себя", "сам") {
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
		targetScope := resolveTargetScopeFromTargeting(item.UITargeting, choiceKind)
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
