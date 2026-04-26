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
	scenarioLocaleData map[string]map[string]map[string]string
	cardLocaleOnce     sync.Once
	cardLocaleData     map[string]cardLocaleDictionaries
)

type cardLocaleDictionaries struct {
	Cards          map[string]string
	WorldBunker    map[string]string
	WorldDisasters map[string]string
	WorldThreats   map[string]string
}

func scenarioLocaleRootCandidates() []string {
	wd, _ := os.Getwd()
	return []string{
		filepath.Join(wd, "locales", "scenario"),
		filepath.Join(wd, "..", "locales", "scenario"),
		filepath.Join(wd, "..", "..", "locales", "scenario"),
	}
}

func loadScenarioLocaleData() {
	scenarioLocaleData = map[string]map[string]map[string]string{}
	for _, scenarioID := range []string{scenarioClassic, scenarioDevTest} {
		scenarioLocaleData[scenarioID] = map[string]map[string]string{}
		for _, locale := range []string{"ru", "en"} {
			scenarioLocaleData[scenarioID][locale] = map[string]string{}
			for _, root := range scenarioLocaleRootCandidates() {
				filePath := filepath.Join(root, scenarioID, locale+".json")
				raw, err := os.ReadFile(filePath)
				if err != nil {
					continue
				}
				parsed := map[string]string{}
				if err := json.Unmarshal(raw, &parsed); err != nil {
					continue
				}
				for key, value := range parsed {
					scenarioLocaleData[scenarioID][locale][key] = value
				}
				break
			}
		}
	}
}

func scenarioText(scenarioID, key, fallback string) string {
	return scenarioTextForLocale(scenarioID, "ru", key, fallback)
}

func scenarioTextForLocale(scenarioID, locale, key, fallback string) string {
	scenarioLocaleOnce.Do(loadScenarioLocaleData)
	locale = normalizeCardLocale(locale)
	if byLocale := scenarioLocaleData[scenarioID]; byLocale != nil {
		if dict := byLocale[locale]; dict != nil {
			if value := dict[key]; value != "" {
				return value
			}
		}
		if locale != "ru" {
			if dict := byLocale["ru"]; dict != nil {
				if value := dict[key]; value != "" {
					return value
				}
			}
		}
	}
	if byLocale := scenarioLocaleData[scenarioClassic]; byLocale != nil {
		dict := byLocale[locale]
		if value := dict[key]; value != "" {
			return value
		}
	}
	if byLocale := scenarioLocaleData[scenarioClassic]; byLocale != nil && locale != "ru" {
		dict := byLocale["ru"]
		if value := dict[key]; value != "" {
			return value
		}
	}
	return fallback
}

func localizedScenarioMeta(meta scenarioMeta, locale string) scenarioMeta {
	out := meta
	out.Name = scenarioTextForLocale(meta.ID, locale, "meta.name", meta.Name)
	out.Description = scenarioTextForLocale(meta.ID, locale, "meta.description", meta.Description)
	return out
}

func localeRootCandidates(parts ...string) []string {
	wd, _ := os.Getwd()
	return []string{
		filepath.Join(append([]string{wd, "locales"}, parts...)...),
		filepath.Join(append([]string{wd, "..", "locales"}, parts...)...),
		filepath.Join(append([]string{wd, "..", "..", "locales"}, parts...)...),
	}
}

func readLocaleObject(filePath string, target any) bool {
	raw, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	return json.Unmarshal(raw, target) == nil
}

func readCardLocaleFile(locale string) map[string]string {
	type cardsFile struct {
		Cards map[string]string `json:"cards"`
	}
	for _, root := range localeRootCandidates("cards") {
		var parsed cardsFile
		if readLocaleObject(filepath.Join(root, locale+".json"), &parsed) {
			return parsed.Cards
		}
	}
	return nil
}

func readWorldLocaleFile(group, locale, field string) map[string]string {
	for _, root := range localeRootCandidates("world", group) {
		parsed := map[string]map[string]string{}
		if readLocaleObject(filepath.Join(root, locale+".json"), &parsed) {
			return parsed[field]
		}
	}
	return nil
}

func loadCardLocaleData() {
	cardLocaleData = map[string]cardLocaleDictionaries{}
	for _, locale := range []string{"ru", "en"} {
		cardLocaleData[locale] = cardLocaleDictionaries{
			Cards:          readCardLocaleFile(locale),
			WorldBunker:    readWorldLocaleFile("bunker", locale, "subtitles"),
			WorldDisasters: readWorldLocaleFile("disasters", locale, "texts"),
			WorldThreats:   readWorldLocaleFile("threats", locale, "subtitles"),
		}
	}
}

