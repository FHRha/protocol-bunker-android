package main

import "testing"

func TestNormalizeIncomingGameAction_LegacyOutcomeActions(t *testing.T) {
	tests := []struct {
		name       string
		actionType string
		payload    map[string]any
		wantAction string
		wantResult string
	}{
		{
			name:       "setOutcome with value survived",
			actionType: "setOutcome",
			payload:    map[string]any{"value": "survived"},
			wantAction: "setBunkerOutcome",
			wantResult: "survived",
		},
		{
			name:       "setBunkerResult with result failed",
			actionType: "setBunkerResult",
			payload:    map[string]any{"result": "failed"},
			wantAction: "setBunkerOutcome",
			wantResult: "failed",
		},
		{
			name:       "setBunkerSurvived alias",
			actionType: "setBunkerSurvived",
			payload:    map[string]any{},
			wantAction: "setBunkerOutcome",
			wantResult: "survived",
		},
		{
			name:       "setBunkerFailed alias",
			actionType: "setBunkerFailed",
			payload:    map[string]any{},
			wantAction: "setBunkerOutcome",
			wantResult: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, payload := normalizeIncomingGameAction(tt.actionType, tt.payload)
			if action != tt.wantAction {
				t.Fatalf("action mismatch: got=%s want=%s", action, tt.wantAction)
			}
			outcome, _ := payload["outcome"].(string)
			if outcome != tt.wantResult {
				t.Fatalf("outcome mismatch: got=%s want=%s", outcome, tt.wantResult)
			}
		})
	}
}
