package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

type specialEffectRule struct {
	ChoiceKind  string `json:"choiceKind"`
	TargetScope string `json:"targetScope"`
}

type specialEffectContract struct {
	EffectRules                  map[string]specialEffectRule `json:"effectRules"`
	CategoryCardSelectionEffects []string                     `json:"categoryCardSelectionEffects"`
}

func TestSpecialEffectContract_GoLoaderParity(t *testing.T) {
	contractPath := filepath.Join("..", "shared-contract", "special_effect_contract.json")
	rawContract, err := os.ReadFile(contractPath)
	if err != nil {
		t.Fatalf("failed to read contract file: %v", err)
	}
	var contract specialEffectContract
	if err := json.Unmarshal(rawContract, &contract); err != nil {
		t.Fatalf("failed to parse contract file: %v", err)
	}

	defs, err := loadImplementedSpecialDefinitionsFromFile(filepath.Join("..", "scenarios", "classic", "SPECIAL_CONDITIONS.json"))
	if err != nil {
		t.Fatalf("failed to load specials definitions: %v", err)
	}
	if len(defs) == 0 {
		t.Fatalf("no implemented specials loaded from json")
	}

	requireCategorySelection := make(map[string]bool, len(contract.CategoryCardSelectionEffects))
	for _, effectType := range contract.CategoryCardSelectionEffects {
		requireCategorySelection[strings.TrimSpace(effectType)] = true
	}

	usedEffects := map[string]bool{}
	for _, def := range defs {
		effectType := strings.TrimSpace(def.Effect.Type)
		rule, ok := contract.EffectRules[effectType]
		if !ok {
			t.Fatalf("missing effect contract for effect=%q (special=%q)", effectType, def.ID)
		}
		usedEffects[effectType] = true

		if strings.TrimSpace(def.ChoiceKind) != strings.TrimSpace(rule.ChoiceKind) {
			t.Fatalf("choiceKind mismatch for effect=%q special=%q: got=%q want=%q", effectType, def.ID, def.ChoiceKind, rule.ChoiceKind)
		}
		if strings.TrimSpace(def.TargetScope) != strings.TrimSpace(rule.TargetScope) {
			t.Fatalf("targetScope mismatch for effect=%q special=%q: got=%q want=%q", effectType, def.ID, def.TargetScope, rule.TargetScope)
		}

		if requireCategorySelection[effectType] {
			category := strings.TrimSpace(asString(def.Effect.Params["category"]))
			if category == "" {
				t.Fatalf("effect=%q special=%q requires non-empty params.category", effectType, def.ID)
			}
		}
	}

	unusedRules := make([]string, 0)
	for effectType := range contract.EffectRules {
		if !usedEffects[effectType] {
			unusedRules = append(unusedRules, effectType)
		}
	}
	if len(unusedRules) > 0 {
		sort.Strings(unusedRules)
		t.Fatalf("contract has unused effect rules: %v", unusedRules)
	}
}
