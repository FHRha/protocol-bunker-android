package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
)

var (
	votingWindowEffectTypes = map[string]bool{
		"banVoteAgainst":                         true,
		"disableVote":                            true,
		"voteWeight":                             true,
		"forceRevote":                            true,
		"doubleVotesAgainst_and_disableSelfVote": true,
	}

	categoryKeyToDeck = map[string]string{
		"profession": coreDecks[0],
		"health":     healthDeckName(),
		"hobby":      coreDecks[2],
		"baggage":    coreDecks[3],
		"facts":      factsDeck,
		"facts1":     factsDeck,
		"facts2":     factsDeck,
		"biology":    coreDecks[4],
	}

	specialDeckCategoryName  = "special"
	devChoiceEffectType      = "devChooseSpecial"
	specialAssetFallbackByID = map[string]string{}
)

type effectiveVoteRecord struct {
	TargetID   string
	Status     string
	ReasonText string
	Weight     int
	Submitted  int64
}

func normalizeSpecialKey(value string) string {
	lower := strings.ToLower(strings.TrimSpace(value))
	lower = strings.ReplaceAll(lower, "\u0451", "\u0435")
	var b strings.Builder
	b.Grow(len(lower))
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func categorySlotKey(categoryKey string) string {
	key := strings.TrimSpace(categoryKey)
	switch key {
	case "facts1", "facts2":
		return key
	default:
		return ""
	}
}

func canonicalCategoryKey(categoryKey string) string {
	key := strings.TrimSpace(categoryKey)
	if key == "" {
		return ""
	}
	if slot := categorySlotKey(key); slot != "" {
		return slot
	}
	return key
}

func specialImageURL(def specialDefinition) string {
	if def.AssetID == "" {
		return ""
	}
	return "/assets/" + def.AssetID
}

func trimCardFileExt(name string) string {
	lower := strings.ToLower(name)
	for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp"} {
		if strings.HasSuffix(lower, ext) {
			return name[:len(name)-len(ext)]
		}
	}
	return name
}

func normalizeSpecialAssetRef(value string) string {
	raw := strings.TrimSpace(value)
	if raw == "" {
		return ""
	}
	raw = strings.ReplaceAll(raw, "\\", "/")
	raw = strings.TrimPrefix(raw, "/")
	raw = strings.TrimPrefix(raw, "./")
	raw = strings.TrimPrefix(raw, "assets/")
	raw = strings.TrimPrefix(raw, "decks/")
	raw = trimCardFileExt(raw)
	return normalizeSpecialKey(raw)
}

func normalizeAssetIDPath(raw string) string {
	value := strings.TrimSpace(sanitizeHumanText(raw, raw))
	if value == "" {
		return ""
	}
	value = strings.ReplaceAll(value, "\\", "/")
	value = strings.TrimPrefix(value, "/")
	if strings.HasPrefix(strings.ToLower(value), "assets/") {
		value = value[len("assets/"):]
	}
	if !strings.HasPrefix(strings.ToLower(value), "decks/") {
		value = "decks/" + strings.TrimPrefix(value, "/")
	}
	return value
}

func specialLookupCandidates(def specialDefinition) []string {
	base := []string{
		def.AssetID,
		def.Title,
		def.ID,
		sanitizeHumanText(def.AssetID, def.AssetID),
		sanitizeHumanText(def.Title, def.Title),
		sanitizeHumanText(def.ID, def.ID),
	}
	expanded := make([]string, 0, len(base)*2)
	for _, item := range base {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		expanded = append(expanded, item)
		slashed := strings.ReplaceAll(item, "\\", "/")
		leaf := filepath.Base(slashed)
		if leaf != "." && leaf != "" && leaf != slashed {
			expanded = append(expanded, leaf, trimCardFileExt(leaf))
		}
	}
	return expanded
}

func copyVotes(input map[string]voteRecord) map[string]voteRecord {
	out := make(map[string]voteRecord, len(input))
	for k, v := range input {
		out[k] = v
	}
	return out
}

func (g *gameSession) playerName(playerID string) string {
	player := g.Players[playerID]
	if player == nil {
		return scenarioUnknownPlayerLabel(g.Scenario)
	}
	return player.Name
}

func (g *gameSession) nextCardInstanceID(playerID string) string {
	g.CardCounter++
	return fmt.Sprintf("%s-%d", playerID, g.CardCounter)
}

func (g *gameSession) assignInitialSpecials() {
	if len(g.Players) == 0 {
		return
	}
	assetLookup := g.buildSpecialAssetLookup()

	for _, playerID := range g.Order {
		player := g.Players[playerID]
		if player == nil {
			continue
		}
		g.assignInitialSpecialsForPlayer(player, assetLookup)
	}
}

func (g *gameSession) buildSpecialAssetLookup() map[string]string {
	assetLookup := map[string]string{}
	addCandidate := func(key string, assetID string) {
		normalized := normalizeSpecialAssetRef(key)
		if normalized == "" || assetID == "" {
			return
		}
		if _, exists := assetLookup[normalized]; !exists {
			assetLookup[normalized] = assetID
		}
	}

	for deckName, cards := range g.DeckPools {
		for _, card := range cards {
			if card.ID == "" {
				continue
			}
			labelRaw := strings.TrimSpace(card.Label)
			labelSanitized := strings.TrimSpace(sanitizeHumanText(card.Label, card.Label))
			deckRaw := strings.TrimSpace(deckName)
			deckSanitized := strings.TrimSpace(sanitizeHumanText(deckName, deckName))
			baseName := filepath.Base(strings.ReplaceAll(card.ID, "\\", "/"))

			addCandidate(card.ID, card.ID)
			addCandidate(strings.TrimPrefix(card.ID, "decks/"), card.ID)
			addCandidate(labelRaw, card.ID)
			addCandidate(labelSanitized, card.ID)
			addCandidate(baseName, card.ID)
			addCandidate(trimCardFileExt(baseName), card.ID)
			for _, deck := range []string{deckRaw, deckSanitized} {
				if deck == "" {
					continue
				}
				addCandidate(deck+"/"+baseName, card.ID)
				addCandidate(deck+"/"+trimCardFileExt(baseName), card.ID)
			}
		}
	}
	return assetLookup
}

func (g *gameSession) resolveSpecialAssetID(def specialDefinition, assetLookup map[string]string) string {
	if len(assetLookup) > 0 {
		for _, candidate := range specialLookupCandidates(def) {
			key := normalizeSpecialAssetRef(candidate)
			if key == "" {
				continue
			}
			if assetID, ok := assetLookup[key]; ok {
				return assetID
			}
		}
	}
	if fallback, ok := specialAssetFallbackByID[strings.TrimSpace(def.ID)]; ok {
		return normalizeAssetIDPath(fallback)
	}
	return normalizeAssetIDPath(def.AssetID)
}

func (g *gameSession) assignInitialSpecialsForPlayer(player *gamePlayer, assetLookup map[string]string) {
	if player == nil {
		return
	}
	if g.IsDev {
		if devChoice, ok := g.buildDevChoiceDefinition(player.PlayerID); ok {
			player.Specials = []specialConditionState{
				{
					InstanceID:     g.nextSpecialInstanceID(player.PlayerID),
					Definition:     devChoice,
					RevealedPublic: true,
					Used:           false,
				},
			}
			return
		}
	}
	def := g.drawSpecialFromPool()
	rawAssetID := def.AssetID
	def.AssetID = g.resolveSpecialAssetID(def, assetLookup)
	if def.AssetID == "" {
		log.Printf("[specials] unresolved asset id=%q title=%q rawAsset=%q", def.ID, def.Title, rawAssetID)
	}
	player.Specials = []specialConditionState{
		{
			InstanceID:     g.nextSpecialInstanceID(player.PlayerID),
			Definition:     def,
			RevealedPublic: g.IsDev,
			Used:           false,
		},
	}
}

func isDevChoiceSpecialID(definitionID string) bool {
	return strings.HasPrefix(strings.TrimSpace(definitionID), "dev-choice-")
}

func copySpecialDefinition(def specialDefinition) specialDefinition {
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
	return copyDef
}

func (g *gameSession) specialCatalogForDev() []specialDefinition {
	if len(g.specialCatalog) > 0 {
		return g.specialCatalog
	}
	return implementedSpecialDefinitions
}

func (g *gameSession) buildDevSpecialOptions() []map[string]any {
	catalog := g.specialCatalogForDev()
	options := make([]map[string]any, 0, len(catalog))
	seen := map[string]bool{}
	for _, def := range catalog {
		if !def.Implemented {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(def.Effect.Type), devChoiceEffectType) {
			continue
		}
		id := strings.TrimSpace(def.ID)
		if id == "" {
			continue
		}
		key := normalizeSpecialKey(id)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		title := strings.TrimSpace(def.Title)
		if title == "" {
			title = id
		}
		options = append(options, map[string]any{
			"id":      id,
			"title":   title,
			"assetId": def.AssetID,
		})
	}
	return options
}

func (g *gameSession) findSpecialDefinitionForDevChoice(lookupRaw string) (specialDefinition, bool) {
	lookup := normalizeSpecialKey(strings.TrimSpace(lookupRaw))
	if lookup == "" {
		return specialDefinition{}, false
	}
	for _, def := range g.specialCatalogForDev() {
		if !def.Implemented {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(def.Effect.Type), devChoiceEffectType) {
			continue
		}
		candidates := specialLookupCandidates(def)
		for _, candidate := range candidates {
			if normalizeSpecialKey(candidate) == lookup {
				return copySpecialDefinition(def), true
			}
		}
	}
	return specialDefinition{}, false
}