func localeKeyFromAssetID(assetID, deck string) string {
	raw := strings.TrimSpace(assetID)
	if raw == "" {
		return ""
	}
	raw = strings.ReplaceAll(raw, "\\", "/")
	raw = strings.TrimPrefix(raw, "/")
	raw = strings.TrimPrefix(raw, "assets/")
	parts := strings.Split(raw, "/")
	if len(parts) > 0 {
		file := parts[len(parts)-1]
		for _, ext := range []string{".png", ".jpg", ".jpeg", ".webp"} {
			if strings.HasSuffix(strings.ToLower(file), ext) {
				file = file[:len(file)-len(ext)]
				break
			}
		}
		if deck == "" && len(parts) >= 2 {
			deck = resolveDeckIDByLabel(parts[len(parts)-2])
		}
		if deck != "" && file != "" {
			return deck + "." + file
		}
	}
	return ""
}

func localeInvariantAssetID(assetID string) string {
	raw := strings.ReplaceAll(strings.TrimSpace(assetID), "\\", "/")
	raw = strings.TrimPrefix(raw, "/")
	raw = strings.TrimPrefix(raw, "assets/")
	parts := strings.Split(raw, "/")
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == "1x" && (parts[i+1] == "ru" || parts[i+1] == "en") {
			parts = append(parts[:i+1], parts[i+2:]...)
			break
		}
	}
	return strings.Join(parts, "/")
}

func localizedCardLabel(locale, assetID, deck, fallback string) string {
	cardLocaleOnce.Do(loadCardLocaleData)
	key := localeKeyFromAssetID(assetID, deck)
	if key == "" {
		return fallback
	}
	locale = normalizeCardLocale(locale)
	if dict := cardLocaleData[locale].Cards; dict != nil {
		if value := dict[key]; value != "" {
			return value
		}
	}
	if locale != "ru" {
		if dict := cardLocaleData["ru"].Cards; dict != nil {
			if value := dict[key]; value != "" {
				return value
			}
		}
	}
	return fallback
}

func localizeWorldCard(locale string, card worldCardView) worldCardView {
	cardLocaleOnce.Do(loadCardLocaleData)
	key := localeKeyFromAssetID(card.ID, card.Kind)
	if key == "" {
		return card
	}
	locale = normalizeCardLocale(locale)
	if title := localizedCardLabel(locale, card.ID, card.Kind, card.Title); title != "" {
		card.Title = title
	}
	var text string
	switch card.Kind {
	case "bunker":
		text = cardLocaleData[locale].WorldBunker[key]
	case "disaster":
		text = cardLocaleData[locale].WorldDisasters[key]
	case "threat":
		text = cardLocaleData[locale].WorldThreats[key]
	}
	if text == "" && locale != "ru" {
		switch card.Kind {
		case "bunker":
			text = cardLocaleData["ru"].WorldBunker[key]
		case "disaster":
			text = cardLocaleData["ru"].WorldDisasters[key]
		case "threat":
			text = cardLocaleData["ru"].WorldThreats[key]
		}
	}
	if text != "" {
		card.Description = text
		if card.Text != "" {
			card.Text = text
		}
	}
	return card
}

func localizeWorldFacedCard(locale string, card worldFacedCardView) worldFacedCardView {
	base := localizeWorldCard(locale, worldCardView{
		Kind:        card.Kind,
		ID:          card.ID,
		Title:       card.Title,
		Description: card.Description,
		Text:        card.Text,
		ImageID:     card.ImageID,
		ImgURL:      card.ImgURL,
	})
	card.Title = base.Title
	card.Description = base.Description
	card.Text = base.Text
	return card
}

func localizeGameViewForLocale(view gameView, locale, scenarioID string) gameView {
	locale = normalizeCardLocale(locale)
	view.LastStage = formatServerTemplate(scenarioTextForLocale(scenarioID, locale, view.LastStageKey, view.LastStage), view.LastStageVars)
	view.Public.ResolutionNote = formatServerTemplate(scenarioTextForLocale(scenarioID, locale, view.Public.ResolutionNoteKey, view.Public.ResolutionNote), view.Public.ResolutionNoteVars)
	for i := range view.You.Specials {
		view.You.Specials[i] = localizeSpecialConditionView(locale, scenarioID, view.You.Specials[i])
	}
	for i := range view.You.Hand {
		view.You.Hand[i].Label = localizedCardLabel(locale, view.You.Hand[i].ID, view.You.Hand[i].Deck, view.You.Hand[i].Label)
	}
	for i := range view.You.Categories {
		for j := range view.You.Categories[i].Cards {
			card := &view.You.Categories[i].Cards[j]
			card.Label = localizedCardLabel(locale, card.ImgURL, card.Deck, card.Label)
		}
	}
	if view.World != nil {
		view.World.Disaster = localizeWorldCard(locale, view.World.Disaster)
		for i := range view.World.Bunker {
			view.World.Bunker[i] = localizeWorldFacedCard(locale, view.World.Bunker[i])
		}
		for i := range view.World.Threats {
			view.World.Threats[i] = localizeWorldFacedCard(locale, view.World.Threats[i])
		}
	}
	for i := range view.Public.Players {
		for j := range view.Public.Players[i].RevealedCards {
			card := &view.Public.Players[i].RevealedCards[j]
			card.Label = localizedCardLabel(locale, card.ID, card.Deck, card.Label)
		}
		for j := range view.Public.Players[i].Categories {
			for k := range view.Public.Players[i].Categories[j].Cards {
				card := &view.Public.Players[i].Categories[j].Cards[k]
				card.Label = localizedCardLabel(locale, card.ImgURL, "", card.Label)
			}
		}
	}
	for i := range view.Public.VotesPublic {
		view.Public.VotesPublic[i].Reason = formatServerTemplate(scenarioTextForLocale(scenarioID, locale, view.Public.VotesPublic[i].ReasonKey, view.Public.VotesPublic[i].Reason), view.Public.VotesPublic[i].ReasonVars)
	}
	return view
}

