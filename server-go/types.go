package main

import (
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	phaseLobby = "lobby"
	phaseGame  = "game"

	scenarioClassic = "classic"
	scenarioDevTest = "dev_test"

	playerAlive      = "alive"
	playerEliminated = "eliminated"
	playerLeftBunker = "left_bunker"
)

var (
	coreDecks      = []string{"Профессия", "Здоровье", "Хобби", "Багаж", "Биология"}
	factsDeck      = "Факты"
	factsSlotOrder = []string{"facts1", "facts2"}
	categoryOrder  = []string{
		"Профессия",
		"Здоровье",
		"Хобби",
		"Багаж",
		"Факт №1",
		"Факт №2",
		"Биология",
		"Особые условия",
	}
)

type scenarioMeta struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	DevOnly     bool   `json:"devOnly,omitempty"`
}

type gameSettings struct {
	EnableRevealDiscussionTimer bool   `json:"enableRevealDiscussionTimer"`
	RevealDiscussionSeconds     int    `json:"revealDiscussionSeconds"`
	EnablePreVoteDiscussionTime bool   `json:"enablePreVoteDiscussionTimer"`
	PreVoteDiscussionSeconds    int    `json:"preVoteDiscussionSeconds"`
	EnablePostVoteDiscussion    bool   `json:"enablePostVoteDiscussionTimer"`
	PostVoteDiscussionSeconds   int    `json:"postVoteDiscussionSeconds"`
	EnablePresenterMode         bool   `json:"enablePresenterMode"`
	ContinuePermission          string `json:"continuePermission"`
	RevealTimeoutAction         string `json:"revealTimeoutAction"`
	RevealsBeforeVoting         int    `json:"revealsBeforeVoting"`
	SpecialUsage                string `json:"specialUsage"`
	MaxPlayers                  int    `json:"maxPlayers"`
	FinalThreatReveal           string `json:"finalThreatReveal"`
}

type manualRulesConfig struct {
	BunkerSlots        int   `json:"bunkerSlots"`
	VotesByRound       []int `json:"votesByRound"`
	TargetReveals      int   `json:"targetReveals"`
	SeedTemplatePlayer *int  `json:"seedTemplatePlayers,omitempty"`
}

type gameRuleset struct {
	PlayerCount int                `json:"playerCount"`
	VotesPerRnd []int              `json:"votesPerRound"`
	TotalExiles int                `json:"totalExiles"`
	BunkerSeats int                `json:"bunkerSeats"`
	RulesetMode string             `json:"rulesetMode"`
	ManualCfg   *manualRulesConfig `json:"manualConfig,omitempty"`
}

type playerSummary struct {
	PlayerID       string `json:"playerId"`
	Name           string `json:"name"`
	Connected      bool   `json:"connected"`
	DisconnectedAt *int64 `json:"disconnectedAt,omitempty"`
	TotalAbsentMS  *int64 `json:"totalAbsentMs,omitempty"`
	CurrentOffMS   *int64 `json:"currentOfflineMs,omitempty"`
	KickRemainMS   *int64 `json:"kickRemainingMs,omitempty"`
	LeftBunker     *bool  `json:"leftBunker,omitempty"`
}

type roomState struct {
	RoomCode            string          `json:"roomCode"`
	Players             []playerSummary `json:"players"`
	HostID              string          `json:"hostId"`
	ControlID           string          `json:"controlId"`
	Phase               string          `json:"phase"`
	ScenarioMeta        scenarioMeta    `json:"scenarioMeta"`
	Settings            gameSettings    `json:"settings"`
	Ruleset             gameRuleset     `json:"ruleset"`
	RulesOverriddenHost bool            `json:"rulesOverriddenByHost"`
	RulesPresetCount    *int            `json:"rulesPresetCount,omitempty"`
	IsDev               *bool           `json:"isDev,omitempty"`
}

type cardRef struct {
	ID        string `json:"id"`
	Deck      string `json:"deck"`
	Instance  string `json:"instanceId,omitempty"`
	Label     string `json:"labelShort,omitempty"`
	Secret    bool   `json:"secret,omitempty"`
	Missing   bool   `json:"missing,omitempty"`
	Revealed  bool   `json:"revealed"`
	SlotKey   string `json:"-"`
	AssetPath string `json:"-"`
}

type publicCategoryCard struct {
	Label    string `json:"labelShort"`
	ImgURL   string `json:"imgUrl,omitempty"`
	Revealed bool   `json:"revealed"`
}

type publicCategorySlot struct {
	Category string               `json:"category"`
	Status   string               `json:"status"`
	Cards    []publicCategoryCard `json:"cards"`
}

type youCategoryCard struct {
	InstanceID string `json:"instanceId"`
	Label      string `json:"labelShort"`
	Revealed   bool   `json:"revealed"`
}