func (g *gameSession) buildDevSpecialFromTemplate(currentDefinitionID string, template specialDefinition) specialDefinition {
	copyDef := copySpecialDefinition(template)
	copyDef.ID = strings.TrimSpace(currentDefinitionID)
	copyDef.Title = strings.TrimSpace(template.Title) + " (DEV)"
	copyDef.Trigger = "active"
	copyDef.Implemented = true
	if copyDef.Effect.Type == "" {
		copyDef.Effect.Type = "none"
	}
	return copyDef
}

func (g *gameSession) buildDevChoiceDefinition(playerID string) (specialDefinition, bool) {
	options := g.buildDevSpecialOptions()
	if len(options) == 0 {
		return specialDefinition{}, false
	}
	return specialDefinition{
		ID:      fmt.Sprintf("dev-choice-%s", playerID),
		Title:   scenarioDevChoiceTitle(g.Scenario),
		Text:    scenarioDevChoiceText(g.Scenario),
		Trigger: "active",
		Effect: specialEffect{
			Type: devChoiceEffectType,
			Params: map[string]any{
				"specialOptions": options,
			},
		},
		Implemented: true,
		ChoiceKind:  "special",
	}, true
}

func (g *gameSession) nextSpecialInstanceID(playerID string) string {
	g.specialCounter++
	return fmt.Sprintf("special-%s-%d", playerID, g.specialCounter)
}

func (g *gameSession) drawSpecialFromPool() specialDefinition {
	if len(g.specialPool) == 0 {
		return specialDefinition{
			ID:          "missing-special",
			Title:       scenarioSpecialNoneTitle(g.Scenario),
			Text:        scenarioSpecialNoneText(g.Scenario),
			Trigger:     "active",
			Implemented: false,
			Effect:      specialEffect{Type: "none"},
			ChoiceKind:  "none",
		}
	}
	index := g.rng.Intn(len(g.specialPool))
	def := g.specialPool[index]
	last := len(g.specialPool) - 1
	g.specialPool[index] = g.specialPool[last]
	g.specialPool[last] = specialDefinition{}
	g.specialPool = g.specialPool[:last]
	return def
}

func (g *gameSession) playerHasRevealedSpecial(player *gamePlayer) bool {
	for _, special := range player.Specials {
		if special.RevealedPublic {
			return true
		}
	}
	return false
}

func (g *gameSession) buildSpecialInstances(player *gamePlayer) []specialConditionInstanceView {
	out := make([]specialConditionInstanceView, 0, len(player.Specials))
	for _, special := range player.Specials {
		choiceKind := effectiveSpecialChoiceKind(special.Definition)
		targetScope := effectiveSpecialTargetScope(special.Definition, choiceKind)
		view := specialConditionInstanceView{
			InstanceID:        special.InstanceID,
			ID:                special.Definition.ID,
			Title:             special.Definition.Title,
			Text:              special.Definition.Text,
			Trigger:           special.Definition.Trigger,
			Implemented:       special.Definition.Implemented,
			RevealedPublic:    special.RevealedPublic,
			Used:              special.Used,
			PendingActivation: special.PendingActivation,
			ImgURL:            specialImageURL(special.Definition),
			Effect: specialConditionEffectView{
				Type:   special.Definition.Effect.Type,
				Params: special.Definition.Effect.Params,
			},
		}
		if choiceKind != "" {
			view.ChoiceKind = choiceKind
			view.NeedsChoice = choiceKind != "none"
		}
		if special.Definition.AllowSelf {
			view.AllowSelfTarget = true
		}
		if targetScope != "" {
			view.TargetScope = targetScope
		}
		out = append(out, view)
	}
	return out
}

func effectiveSpecialChoiceKind(def specialDefinition) string {
	kind := strings.TrimSpace(strings.ToLower(def.ChoiceKind))
	switch kind {
	case "player", "neighbor", "category", "bunker", "special", "none":
		return kind
	}
	effectType := strings.TrimSpace(strings.ToLower(def.Effect.Type))
	switch effectType {
	case "banvoteagainst",
		"disablevote",
		"doublevotesagainst_and_disableselfvote",
		"replacerevealedcard",
		"discardrevealedanddealhidden",
		"stealbaggage_and_givespecial":
		return "player"
	case "swaprevealedwithneighbor":
		return "neighbor"
	case "forcerevealcategoryforall":
		return "category"
	case "devchoosespecial":
		return "special"
	case "replacebunkercard", "discardbunkercard", "stealbunkercardtoexiled":
		return "bunker"
	}
	if strings.Contains(effectType, "bunker") {
		return "bunker"
	}
	targetRaw := strings.TrimSpace(strings.ToLower(asString(def.Effect.Params["target"])))
	if strings.Contains(targetRaw, "bunker") {
		return "bunker"
	}
	return "none"
}

func effectiveSpecialTargetScope(def specialDefinition, choiceKind string) string {
	scope := strings.TrimSpace(strings.ToLower(def.TargetScope))
	switch scope {
	case "self", "neighbors", "any_alive", "any_including_self":
		return scope
	}
	switch strings.TrimSpace(strings.ToLower(def.Effect.Type)) {
	case "swaprevealedwithneighbor":
		return "neighbors"
	case "banvoteagainst",
		"disablevote",
		"doublevotesagainst_and_disableselfvote",
		"replacerevealedcard",
		"discardrevealedanddealhidden",
		"stealbaggage_and_givespecial":
		return "any_alive"
	}
	if choiceKind == "neighbor" {
		return "neighbors"
	}
	if choiceKind == "player" {
		return "any_alive"
	}
	return ""
}

func (g *gameSession) hasRevealedBunkerCards() bool {
	if g.IsDev {
		return len(g.World.Bunker) > 0
	}
	for _, card := range g.World.Bunker {
		if card.IsRevealed {
			return true
		}
	}
	return false
}

func (g *gameSession) findPlayerSpecialByInstance(player *gamePlayer, specialInstanceID string) *specialConditionState {
	if player == nil || specialInstanceID == "" {
		return nil
	}
	for i := range player.Specials {
		if player.Specials[i].InstanceID == specialInstanceID {
			return &player.Specials[i]
		}
	}
	return nil
}

func (g *gameSession) revealedBunkerIndices() []int {
	if g.IsDev {
		indices := make([]int, 0, len(g.World.Bunker))
		for idx := range g.World.Bunker {
			indices = append(indices, idx)
		}
		return indices
	}
	indices := make([]int, 0, len(g.World.Bunker))
	for idx, card := range g.World.Bunker {
		if card.IsRevealed {
			indices = append(indices, idx)
		}
	}
	return indices
}

func (g *gameSession) resolveBunkerIndex(payload map[string]any, allowRandom bool) (int, localizedRuntimeError) {
	raw, hasValue := payload["bunkerIndex"]
	missingSelection := !hasValue || raw == nil
	if !missingSelection {
		if rawText, ok := raw.(string); ok && strings.TrimSpace(rawText) == "" {
			missingSelection = true
		}
	}
	if missingSelection {
		if !allowRandom {
			return -1, localizedRuntimeError{Message: "You need to select a bunker card.", Key: "error.bunker.pickRequired"}
		}
		revealed := g.revealedBunkerIndices()
		if len(revealed) == 0 {
			return -1, localizedRuntimeError{Message: "There are no revealed bunker cards.", Key: "error.bunker.noRevealed"}
		}
		return revealed[g.rng.Intn(len(revealed))], localizedRuntimeError{}
	}
	index := asInt(raw, -1)
	if index < 0 || index >= len(g.World.Bunker) {
		return -1, localizedRuntimeError{Message: "Invalid bunker card.", Key: "error.bunker.invalid"}
	}
	if !g.IsDev && !g.World.Bunker[index].IsRevealed {
		return -1, localizedRuntimeError{Message: "You can only select a revealed bunker card.", Key: "error.bunker.revealedOnly"}
	}
	return index, localizedRuntimeError{}
}

func (g *gameSession) clearActiveTimer() {
	g.ActiveTimer = nil
}

func (g *gameSession) schedulePhaseTimer(kind string, seconds int) {
	if seconds <= 0 {
		return
	}
	g.ActiveTimer = &gameTimerState{
		Kind:   kind,
		EndsAt: time.Now().Add(time.Duration(seconds) * time.Second).UnixMilli(),
	}
}

func (g *gameSession) scheduleRevealTimeoutIfNeeded() {
	if g.Phase != scenarioPhaseReveal || g.CurrentTurnID == "" {
		return
	}
	if !g.Settings.EnableRevealDiscussionTimer {
		return
	}
	g.schedulePhaseTimer(scenarioPhaseRevealDiscussion, g.Settings.RevealDiscussionSeconds)
}

func (g *gameSession) currentTimer() *gameTimerState {
	if g.ActiveTimer == nil {
		return nil
	}
	copy := *g.ActiveTimer
	return &copy
}

func (g *gameSession) handleTimerExpired(now int64) gameActionResult {
	if g.ActiveTimer == nil {
		return gameActionResult{}
	}
	if now < g.ActiveTimer.EndsAt {
		return gameActionResult{}
	}
	kind := g.ActiveTimer.Kind
	g.clearActiveTimer()

	switch kind {
	case scenarioPhaseRevealDiscussion:
		if g.Phase == scenarioPhaseReveal {
			return g.handleRevealPhaseTimeout()
		}
		if g.Phase == scenarioPhaseRevealDiscussion {
			return g.advanceAfterDiscussion()
		}
	case "pre_vote":
		if g.Phase == scenarioPhaseRevealDiscussion {
			return g.advanceAfterDiscussion()
		}
	case "post_vote":
		if g.Phase == scenarioPhaseVoting && g.VotePhase == votePhaseSpecialWindow {
			return g.finalizeVoting(g.HostID)
		}
	case timerKindResolutionAuto:
		if g.Phase == scenarioPhaseResolution {
			if g.shouldEnd() {
				return g.finishGame("auto_after_resolution")
			}
			if g.VotesRemaining > 0 {
				g.startVoting()
				return gameActionResult{
					StateChanged: true,
					Events:       []gameEvent{g.makeEvent("votingStart", g.LastStageText)},
				}
			}
			return g.startNextRoundOrEnd()
		}
	}
	return gameActionResult{}
}