func localizeSpecialConditionView(locale, scenarioID string, special specialConditionInstanceView) specialConditionInstanceView {
	if strings.EqualFold(strings.TrimSpace(special.Effect.Type), devChoiceEffectType) {
		special.Title = scenarioTextForLocale(scenarioID, locale, "dev.choice.title", special.Title)
		special.Text = scenarioTextForLocale(scenarioID, locale, "dev.choice.text", special.Text)
		localizeSpecialOptionsParam(locale, scenarioID, special.Effect.Params)
		return special
	}

	hasDevSuffix := strings.HasSuffix(strings.TrimSpace(special.Title), " (DEV)")
	for _, key := range []string{
		normalizeSpecialConditionLocaleKey(special.ImgURL),
		normalizeSpecialConditionLocaleKey(special.ID),
		normalizeSpecialConditionLocaleKey(strings.TrimSuffix(strings.TrimSpace(special.Title), " (DEV)")),
	} {
		if key == "" {
			continue
		}
		if entry, ok := specialConditionLocaleEntryFor(scenarioID, locale, key); ok {
			if entry.Title != "" {
				special.Title = entry.Title
				if hasDevSuffix {
					special.Title += " (DEV)"
				}
			}
			if entry.Text != "" {
				special.Text = entry.Text
			}
			break
		}
	}
	localizeSpecialOptionsParam(locale, scenarioID, special.Effect.Params)
	return special
}

func localizeSpecialOptionsParam(locale, scenarioID string, params map[string]any) {
	if len(params) == 0 {
		return
	}
	rawOptions, ok := params["specialOptions"]
	if !ok {
		return
	}
	options, ok := rawOptions.([]map[string]any)
	if ok {
		for i := range options {
			localizeSpecialOptionMap(locale, scenarioID, options[i])
		}
		return
	}
	anyOptions, ok := rawOptions.([]any)
	if !ok {
		return
	}
	for _, item := range anyOptions {
		if option, ok := item.(map[string]any); ok {
			localizeSpecialOptionMap(locale, scenarioID, option)
		}
	}
}

func localizeSpecialOptionMap(locale, scenarioID string, option map[string]any) {
	if len(option) == 0 {
		return
	}
	for _, key := range []string{
		normalizeSpecialConditionLocaleKey(asString(option["assetId"])),
		normalizeSpecialConditionLocaleKey(asString(option["id"])),
		normalizeSpecialConditionLocaleKey(asString(option["title"])),
	} {
		if key == "" {
			continue
		}
		if entry, ok := specialConditionLocaleEntryFor(scenarioID, locale, key); ok && entry.Title != "" {
			option["title"] = entry.Title
			return
		}
	}
}

func localizeSpecialTitleVar(locale, scenarioID string, vars map[string]any) {
	if len(vars) == 0 {
		return
	}
	for _, key := range []string{
		normalizeSpecialConditionLocaleKey(asString(vars["assetId"])),
		normalizeSpecialConditionLocaleKey(asString(vars["specialId"])),
		normalizeSpecialConditionLocaleKey(asString(vars["title"])),
	} {
		if key == "" {
			continue
		}
		if entry, ok := specialConditionLocaleEntryFor(scenarioID, locale, key); ok && entry.Title != "" {
			vars["title"] = entry.Title
			return
		}
	}
}

func localizeGameEventForLocale(event gameEvent, locale, scenarioID string) gameEvent {
	out := event
	vars := cloneLocalizedVars(out.MessageVars)
	switch out.MessageKey {
	case "event.devChoice.selected":
		localizeSpecialTitleVar(locale, scenarioID, vars)
	}
	if out.MessageKey != "" {
		if strings.HasPrefix(out.MessageKey, "info.") || strings.HasPrefix(out.MessageKey, "error.") {
			out.Message = formatServerTemplate(serverText(locale, out.MessageKey, out.Message), vars)
		} else {
			out.Message = formatServerTemplate(scenarioTextForLocale(scenarioID, locale, out.MessageKey, out.Message), vars)
		}
	}
	out.MessageVars = vars
	return out
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