type youCategorySlot struct {
	Category string            `json:"category"`
	Cards    []youCategoryCard `json:"cards"`
}

type publicPlayerView struct {
	PlayerID        string               `json:"playerId"`
	Name            string               `json:"name"`
	Status          string               `json:"status"`
	Connected       bool                 `json:"connected"`
	DisconnectedAt  *int64               `json:"disconnectedAt,omitempty"`
	TotalAbsentMS   *int64               `json:"totalAbsentMs,omitempty"`
	CurrentOffMS    *int64               `json:"currentOfflineMs,omitempty"`
	KickRemainMS    *int64               `json:"kickRemainingMs,omitempty"`
	LeftBunker      bool                 `json:"leftBunker,omitempty"`
	RevealedCards   []cardRef            `json:"revealedCards"`
	RevealedCount   int                  `json:"revealedCount"`
	TotalCards      int                  `json:"totalCards"`
	SpecialRevealed bool                 `json:"specialRevealed"`
	Categories      []publicCategorySlot `json:"categories"`
}

type votingView struct {
	HasVoted bool `json:"hasVoted"`
}

type votingProgress struct {
	Voted int `json:"voted"`
	Total int `json:"total"`
}

type votePublic struct {
	VoterID    string `json:"voterId"`
	VoterName  string `json:"voterName"`
	TargetID   string `json:"targetId,omitempty"`
	TargetName string `json:"targetName,omitempty"`
	Status     string `json:"status"`
	Reason     string `json:"reason,omitempty"`
	Submitted  int64  `json:"submittedAt,omitempty"`
}

type roundRulesPublic struct {
	NoTalkUntilVoting bool   `json:"noTalkUntilVoting,omitempty"`
	ForcedCategory    string `json:"forcedRevealCategory,omitempty"`
}

type gameTimerState struct {
	Kind   string `json:"kind"`
	EndsAt int64  `json:"endsAt"`
}

type specialConditionEffectView struct {
	Type   string         `json:"type"`
	Params map[string]any `json:"params,omitempty"`
}

type specialConditionInstanceView struct {
	InstanceID      string                     `json:"instanceId"`
	ID              string                     `json:"id"`
	Title           string                     `json:"title"`
	Text            string                     `json:"text"`
	Trigger         string                     `json:"trigger"`
	Effect          specialConditionEffectView `json:"effect"`
	Implemented     bool                       `json:"implemented"`
	RevealedPublic  bool                       `json:"revealedPublic"`
	Used            bool                       `json:"used"`
	ImgURL          string                     `json:"imgUrl,omitempty"`
	NeedsChoice     bool                       `json:"needsChoice,omitempty"`
	ChoiceKind      string                     `json:"choiceKind,omitempty"`
	AllowSelfTarget bool                       `json:"allowSelfTarget,omitempty"`
	TargetScope     string                     `json:"targetScope,omitempty"`
}

type worldCardView struct {
	Kind        string `json:"kind"`
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Text        string `json:"text,omitempty"`
	ImageID     string `json:"imageId,omitempty"`
}

type worldFacedCardView struct {
	Kind            string `json:"kind"`
	ID              string `json:"id"`
	Title           string `json:"title"`
	Description     string `json:"description"`
	Text            string `json:"text,omitempty"`
	ImageID         string `json:"imageId,omitempty"`
	IsRevealed      bool   `json:"isRevealed"`
	RevealedAtRound *int   `json:"revealedAtRound,omitempty"`
	RevealedBy      string `json:"revealedBy,omitempty"`
}

type worldStateView struct {
	Disaster worldCardView        `json:"disaster"`
	Bunker   []worldFacedCardView `json:"bunker"`
	Threats  []worldFacedCardView `json:"threats"`
	Counts   struct {
		Bunker  int `json:"bunker"`
		Threats int `json:"threats"`
	} `json:"counts"`
}

type worldEventView struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Round int    `json:"round"`
}

type postGameStateView struct {
	IsActive  bool   `json:"isActive"`
	EnteredAt int64  `json:"enteredAt"`
	Outcome   string `json:"outcome,omitempty"`
	DecidedBy string `json:"decidedBy,omitempty"`
	DecidedAt int64  `json:"decidedAt,omitempty"`
}

type threatModifierView struct {
	Delta      int      `json:"delta"`
	Reasons    []string `json:"reasons"`
	BaseCount  int      `json:"baseCount"`
	FinalCount int      `json:"finalCount"`
}