func (g *gameSession) enterRevealDiscussion() gameActionResult {
	g.clearActiveTimer()
	g.Phase = scenarioPhaseRevealDiscussion
	g.VotePhase = ""

	roundComplete := g.allAliveRevealed()
	if roundComplete && g.VotesRemaining > 0 && g.Settings.EnablePreVoteDiscussionTime {
		g.schedulePhaseTimer("pre_vote", g.Settings.PreVoteDiscussionSeconds)
		return gameActionResult{}
	}
	if g.Settings.EnableRevealDiscussionTimer {
		g.schedulePhaseTimer(scenarioPhaseRevealDiscussion, g.Settings.RevealDiscussionSeconds)
	}
	return gameActionResult{}
}

func (g *gameSession) advanceAfterDiscussion() gameActionResult {
	if g.Phase != scenarioPhaseRevealDiscussion {
		return runtimeError("Cannot continue the turn right now.", g.scenarioTextKey("error.continue.notNow", "classic.auto.078"), nil)
	}
	g.clearActiveTimer()

	if !g.allAliveRevealed() {
		nextTurn := g.nextUnrevealedAliveAfter(g.CurrentTurnID)
		if nextTurn == "" {
			nextTurn = g.firstAliveID()
		}
		g.CurrentTurnID = nextTurn
		g.Phase = scenarioPhaseReveal
		g.setLastStage(scenarioText(g.Scenario, "event.reveal.nextTurn", "Next reveal turn."), g.scenarioTextKey("event.reveal.nextTurn", "event.reveal.nextTurn"), nil)
		g.scheduleRevealTimeoutIfNeeded()
		return gameActionResult{
			StateChanged: true,
			Events:       []gameEvent{g.makeEventLocalized("info", g.LastStageText, g.LastStageKey, g.LastStageVars)},
		}
	}
	if g.VotesRemaining > 0 {
		g.startVoting()
		return gameActionResult{
			StateChanged: true,
			Events:       []gameEvent{g.makeEvent("votingStart", g.LastStageText)},
		}
	}
	return g.startNextRoundOrEnd()
}

func (g *gameSession) handleRevealPhaseTimeout() gameActionResult {
	actorID := g.CurrentTurnID
	if actorID == "" {
		return gameActionResult{}
	}
	player := g.Players[actorID]
	if player == nil || player.Status != playerAlive {
		nextTurn := g.nextUnrevealedAliveAfter(actorID)
		g.CurrentTurnID = nextTurn
		g.scheduleRevealTimeoutIfNeeded()
		return gameActionResult{StateChanged: nextTurn != ""}
	}

	if g.Settings.RevealTimeoutAction == "random_card" {
		hidden := make([]int, 0, len(player.Hand))
		for i := range player.Hand {
			if !player.Hand[i].Revealed {
				hidden = append(hidden, i)
			}
		}
		if len(hidden) > 0 {
			cardIndex := hidden[g.rng.Intn(len(hidden))]
			card := &player.Hand[cardIndex]
			card.Revealed = true
			if card.Deck == healthDeckName() && g.FirstHealthRevealerID == "" {
				g.FirstHealthRevealerID = actorID
			}
			enterResult := g.enterRevealDiscussion()
			events := []gameEvent{g.makeEventLocalized("info", scenarioText(g.Scenario, "event.reveal.timeoutAuto", fmt.Sprintf("%s auto-revealed a card due to timeout.", player.Name)), "event.reveal.timeoutAuto", map[string]any{"name": player.Name})}
			events = append(events, enterResult.Events...)
			return gameActionResult{StateChanged: true, Events: events}
		}
	}

	g.RevealedThisRnd[actorID] = true
	g.LastRevealerID = actorID
	if g.allAliveRevealed() {
		if g.VotesRemaining > 0 {
			g.startVoting()
			g.setLastStage(
				scenarioText(g.Scenario, "event.voting.startedAfterReveal", "Voting started after the reveal phase."),
				"event.voting.startedAfterReveal",
				nil,
			)
			return gameActionResult{
				StateChanged: true,
				Events: []gameEvent{
					g.makeEventLocalized("info", g.LastStageText, g.LastStageKey, g.LastStageVars),
					g.makeEventLocalized("votingStart", g.LastStageText, g.LastStageKey, g.LastStageVars),
				},
			}
		}
		next := g.startNextRoundOrEnd()
		if next.StateChanged && g.LastStageKey != "" {
			next.Events = append([]gameEvent{g.makeEventLocalized("info", g.LastStageText, g.LastStageKey, g.LastStageVars)}, next.Events...)
		}
		return next
	}

	nextTurn := g.nextUnrevealedAliveAfter(actorID)
	if nextTurn == "" {
		nextTurn = g.firstAliveID()
	}
	g.CurrentTurnID = nextTurn
	g.Phase = scenarioPhaseReveal
	g.setLastStage(scenarioText(g.Scenario, "event.reveal.nextTurn", "Next reveal turn."), g.scenarioTextKey("event.reveal.nextTurn", "event.reveal.nextTurn"), nil)
	g.scheduleRevealTimeoutIfNeeded()
	return gameActionResult{
		StateChanged: true,
		Events:       []gameEvent{g.makeEventLocalized("info", g.LastStageText, g.LastStageKey, g.LastStageVars)},
	}
}

func (g *gameSession) markVoteWasted(voterID, reason string) {
	if reason == "" {
		reason = scenarioText(g.Scenario, "vote.spent", "Vote spent.")
	}
	g.AutoWastedVoters[voterID] = true
	g.VoteDisabled[voterID] = reason
	g.Votes[voterID] = voteRecord{
		TargetID:   "",
		Submitted:  time.Now().UnixMilli(),
		IsValid:    false,
		ReasonText: reason,
	}
}

func (g *gameSession) markVoteSelf(voterID string) {
	if voterID == "" {
		return
	}
	if _, ok := g.Players[voterID]; !ok {
		return
	}
	g.AutoSelfVoteVoters[voterID] = true
	g.Votes[voterID] = voteRecord{
		TargetID:  voterID,
		Submitted: time.Now().UnixMilli(),
		IsValid:   true,
	}
}

func (g *gameSession) enterVoteSpecialWindow() gameActionResult {
	g.BaseVotes = copyVotes(g.Votes)
	g.VotePhase = votePhaseSpecialWindow
	g.setLastStage(
		scenarioText(g.Scenario, "event.voting.specialWindow", "Vote collection complete. Special-condition window is open."),
		g.scenarioTextKey("event.voting.specialWindow", "event.voting.specialWindow"),
		nil,
	)
	g.clearActiveTimer()
	if g.Settings.EnablePostVoteDiscussion {
		g.schedulePhaseTimer("post_vote", g.Settings.PostVoteDiscussionSeconds)
	}
	return gameActionResult{
		StateChanged: true,
		Events:       []gameEvent{g.makeEventLocalized("info", g.LastStageText, g.LastStageKey, g.LastStageVars)},
	}
}

func (g *gameSession) resetVotesForRevote() {
	g.Votes = map[string]voteRecord{}
	g.BaseVotes = map[string]voteRecord{}
	for voterID := range g.AutoWastedVoters {
		g.markVoteWasted(voterID, "Vote spent.")
	}
	// Secret forced-self vote applies only to one voting attempt.
	g.AutoSelfVoteVoters = map[string]bool{}
}

func (g *gameSession) startTieBreakRevote(candidates []string) {
	g.TieBreakUsed = true
	g.VoteCandidates = map[string]bool{}
	for _, candidateID := range candidates {
		g.VoteCandidates[candidateID] = true
	}
	g.RevoteDisallowByVoter = map[string]map[string]bool{}
	g.resetVotesForRevote()
	g.VotePhase = votePhaseVoting
	g.setLastStage(scenarioText(g.Scenario, "event.tie.revote", "Tie. Revote between top candidates."), g.scenarioTextKey("event.tie.revote", "classic.auto.061"), nil)
}

func (g *gameSession) buildEffectiveVotes(source map[string]voteRecord) map[string]effectiveVoteRecord {
	out := make(map[string]effectiveVoteRecord, len(g.Players))
	for _, playerID := range g.Order {
		player := g.Players[playerID]
		if player == nil {
			continue
		}
		record, ok := source[playerID]
		if !ok {
			out[playerID] = effectiveVoteRecord{Status: "not_voted", Weight: 0}
			continue
		}

		info := effectiveVoteRecord{
			TargetID:   record.TargetID,
			Status:     "voted",
			ReasonText: record.ReasonText,
			Weight:     1,
			Submitted:  record.Submitted,
		}
		if !record.IsValid {
			info.Status = "invalid"
		}
		if info.TargetID == "" {
			info.Status = "not_voted"
		}
		if reason, blocked := g.VoteDisabled[playerID]; blocked {
			info.Status = "invalid"
			if info.ReasonText == "" {
				info.ReasonText = reason
			}
		}
		if info.TargetID != "" {
			if disallow := g.RevoteDisallowByVoter[playerID]; disallow != nil && disallow[info.TargetID] {
				info.Status = "invalid"
				if info.ReasonText == "" {
					info.ReasonText = "You cannot vote for this candidate."
				}
			}
		}
		if info.TargetID != "" && !g.VoteCandidates[info.TargetID] {
			info.Status = "invalid"
			if info.ReasonText == "" {
				info.ReasonText = "Invalid candidate."
			}
		}
		if info.TargetID != "" {
			target := g.Players[info.TargetID]
			if target != nil && target.BannedAgainst[playerID] {
				info.Status = "invalid"
				if info.ReasonText == "" {
					info.ReasonText = "You cannot vote against this player."
				}
			}
		}

		if info.Status == "voted" && info.TargetID != "" {
			if weight, ok := g.VoteWeights[playerID]; ok && weight > 0 {
				info.Weight = weight
			}
			if g.DoubleAgainst != "" && g.DoubleAgainst == info.TargetID {
				info.Weight *= 2
			}
		} else {
			info.TargetID = ""
			info.Weight = 0
		}
		out[playerID] = info
	}
	return out
}

