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
		"profession": "Профессия",
		"health":     "Здоровье",
		"hobby":      "Хобби",
		"baggage":    "Багаж",
		"facts":      "Факты",
		"facts1":     "Факты",
		"facts2":     "Факты",
		"biology":    "Биология",
	}

	categoryKeyToSlot = map[string]string{
		"facts1": "facts1",
		"facts2": "facts2",
	}

	categoryKeyToLabel = map[string]string{
		"profession": "Профессия",
		"health":     "Здоровье",
		"hobby":      "Хобби",
		"baggage":    "Багаж",
		"facts":      "Факт №1",
		"facts1":     "Факт №1",
		"facts2":     "Факт №2",
		"biology":    "Биология",
	}

	specialDeckCategoryName = "Особые условия"
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
	lower = strings.ReplaceAll(lower, "ё", "е")
	var b strings.Builder
	b.Grow(len(lower))
	for _, r := range lower {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
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
		return "игрок"
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
		normalizedDeck := normalizeSpecialKey(deckName)
		if normalizedDeck == "" || (!strings.Contains(normalizedDeck, normalizeSpecialKey(specialDeckCategoryName)) && !strings.Contains(normalizedDeck, "special")) {
			continue
		}
		for _, card := range cards {
			if card.ID == "" {
				continue
			}
			label := strings.TrimSpace(sanitizeHumanText(card.Label, card.Label))
			deck := strings.TrimSpace(sanitizeHumanText(deckName, deckName))
			baseName := filepath.Base(strings.ReplaceAll(card.ID, "\\", "/"))

			addCandidate(card.ID, card.ID)
			addCandidate(strings.TrimPrefix(card.ID, "decks/"), card.ID)
			addCandidate(label, card.ID)
			addCandidate(baseName, card.ID)
			addCandidate(trimCardFileExt(baseName), card.ID)
			addCandidate(deck+"/"+baseName, card.ID)
			addCandidate(deck+"/"+trimCardFileExt(baseName), card.ID)
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
	return normalizeAssetIDPath(def.AssetID)
}

func (g *gameSession) assignInitialSpecialsForPlayer(player *gamePlayer, assetLookup map[string]string) {
	if player == nil {
		return
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
	if g.IsDev {
		if devChoice, ok := g.buildDevChoiceDefinition(player.PlayerID); ok {
			player.Specials = append(player.Specials, specialConditionState{
				InstanceID:     g.nextSpecialInstanceID(player.PlayerID),
				Definition:     devChoice,
				RevealedPublic: true,
				Used:           false,
			})
		}
	}
}

func (g *gameSession) buildDevChoiceDefinition(playerID string) (specialDefinition, bool) {
	var base specialDefinition
	found := false
	for _, item := range implementedSpecialDefinitions {
		if item.Effect.Type == "forceRevealCategoryForAll" {
			base = item
			found = true
			break
		}
	}
	if !found {
		for _, item := range implementedSpecialDefinitions {
			if item.ChoiceKind != "" && item.ChoiceKind != "none" {
				base = item
				found = true
				break
			}
		}
	}
	if !found {
		return specialDefinition{}, false
	}

	copyDef := base
	copyDef.ID = fmt.Sprintf("dev-choice-%s", playerID)
	copyDef.Title = strings.TrimSpace(base.Title) + "(DEV)"
	copyDef.Implemented = true
	if base.Effect.Type == "forceRevealCategoryForAll" {
		copyDef.Requires = nil
		copyDef.ChoiceKind = "category"
		copyDef.TargetScope = ""
		copyDef.AllowSelf = false
	}
	return copyDef, true
}

func (g *gameSession) nextSpecialInstanceID(playerID string) string {
	g.specialCounter++
	return fmt.Sprintf("special-%s-%d", playerID, g.specialCounter)
}

func (g *gameSession) drawSpecialFromPool() specialDefinition {
	if len(g.specialPool) == 0 {
		return specialDefinition{
			ID:          "missing-special",
			Title:       "Нет доступного условия",
			Text:        "Колода особых условий пуста.",
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
		view := specialConditionInstanceView{
			InstanceID:     special.InstanceID,
			ID:             special.Definition.ID,
			Title:          special.Definition.Title,
			Text:           special.Definition.Text,
			Trigger:        special.Definition.Trigger,
			Implemented:    special.Definition.Implemented,
			RevealedPublic: special.RevealedPublic,
			Used:           special.Used,
			ImgURL:         specialImageURL(special.Definition),
			Effect: specialConditionEffectView{
				Type:   special.Definition.Effect.Type,
				Params: special.Definition.Effect.Params,
			},
		}
		if special.Definition.ChoiceKind != "" {
			view.ChoiceKind = special.Definition.ChoiceKind
			view.NeedsChoice = special.Definition.ChoiceKind != "none"
		}
		if special.Definition.AllowSelf {
			view.AllowSelfTarget = true
		}
		if special.Definition.TargetScope != "" {
			view.TargetScope = special.Definition.TargetScope
		}
		out = append(out, view)
	}
	return out
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
		return gameActionResult{Error: "Сейчас нельзя продолжить ход."}
	}
	g.clearActiveTimer()

	if !g.allAliveRevealed() {
		nextTurn := g.nextUnrevealedAliveAfter(g.CurrentTurnID)
		if nextTurn == "" {
			nextTurn = g.firstAliveID()
		}
		g.CurrentTurnID = nextTurn
		g.Phase = scenarioPhaseReveal
		g.LastStageText = "Следующий ход раскрытия."
		g.scheduleRevealTimeoutIfNeeded()
		return gameActionResult{
			StateChanged: true,
			Events:       []gameEvent{g.makeEvent("info", g.LastStageText)},
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
			if card.Deck == "Здоровье" && g.FirstHealthRevealerID == "" {
				g.FirstHealthRevealerID = actorID
			}
			g.RevealedThisRnd[actorID] = true
			g.LastRevealerID = actorID
			g.LastStageText = fmt.Sprintf("Таймаут: %s автоматически раскрыл карту.", player.Name)
			enterResult := g.enterRevealDiscussion()
			events := []gameEvent{g.makeEvent("info", g.LastStageText)}
			events = append(events, enterResult.Events...)
			return gameActionResult{StateChanged: true, Events: events}
		}
	}

	g.RevealedThisRnd[actorID] = true
	g.LastRevealerID = actorID
	if g.allAliveRevealed() {
		if g.VotesRemaining > 0 {
			g.startVoting()
			g.LastStageText = fmt.Sprintf("Таймаут: ход %s пропущен.", player.Name)
			return gameActionResult{
				StateChanged: true,
				Events: []gameEvent{
					g.makeEvent("info", g.LastStageText),
					g.makeEvent("votingStart", g.LastStageText),
				},
			}
		}
		next := g.startNextRoundOrEnd()
		if next.StateChanged {
			next.Events = append([]gameEvent{g.makeEvent("info", fmt.Sprintf("Таймаут: ход %s пропущен.", player.Name))}, next.Events...)
		}
		return next
	}

	nextTurn := g.nextUnrevealedAliveAfter(actorID)
	if nextTurn == "" {
		nextTurn = g.firstAliveID()
	}
	g.CurrentTurnID = nextTurn
	g.Phase = scenarioPhaseReveal
	g.LastStageText = fmt.Sprintf("Таймаут: ход %s пропущен.", player.Name)
	g.scheduleRevealTimeoutIfNeeded()
	return gameActionResult{
		StateChanged: true,
		Events:       []gameEvent{g.makeEvent("info", g.LastStageText)},
	}
}

func (g *gameSession) markVoteWasted(voterID, reason string) {
	if reason == "" {
		reason = "Голос заблокирован."
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

func (g *gameSession) enterVoteSpecialWindow() gameActionResult {
	g.BaseVotes = copyVotes(g.Votes)
	g.VotePhase = votePhaseSpecialWindow
	g.LastStageText = "Сбор голосов завершён. Окно спецусловий."
	g.clearActiveTimer()
	if g.Settings.EnablePostVoteDiscussion {
		g.schedulePhaseTimer("post_vote", g.Settings.PostVoteDiscussionSeconds)
	}
	return gameActionResult{
		StateChanged: true,
		Events:       []gameEvent{g.makeEvent("info", g.LastStageText)},
	}
}

func (g *gameSession) resetVotesForRevote() {
	g.Votes = map[string]voteRecord{}
	g.BaseVotes = map[string]voteRecord{}
	for voterID := range g.AutoWastedVoters {
		g.markVoteWasted(voterID, "Голос потрачен.")
	}
}

func (g *gameSession) startTieBreakRevote(candidates []string) {
	g.TieBreakUsed = true
	g.VoteCandidates = map[string]bool{}
	for _, candidateID := range candidates {
		g.VoteCandidates[candidateID] = true
	}
	g.RevoteDisallow = map[string]bool{}
	g.resetVotesForRevote()
	g.VotePhase = votePhaseVoting
	g.clearActiveTimer()
	g.LastStageText = "Ничья. Переголосование между топ-кандидатами."
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
		if info.TargetID != "" && g.RevoteDisallow[info.TargetID] {
			info.Status = "invalid"
			if info.ReasonText == "" {
				info.ReasonText = "Нельзя голосовать за этого кандидата."
			}
		}
		if info.TargetID != "" && !g.VoteCandidates[info.TargetID] {
			info.Status = "invalid"
			if info.ReasonText == "" {
				info.ReasonText = "Кандидат недоступен."
			}
		}
		if info.TargetID != "" {
			target := g.Players[info.TargetID]
			if target != nil && target.BannedAgainst[playerID] {
				info.Status = "invalid"
				if info.ReasonText == "" {
					info.ReasonText = "Голос против этого игрока запрещён."
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
	delete(g.AutoWastedVoters, targetID)
	delete(g.RevoteDisallow, targetID)
	if g.DoubleAgainst == targetID {
		g.DoubleAgainst = ""
	}
}

func (g *gameSession) applyElimination(targetID string) {
	target := g.Players[targetID]
	if target == nil || target.Status != playerAlive {
		return
	}
	target.Status = playerEliminated
	g.TotalExiles++
	g.LastEliminatedID = targetID
	if g.VotesRemaining > 0 {
		g.VotesRemaining--
	}
	g.handleOnOwnerEliminated(target)
	g.handleSecretEliminationTriggers(targetID)
	g.removeFromVoting(targetID)
}

func (g *gameSession) handleOnOwnerEliminated(player *gamePlayer) {
	for i := range player.Specials {
		special := &player.Specials[i]
		if special.Used || special.Definition.Trigger != "onOwnerEliminated" || !special.Definition.Implemented {
			continue
		}
		special.Used = true
		if special.Definition.Effect.Type == "addFinalThreat" {
			threatKey := asString(special.Definition.Effect.Params["threatKey"])
			if strings.TrimSpace(threatKey) == "" {
				threatKey = special.Definition.ID
			}
			g.FinalThreats = append(g.FinalThreats, threatKey)
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
	for _, card := range player.Hand {
		if card.Deck != "Биология" || !card.Revealed {
			continue
		}
		match := ageRegexp.FindString(card.Label)
		if match == "" {
			return 0, false
		}
		age := asInt(match, -1)
		if age < 1 || age > 120 {
			return 0, false
		}
		return age, true
	}
	return 0, false
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

func (g *gameSession) handleSecretEliminationTriggers(eliminatedID string) {
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
				player.ForcedWastedVoteNext = true
			}
		}
	}
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
		return "", "", "Недопустимый сосед."
	}
	if side == "left" && left != "" {
		return left, "left", ""
	}
	if side == "right" && right != "" {
		return right, "right", ""
	}
	return "", "", "Сосед не найден."
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
	slot := categoryKeyToSlot[categoryKey]
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
	if _, ok := categoryKeyToLabel[raw]; ok {
		return raw
	}

	normalized := normalizeSpecialKey(raw)
	if normalized == "" {
		return ""
	}
	for key, label := range categoryKeyToLabel {
		if normalizeSpecialKey(key) == normalized || normalizeSpecialKey(label) == normalized {
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

func (g *gameSession) drawCardFromDeck(deckName string) (assetCard, bool) {
	pool := g.DeckPools[deckName]
	card, nextPool, ok := drawRandomCard(pool, g.rng)
	g.DeckPools[deckName] = nextPool
	return card, ok
}

func (g *gameSession) applySpecial(actorID, specialInstanceID string, payload map[string]any) gameActionResult {
	player := g.Players[actorID]
	if player == nil || player.Status != playerAlive {
		return gameActionResult{Error: "Игрок не найден."}
	}
	if g.Phase == scenarioPhaseEnded {
		return gameActionResult{Error: "Игра уже завершена."}
	}
	if specialInstanceID == "" {
		return gameActionResult{Error: "Особое условие не найдено."}
	}

	var special *specialConditionState
	for i := range player.Specials {
		if player.Specials[i].InstanceID == specialInstanceID {
			special = &player.Specials[i]
			break
		}
	}
	if special == nil {
		return gameActionResult{Error: "Особое условие не найдено."}
	}
	if !special.Definition.Implemented {
		return gameActionResult{Error: "Эта карта ещё не реализована."}
	}
	if special.Used {
		return gameActionResult{Error: "Эта карта уже использована."}
	}

	if !g.IsDev && g.Settings.SpecialUsage == "only_during_voting" && g.Phase != scenarioPhaseVoting {
		return gameActionResult{Error: "Особые условия можно использовать только во время голосования."}
	}
	if special.Definition.Trigger == "onOwnerEliminated" || special.Definition.Trigger == "secret_onEliminate" {
		return gameActionResult{Error: "Эта карта срабатывает автоматически."}
	}

	effectivePayload := map[string]any{}
	for k, v := range payload {
		effectivePayload[k] = v
	}
	choiceKind := special.Definition.ChoiceKind
	targetScope := special.Definition.TargetScope

	if choiceKind != "" && choiceKind != "none" && len(payload) == 0 {
		if !(choiceKind == "category" && asString(special.Definition.Effect.Params["category"]) != "") {
			return gameActionResult{Error: "Нужен выбор для применения карты."}
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
			return gameActionResult{Error: "Нужно выбрать цель."}
		}
		if !slices.Contains(candidates, targetID) {
			return gameActionResult{Error: "Недопустимая цель."}
		}
	}

	if choiceKind == "player" && !special.Definition.AllowSelf {
		targetID := strings.TrimSpace(asString(effectivePayload["targetPlayerId"]))
		if targetID != "" && targetID == actorID {
			return gameActionResult{Error: "Нельзя выбрать себя."}
		}
	}

	if errText := g.validateSpecialRequires(player, special, effectivePayload); errText != "" {
		return gameActionResult{Error: errText}
	}

	result := g.applySpecialEffect(player, special, effectivePayload)
	if result.Error != "" {
		return result
	}
	if !result.StateChanged {
		return result
	}
	if !special.RevealedPublic {
		special.RevealedPublic = true
		result.Events = append(result.Events, g.makeEvent("info", fmt.Sprintf("%s применяет особое условие: %s.", player.Name, special.Definition.Title)))
	}
	return result
}

func (g *gameSession) validateSpecialRequires(player *gamePlayer, special *specialConditionState, payload map[string]any) string {
	for _, requirement := range special.Definition.Requires {
		switch requirement {
		case "phase=voting":
			if !g.IsDev && g.Phase != scenarioPhaseVoting {
				return "Эту карту можно использовать только в фазе голосования."
			}
		case "phase=reveal":
			if !g.IsDev && g.Phase != scenarioPhaseReveal {
				return "Эту карту можно использовать только в фазе раскрытия."
			}
		case "votingStarted":
			if !g.IsDev && len(g.Votes) == 0 && len(g.BaseVotes) == 0 {
				return "Голосование ещё не началось."
			}
		case "targetHasBaggage":
			targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
			target := g.Players[targetID]
			if target == nil || len(g.getCardsByCategoryKey(target, "baggage", false)) == 0 {
				return "У выбранного игрока нет багажа."
			}
		case "targetHasRevealedHealth":
			targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
			target := g.Players[targetID]
			if target == nil || len(g.getCardsByCategoryKey(target, "health", !g.IsDev)) == 0 {
				return "У выбранного игрока нет раскрытого здоровья."
			}
		case "targetHasRevealedProfession":
			targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
			target := g.Players[targetID]
			if target == nil || len(g.getCardsByCategoryKey(target, "profession", !g.IsDev)) == 0 {
				return "У выбранного игрока нет раскрытой профессии."
			}
		case "targetHasRevealedSameCategory":
			categoryKey := asString(special.Definition.Effect.Params["category"])
			targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
			target := g.Players[targetID]
			if target == nil || len(g.getCardsByCategoryKey(target, categoryKey, !g.IsDev)) == 0 {
				return "У соседа нет раскрытой карты этой категории."
			}
		case "needsNeighborIndexing":
			if len(g.Order) <= 1 {
				return "Недостаточно игроков для соседей."
			}
		case "ageFieldAvailable", "someRevealedAges":
			_, _, ok := g.computeAgeExtremes()
			if !ok {
				return "Возраст ещё не раскрыт ни у одного игрока."
			}
		case "trackFirstRevealHealth":
			if g.FirstHealthRevealerID == "" {
				return "Ещё нет первого раскрытия здоровья."
			}
		}
	}
	return ""
}

func (g *gameSession) applySpecialEffect(player *gamePlayer, special *specialConditionState, payload map[string]any) gameActionResult {
	def := special.Definition
	effectType := def.Effect.Type

	if !g.IsDev && votingWindowEffectTypes[effectType] && g.VotePhase != votePhaseSpecialWindow {
		return gameActionResult{Error: "Эту карту можно использовать только в окне спецусловий голосования."}
	}

	switch effectType {
	case "banVoteAgainst":
		if g.Phase != scenarioPhaseVoting {
			return gameActionResult{Error: "Сейчас нет голосования."}
		}
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return gameActionResult{Error: "Цель не в игре."}
		}
		if targetID == player.PlayerID && !def.AllowSelf {
			return gameActionResult{Error: "Нельзя выбрать себя."}
		}
		player.BannedAgainst[targetID] = true
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s использует карту \"%s\".", player.Name, def.Title))}}

	case "voteWeight":
		if g.Phase != scenarioPhaseVoting {
			return gameActionResult{Error: "Сейчас нет голосования."}
		}
		weight := asInt(def.Effect.Params["weight"], 2)
		if weight <= 0 {
			weight = 2
		}
		g.VoteWeights[player.PlayerID] = weight
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s усиливает свой голос.", player.Name))}}

	case "disableVote":
		if g.Phase != scenarioPhaseVoting {
			return gameActionResult{Error: "Сейчас нет голосования."}
		}
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return gameActionResult{Error: "Цель не в игре."}
		}
		if targetID == player.PlayerID && !def.AllowSelf {
			return gameActionResult{Error: "Нельзя выбрать себя."}
		}
		g.markVoteWasted(targetID, "Голос заблокирован.")
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s блокирует голос игрока %s.", player.Name, target.Name))}}

	case "doubleVotesAgainst_and_disableSelfVote":
		if g.Phase != scenarioPhaseVoting {
			return gameActionResult{Error: "Сейчас нет голосования."}
		}
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return gameActionResult{Error: "Цель не в игре."}
		}
		if targetID == player.PlayerID && !def.AllowSelf {
			return gameActionResult{Error: "Нельзя выбрать себя."}
		}
		g.DoubleAgainst = targetID
		g.markVoteWasted(player.PlayerID, "Ваш голос потрачен.")
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s усиливает голоса против %s.", player.Name, target.Name))}}

	case "forceRevote":
		if g.Phase != scenarioPhaseVoting {
			return gameActionResult{Error: "Сейчас нет голосования."}
		}
		source := g.currentVoteSource()
		if len(source) == 0 {
			return gameActionResult{Error: "Нет данных голосования."}
		}
		if asBool(def.Effect.Params["disallowPreviousCandidate"]) {
			_, topCandidates := g.computeVoteTotals(source)
			g.RevoteDisallow = map[string]bool{}
			for _, candidateID := range topCandidates {
				g.RevoteDisallow[candidateID] = true
			}
		}
		g.resetVotesForRevote()
		g.VotePhase = votePhaseVoting
		g.clearActiveTimer()
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s запускает переголосование.", player.Name))}}

	case "swapRevealedWithNeighbor":
		categoryKey := asString(def.Effect.Params["category"])
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return gameActionResult{Error: "Сосед не найден."}
		}
		yourCard := g.getFirstCardForSpecial(player, categoryKey)
		theirCard := g.getFirstCardForSpecial(target, categoryKey)
		if yourCard == nil || theirCard == nil {
			return gameActionResult{Error: "Нужны раскрытые карты у обоих игроков."}
		}
		yourCard.CardID, theirCard.CardID = theirCard.CardID, yourCard.CardID
		yourCard.Label, theirCard.Label = theirCard.Label, yourCard.Label
		yourCard.Missing, theirCard.Missing = theirCard.Missing, yourCard.Missing
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s меняется раскрытой картой с %s.", player.Name, target.Name))}}

	case "replaceRevealedCard":
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		categoryKey := asString(def.Effect.Params["category"])
		deckName := categoryKeyToDeck[categoryKey]
		if target == nil || target.Status != playerAlive {
			return gameActionResult{Error: "Цель не в игре."}
		}
		if deckName == "" {
			return gameActionResult{Error: "Неизвестная категория."}
		}
		revealedCard := g.getFirstCardForSpecial(target, categoryKey)
		if revealedCard == nil {
			return gameActionResult{Error: "У цели нет раскрытой карты этой категории."}
		}
		newCard, ok := g.drawCardFromDeck(deckName)
		if !ok {
			return gameActionResult{Error: fmt.Sprintf("В колоде категории \"%s\" больше нет карт.", deckName)}
		}
		revealedCard.CardID = newCard.ID
		revealedCard.Label = newCard.Label
		revealedCard.Missing = false
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s заменяет раскрытую карту у %s.", player.Name, target.Name))}}

	case "discardRevealedAndDealHidden":
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		categoryKey := asString(def.Effect.Params["category"])
		deckName := categoryKeyToDeck[categoryKey]
		if target == nil || target.Status != playerAlive {
			return gameActionResult{Error: "Цель не в игре."}
		}
		if deckName == "" {
			return gameActionResult{Error: "Неизвестная категория."}
		}
		revealedCard := g.getFirstCardForSpecial(target, categoryKey)
		if revealedCard == nil {
			return gameActionResult{Error: "У цели нет раскрытой карты этой категории."}
		}
		newCard, ok := g.drawCardFromDeck(deckName)
		if !ok {
			return gameActionResult{Error: fmt.Sprintf("В колоде категории \"%s\" больше нет карт.", deckName)}
		}
		revealedCard.CardID = newCard.ID
		revealedCard.Label = newCard.Label
		revealedCard.Missing = false
		revealedCard.Revealed = false
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s сбрасывает раскрытую карту у %s.", player.Name, target.Name))}}

	case "redealAllRevealed":
		categoryKey := asString(def.Effect.Params["category"])
		deckName := categoryKeyToDeck[categoryKey]
		if deckName == "" {
			return gameActionResult{Error: "Неизвестная категория."}
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
			return gameActionResult{Error: "Нет раскрытых карт для перераздачи."}
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
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s перераздаёт раскрытые карты категории %s.", player.Name, deckName))}}

	case "forceRevealCategoryForAll":
		categoryKey := resolveCategoryKey(asString(payload["category"]))
		if categoryKey == "" {
			categoryKey = resolveCategoryKey(asString(def.Effect.Params["category"]))
		}
		if categoryKey == "" {
			return gameActionResult{Error: "Нужно выбрать категорию."}
		}
		forcedLabel := categoryKeyToLabel[categoryKey]
		if forcedLabel == "" {
			deckName := categoryKeyToDeck[categoryKey]
			if deckName == "" {
				deckName = categoryKey
			}
			forcedLabel = deckName
		}
		g.RoundRules.ForcedCategory = forcedLabel
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s требует раскрыть категорию %s.", player.Name, forcedLabel))}}

	case "setRoundRule":
		if raw, ok := def.Effect.Params["noTalkUntilVoting"]; ok {
			g.RoundRules.NoTalkUntilVoting = asBool(raw)
		} else {
			g.RoundRules.NoTalkUntilVoting = true
		}
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s вводит правило раунда.", player.Name))}}

	case "stealBaggage_and_giveSpecial":
		targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
		target := g.Players[targetID]
		if target == nil || target.Status != playerAlive {
			return gameActionResult{Error: "Цель не в игре."}
		}
		if targetID == player.PlayerID && !def.AllowSelf {
			return gameActionResult{Error: "Нельзя выбрать себя."}
		}
		targetBaggage := g.getCardsByCategoryKey(target, "baggage", false)
		if len(targetBaggage) == 0 {
			return gameActionResult{Error: "У цели нет багажа."}
		}
		giveCount := asInt(def.Effect.Params["giveSpecialCount"], 1)
		if giveCount < 1 {
			giveCount = 1
		}
		if len(g.specialPool) < giveCount {
			return gameActionResult{Error: "В колоде особых условий больше нет карт."}
		}

		stolen := targetBaggage[0]
		target.Hand = slices.DeleteFunc(target.Hand, func(card handCard) bool { return card.InstanceID == stolen.InstanceID })
		player.Hand = append(player.Hand, handCard{
			InstanceID: g.nextCardInstanceID(player.PlayerID),
			CardID:     stolen.CardID,
			Deck:       stolen.Deck,
			SlotKey:    stolen.SlotKey,
			Label:      stolen.Label,
			Revealed:   stolen.Revealed,
			Missing:    stolen.Missing,
		})
		for i := 0; i < giveCount; i++ {
			defToGive := g.drawSpecialFromPool()
			target.Specials = append(target.Specials, specialConditionState{
				InstanceID:     g.nextSpecialInstanceID(target.PlayerID),
				Definition:     defToGive,
				RevealedPublic: false,
				Used:           false,
			})
		}
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s забирает багаж у %s.", player.Name, target.Name))}}

	case "addFinalThreat":
		threatKey := asString(def.Effect.Params["threatKey"])
		if strings.TrimSpace(threatKey) == "" {
			threatKey = def.ID
		}
		g.FinalThreats = append(g.FinalThreats, threatKey)
		special.Used = true
		return gameActionResult{StateChanged: true, Events: []gameEvent{g.makeEvent("info", fmt.Sprintf("%s добавляет угрозу в финал.", player.Name))}}

	default:
		return gameActionResult{Error: "Эффект не поддерживается"}
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