type gameView struct {
	Phase      string             `json:"phase"`
	Round      int                `json:"round"`
	Categories []string           `json:"categoryOrder"`
	LastStage  string             `json:"lastStageText,omitempty"`
	Ruleset    gameRuleset        `json:"ruleset"`
	World      *worldStateView    `json:"world,omitempty"`
	WorldEvent *worldEventView    `json:"worldEvent,omitempty"`
	PostGame   *postGameStateView `json:"postGame,omitempty"`
	You        struct {
		PlayerID   string                         `json:"playerId"`
		Name       string                         `json:"name"`
		Hand       []cardRef                      `json:"hand"`
		Categories []youCategorySlot              `json:"categories"`
		Specials   []specialConditionInstanceView `json:"specialConditions"`
	} `json:"you"`
	Public struct {
		Players             []publicPlayerView  `json:"players"`
		RevealedThisRound   []string            `json:"revealedThisRound"`
		RoundRevealedCount  int                 `json:"roundRevealedCount,omitempty"`
		RoundTotalAlive     int                 `json:"roundTotalAlive,omitempty"`
		CurrentTurnPlayerID *string             `json:"currentTurnPlayerId,omitempty"`
		VotesRemaining      int                 `json:"votesRemainingInRound,omitempty"`
		VotesTotal          int                 `json:"votesTotalThisRound,omitempty"`
		RevealLimit         int                 `json:"revealLimit,omitempty"`
		Voting              *votingView         `json:"voting,omitempty"`
		VotePhase           *string             `json:"votePhase,omitempty"`
		VotesPublic         []votePublic        `json:"votesPublic,omitempty"`
		VotingProgress      *votingProgress     `json:"votingProgress,omitempty"`
		ThreatModifier      *threatModifierView `json:"threatModifier,omitempty"`
		CanOpenVotingModal  bool                `json:"canOpenVotingModal,omitempty"`
		CanContinue         bool                `json:"canContinue,omitempty"`
		VoteModalOpen       bool                `json:"voteModalOpen,omitempty"`
		ActiveTimer         *gameTimerState     `json:"activeTimer,omitempty"`
		LastEliminated      string              `json:"lastEliminated,omitempty"`
		Winners             []string            `json:"winners,omitempty"`
		ResolutionNote      string              `json:"resolutionNote,omitempty"`
		RoundRules          *roundRulesPublic   `json:"roundRules,omitempty"`
	} `json:"public"`
}

type gameEvent struct {
	ID        string `json:"id"`
	Kind      string `json:"kind"`
	Message   string `json:"message"`
	CreatedAt int64  `json:"createdAt"`
}

type wsServerMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type clientHelloPayload struct {
	Name        string `json:"name"`
	RoomCode    string `json:"roomCode,omitempty"`
	Create      bool   `json:"create,omitempty"`
	ScenarioID  string `json:"scenarioId,omitempty"`
	PlayerToken string `json:"playerToken,omitempty"`
	TabID       string `json:"tabId,omitempty"`
	SessionID   string `json:"sessionId,omitempty"`
}

type player struct {
	ID                string
	Name              string
	Token             string
	TabID             string
	SessionID         string
	Connection        *websocket.Conn
	Connected         bool
	DisconnectedAtMS  *int64
	TotalAbsentMS     int64
	DisconnectTimer   *time.Timer
	DisconnectVersion int64
	NeedsFullState    bool
	NeedsFullGameView bool
}

type room struct {
	Code              string
	CreatedAtMS       int64
	HostID            string
	ControlID         string
	Phase             string
	Scenario          scenarioMeta
	Settings          gameSettings
	Ruleset           gameRuleset
	RulesOverridden   bool
	RulesPresetCount  *int
	Players           map[string]*player
	PlayersByToken    map[string]string
	PlayersBySession  map[string]string
	JoinOrder         []string
	Game              roomGame
	IsDev             bool
	LastRoomStateSent bool
	GameTimer         *time.Timer
	GameTimerVersion  int64
}

type roomGame interface {
	handleAction(actorID, actionType string, payload map[string]any) gameActionResult
	handleTimerExpired(now int64) gameActionResult
	buildGameView(room *room, playerID string) gameView
	currentTimer() *gameTimerState
	setHostID(hostID string)
	playerStatus(playerID string) (string, bool)
}

type server struct {
	cfg                config
	mu                 sync.Mutex
	rooms              map[string]*room
	connToID           map[*websocket.Conn]connInfo
	assets             assetCatalog
	specialDefinitions []specialDefinition
	upgrader           websocket.Upgrader
}

type connInfo struct {
	RoomCode string
	PlayerID string
}

type config struct {
	Host                string
	Port                int
	AssetsRoot          string
	ClientDistRoot      string
	IdentityMode        string
	ScenariosSourceRoot string
	SpecialsFile        string
	EnableDevScenarios  bool
}

type assetCard struct {
	ID    string
	Deck  string
	Label string
}

type assetCatalog struct {
	Decks map[string][]assetCard
}

type httpResponseErr struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type serverWithWS struct {
	httpHandler http.Handler
	server      *server
}