func (g *gameSession) computeVoteTotals(source map[string]voteRecord) (map[string]int, []string) {
	totals := map[string]int{}
	for candidateID := range g.VoteCandidates {
		totals[candidateID] = 0
	}
	effective := g.buildEffectiveVotes(source)
	for voterID, info := range effective {
		if info.Status != "voted" || info.TargetID == "" {
			continue
		}
		voter := g.Players[voterID]
		if voter == nil || voter.Status != playerAlive {
			continue
		}
		totals[info.TargetID] += info.Weight
	}
	maxVotes := 0
	top := make([]string, 0, len(totals))
	for candidateID, count := range totals {
		if count > maxVotes {
			maxVotes = count
			top = []string{candidateID}
			continue
		}
		if count == maxVotes {
			top = append(top, candidateID)
		}
	}
	if len(top) == 0 {
		for candidateID := range g.VoteCandidates {
			top = append(top, candidateID)
		}
	}
	slices.Sort(top)
	return totals, top
}

func (g *gameSession) removeFromVoting(targetID string) {
	delete(g.VoteCandidates, targetID)
	delete(g.Votes, targetID)
	delete(g.BaseVotes, targetID)
	delete(g.VoteResults, targetID)
	delete(g.VoteWeights, targetID)
	delete(g.VoteDisabled, targetID)
	delete(g.AutoSelfVoteVoters, targetID)
	delete(g.AutoWastedVoters, targetID)
	for voterID, restricted := range g.RevoteDisallowByVoter {
		if restricted == nil {
			continue
		}
		delete(restricted, targetID)
		if len(restricted) == 0 {
			delete(g.RevoteDisallowByVoter, voterID)
		}
	}
	if g.DoubleAgainst == targetID {
		g.DoubleAgainst = ""
	}
}

func (g *gameSession) applyElimination(targetID string) []gameEvent {
	events := make([]gameEvent, 0, 2)
	target := g.Players[targetID]
	if target == nil || target.Status != playerAlive {
		return events
	}
	target.Status = playerEliminated
	g.TotalExiles++
	g.LastEliminatedID = targetID
	if g.VotesRemaining > 0 {
		g.VotesRemaining--
	}
	g.handleOnOwnerEliminated(target)
	events = append(events, g.handleSecretEliminationTriggers(targetID)...)
	g.removeFromVoting(targetID)
	return events
}

func (g *gameSession) handleOnOwnerEliminated(player *gamePlayer) {
	for i := range player.Specials {
		special := &player.Specials[i]
		if special.Used || special.Definition.Trigger != "onOwnerEliminated" || !special.Definition.Implemented {
			continue
		}
		choiceKind := effectiveSpecialChoiceKind(special.Definition)
		if choiceKind != "none" {
			if choiceKind == "bunker" && !g.hasRevealedBunkerCards() {
				special.Used = true
				special.PendingActivation = false
				continue
			}
			special.PendingActivation = true
			if !special.RevealedPublic {
				special.RevealedPublic = true
			}
			continue
		}

		result := g.applySpecialEffect(player, special, map[string]any{})
		if result.Error != "" {
			continue
		}
		if !special.RevealedPublic {
			special.RevealedPublic = true
		}
	}
}

func (g *gameSession) neighborIDs(playerID string, includeEliminatedID string) (string, string) {
	aliveOrTarget := func(id string) bool {
		if id == includeEliminatedID {
			return true
		}
		player := g.Players[id]
		return player != nil && player.Status == playerAlive
	}
	index := slices.Index(g.Order, playerID)
	if index < 0 || len(g.Order) == 0 {
		return "", ""
	}
	left := ""
	right := ""
	for step := 1; step <= len(g.Order); step++ {
		candidate := g.Order[(index-step+len(g.Order))%len(g.Order)]
		if aliveOrTarget(candidate) {
			left = candidate
			break
		}
	}
	for step := 1; step <= len(g.Order); step++ {
		candidate := g.Order[(index+step)%len(g.Order)]
		if aliveOrTarget(candidate) {
			right = candidate
			break
		}
	}
	return left, right
}

func (g *gameSession) parseRevealedAge(player *gamePlayer) (int, bool) {
	var ageRegexp = regexp.MustCompile(`\d{1,3}`)
	cards := g.getCardsByCategoryKey(player, "biology", true)
	if len(cards) == 0 {
		return 0, false
	}
	match := ageRegexp.FindString(cards[0].Label)
	if match == "" {
		return 0, false
	}
	age := asInt(match, -1)
	if age < 1 || age > 120 {
		return 0, false
	}
	return age, true
}

func (g *gameSession) computeAgeExtremes() (youngestID string, oldestID string, ok bool) {
	youngestAge := 1 << 30
	oldestAge := -1
	for _, playerID := range g.Order {
		player := g.Players[playerID]
		if player == nil || player.Status != playerAlive {
			continue
		}
		age, hasAge := g.parseRevealedAge(player)
		if !hasAge {
			continue
		}
		if age < youngestAge {
			youngestAge = age
			youngestID = playerID
		}
		if age > oldestAge {
			oldestAge = age
			oldestID = playerID
		}
	}
	if youngestID == "" || oldestID == "" {
		return "", "", false
	}
	return youngestID, oldestID, true
}

func (g *gameSession) handleSecretEliminationTriggers(eliminatedID string) []gameEvent {
	events := make([]gameEvent, 0, 2)
	youngestID, oldestID, hasAgeExtremes := g.computeAgeExtremes()
	for _, playerID := range g.Order {
		player := g.Players[playerID]
		if player == nil {
			continue
		}
		for i := range player.Specials {
			special := &player.Specials[i]
			if special.Used || special.Definition.Trigger != "secret_onEliminate" || !special.Definition.Implemented {
				continue
			}
			condition := asString(special.Definition.Effect.Params["condition"])
			triggered := false
			switch condition {
			case "leftNeighborEliminated":
				left, _ := g.neighborIDs(playerID, eliminatedID)
				triggered = left == eliminatedID
			case "rightNeighborEliminated":
				_, right := g.neighborIDs(playerID, eliminatedID)
				triggered = right == eliminatedID
			case "youngestByRevealedAgeEliminated":
				triggered = hasAgeExtremes && youngestID == eliminatedID
			case "oldestByRevealedAgeEliminated":
				triggered = hasAgeExtremes && oldestID == eliminatedID
			case "firstRevealedHealthEliminated":
				triggered = g.FirstHealthRevealerID != "" && g.FirstHealthRevealerID == eliminatedID
			}
			if triggered {
				special.Used = true
				special.RevealedPublic = true
				player.ForcedSelfVoteNext = true
				player.ForcedWastedVoteNext = false
				events = append(
					events,
					g.makeEventLocalized(
						"info",
						scenarioText(g.Scenario, "event.special.secretForcedSelfVote", fmt.Sprintf("%s triggers a secret effect: %s.", player.Name, special.Definition.Title)),
						"event.special.secretForcedSelfVote",
						map[string]any{"name": player.Name, "title": special.Definition.Title},
					),
				)
			}
		}
	}
	return events
}

func (g *gameSession) resolveNeighborChoice(actorID string, payload map[string]any) (string, string, string) {
	left, right := g.neighborIDs(actorID, "")
	targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
	side := strings.TrimSpace(strings.ToLower(asString(payload["side"])))

	if targetID != "" {
		if targetID == left {
			return left, "left", ""
		}
		if targetID == right {
			return right, "right", ""
		}
		return "", "", scenarioText(g.Scenario, "error.neighbor.invalid", "Invalid neighbor.")
	}
	if side == "left" && left != "" {
		return left, "left", ""
	}
	if side == "right" && right != "" {
		return right, "right", ""
	}
	return "", "", scenarioText(g.Scenario, "error.neighbor.notFound", "Neighbor not found.")
}
func (g *gameSession) getTargetCandidates(scope, actorID string) []string {
	switch scope {
	case "self":
		return []string{actorID}
	case "neighbors":
		left, right := g.neighborIDs(actorID, "")
		out := make([]string, 0, 2)
		if left != "" {
			out = append(out, left)
		}
		if right != "" && right != left {
			out = append(out, right)
		}
		return out
	case "any_alive", "any_including_self":
		return g.aliveIDs()
	default:
		return nil
	}
}

func (g *gameSession) getCardsByCategoryKey(player *gamePlayer, categoryKey string, onlyRevealed bool) []*handCard {
	deckName := categoryKeyToDeck[categoryKey]
	if deckName == "" {
		deckName = categoryKey
	}
	slot := categorySlotKey(categoryKey)
	out := make([]*handCard, 0, 2)
	for i := range player.Hand {
		card := &player.Hand[i]
		if card.Deck != deckName {
			continue
		}
		if slot != "" && card.SlotKey != slot {
			continue
		}
		if onlyRevealed && !card.Revealed {
			continue
		}
		out = append(out, card)
	}
	return out
}

func resolveCategoryKey(input string) string {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return ""
	}
	if _, ok := categoryKeyToDeck[raw]; ok {
		return raw
	}
	if slot := categorySlotKey(raw); slot != "" {
		return raw
	}

	normalized := normalizeSpecialKey(raw)
	if normalized == "" {
		return ""
	}
	for key := range categoryKeyToDeck {
		if normalizeSpecialKey(key) == normalized || normalizeSpecialKey(canonicalCategoryKey(key)) == normalized {
			return key
		}
		if deckName := categoryKeyToDeck[key]; deckName != "" && normalizeSpecialKey(deckName) == normalized {
			return key
		}
	}
	return ""
}

