package main

import "math"

const (
	minClassicPlayers          = 4
	maxClassicPlayers          = 16
	manualMaxRounds            = 64
	manualMaxVotesPerRound     = 9
	manualMinTargetReveals     = 5
	manualMaxTargetReveals     = 7
	manualDefaultTargetReveals = 7
	defaultScenarioID          = scenarioClassic
	defaultScenarioName        = "Classic Bunker"
	defaultScenarioDescription = "Базовый сценарий: раунды, голосование и особые условия."
)

var rulesetTable = map[int]struct {
	VotesPerRound []int
	TotalExiles   int
	BunkerSeats   int
}{
	4:  {VotesPerRound: []int{0, 0, 0, 1, 1}, TotalExiles: 2, BunkerSeats: 2},
	5:  {VotesPerRound: []int{0, 0, 1, 1, 1}, TotalExiles: 3, BunkerSeats: 2},
	6:  {VotesPerRound: []int{0, 0, 1, 1, 1}, TotalExiles: 3, BunkerSeats: 3},
	7:  {VotesPerRound: []int{0, 1, 1, 1, 1}, TotalExiles: 4, BunkerSeats: 3},
	8:  {VotesPerRound: []int{0, 1, 1, 1, 1}, TotalExiles: 4, BunkerSeats: 4},
	9:  {VotesPerRound: []int{0, 1, 1, 1, 2}, TotalExiles: 5, BunkerSeats: 4},
	10: {VotesPerRound: []int{0, 1, 1, 1, 2}, TotalExiles: 5, BunkerSeats: 5},
	11: {VotesPerRound: []int{0, 1, 1, 2, 2}, TotalExiles: 6, BunkerSeats: 5},
	12: {VotesPerRound: []int{0, 1, 1, 2, 2}, TotalExiles: 6, BunkerSeats: 6},
	13: {VotesPerRound: []int{0, 1, 2, 2, 2}, TotalExiles: 7, BunkerSeats: 6},
	14: {VotesPerRound: []int{0, 1, 2, 2, 2}, TotalExiles: 7, BunkerSeats: 7},
	15: {VotesPerRound: []int{0, 2, 2, 2, 2}, TotalExiles: 8, BunkerSeats: 7},
	16: {VotesPerRound: []int{0, 2, 2, 2, 2}, TotalExiles: 8, BunkerSeats: 8},
}

func defaultSettings() gameSettings {
	return gameSettings{
		EnableRevealDiscussionTimer: false,
		RevealDiscussionSeconds:     60,
		EnablePreVoteDiscussionTime: false,
		PreVoteDiscussionSeconds:    60,
		EnablePostVoteDiscussion:    false,
		PostVoteDiscussionSeconds:   45,
		EnablePresenterMode:         false,
		ContinuePermission:          "revealer_only",
		RevealTimeoutAction:         "random_card",
		RevealsBeforeVoting:         2,
		SpecialUsage:                "anytime",
		MaxPlayers:                  12,
		FinalThreatReveal:           "host",
	}
}

func defaultSettingsForScenario(scenarioID string) gameSettings {
	settings := defaultSettings()
	if scenarioID == scenarioDevTest {
		settings.ContinuePermission = "host_only"
	}
	return settings
}

func defaultScenarioClassic() scenarioMeta {
	return scenarioMeta{
		ID:          defaultScenarioID,
		Name:        defaultScenarioName,
		Description: defaultScenarioDescription,
	}
}

func clampInt(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

func safeRoundInt(raw float64, fallback int) int {
	if math.IsNaN(raw) || math.IsInf(raw, 0) {
		return fallback
	}
	return int(math.Round(raw))
}

func getRulesetForPlayerCount(count int) gameRuleset {
	clamped := clampInt(count, minClassicPlayers, maxClassicPlayers)
	entry, ok := rulesetTable[clamped]
	if !ok {
		entry = rulesetTable[minClassicPlayers]
		clamped = minClassicPlayers
	}

	votes := make([]int, 0, len(entry.VotesPerRound))
	votes = append(votes, entry.VotesPerRound...)
	return gameRuleset{
		PlayerCount: clamped,
		VotesPerRnd: votes,
		TotalExiles: entry.TotalExiles,
		BunkerSeats: entry.BunkerSeats,
		RulesetMode: "preset",
	}
}

func buildAutoRuleset(playerCount int) gameRuleset {
	ruleset := getRulesetForPlayerCount(playerCount)
	ruleset.RulesetMode = "auto"
	ruleset.ManualCfg = nil
	return ruleset
}

func normalizeVotesByRound(votes []int) []int {
	out := make([]int, 0, len(votes))
	for i, vote := range votes {
		if i >= manualMaxRounds {
			break
		}
		out = append(out, clampInt(vote, 0, manualMaxVotesPerRound))
	}
	if len(out) == 0 {
		out = append(out, 0)
	}
	return out
}

func seedManualConfigFromPreset(presetCount int) manualRulesConfig {
	preset := getRulesetForPlayerCount(presetCount)
	template := clampInt(presetCount, minClassicPlayers, maxClassicPlayers)
	return manualRulesConfig{
		BunkerSlots:        clampInt(preset.BunkerSeats, 1, maxClassicPlayers),
		VotesByRound:       normalizeVotesByRound(preset.VotesPerRnd),
		TargetReveals:      manualDefaultTargetReveals,
		SeedTemplatePlayer: &template,
	}
}

func normalizeManualConfig(input manualRulesConfig, fallbackPreset int) manualRulesConfig {
	template := fallbackPreset
	if input.SeedTemplatePlayer != nil {
		template = *input.SeedTemplatePlayer
	}
	template = clampInt(template, minClassicPlayers, maxClassicPlayers)

	votes := normalizeVotesByRound(input.VotesByRound)
	targetReveals := input.TargetReveals
	if targetReveals == 0 {
		targetReveals = manualDefaultTargetReveals
	}
	targetReveals = clampInt(targetReveals, manualMinTargetReveals, manualMaxTargetReveals)

	return manualRulesConfig{
		BunkerSlots:        clampInt(input.BunkerSlots, 1, maxClassicPlayers),
		VotesByRound:       votes,
		TargetReveals:      targetReveals,
		SeedTemplatePlayer: &template,
	}
}

func requiredVotes(playerCount, bunkerSlots int) int {
	return maxInt(0, clampInt(playerCount, 0, 64)-clampInt(bunkerSlots, 1, maxClassicPlayers))
}

func buildManualRuleset(input manualRulesConfig, playerCount int) gameRuleset {
	normalized := normalizeManualConfig(input, playerCount)
	effectiveCount := clampInt(playerCount, minClassicPlayers, maxClassicPlayers)
	return gameRuleset{
		PlayerCount: effectiveCount,
		VotesPerRnd: append([]int(nil), normalized.VotesByRound...),
		TotalExiles: requiredVotes(effectiveCount, normalized.BunkerSlots),
		BunkerSeats: normalized.BunkerSlots,
		RulesetMode: "manual",
		ManualCfg:   &normalized,
	}
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
