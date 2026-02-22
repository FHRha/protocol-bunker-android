package main

type specialEffect struct {
	Type   string
	Params map[string]any
}

type specialDefinition struct {
	ID          string
	Title       string
	Text        string
	Trigger     string
	Effect      specialEffect
	Implemented bool
	Requires    []string
	ChoiceKind  string
	TargetScope string
	AllowSelf   bool
	AssetID     string
}

type specialConditionState struct {
	InstanceID     string
	Definition     specialDefinition
	RevealedPublic bool
	Used           bool
}

var implementedSpecialDefinitions = []specialDefinition{
	{
		ID:      "ban_vote_against",
		Title:   "Будь Другом",
		Text:    "Выбранный игрок до конца игры не голосует против тебя.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "banVoteAgainst",
			Params: map[string]any{"scope": "untilGameEnd"},
		},
		Implemented: true,
		Requires:    []string{"phase=voting"},
		ChoiceKind:  "player",
		TargetScope: "any_alive",
	},
	{
		ID:      "vote_weight",
		Title:   "Громкий Голос",
		Text:    "Твой голос считается за два в этом голосовании.",
		Trigger: "onVote",
		Effect: specialEffect{
			Type:   "voteWeight",
			Params: map[string]any{"weight": 2, "scope": "thisVoting"},
		},
		Implemented: true,
		Requires:    []string{"phase=voting"},
		ChoiceKind:  "none",
	},
	{
		ID:      "redeal_baggage",
		Title:   "Давайте На Чистоту Багажа",
		Text:    "Перераздать все раскрытые карты багажа между живыми игроками.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "redealAllRevealed",
			Params: map[string]any{"category": "baggage", "scope": "aliveOnly"},
		},
		Implemented: true,
		Requires:    []string{"phase=any"},
		ChoiceKind:  "none",
	},
	{
		ID:      "redeal_biology",
		Title:   "Давайте На Чистоту Биология",
		Text:    "Перераздать все раскрытые карты биологии между живыми игроками.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "redealAllRevealed",
			Params: map[string]any{"category": "biology", "scope": "aliveOnly"},
		},
		Implemented: true,
		Requires:    []string{"phase=any"},
		ChoiceKind:  "none",
	},
	{
		ID:      "redeal_health",
		Title:   "Давайте На Чистоту Здоровья",
		Text:    "Перераздать все раскрытые карты здоровья между живыми игроками.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "redealAllRevealed",
			Params: map[string]any{"category": "health", "scope": "aliveOnly"},
		},
		Implemented: true,
		Requires:    []string{"phase=any"},
		ChoiceKind:  "none",
	},
	{
		ID:      "redeal_facts",
		Title:   "Давайте На Чистоту Фактов",
		Text:    "Перераздать раскрытые факты между живыми игроками.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "redealAllRevealed",
			Params: map[string]any{"category": "facts", "scope": "aliveOnly"},
		},
		Implemented: true,
		Requires:    []string{"phase=any"},
		ChoiceKind:  "none",
	},
	{
		ID:      "redeal_hobby",
		Title:   "Давайте На Чистоту Хобби",
		Text:    "Перераздать раскрытые карты хобби между живыми игроками.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "redealAllRevealed",
			Params: map[string]any{"category": "hobby", "scope": "aliveOnly"},
		},
		Implemented: true,
		Requires:    []string{"phase=any"},
		ChoiceKind:  "none",
	},
	{
		ID:      "disable_vote",
		Title:   "Дискредитация",
		Text:    "Лиши выбранного игрока голоса в текущем голосовании.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "disableVote",
			Params: map[string]any{"scope": "thisVoting"},
		},
		Implemented: true,
		Requires:    []string{"phase=voting"},
		ChoiceKind:  "player",
		TargetScope: "any_alive",
	},
	{
		ID:      "protect_left",
		Title:   "Защити Игрока Слева",
		Text:    "Если левый сосед исключён, твой следующий голос тратится.",
		Trigger: "secret_onEliminate",
		Effect: specialEffect{
			Type:   "forcedWastedVoteOnNextVoting",
			Params: map[string]any{"condition": "leftNeighborEliminated"},
		},
		Implemented: true,
		Requires:    []string{"needsNeighborIndexing"},
		ChoiceKind:  "none",
	},
	{
		ID:      "protect_right",
		Title:   "Защити Игрока Справа",
		Text:    "Если правый сосед исключён, твой следующий голос тратится.",
		Trigger: "secret_onEliminate",
		Effect: specialEffect{
			Type:   "forcedWastedVoteOnNextVoting",
			Params: map[string]any{"condition": "rightNeighborEliminated"},
		},
		Implemented: true,
		Requires:    []string{"needsNeighborIndexing"},
		ChoiceKind:  "none",
	},
	{
		ID:      "protect_youngest",
		Title:   "Защити Младшего",
		Text:    "Если исключён самый младший по раскрытому возрасту, твой следующий голос тратится.",
		Trigger: "secret_onEliminate",
		Effect: specialEffect{
			Type:   "forcedWastedVoteOnNextVoting",
			Params: map[string]any{"condition": "youngestByRevealedAgeEliminated"},
		},
		Implemented: true,
		Requires:    []string{"ageFieldAvailable", "someRevealedAges"},
		ChoiceKind:  "none",
	},
	{
		ID:      "protect_brave",
		Title:   "Защити Смелого",
		Text:    "Если исключён первый раскрывший здоровье, твой следующий голос тратится.",
		Trigger: "secret_onEliminate",
		Effect: specialEffect{
			Type:   "forcedWastedVoteOnNextVoting",
			Params: map[string]any{"condition": "firstRevealedHealthEliminated"},
		},
		Implemented: true,
		Requires:    []string{"trackFirstRevealHealth"},
		ChoiceKind:  "none",
	},
	{
		ID:      "protect_oldest",
		Title:   "Защити Старшего",
		Text:    "Если исключён самый старший по раскрытому возрасту, твой следующий голос тратится.",
		Trigger: "secret_onEliminate",
		Effect: specialEffect{
			Type:   "forcedWastedVoteOnNextVoting",
			Params: map[string]any{"condition": "oldestByRevealedAgeEliminated"},
		},
		Implemented: true,
		Requires:    []string{"ageFieldAvailable", "someRevealedAges"},
		ChoiceKind:  "none",
	},
	{
		ID:      "compromat",
		Title:   "Компромат",
		Text:    "Голоса против выбранного игрока считаются двойными, твой голос тратится.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "doubleVotesAgainst_and_disableSelfVote",
			Params: map[string]any{"scope": "thisVoting"},
		},
		Implemented: true,
		Requires:    []string{"phase=voting"},
		ChoiceKind:  "player",
		TargetScope: "any_alive",
	},
	{
		ID:      "need_more",
		Title:   "Мне Нужнее",
		Text:    "Забери багаж у выбранного игрока и выдай ему новое спецусловие.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "stealBaggage_and_giveSpecial",
			Params: map[string]any{"giveSpecialCount": 1},
		},
		Implemented: true,
		Requires:    []string{"targetHasBaggage"},
		ChoiceKind:  "player",
		TargetScope: "any_alive",
	},
	{
		ID:      "silence",
		Title:   "Молчание",
		Text:    "До голосования нельзя говорить.",
		Trigger: "onRevealOrActive",
		Effect: specialEffect{
			Type:   "setRoundRule",
			Params: map[string]any{"noTalkUntilVoting": true, "scope": "thisRound"},
		},
		Implemented: true,
		Requires:    []string{"phase=reveal"},
		ChoiceKind:  "none",
	},
	{
		ID:      "swap_baggage",
		Title:   "Обмен Карт Багаж",
		Text:    "Поменяйся раскрытой картой багажа с соседом.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "swapRevealedWithNeighbor",
			Params: map[string]any{"category": "baggage"},
		},
		Implemented: true,
		Requires:    []string{"targetHasRevealedSameCategory", "needsNeighborIndexing"},
		ChoiceKind:  "neighbor",
		TargetScope: "neighbors",
	},
	{
		ID:      "swap_biology",
		Title:   "Обмен Карт Биология",
		Text:    "Поменяйся раскрытой картой биологии с соседом.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "swapRevealedWithNeighbor",
			Params: map[string]any{"category": "biology"},
		},
		Implemented: true,
		Requires:    []string{"targetHasRevealedSameCategory", "needsNeighborIndexing"},
		ChoiceKind:  "neighbor",
		TargetScope: "neighbors",
	},
	{
		ID:      "swap_health",
		Title:   "Обмен Карт Здоровье",
		Text:    "Поменяйся раскрытой картой здоровья с соседом.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "swapRevealedWithNeighbor",
			Params: map[string]any{"category": "health"},
		},
		Implemented: true,
		Requires:    []string{"targetHasRevealedSameCategory", "needsNeighborIndexing"},
		ChoiceKind:  "neighbor",
		TargetScope: "neighbors",
	},
	{
		ID:      "swap_facts",
		Title:   "Обмен Карт Фактов",
		Text:    "Поменяйся раскрытой картой факта с соседом.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "swapRevealedWithNeighbor",
			Params: map[string]any{"category": "facts"},
		},
		Implemented: true,
		Requires:    []string{"targetHasRevealedSameCategory", "needsNeighborIndexing"},
		ChoiceKind:  "neighbor",
		TargetScope: "neighbors",
	},
	{
		ID:      "swap_hobby",
		Title:   "Обмен Карт Хобби",
		Text:    "Поменяйся раскрытой картой хобби с соседом.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "swapRevealedWithNeighbor",
			Params: map[string]any{"category": "hobby"},
		},
		Implemented: true,
		Requires:    []string{"targetHasRevealedSameCategory", "needsNeighborIndexing"},
		ChoiceKind:  "neighbor",
		TargetScope: "neighbors",
	},
	{
		ID:      "plan_b",
		Title:   "План Б",
		Text:    "Запускает переголосование. Лидеров предыдущего голоса нельзя выбрать.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "forceRevote",
			Params: map[string]any{"disallowPreviousCandidate": true},
		},
		Implemented: true,
		Requires:    []string{"phase=voting", "votingStarted"},
		ChoiceKind:  "none",
	},
	{
		ID:      "replace_health",
		Title:   "Просроченные Таблетки",
		Text:    "Замени раскрытую карту здоровья у выбранного игрока.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "replaceRevealedCard",
			Params: map[string]any{"category": "health", "source": "deckRandom"},
		},
		Implemented: true,
		Requires:    []string{"targetHasRevealedHealth"},
		ChoiceKind:  "player",
		TargetScope: "any_alive",
	},
	{
		ID:      "force_category",
		Title:   "Прямой Вопрос",
		Text:    "Выбери категорию, которую все обязаны раскрыть в этом раунде.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "forceRevealCategoryForAll",
			Params: map[string]any{"scope": "thisRound", "onlyIfHidden": true},
		},
		Implemented: true,
		Requires:    []string{"phase=reveal"},
		ChoiceKind:  "category",
	},
	{
		ID:      "final_threat",
		Title:   "Тайная Угроза",
		Text:    "При исключении владельца добавляет финальную угрозу.",
		Trigger: "onOwnerEliminated",
		Effect: specialEffect{
			Type:   "addFinalThreat",
			Params: map[string]any{"threatKey": "raiders_know_bunker"},
		},
		Implemented: true,
		Requires:    []string{"ownerEliminated"},
		ChoiceKind:  "none",
	},
	{
		ID:      "replace_profession",
		Title:   "Фейковый Диплом",
		Text:    "Замени раскрытую карту профессии у выбранного игрока.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "replaceRevealedCard",
			Params: map[string]any{"category": "profession", "source": "deckRandom"},
		},
		Implemented: true,
		Requires:    []string{"targetHasRevealedProfession"},
		ChoiceKind:  "player",
		TargetScope: "any_alive",
	},
	{
		ID:      "discard_health",
		Title:   "Хорошие Таблетки",
		Text:    "Сбрасывает раскрытую карту здоровья и выдаёт новую скрытую.",
		Trigger: "active",
		Effect: specialEffect{
			Type:   "discardRevealedAndDealHidden",
			Params: map[string]any{"category": "health", "dealFrom": "healthDeck"},
		},
		Implemented: true,
		Requires:    []string{"targetHasRevealedHealth"},
		ChoiceKind:  "player",
		TargetScope: "any_alive",
	},
}