func (g *gameSession) getFirstRevealedCard(player *gamePlayer, categoryKey string) *handCard {
	cards := g.getCardsByCategoryKey(player, categoryKey, true)
	if len(cards) == 0 {
		return nil
	}
	return cards[0]
}

func (g *gameSession) getFirstCardForSpecial(player *gamePlayer, categoryKey string) *handCard {
	if player == nil {
		return nil
	}
	if card := g.getFirstRevealedCard(player, categoryKey); card != nil {
		return card
	}
	if !g.IsDev {
		return nil
	}
	cards := g.getCardsByCategoryKey(player, categoryKey, false)
	if len(cards) == 0 {
		return nil
	}
	return cards[0]
}

func (g *gameSession) getCardByCategoryInstance(player *gamePlayer, categoryKey, instanceID string) *handCard {
	if player == nil || strings.TrimSpace(instanceID) == "" {
		return nil
	}
	cards := g.getCardsByCategoryKey(player, categoryKey, false)
	for _, card := range cards {
		if card.InstanceID == instanceID {
			return card
		}
	}
	return nil
}

func (g *gameSession) drawCardFromDeck(deckName string) (assetCard, bool) {
	pool := g.DeckPools[deckName]
	card, nextPool, ok := drawRandomCard(pool, g.rng)
	g.DeckPools[deckName] = nextPool
	return card, ok
}

func (g *gameSession) applySpecial(actorID, specialInstanceID string, payload map[string]any) gameActionResult {
	player := g.Players[actorID]
	if player == nil || player.Status != playerAlive {
		return runtimeError("You are not an active player.", "error.player.excluded", nil)
	}
	if g.Phase == scenarioPhaseEnded {
		return runtimeError("Game has already ended.", "error.game.alreadyEnded", nil)
	}
	if specialInstanceID == "" {
		return runtimeError("Special condition not found.", "error.special.notFound", nil)
	}

	special := g.findPlayerSpecialByInstance(player, specialInstanceID)
	if special == nil {
		return runtimeError("Special condition not found.", "error.special.notFound", nil)
	}
	if !special.Definition.Implemented {
		return runtimeError("This card is not implemented yet.", "error.special.unimplemented", nil)
	}
	if special.Used {
		return runtimeError("This card has already been used.", "error.special.alreadyUsed", nil)
	}

	if !g.IsDev && g.Settings.SpecialUsage == "only_during_voting" && g.Phase != scenarioPhaseVoting {
		return runtimeError("Special conditions can only be used during voting.", "error.special.onlyVoting", nil)
	}
	if special.Definition.Trigger == "onOwnerEliminated" || special.Definition.Trigger == "secret_onEliminate" {
		return runtimeError("This card triggers automatically.", "error.special.autoTrigger", nil)
	}

	effectivePayload := map[string]any{}
	for k, v := range payload {
		effectivePayload[k] = v
	}
	choiceKind := effectiveSpecialChoiceKind(special.Definition)
	targetScope := effectiveSpecialTargetScope(special.Definition, choiceKind)

	if choiceKind != "" && choiceKind != "none" && len(payload) == 0 {
		allowImplicitChoice := choiceKind == "category" && asString(special.Definition.Effect.Params["category"]) != ""
		if choiceKind == "bunker" {
			effectType := strings.TrimSpace(special.Definition.Effect.Type)
			allowImplicitChoice = effectType == "discardBunkerCard" || effectType == "stealBunkerCardToExiled"
		}
		if !allowImplicitChoice {
			return runtimeError("A choice is required to apply this card.", "error.special.payloadRequired", nil)
		}
	}

	if targetScope == "neighbors" {
		targetID, side, errText := g.resolveNeighborChoice(actorID, effectivePayload)
		if errText != "" {
			return gameActionResult{Error: errText}
		}
		effectivePayload["targetPlayerId"] = targetID
		effectivePayload["side"] = side
	} else if targetScope == "self" {
		effectivePayload["targetPlayerId"] = actorID
	} else if targetScope == "any_alive" || targetScope == "any_including_self" {
		candidates := g.getTargetCandidates(targetScope, actorID)
		targetID := strings.TrimSpace(asString(effectivePayload["targetPlayerId"]))
		if targetID == "" {
			return runtimeError("You need to choose a target.", "error.target.required", nil)
		}
		if !slices.Contains(candidates, targetID) {
			return runtimeError("Invalid target.", "error.target.invalid", nil)
		}
	}

	if choiceKind == "player" && !special.Definition.AllowSelf {
		targetID := strings.TrimSpace(asString(effectivePayload["targetPlayerId"]))
		if targetID != "" && targetID == actorID {
			return runtimeError("You cannot choose yourself.", "error.target.cannotSelf", nil)
		}
	}

	if validationErr := g.validateSpecialRequires(special, effectivePayload); !validationErr.isEmpty() {
		return validationErr.asActionResult()
	}

	wasDevChoiceCard := g.IsDev && isDevChoiceSpecialID(special.Definition.ID)
	effectTypeBeforeApply := strings.TrimSpace(strings.ToLower(special.Definition.Effect.Type))
	result := g.applySpecialEffect(player, special, effectivePayload)
	if result.Error != "" {
		return result
	}
	if !result.StateChanged {
		return result
	}
	special = g.findPlayerSpecialByInstance(player, specialInstanceID)
	if special == nil {
		return result
	}
	if wasDevChoiceCard && effectTypeBeforeApply != strings.ToLower(devChoiceEffectType) {
		if chooser, ok := g.buildDevChoiceDefinition(player.PlayerID); ok {
			special.Definition = chooser
			special.Used = false
			special.PendingActivation = false
			special.RevealedPublic = true
			result.StateChanged = true
		}
	}
	if !special.RevealedPublic {
		special.RevealedPublic = true
		result.Events = append(result.Events, g.makeEventLocalized("info", scenarioText(g.Scenario, "event.special.applied", fmt.Sprintf("%s applies a special condition: %s.", player.Name, special.Definition.Title)), "event.special.applied", map[string]any{"name": player.Name, "title": special.Definition.Title}))
	}
	return result
}

func (g *gameSession) applySpecialWithPending(actorID, specialInstanceID string, payload map[string]any) gameActionResult {
	player := g.Players[actorID]
	if player == nil {
		return runtimeError("Player special condition was not found.", "error.special.playerMissing", nil)
	}
	if specialInstanceID == "" {
		return runtimeError("Special condition not found.", "error.special.notFound", nil)
	}

	special := g.findPlayerSpecialByInstance(player, specialInstanceID)
	if special == nil {
		return runtimeError("Special condition not found.", "error.special.notFound", nil)
	}
	if !special.Definition.Implemented {
		return runtimeError("This card is not implemented yet.", "error.special.unimplemented", nil)
	}
	if special.Used {
		return runtimeError("This card has already been used.", "error.special.alreadyUsed", nil)
	}

	trigger := special.Definition.Trigger
	if trigger == "secret_onEliminate" {
		return runtimeError("This card triggers automatically.", "error.special.autoTrigger", nil)
	}
	if trigger == "onOwnerEliminated" {
		if player.Status != playerEliminated {
			return runtimeError("This card can only be activated after the owner is eliminated.", "error.special.autoTrigger", nil)
		}
	} else {
		if player.Status != playerAlive {
			return runtimeError("You are not an active player.", "error.player.excluded", nil)
		}
		if g.Phase == scenarioPhaseEnded {
			return runtimeError("Game has already ended.", "error.game.alreadyEnded", nil)
		}
		if !g.IsDev && g.Settings.SpecialUsage == "only_during_voting" && g.Phase != scenarioPhaseVoting {
			return runtimeError("Special conditions can only be used during voting.", "error.special.onlyVoting", nil)
		}
	}

	effectivePayload := map[string]any{}
	for k, v := range payload {
		effectivePayload[k] = v
	}
	choiceKind := effectiveSpecialChoiceKind(special.Definition)
	targetScope := effectiveSpecialTargetScope(special.Definition, choiceKind)

	if choiceKind != "" && choiceKind != "none" && len(payload) == 0 {
		allowImplicitChoice := choiceKind == "category" && asString(special.Definition.Effect.Params["category"]) != ""
		if choiceKind == "bunker" {
			effectType := strings.TrimSpace(special.Definition.Effect.Type)
			allowImplicitChoice = effectType == "discardBunkerCard" || effectType == "stealBunkerCardToExiled"
		}
		if !allowImplicitChoice {
			return runtimeError("A choice is required to apply this card.", "error.special.payloadRequired", nil)
		}
	}

	if targetScope == "neighbors" {
		targetID, side, errText := g.resolveNeighborChoice(actorID, effectivePayload)
		if errText != "" {
			return gameActionResult{Error: errText}
		}
		effectivePayload["targetPlayerId"] = targetID
		effectivePayload["side"] = side
	} else if targetScope == "self" {
		effectivePayload["targetPlayerId"] = actorID
	} else if targetScope == "any_alive" || targetScope == "any_including_self" {
		candidates := g.getTargetCandidates(targetScope, actorID)
		targetID := strings.TrimSpace(asString(effectivePayload["targetPlayerId"]))
		if targetID == "" {
			return runtimeError("You need to choose a target.", "error.target.required", nil)
		}
		if !slices.Contains(candidates, targetID) {
			return runtimeError("Invalid target.", "error.target.invalid", nil)
		}
	}

	if choiceKind == "player" && !special.Definition.AllowSelf {
		targetID := strings.TrimSpace(asString(effectivePayload["targetPlayerId"]))
		if targetID != "" && targetID == actorID {
			return runtimeError("You cannot choose yourself.", "error.target.cannotSelf", nil)
		}
	}

	if validationErr := g.validateSpecialRequires(special, effectivePayload); !validationErr.isEmpty() {
		return validationErr.asActionResult()
	}

	wasDevChoiceCard := g.IsDev && isDevChoiceSpecialID(special.Definition.ID)
	effectTypeBeforeApply := strings.TrimSpace(strings.ToLower(special.Definition.Effect.Type))
	result := g.applySpecialEffect(player, special, effectivePayload)
	if result.Error != "" {
		return result
	}
	special = g.findPlayerSpecialByInstance(player, specialInstanceID)
	if special == nil {
		if !result.StateChanged {
			return result
		}
		return result
	}
	if wasDevChoiceCard && result.StateChanged && effectTypeBeforeApply != strings.ToLower(devChoiceEffectType) {
		if chooser, ok := g.buildDevChoiceDefinition(player.PlayerID); ok {
			special.Definition = chooser
			special.Used = false
			special.PendingActivation = false
			special.RevealedPublic = true
			result.StateChanged = true
		}
	}
	if trigger == "onOwnerEliminated" && special.PendingActivation {
		special.PendingActivation = false
		result.StateChanged = true
	}
	loggedSpecialUseEvent := false
	if trigger == "onOwnerEliminated" {
		result.Events = append(
			result.Events,
			g.makeEventLocalized("info", scenarioText(g.Scenario, "event.special.applied", fmt.Sprintf("%s applies a special condition: %s.", player.Name, special.Definition.Title)), "event.special.applied", map[string]any{"name": player.Name, "title": special.Definition.Title}),
		)
		loggedSpecialUseEvent = true
	}
	if !special.RevealedPublic {
		special.RevealedPublic = true
		if !loggedSpecialUseEvent {
			result.Events = append(
				result.Events,
				g.makeEventLocalized("info", scenarioText(g.Scenario, "event.special.applied", fmt.Sprintf("%s applies a special condition: %s.", player.Name, special.Definition.Title)), "event.special.applied", map[string]any{"name": player.Name, "title": special.Definition.Title}),
			)
		}
		result.StateChanged = true
	}
	if !result.StateChanged {
		return result
	}
	return result
}

type localizedRuntimeError struct {
	Message string
	Key     string
	Vars    map[string]any
}

func runtimeError(message, key string, vars map[string]any) gameActionResult {
	return gameActionResult{
		Error:     message,
		ErrorKey:  key,
		ErrorVars: vars,
	}
}

func (e localizedRuntimeError) isEmpty() bool {
	return e.Message == "" && e.Key == "" && len(e.Vars) == 0
}

func (e localizedRuntimeError) asActionResult() gameActionResult {
	return runtimeError(e.Message, e.Key, e.Vars)
}

func (g *gameSession) validateSpecialRequires(special *specialConditionState, payload map[string]any) localizedRuntimeError {
	for _, requirement := range special.Definition.Requires {
		switch requirement {
		case "phase=voting":
			if !g.IsDev && g.Phase != scenarioPhaseVoting {
				return localizedRuntimeError{Message: "This card can only be used during voting.", Key: "error.special.onlyVoting"}
			}
		case "phase=reveal":
			if !g.IsDev && g.Phase != scenarioPhaseReveal {
				return localizedRuntimeError{Message: "This card can only be used during reveal.", Key: "validate.phase.reveal"}
			}
		case "votingStarted":
			if !g.IsDev && len(g.Votes) == 0 && len(g.BaseVotes) == 0 {
				return localizedRuntimeError{Message: "Voting has not started yet.", Key: "error.voting.notNow"}
			}
		case "targetHasBaggage":
			targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
			target := g.Players[targetID]
			if target == nil || len(g.getCardsByCategoryKey(target, "baggage", false)) == 0 {
				return localizedRuntimeError{Message: "The selected player has no baggage.", Key: "validate.target.noBaggage"}
			}
		case "targetHasRevealedHealth":
			targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
			target := g.Players[targetID]
			if target == nil || len(g.getCardsByCategoryKey(target, "health", !g.IsDev)) == 0 {
				return localizedRuntimeError{Message: "The selected player has no revealed health card.", Key: "validate.target.noRevealedHealth"}
			}
		case "targetHasRevealedProfession":
			targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
			target := g.Players[targetID]
			if target == nil || len(g.getCardsByCategoryKey(target, "profession", !g.IsDev)) == 0 {
				return localizedRuntimeError{Message: "The selected player has no revealed profession card.", Key: "validate.target.noRevealedProfession"}
			}
		case "targetHasRevealedSameCategory":
			categoryKey := asString(special.Definition.Effect.Params["category"])
			targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
			target := g.Players[targetID]
			if target == nil || len(g.getCardsByCategoryKey(target, categoryKey, !g.IsDev)) == 0 {
				return localizedRuntimeError{Message: "The neighbor has no revealed card in this category.", Key: "validate.neighbor.noRevealedCard"}
			}
		case "needsNeighborIndexing":
			if len(g.Order) <= 1 {
				return localizedRuntimeError{Message: "Not enough players for neighbor targeting.", Key: "validate.neighbor.notEnoughPlayers"}
			}
		case "ageFieldAvailable", "someRevealedAges":
			_, _, ok := g.computeAgeExtremes()
			if !ok {
				return localizedRuntimeError{Message: "No one has revealed age yet.", Key: "validate.age.noneRevealed"}
			}
		case "trackFirstRevealHealth":
			if g.FirstHealthRevealerID == "" {
				return localizedRuntimeError{Message: "The first health reveal has not happened yet.", Key: "validate.health.firstRevealMissing"}
			}
		}
	}
	return localizedRuntimeError{}
}

func (g *gameSession) applySpecialEffect(player *gamePlayer, special *specialConditionState, payload map[string]any) gameActionResult {
	def := special.Definition
	effectType := def.Effect.Type

	if !g.IsDev && votingWindowEffectTypes[effectType] && g.VotePhase != votePhaseSpecialWindow {
		return runtimeError("This card can only be used in the voting special-condition window.", "error.special.onlyVoting", nil)
	}

	switch effectType {
	case devChoiceEffectType:
		if !g.IsDev {
			return runtimeError("This action is only available in dev mode.", "error.dev.only", nil)
		}
		selectedID := strings.TrimSpace(asString(payload["specialId"]))
		if selectedID == "" {
			return runtimeError("You need to choose a special condition.", "error.special.choiceRequired", nil)
		}
		selectedDef, ok := g.findSpecialDefinitionForDevChoice(selectedID)
		if !ok {
			return runtimeError("The selected special condition was not found.", "error.special.choiceNotFound", nil)
		}
		assetLookup := g.buildSpecialAssetLookup()
		nextDef := g.buildDevSpecialFromTemplate(special.Definition.ID, selectedDef)
		resolvedAssetID := g.resolveSpecialAssetID(selectedDef, assetLookup)
		if strings.TrimSpace(resolvedAssetID) == "" {
			resolvedAssetID = g.resolveSpecialAssetID(nextDef, assetLookup)
		}
		nextDef.AssetID = resolvedAssetID
		special.Definition = nextDef
		special.Used = false
		special.PendingActivation = false
		special.RevealedPublic = true
		return gameActionResult{
			StateChanged: true,
			Events: []gameEvent{
				g.makeEventLocalized("info", scenarioText(g.Scenario, "event.devChoice.selected", fmt.Sprintf("%s selected for the DEV card: %s.", player.Name, selectedDef.Title)), "event.devChoice.selected", map[string]any{"name": player.Name, "title": selectedDef.Title, "assetId": resolvedAssetID}),
			},
		}
	case "banVoteAgainst":
		if g.Phase != scenarioPhaseVoting {
			return runtimeError("There is no voting right now.", "error.voting.notNow", nil)
		}
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		if targetID == "" {
			return runtimeError("You need to choose a target.", "error.target.required", nil)
		}
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return runtimeError("The target is not in the game.", "error.target.notAlive", nil)
		}
		if targetID == player.PlayerID && !def.AllowSelf {
			return runtimeError("You cannot choose yourself.", "error.target.cannotSelf", nil)
		}
		player.BannedAgainst[targetID] = true
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.vote.banAgainst", fmt.Sprintf("%s forbids voting against %s for themselves.", player.Name, target.Name)),
			"event.vote.banAgainst",
			map[string]any{"name": player.Name, "target": target.Name},
		)}}

	case "voteWeight":
		if g.Phase != scenarioPhaseVoting {
			return runtimeError("There is no voting right now.", "error.voting.notNow", nil)
		}
		weight := asInt(def.Effect.Params["weight"], 2)
		if weight <= 0 {
			weight = 2
		}
		g.VoteWeights[player.PlayerID] = weight
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.vote.weightBoost", fmt.Sprintf("%s boosts their vote.", player.Name)),
			"event.vote.weightBoost",
			map[string]any{"name": player.Name},
		)}}

	case "disableVote":
		if g.Phase != scenarioPhaseVoting {
			return runtimeError("There is no voting right now.", "error.voting.notNow", nil)
		}
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		if targetID == "" {
			return runtimeError("You need to choose a target.", "error.target.required", nil)
		}
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return runtimeError("The target is not in the game.", "error.target.notAlive", nil)
		}
		if targetID == player.PlayerID && !def.AllowSelf {
			return runtimeError("You cannot choose yourself.", "error.target.cannotSelf", nil)
		}
		g.markVoteWasted(targetID, "vote.blocked.bySpecial")
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.vote.disable", fmt.Sprintf("%s blocks %s's vote.", player.Name, target.Name)),
			"event.vote.disable",
			map[string]any{"name": player.Name, "target": target.Name},
		)}}

	case "doubleVotesAgainst_and_disableSelfVote":
		if g.Phase != scenarioPhaseVoting {
			return runtimeError("There is no voting right now.", "error.voting.notNow", nil)
		}
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		if targetID == "" {
			return runtimeError("You need to choose a target.", "error.target.required", nil)
		}
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return runtimeError("The target is not in the game.", "error.target.notAlive", nil)
		}
		if targetID == player.PlayerID && !def.AllowSelf {
			return runtimeError("You cannot choose yourself.", "error.target.cannotSelf", nil)
		}
		g.DoubleAgainst = targetID
		g.markVoteWasted(player.PlayerID, "vote.spent.bySpecial")
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.vote.doubleAgainst", fmt.Sprintf("%s boosts votes against %s.", player.Name, target.Name)),
			"event.vote.doubleAgainst",
			map[string]any{"name": player.Name, "target": target.Name},
		)}}

	case "forceRevote":
		if g.Phase != scenarioPhaseVoting {
			return runtimeError("There is no voting right now.", "error.voting.notNow", nil)
		}
		source := g.currentVoteSource()
		if len(source) == 0 {
			return runtimeError("Voting data is unavailable.", "error.voting.noSource", nil)
		}
		if asBool(def.Effect.Params["disallowPreviousCandidate"]) {
			disallowByVoter := make(map[string]map[string]bool, len(source))
			for voterID, record := range source {
				if !record.IsValid || strings.TrimSpace(record.TargetID) == "" {
					continue
				}
				if !g.VoteCandidates[record.TargetID] {
					continue
				}
				disallowByVoter[voterID] = map[string]bool{record.TargetID: true}
			}
			g.RevoteDisallowByVoter = disallowByVoter
		} else {
			g.RevoteDisallowByVoter = map[string]map[string]bool{}
		}
		g.resetVotesForRevote()
		g.VotePhase = votePhaseVoting
		g.clearActiveTimer()
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.vote.forceRevote", fmt.Sprintf("%s starts a revote.", player.Name)),
			"event.vote.forceRevote",
			map[string]any{"name": player.Name},
		)}}

	case "swapRevealedWithNeighbor":
		categoryKey := asString(def.Effect.Params["category"])
		if strings.TrimSpace(categoryKey) == "" {
			return runtimeError("You need to choose a category.", "error.category.required", nil)
		}
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		if targetID == "" {
			return runtimeError("You need to choose a target.", "error.target.required", nil)
		}
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return runtimeError("Invalid target.", "error.target.invalid", nil)
		}
		requestedActorCardID := strings.TrimSpace(asString(payload["actorCardId"]))
		requestedTargetCardID := strings.TrimSpace(asString(payload["targetCardId"]))

		var yourCard *handCard
		if requestedActorCardID != "" {
			yourCard = g.getCardByCategoryInstance(player, categoryKey, requestedActorCardID)
			if yourCard == nil {
				return runtimeError("You need to choose one of your cards in this category.", "error.target.invalid", map[string]any{"field": "actorCardId"})
			}
			if !g.IsDev && !yourCard.Revealed {
				return runtimeError("You can only choose a revealed card in this category.", "error.target.noRevealedCategory", map[string]any{"field": "actorCardId"})
			}
		} else {
			yourCard = g.getFirstCardForSpecial(player, categoryKey)
			if yourCard == nil {
				return runtimeError("You do not have a revealed card in this category.", "error.target.noRevealedCategory", map[string]any{"field": "actorCardId"})
			}
		}

		var revealedCard *handCard
		if requestedTargetCardID != "" {
			revealedCard = g.getCardByCategoryInstance(target, categoryKey, requestedTargetCardID)
			if revealedCard == nil {
				return runtimeError("The target does not have that card in this category.", "error.target.invalid", map[string]any{"field": "targetCardId"})
			}
			if !g.IsDev && !revealedCard.Revealed {
				return runtimeError("The target has no revealed card in that category.", "error.target.noRevealedCategory", map[string]any{"field": "targetCardId"})
			}
		} else {
			revealedCard = g.getFirstCardForSpecial(target, categoryKey)
			if revealedCard == nil {
				return runtimeError("The target has no revealed card in that category.", "error.target.noRevealedCategory", nil)
			}
		}

		yourCard.CardID, revealedCard.CardID = revealedCard.CardID, yourCard.CardID
		yourCard.Label, revealedCard.Label = revealedCard.Label, yourCard.Label
		yourCard.Missing, revealedCard.Missing = revealedCard.Missing, yourCard.Missing
		yourCard.Revealed, revealedCard.Revealed = revealedCard.Revealed, yourCard.Revealed
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.swap.revealedWithNeighbor", fmt.Sprintf("%s swaps a revealed card with %s.", player.Name, target.Name)),
			"event.swap.revealedWithNeighbor",
			map[string]any{"name": player.Name, "target": target.Name},
		)}}

	case "replaceRevealedCard":
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		categoryKey := asString(def.Effect.Params["category"])
		deckName := categoryKeyToDeck[categoryKey]
		if target == nil || target.Status != playerAlive {
			return runtimeError("The target is not in the game.", "error.target.notAlive", nil)
		}
		if deckName == "" {
			return runtimeError("Deck is unavailable for this category.", "error.deck.unavailable", map[string]any{"category": categoryKey})
		}
		requestedTargetCardID := strings.TrimSpace(asString(payload["targetCardId"]))
		var revealedCard *handCard
		if requestedTargetCardID != "" {
			revealedCard = g.getCardByCategoryInstance(target, categoryKey, requestedTargetCardID)
			if revealedCard == nil {
				return runtimeError("The target does not have that card in this category.", "error.target.invalid", map[string]any{"field": "targetCardId"})
			}
			if !g.IsDev && !revealedCard.Revealed {
				return runtimeError("The target has no revealed card in that category.", "error.target.noRevealedCategory", map[string]any{"field": "targetCardId"})
			}
		} else {
			revealedCard = g.getFirstCardForSpecial(target, categoryKey)
		}
		if revealedCard == nil {
			return runtimeError("The target has no revealed card in that category.", "error.target.noRevealedCategory", nil)
		}
		newCard, ok := g.drawCardFromDeck(deckName)
		if !ok {
			return runtimeError(scenarioText(g.Scenario, "error.deck.emptyCategory", fmt.Sprintf("No more cards remain in category \"%s\".", deckName)), "error.deck.emptyCategory", map[string]any{"deckName": deckName})
		}
		revealedCard.CardID = newCard.ID
		revealedCard.Label = newCard.Label
		revealedCard.Missing = false
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.card.replaced", fmt.Sprintf("%s replaces %s's revealed card.", player.Name, target.Name)),
			"event.card.replaced",
			map[string]any{"name": player.Name, "target": target.Name},
		)}}

	case "discardRevealedAndDealHidden":
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		categoryKey := asString(def.Effect.Params["category"])
		deckName := categoryKeyToDeck[categoryKey]
		if target == nil || target.Status != playerAlive {
			return runtimeError("The target is not in the game.", "error.target.notAlive", nil)
		}
		if deckName == "" {
			return runtimeError("Deck is unavailable for this category.", "error.deck.unavailable", map[string]any{"category": categoryKey})
		}
		requestedTargetCardID := strings.TrimSpace(asString(payload["targetCardId"]))
		var revealedCard *handCard
		if requestedTargetCardID != "" {
			revealedCard = g.getCardByCategoryInstance(target, categoryKey, requestedTargetCardID)
			if revealedCard == nil {
				return runtimeError("The target does not have that card in this category.", "error.target.invalid", map[string]any{"field": "targetCardId"})
			}
			if !g.IsDev && !revealedCard.Revealed {
				return runtimeError("The target has no revealed card in that category.", "error.target.noRevealedCategory", map[string]any{"field": "targetCardId"})
			}
		} else {
			revealedCard = g.getFirstCardForSpecial(target, categoryKey)
		}
		if revealedCard == nil {
			return runtimeError("The target has no revealed card in that category.", "error.target.noRevealedCategory", nil)
		}
		newCard, ok := g.drawCardFromDeck(deckName)
		if !ok {
			return runtimeError(scenarioText(g.Scenario, "error.deck.emptyCategory", fmt.Sprintf("No more cards remain in category \"%s\".", deckName)), "error.deck.emptyCategory", map[string]any{"deckName": deckName})
		}
		revealedCard.CardID = newCard.ID
		revealedCard.Label = newCard.Label
		revealedCard.Missing = false
		revealedCard.Revealed = false
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.card.discardedDealHidden", fmt.Sprintf("%s discards %s's revealed card and deals a hidden one.", player.Name, target.Name)),
			"event.card.discardedDealHidden",
			map[string]any{"name": player.Name, "target": target.Name},
		)}}

	case "redealAllRevealed":
		categoryKey := asString(def.Effect.Params["category"])
		deckName := categoryKeyToDeck[categoryKey]
		if deckName == "" {
			return runtimeError("Deck is unavailable for this category.", "error.deck.unavailable", map[string]any{"category": categoryKey})
		}
		revealedSlots := make([]*handCard, 0, 16)
		for _, targetID := range g.Order {
			target := g.Players[targetID]
			if target == nil || target.Status != playerAlive {
				continue
			}
			targetCards := g.getCardsByCategoryKey(target, categoryKey, true)
			if g.IsDev && len(targetCards) == 0 {
				targetCards = g.getCardsByCategoryKey(target, categoryKey, false)
			}
			revealedSlots = append(revealedSlots, targetCards...)
		}
		if len(revealedSlots) == 0 {
			return runtimeError(scenarioText(g.Scenario, "error.redeal.none", "There are no revealed cards to redeal."), "error.redeal.none", nil)
		}
		shuffled := append([]*handCard(nil), revealedSlots...)
		for i := len(shuffled) - 1; i > 0; i-- {
			j := g.rng.Intn(i + 1)
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		}
		for i := range revealedSlots {
			from := shuffled[i]
			to := revealedSlots[i]
			to.CardID = from.CardID
			to.Label = from.Label
			to.Missing = from.Missing
		}
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.redeal.category", fmt.Sprintf("%s redeals revealed cards in category %s.", player.Name, deckName)),
			"event.redeal.category",
			map[string]any{"name": player.Name, "deckName": deckName},
		)}}

	case "replaceBunkerCard":
		index, bunkerErr := g.resolveBunkerIndex(payload, false)
		if !bunkerErr.isEmpty() {
			return bunkerErr.asActionResult()
		}
		bunkerDeck, _, _ := resolveWorldDeckNames(g.DeckPools)
		if bunkerDeck == "" {
			return runtimeError("Bunker deck is unavailable.", "error.deck.unavailable", map[string]any{"deckName": "bunker"})
		}
		replacement, nextPool, ok := drawRandomCard(g.DeckPools[bunkerDeck], g.rng)
		if !ok {
			return runtimeError(scenarioText(g.Scenario, "error.bunker.noReplacement", "No bunker cards are available for replacement."), "error.bunker.noReplacement", nil)
		}
		g.DeckPools[bunkerDeck] = nextPool
		target := &g.World.Bunker[index]
		target.ID = replacement.ID
		target.Title = replacement.Label
		target.Description = replacement.Label
		target.Text = ""
		target.ImageID = replacement.ID
		target.ImgURL = "/assets/" + replacement.ID
		target.IsRevealed = true
		target.RevealedBy = player.PlayerID
		if target.RevealedAtRound == nil {
			revealRound := g.Round
			target.RevealedAtRound = &revealRound
		}
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.bunker.replaced", fmt.Sprintf("%s replaces a bunker card.", player.Name)),
			"event.bunker.replaced",
			map[string]any{"name": player.Name},
		)}}

	case "discardBunkerCard":
		index, bunkerErr := g.resolveBunkerIndex(payload, true)
		if !bunkerErr.isEmpty() {
			return bunkerErr.asActionResult()
		}
		target := &g.World.Bunker[index]
		target.ID = fmt.Sprintf("bunker-discarded-%d-%d", time.Now().UnixMilli(), g.rng.Intn(1_000_000))
		target.Title = scenarioText(g.Scenario, "world.bunker.lost.title", "Bunker card lost")
		target.Description = scenarioText(g.Scenario, "world.bunker.lost.description", "The bunker card was discarded by a special condition.")
		target.Text = ""
		target.ImageID = ""
		target.ImgURL = ""
		target.IsRevealed = true
		target.RevealedBy = player.PlayerID
		if target.RevealedAtRound == nil {
			revealRound := g.Round
			target.RevealedAtRound = &revealRound
		}
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.bunker.discardedBySpecial", fmt.Sprintf("Special condition \"%s\" discards a bunker card.", special.Definition.Title)),
			"event.bunker.discardedBySpecial",
			map[string]any{"title": special.Definition.Title},
		)}}

	case "stealBunkerCardToExiled":
		index, bunkerErr := g.resolveBunkerIndex(payload, true)
		if !bunkerErr.isEmpty() {
			return bunkerErr.asActionResult()
		}
		target := &g.World.Bunker[index]
		target.ID = fmt.Sprintf("bunker-stolen-%d-%d", time.Now().UnixMilli(), g.rng.Intn(1_000_000))
		target.Title = scenarioText(g.Scenario, "world.bunker.removed.title", "Bunker card removed")
		target.Description = scenarioText(g.Scenario, "world.bunker.removed.description", "The bunker card was removed from play by a special condition.")
		target.Text = ""
		target.ImageID = ""
		target.ImgURL = ""
		target.IsRevealed = true
		target.RevealedBy = player.PlayerID
		if target.RevealedAtRound == nil {
			revealRound := g.Round
			target.RevealedAtRound = &revealRound
		}
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.bunker.removedBySpecial", fmt.Sprintf("Card \"%s\" applied: the bunker card was removed from the table.", special.Definition.Title)),
			"event.bunker.removedBySpecial",
			map[string]any{"title": special.Definition.Title},
		)}}

	case "forceRevealCategoryForAll":
		categoryKey := resolveCategoryKey(asString(payload["category"]))
		if categoryKey == "" {
			categoryKey = resolveCategoryKey(asString(def.Effect.Params["category"]))
		}
		if categoryKey == "" {
			return runtimeError("You need to choose a category.", "error.category.required", nil)
		}
		forcedLabel := canonicalCategoryKey(categoryKey)
		if forcedLabel == "" {
			deckName := categoryKeyToDeck[categoryKey]
			if deckName == "" {
				deckName = categoryKey
			}
			forcedLabel = deckName
		}
		g.RoundRules.ForcedCategory = forcedLabel
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.category.forcedAll", fmt.Sprintf("%s forces everyone to reveal category %s.", player.Name, forcedLabel)),
			"event.category.forcedAll",
			map[string]any{"name": player.Name, "category": forcedLabel},
		)}}

	case "setRoundRule":
		if raw, ok := def.Effect.Params["noTalkUntilVoting"]; ok {
			g.RoundRules.NoTalkUntilVoting = asBool(raw)
		} else {
			g.RoundRules.NoTalkUntilVoting = true
		}
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.round.ruleSet", fmt.Sprintf("%s sets a round rule.", player.Name)),
			"event.round.ruleSet",
			map[string]any{"name": player.Name},
		)}}

	case "stealBaggage_and_giveSpecial":
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return runtimeError("The target is not in the game.", "error.target.notAlive", nil)
		}
		if targetID == player.PlayerID && !def.AllowSelf {
			return runtimeError("You cannot choose yourself.", "error.target.cannotSelf", nil)
		}
		targetBaggage := g.getCardsByCategoryKey(target, "baggage", false)
		if len(targetBaggage) == 0 {
			return runtimeError("The target has no baggage cards.", "error.baggage.none", nil)
		}
		requestedBaggageCardID := strings.TrimSpace(asString(payload["baggageCardId"]))
		stolen := targetBaggage[0]
		if requestedBaggageCardID != "" {
			stolen = nil
			for _, card := range targetBaggage {
				if card.InstanceID == requestedBaggageCardID {
					stolen = card
					break
				}
			}
			if stolen == nil {
				return runtimeError("You need to choose a specific baggage card.", "error.baggage.pickSpecific", nil)
			}
		}
		giveCount := asInt(def.Effect.Params["giveSpecialCount"], 1)
		if giveCount < 1 {
			giveCount = 1
		}
		if len(g.specialPool) < giveCount {
			return runtimeError("There are no special conditions to give.", "error.special.poolEmpty", nil)
		}

		stolenCard := *stolen
		stolenVisible := stolenCard.Revealed || g.IsDev
		target.Hand = slices.DeleteFunc(target.Hand, func(card handCard) bool { return card.InstanceID == stolenCard.InstanceID })
		player.Hand = append(player.Hand, handCard{
			InstanceID:         g.nextCardInstanceID(player.PlayerID),
			CardID:             stolenCard.CardID,
			Deck:               stolenCard.Deck,
			SlotKey:            stolenCard.SlotKey,
			Label:              stolenCard.Label,
			Revealed:           stolenVisible,
			Missing:            stolenCard.Missing,
			PublicBackCategory: stolenCard.PublicBackCategory,
		})
		specialAssetID := strings.TrimSpace(g.resolveSpecialAssetID(def, g.buildSpecialAssetLookup()))
		target.Hand = append(target.Hand, handCard{
			InstanceID:         g.nextCardInstanceID(target.PlayerID),
			CardID:             specialAssetID,
			Deck:               stolenCard.Deck,
			SlotKey:            stolenCard.SlotKey,
			Label:              def.Title,
			Revealed:           false,
			Missing:            specialAssetID == "",
			PublicBackCategory: specialDeckCategoryName,
		})
		for i := 0; i < giveCount; i++ {
			defToGive := g.drawSpecialFromPool()
			defToGive.AssetID = g.resolveSpecialAssetID(defToGive, g.buildSpecialAssetLookup())
			target.Specials = append(target.Specials, specialConditionState{
				InstanceID:     g.nextSpecialInstanceID(target.PlayerID),
				Definition:     defToGive,
				RevealedPublic: false,
				Used:           false,
			})
		}
		special.Used = true
		if !isDevChoiceSpecialID(special.Definition.ID) {
			player.Specials = slices.DeleteFunc(player.Specials, func(item specialConditionState) bool {
				return item.InstanceID == special.InstanceID
			})
		}
		imgURL := ""
		if stolenVisible && stolenCard.CardID != "" {
			imgURL = "/assets/" + stolenCard.CardID
		}
		player.SpecialCategoryProxyCards = []publicCategoryCard{
			{
				Label:        stolenCard.Label,
				ImgURL:       imgURL,
				Revealed:     stolenVisible,
				Hidden:       !stolenVisible,
				BackCategory: canonicalCategoryKey("baggage"),
			},
		}
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.special.stealBaggageAndGive", fmt.Sprintf("%s steals baggage from %s and gives a special condition.", player.Name, target.Name)),
			"event.special.stealBaggageAndGive",
			map[string]any{"name": player.Name, "target": target.Name},
		)}}

	case "addFinalThreat":
		threatKey := asString(def.Effect.Params["threatKey"])
		if strings.TrimSpace(threatKey) == "" {
			threatKey = def.ID
		}
		g.FinalThreats = append(g.FinalThreats, threatKey)
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEventLocalized(
			"info",
			scenarioText(g.Scenario, "event.finalThreat.added", fmt.Sprintf("%s adds a threat to the finale.", player.Name)),
			"event.finalThreat.added",
			map[string]any{"name": player.Name},
		)}}

	default:
		return runtimeError("Effect is not supported.", "error.effect.unsupported", nil)
	}
}

func asBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return strings.EqualFold(v, "true") || v == "1"
	default:
		return false
	}
}
