package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pathHealth                = "/health"
	pathAPIScenarios          = "/api/scenarios"
	pathAssets                = "/assets/"
	writeTimeoutSeconds       = 2
	readLimitBytes      int64 = 1 << 20
	disconnectGraceMS   int64 = 300000
)

func normalizeOutcomeToken(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "survived", "success", "alive", "bunker_survived":
		return "survived"
	case "failed", "dead", "not_survived", "bunker_failed":
		return "failed"
	default:
		return strings.TrimSpace(raw)
	}
}

func normalizeIncomingGameAction(actionType string, payload map[string]any) (string, map[string]any) {
	if payload == nil {
		payload = map[string]any{}
	}

	action := strings.TrimSpace(actionType)
	switch action {
	case "setOutcome", "setBunkerResult":
		if _, ok := payload["outcome"]; !ok {
			if fromValue := normalizeOutcomeToken(asString(payload["value"])); fromValue != "" {
				payload["outcome"] = fromValue
			} else if fromResult := normalizeOutcomeToken(asString(payload["result"])); fromResult != "" {
				payload["outcome"] = fromResult
			}
		}
		if raw, ok := payload["outcome"].(string); ok {
			payload["outcome"] = normalizeOutcomeToken(raw)
		}
		return "setBunkerOutcome", payload
	case "setBunkerSurvived":
		payload["outcome"] = "survived"
		return "setBunkerOutcome", payload
	case "setBunkerFailed":
		payload["outcome"] = "failed"
		return "setBunkerOutcome", payload
	default:
		return action, payload
	}
}

func summarizePayload(payload map[string]any) string {
	if len(payload) == 0 {
		return "{}"
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return "{invalid_json_payload}"
	}
	const limit = 360
	if len(encoded) <= limit {
		return string(encoded)
	}
	return string(encoded[:limit]) + "...(truncated)"
}

func summarizeGameState(game roomGame) string {
	if game == nil {
		return "phase=none"
	}
	var g *gameSession
	switch current := game.(type) {
	case *gameSession:
		g = current
	case *devTestSession:
		g = current.core
	default:
		return "phase=unknown"
	}
	if g == nil {
		return "phase=none"
	}
	alive := 0
	for _, playerID := range g.Order {
		player := g.Players[playerID]
		if player != nil && player.Status == playerAlive {
			alive++
		}
	}

	timerInfo := "none"
	if g.ActiveTimer != nil {
		remaining := g.ActiveTimer.EndsAt - time.Now().UnixMilli()
		if remaining < 0 {
			remaining = 0
		}
		timerInfo = fmt.Sprintf("%s/%dms", g.ActiveTimer.Kind, remaining)
	}

	return fmt.Sprintf(
		"phase=%s votePhase=%s round=%d turn=%s alive=%d revealed=%d votes=%d baseVotes=%d voteRemaining=%d timer=%s",
		g.Phase,
		g.VotePhase,
		g.Round,
		g.CurrentTurnID,
		alive,
		len(g.RevealedThisRnd),
		len(g.Votes),
		len(g.BaseVotes),
		g.VotesRemaining,
		timerInfo,
	)
}

func newServer(cfg config) (*server, error) {
	if cfg.Host == "" {
		cfg.Host = "0.0.0.0"
	}
	if cfg.Port <= 0 {
		cfg.Port = 8080
	}
	if cfg.AssetsRoot == "" {
		cfg.AssetsRoot = filepath.Clean(filepath.Join("..", "assets"))
	}
	if cfg.ClientDistRoot == "" {
		cfg.ClientDistRoot = filepath.Clean(filepath.Join("..", "client", "dist"))
	}
	if cfg.ScenariosSourceRoot == "" {
		cfg.ScenariosSourceRoot = filepath.Clean(filepath.Join("..", "scenarios"))
	}
	cfg.IdentityMode = normalizeIdentityMode(cfg.IdentityMode)
	assets, err := loadAssetCatalog(cfg.AssetsRoot)
	if err != nil {
		return nil, err
	}
	if len(assets.Decks) == 0 {
		return nil, fmt.Errorf("assets catalog is empty in %s", cfg.AssetsRoot)
	}
	if _, err := os.Stat(filepath.Join(cfg.ClientDistRoot, "index.html")); err != nil {
		return nil, fmt.Errorf("client dist index not found at %s: %w", cfg.ClientDistRoot, err)
	}
	specialDefinitions := cloneSpecialDefinitions(implementedSpecialDefinitions)
	if specialsPath := resolveSpecialsFile(cfg); specialsPath != "" {
		loadedDefs, loadErr := loadImplementedSpecialDefinitionsFromFile(specialsPath)
		if loadErr != nil {
			log.Printf("[specials] failed to load %s: %v (fallback to built-in)", specialsPath, loadErr)
		} else {
			specialDefinitions = cloneSpecialDefinitions(loadedDefs)
			log.Printf("[specials] loaded %d implemented definitions from %s", len(specialDefinitions), specialsPath)
		}
	}

	srv := &server{
		cfg:                cfg,
		rooms:              map[string]*room{},
		connToID:           map[*websocket.Conn]connInfo{},
		assets:             assets,
		specialDefinitions: specialDefinitions,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
	return srv, nil
}

func (s *server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(pathHealth, s.handleHealth)
	mux.HandleFunc(pathAPIScenarios, s.handleScenarios)
	mux.Handle(pathAssets, s.assetsFileServer())
	mux.HandleFunc("/", s.handleClient)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if websocket.IsWebSocketUpgrade(r) {
			s.handleWebSocket(w, r)
			return
		}
		mux.ServeHTTP(w, r)
	})
}

func (s *server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"service": "protocol-bunker-host",
		"port":    s.cfg.Port,
		"mode":    "lan_only",
	})
}

func (s *server) handleScenarios(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, http.StatusOK, availableScenarios(s.cfg.EnableDevScenarios))
}

func (s *server) assetsFileServer() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, pathAssets)
		if decoded, err := url.PathUnescape(rel); err == nil {
			rel = decoded
		}
		rel = strings.TrimPrefix(path.Clean("/"+rel), "/")
		if rel == "." || rel == "" || strings.HasPrefix(rel, "..") {
			http.NotFound(w, r)
			return
		}

		if serveExistingFile(w, r, filepath.Join(s.cfg.ClientDistRoot, "assets", filepath.FromSlash(rel))) {
			return
		}
		if serveExistingFile(w, r, filepath.Join(s.cfg.AssetsRoot, filepath.FromSlash(rel))) {
			return
		}
		http.NotFound(w, r)
	})
}

func serveExistingFile(w http.ResponseWriter, r *http.Request, filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil || info.IsDir() {
		return false
	}
	http.ServeFile(w, r, filePath)
	return true
}

func (s *server) handleClient(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	clean := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if clean == "." {
		clean = "index.html"
	}
	if strings.HasPrefix(clean, "api"+string(filepath.Separator)) {
		http.NotFound(w, r)
		return
	}

	targetPath := filepath.Join(s.cfg.ClientDistRoot, clean)
	stat, err := os.Stat(targetPath)
	if err == nil && !stat.IsDir() {
		if clean == "index.html" {
			s.serveIndexHtml(w, r)
			return
		}
		http.ServeFile(w, r, targetPath)
		return
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		http.Error(w, "failed to read client file", http.StatusInternalServerError)
		return
	}

	s.serveIndexHtml(w, r)
}

func (s *server) serveIndexHtml(w http.ResponseWriter, r *http.Request) {
	indexPath := filepath.Join(s.cfg.ClientDistRoot, "index.html")
	if s.cfg.IdentityMode == "prod" {
		http.ServeFile(w, r, indexPath)
		return
	}

	body, err := os.ReadFile(indexPath)
	if err != nil {
		http.Error(w, "failed to read client index", http.StatusInternalServerError)
		return
	}

	page := injectIdentityModeScript(string(body), s.cfg.IdentityMode)
	w.Header().Set("content-type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write([]byte(page))
}

func injectIdentityModeScript(page string, mode string) string {
	if strings.Contains(page, "__BUNKER_IDENTITY_MODE__") {
		return page
	}
	normalizedMode := normalizeIdentityMode(mode)
	devTabFlag := "false"
	if normalizedMode == "dev_tab" {
		devTabFlag = "true"
	}
	script := fmt.Sprintf(
		`<script>window.__BUNKER_IDENTITY_MODE__=%q;window.__BUNKER_DEV_TAB_IDENTITY__=%s;</script>`,
		normalizedMode,
		devTabFlag,
	)
	if idx := strings.Index(strings.ToLower(page), "</head>"); idx >= 0 {
		return page[:idx] + script + page[idx:]
	}
	return script + page
}

func (s *server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	conn.SetReadLimit(readLimitBytes)

	go func() {
		defer func() {
			_ = conn.Close()
			s.handleDisconnect(conn)
		}()

		for {
			var raw map[string]any
			if err := conn.ReadJSON(&raw); err != nil {
				return
			}
			s.handleMessage(conn, raw)
		}
	}()
}

func (s *server) handleMessage(conn *websocket.Conn, raw map[string]any) {
	msgType, _ := raw["type"].(string)
	if msgType == "" {
		log.Printf("[ws] dropped message without type payload=%s", summarizePayload(raw))
		s.sendError(conn, "Неверный формат сообщения")
		return
	}

	payload := map[string]any{}
	if incomingPayload, ok := raw["payload"].(map[string]any); ok {
		payload = incomingPayload
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	switch msgType {
	case "hello":
		s.handleHelloLocked(conn, payload)
	case "resume":
		s.handleResumeLocked(conn, payload)
	case "startGame":
		s.handleStartGameLocked(conn)
	case "updateSettings":
		s.handleUpdateSettingsLocked(conn, payload)
	case "updateRules":
		s.handleUpdateRulesLocked(conn, payload)
	case "requestHostTransfer":
		s.handleHostTransferLocked(conn)
	case "kickFromLobby":
		s.handleKickFromLobbyLocked(conn, payload)
	case "ping":
		s.writeWSLocked(conn, wsServerMessage{Type: "pong", Payload: map[string]any{}})
	case "revealCard", "vote", "finalizeVoting", "applySpecial", "revealWorldThreat", "setBunkerOutcome", "setOutcome", "setBunkerResult", "setBunkerSurvived", "setBunkerFailed", "continueRound", "devSkipRound", "devKickPlayer", "devAddPlayer", "devRemovePlayer", "markLeftBunker":
		s.handleGameActionLocked(conn, msgType, payload)
	default:
		log.Printf("[ws] unknown message type=%s payload=%s", msgType, summarizePayload(payload))
		s.sendErrorLocked(conn, "Неизвестное сообщение")
	}
}

func (s *server) handleHelloLocked(conn *websocket.Conn, payload map[string]any) {
	var hello clientHelloPayload
	if !decodePayload(payload, &hello) {
		s.sendErrorLocked(conn, "Неверный формат hello")
		return
	}
	hello.Name = strings.TrimSpace(hello.Name)
	if hello.Name == "" {
		s.sendErrorLocked(conn, "Имя игрока обязательно")
		return
	}

	if hello.Create {
		if err := s.createRoomAndAttachLocked(conn, hello); err != nil {
			s.sendErrorLocked(conn, err.Error())
		}
		return
	}

	roomCode := strings.ToUpper(strings.TrimSpace(hello.RoomCode))
	if roomCode == "" {
		s.sendErrorLocked(conn, "Нужен roomCode")
		return
	}
	room := s.rooms[roomCode]
	if room == nil {
		s.sendErrorLocked(conn, "Комната не найдена")
		return
	}

	var existing *player
	if hello.PlayerToken != "" {
		if playerID, ok := room.PlayersByToken[hello.PlayerToken]; ok {
			existing = room.Players[playerID]
		}
	}
	if existing == nil && hello.SessionID != "" {
		if playerID, ok := room.PlayersBySession[hello.SessionID]; ok {
			existing = room.Players[playerID]
		}
	}

	if existing == nil && room.Phase == phaseGame {
		s.sendErrorLocked(conn, "Не удалось восстановить игрока. Перезайдите в комнату.")
		return
	}
	if existing == nil && len(room.Players) >= s.effectiveMaxPlayers(room) {
		maxPlayers := s.effectiveMaxPlayers(room)
		s.writeWSLocked(conn, wsServerMessage{
			Type: "error",
			Payload: map[string]any{
				"message":    fmt.Sprintf("Комната заполнена (макс %d).", maxPlayers),
				"code":       "ROOM_FULL",
				"maxPlayers": maxPlayers,
			},
		})
		return
	}

	joined := s.attachPlayerLocked(room, hello, conn, existing)
	if joined == nil {
		s.sendErrorLocked(conn, "Не удалось подключить игрока")
		return
	}
	s.broadcastRoomStateLocked(room)
	if room.Phase == phaseGame && room.Game != nil {
		s.broadcastGameViewsLocked(room)
	}
}

func (s *server) createRoomAndAttachLocked(conn *websocket.Conn, hello clientHelloPayload) error {
	scenarioID := strings.TrimSpace(hello.ScenarioID)
	if scenarioID == "" {
		return fmt.Errorf("Нужен scenarioId")
	}

	var scenario scenarioMeta
	scenarioFound := false
	for _, meta := range availableScenarios(s.cfg.EnableDevScenarios) {
		if meta.ID == scenarioID {
			scenario = meta
			scenarioFound = true
			break
		}
	}
	if !scenarioFound {
		return fmt.Errorf("Сценарий не найден")
	}

	roomCode := s.generateRoomCodeLocked()
	initialRuleset := buildAutoRuleset(minClassicPlayers)
	room := &room{
		Code:             roomCode,
		CreatedAtMS:      time.Now().UnixMilli(),
		Phase:            phaseLobby,
		Scenario:         scenario,
		Settings:         defaultSettingsForScenario(scenario.ID),
		Ruleset:          initialRuleset,
		RulesOverridden:  false,
		Players:          map[string]*player{},
		PlayersByToken:   map[string]string{},
		PlayersBySession: map[string]string{},
		JoinOrder:        []string{},
		IsDev:            scenario.ID == scenarioDevTest,
	}
	s.rooms[roomCode] = room
	log.Printf("[room] created room=%s scenario=%s", room.Code, room.Scenario.ID)

	joined := s.attachPlayerLocked(room, hello, conn, nil)
	if joined == nil {
		delete(s.rooms, room.Code)
		return fmt.Errorf("не удалось подключить создателя комнаты")
	}
	s.broadcastRoomStateLocked(room)
	return nil
}

func (s *server) handleResumeLocked(conn *websocket.Conn, payload map[string]any) {
	roomCode := strings.ToUpper(strings.TrimSpace(asString(payload["roomCode"])))
	sessionID := strings.TrimSpace(asString(payload["sessionId"]))
	if roomCode == "" || sessionID == "" {
		s.sendErrorLocked(conn, "Неверный resume payload")
		return
	}
	room := s.rooms[roomCode]
	if room == nil {
		s.sendErrorLocked(conn, "Комната не найдена")
		return
	}

	playerID, ok := room.PlayersBySession[sessionID]
	if !ok {
		s.sendErrorLocked(conn, "Не удалось восстановить игрока.")
		return
	}
	existing := room.Players[playerID]
	if existing == nil {
		s.sendErrorLocked(conn, "Не удалось восстановить игрока.")
		return
	}

	hello := clientHelloPayload{
		Name:        existing.Name,
		RoomCode:    room.Code,
		PlayerToken: existing.Token,
		TabID:       existing.TabID,
		SessionID:   sessionID,
	}
	joined := s.attachPlayerLocked(room, hello, conn, existing)
	if joined == nil {
		s.sendErrorLocked(conn, "Не удалось восстановить игрока.")
		return
	}
	s.broadcastRoomStateLocked(room)
	if room.Phase == phaseGame && room.Game != nil {
		s.broadcastGameViewsLocked(room)
	}
}

func (s *server) handleStartGameLocked(conn *websocket.Conn) {
	room, playerID := s.roomAndPlayerByConnLocked(conn)
	if room == nil {
		s.sendErrorLocked(conn, "Вы не в комнате")
		return
	}
	s.pruneLobbyDisconnectedPlayersLocked(room)
	if s.rooms[room.Code] == nil {
		s.sendErrorLocked(conn, "Room not found.")
		return
	}
	if room.Phase != phaseLobby {
		s.sendErrorLocked(conn, "Игра уже запущена")
		return
	}
	if playerID != room.ControlID {
		s.sendErrorLocked(conn, "Только CONTROL может начать игру")
		return
	}
	if room.Scenario.ID == scenarioClassic && len(room.Players) < minClassicPlayers {
		s.sendErrorLocked(conn, fmt.Sprintf("Для Classic нужно минимум %d игроков.", minClassicPlayers))
		return
	}

	players := make([]*player, 0, len(room.JoinOrder))
	for _, id := range room.JoinOrder {
		if p := room.Players[id]; p != nil {
			players = append(players, p)
		}
	}
	if room.Scenario.ID == scenarioDevTest {
		room.Game = newDevTestSession(
			room.Code,
			room.HostID,
			room.Settings,
			room.Ruleset,
			players,
			s.assets,
			s.specialDefinitions,
			time.Now().UnixNano(),
		)
	} else {
		room.Game = newGameSession(
			room.Code,
			room.HostID,
			room.Scenario.ID,
			room.Settings,
			room.Ruleset,
			players,
			s.assets,
			s.specialDefinitions,
			time.Now().UnixNano(),
		)
	}
	room.Phase = phaseGame
	for _, p := range room.Players {
		p.NeedsFullState = true
		p.NeedsFullGameView = true
	}

	s.broadcastRoomStateLocked(room)
	s.broadcastGameViewsLocked(room)
	s.broadcastEventLocked(room, gameEvent{
		ID:        fmt.Sprintf("%s-%d", room.Code, time.Now().UnixMilli()),
		Kind:      "roundStart",
		Message:   "Игра началась.",
		CreatedAt: time.Now().UnixMilli(),
	})
	s.rescheduleRoomGameTimerLocked(room)
}

func (s *server) handleUpdateSettingsLocked(conn *websocket.Conn, payload map[string]any) {
	room, playerID := s.roomAndPlayerByConnLocked(conn)
	if room == nil {
		s.sendErrorLocked(conn, "Вы не в комнате")
		return
	}
	if room.Phase != phaseLobby {
		s.sendErrorLocked(conn, "Настройки доступны только в лобби.")
		return
	}
	if playerID != room.ControlID {
		s.sendErrorLocked(conn, "Только CONTROL может менять настройки.")
		return
	}

	next := room.Settings
	if !decodePayload(payload, &next) {
		s.sendErrorLocked(conn, "Неверный формат настроек")
		return
	}

	minAllowed := 2
	if room.Scenario.ID == scenarioClassic {
		minAllowed = minClassicPlayers
	}
	next.MaxPlayers = clampInt(next.MaxPlayers, minAllowed, maxClassicPlayers)
	next.EnablePresenterMode = false
	if next.MaxPlayers < len(room.Players) {
		s.sendErrorLocked(conn, "Лимит игроков меньше текущего числа.")
		return
	}
	room.Settings = next
	s.broadcastRoomStateLocked(room)
}

func (s *server) handleUpdateRulesLocked(conn *websocket.Conn, payload map[string]any) {
	room, playerID := s.roomAndPlayerByConnLocked(conn)
	if room == nil {
		s.sendErrorLocked(conn, "Вы не в комнате")
		return
	}
	if room.Scenario.ID != scenarioClassic {
		s.sendErrorLocked(conn, "Правила доступны только для Classic.")
		return
	}
	if room.Phase != phaseLobby {
		s.sendErrorLocked(conn, "Правила можно менять только в лобби.")
		return
	}
	if playerID != room.ControlID {
		s.sendErrorLocked(conn, "Только CONTROL может менять правила.")
		return
	}

	mode := strings.ToLower(strings.TrimSpace(asString(payload["mode"])))
	switch mode {
	case "auto":
		room.RulesOverridden = false
		room.RulesPresetCount = nil
		room.Ruleset = buildAutoRuleset(len(room.Players))
	case "manual":
		room.RulesOverridden = true
		presetCount := clampInt(asInt(payload["presetPlayerCount"], len(room.Players)), minClassicPlayers, maxClassicPlayers)
		room.RulesPresetCount = &presetCount

		manualRaw, hasManual := payload["manualConfig"]
		if !hasManual {
			seed := seedManualConfigFromPreset(presetCount)
			room.Ruleset = buildManualRuleset(seed, len(room.Players))
			break
		}
		manualMap, ok := manualRaw.(map[string]any)
		if !ok {
			s.sendErrorLocked(conn, "Неверный manualConfig")
			return
		}
		var manualCfg manualRulesConfig
		if !decodePayload(manualMap, &manualCfg) {
			s.sendErrorLocked(conn, "Неверный manualConfig")
			return
		}
		manualCfg = normalizeManualConfig(manualCfg, presetCount)
		room.Ruleset = buildManualRuleset(manualCfg, len(room.Players))
		room.RulesPresetCount = manualCfg.SeedTemplatePlayer
	default:
		s.sendErrorLocked(conn, "Неизвестный режим правил.")
		return
	}

	s.broadcastRoomStateLocked(room)
}

func (s *server) handleHostTransferLocked(conn *websocket.Conn) {
	room, playerID := s.roomAndPlayerByConnLocked(conn)
	if room == nil {
		s.sendErrorLocked(conn, "Вы не в комнате")
		return
	}
	if playerID != room.ControlID {
		s.sendErrorLocked(conn, "Только CONTROL может передать роль.")
		return
	}
	next := s.pickNextHostLocked(room, room.HostID)
	if next == "" {
		s.sendErrorLocked(conn, "Нет другого игрока для передачи роли.")
		return
	}
	s.transferHostLocked(room, next, "manual")
}

func (s *server) handleKickFromLobbyLocked(conn *websocket.Conn, payload map[string]any) {
	room, playerID := s.roomAndPlayerByConnLocked(conn)
	if room == nil {
		s.sendErrorLocked(conn, "Вы не в комнате")
		return
	}
	if room.Phase != phaseLobby {
		s.sendErrorLocked(conn, "Команда доступна только в лобби.")
		return
	}
	if playerID != room.ControlID {
		s.sendErrorLocked(conn, "Только CONTROL может кикать игроков.")
		return
	}

	targetID := strings.TrimSpace(asString(payload["targetPlayerId"]))
	if targetID == "" {
		s.sendErrorLocked(conn, "Нужно указать targetPlayerId.")
		return
	}
	if targetID == room.HostID {
		s.sendErrorLocked(conn, "Нельзя кикнуть хоста.")
		return
	}
	target := room.Players[targetID]
	if target == nil {
		s.sendErrorLocked(conn, "Игрок не найден.")
		return
	}
	if target.Connection != nil {
		_ = target.Connection.Close()
	}
	s.removeLobbyPlayerLocked(room, targetID)
	if s.rooms[room.Code] != nil {
		s.broadcastRoomStateLocked(room)
	}
}

func (s *server) handleGameActionLocked(conn *websocket.Conn, actionType string, payload map[string]any) {
	room, playerID := s.roomAndPlayerByConnLocked(conn)
	if room == nil {
		s.sendErrorLocked(conn, "Вы не в комнате")
		return
	}
	if room.Phase != phaseGame || room.Game == nil {
		s.sendErrorLocked(conn, "Игра не найдена")
		return
	}
	actionType, payload = normalizeIncomingGameAction(actionType, payload)
	beforeState := summarizeGameState(room.Game)

	controlOnly := map[string]bool{
		"finalizeVoting":   true,
		"setBunkerOutcome": true,
		"devSkipRound":     true,
		"devKickPlayer":    true,
		"devAddPlayer":     true,
		"devRemovePlayer":  true,
		"markLeftBunker":   true,
	}
	if controlOnly[actionType] && playerID != room.ControlID {
		s.sendErrorLocked(conn, "Действие доступно только роли CONTROL.")
		return
	}
	log.Printf(
		"[action] room=%s player=%s scenario=%s type=%s payload=%s state_before=%s",
		room.Code,
		playerID,
		room.Scenario.ID,
		actionType,
		summarizePayload(payload),
		beforeState,
	)

	result := room.Game.handleAction(playerID, actionType, payload)
	if result.Error != "" {
		log.Printf(
			"[action] room=%s player=%s type=%s error=%s payload=%s state_after=%s",
			room.Code,
			playerID,
			actionType,
			result.Error,
			summarizePayload(payload),
			summarizeGameState(room.Game),
		)
		s.sendErrorLocked(conn, result.Error)
		return
	}
	if !result.StateChanged {
		log.Printf(
			"[action] room=%s player=%s type=%s no_state_change state_after=%s",
			room.Code,
			playerID,
			actionType,
			summarizeGameState(room.Game),
		)
		return
	}
	log.Printf(
		"[action] room=%s player=%s type=%s state_changed events=%d state_after=%s",
		room.Code,
		playerID,
		actionType,
		len(result.Events),
		summarizeGameState(room.Game),
	)

	s.broadcastGameViewsLocked(room)
	for _, event := range result.Events {
		s.broadcastEventLocked(room, event)
	}
	s.rescheduleRoomGameTimerLocked(room)
}

func (s *server) attachPlayerLocked(room *room, hello clientHelloPayload, conn *websocket.Conn, existing *player) *player {
	now := time.Now().UnixMilli()
	if existing == nil {
		playerID := s.newPlayerIDLocked(room)
		token := randomToken(24)
		existing = &player{
			ID:                playerID,
			Name:              hello.Name,
			Token:             token,
			TabID:             hello.TabID,
			SessionID:         strings.TrimSpace(hello.SessionID),
			Connection:        conn,
			Connected:         true,
			NeedsFullState:    true,
			NeedsFullGameView: true,
		}
		if existing.SessionID == "" {
			existing.SessionID = randomToken(12)
		}
		room.Players[playerID] = existing
		room.PlayersByToken[token] = playerID
		room.PlayersBySession[existing.SessionID] = playerID
		room.JoinOrder = append(room.JoinOrder, playerID)
		if room.HostID == "" {
			room.HostID = playerID
		}
		if room.ControlID == "" {
			room.ControlID = playerID
		}
	} else {
		if existing.Connection != nil && existing.Connection != conn {
			delete(s.connToID, existing.Connection)
			_ = existing.Connection.Close()
		}
		s.finalizeReconnectLocked(existing, now)
		existing.Name = hello.Name
		if sid := strings.TrimSpace(hello.SessionID); sid != "" {
			if existing.SessionID != "" {
				delete(room.PlayersBySession, existing.SessionID)
			}
			existing.SessionID = sid
			room.PlayersBySession[sid] = existing.ID
		}
		existing.Connection = conn
		existing.Connected = true
		existing.NeedsFullState = true
		existing.NeedsFullGameView = true
	}

	s.connToID[conn] = connInfo{RoomCode: room.Code, PlayerID: existing.ID}
	s.writeWSLocked(conn, wsServerMessage{
		Type: "helloAck",
		Payload: map[string]any{
			"playerId":    existing.ID,
			"playerToken": existing.Token,
		},
	})

	return existing
}

func (s *server) roomAndPlayerByConnLocked(conn *websocket.Conn) (*room, string) {
	info, ok := s.connToID[conn]
	if !ok {
		return nil, ""
	}
	room := s.rooms[info.RoomCode]
	if room == nil {
		return nil, ""
	}
	if room.Players[info.PlayerID] == nil {
		return nil, ""
	}
	return room, info.PlayerID
}

func (s *server) newPlayerIDLocked(room *room) string {
	for {
		id := "p_" + randomToken(6)
		if room.Players[id] == nil {
			return id
		}
	}
}

func (s *server) generateRoomCodeLocked() string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	for {
		builder := strings.Builder{}
		builder.Grow(4)
		seed := randomToken(6)
		for i := 0; i < 4; i++ {
			builder.WriteByte(alphabet[int(seed[i])%len(alphabet)])
		}
		code := builder.String()
		if s.rooms[code] == nil {
			return code
		}
	}
}

func (s *server) effectiveMaxPlayers(room *room) int {
	maxPlayers := room.Settings.MaxPlayers
	if room.Scenario.ID == scenarioClassic {
		return clampInt(maxPlayers, minClassicPlayers, maxClassicPlayers)
	}
	return clampInt(maxPlayers, 2, 64)
}

func (s *server) removeLobbyPlayerLocked(room *room, playerID string) {
	player := room.Players[playerID]
	if player == nil {
		return
	}
	log.Printf("[lobby] remove player room=%s id=%s name=%s connected=%v", room.Code, player.ID, player.Name, player.Connected)

	if player.Connection != nil {
		delete(s.connToID, player.Connection)
	}
	delete(room.Players, playerID)
	delete(room.PlayersByToken, player.Token)
	if player.SessionID != "" {
		delete(room.PlayersBySession, player.SessionID)
	}
	room.JoinOrder = slices.DeleteFunc(room.JoinOrder, func(id string) bool { return id == playerID })

	if len(room.Players) == 0 {
		s.cancelRoomGameTimerLocked(room)
		delete(s.rooms, room.Code)
		return
	}

	if room.HostID == playerID {
		room.HostID = s.pickNextHostLocked(room, playerID)
	}
	if room.ControlID == playerID {
		room.ControlID = room.HostID
	}
	if room.HostID == "" {
		room.HostID = room.JoinOrder[0]
	}
	if room.ControlID == "" {
		room.ControlID = room.HostID
	}
	if room.Scenario.ID == scenarioClassic && !room.RulesOverridden {
		room.Ruleset = buildAutoRuleset(len(room.Players))
	}
}

func (s *server) pruneLobbyDisconnectedPlayersLocked(room *room) {
	if room == nil || room.Phase != phaseLobby {
		return
	}

	ids := append([]string(nil), room.JoinOrder...)
	for _, playerID := range ids {
		player := room.Players[playerID]
		if player == nil {
			continue
		}
		if player.Connected && player.Connection != nil {
			continue
		}
		log.Printf(
			"[lobby] prune disconnected room=%s id=%s name=%s connected=%v hasConn=%v",
			room.Code,
			player.ID,
			player.Name,
			player.Connected,
			player.Connection != nil,
		)
		s.removeLobbyPlayerLocked(room, playerID)
		if s.rooms[room.Code] == nil {
			return
		}
	}
}

func (s *server) pickNextHostLocked(room *room, excludeID string) string {
	for _, id := range room.JoinOrder {
		if id == excludeID {
			continue
		}
		if room.Players[id] != nil {
			return id
		}
	}
	return ""
}

func (s *server) transferHostLocked(room *room, newHostID string, reason string) {
	if room.Players[newHostID] == nil || room.HostID == newHostID {
		return
	}
	room.HostID = newHostID
	room.ControlID = newHostID
	if room.Game != nil {
		room.Game.setHostID(newHostID)
	}
	s.broadcastRoomStateLocked(room)
	for _, p := range room.Players {
		if p.Connection == nil {
			continue
		}
		s.writeWSLocked(p.Connection, wsServerMessage{
			Type: "hostChanged",
			Payload: map[string]any{
				"newHostId": newHostID,
				"reason":    reason,
			},
		})
	}
}

func (s *server) clearDisconnectTimerLocked(player *player) {
	if player == nil || player.DisconnectTimer == nil {
		return
	}
	player.DisconnectTimer.Stop()
	player.DisconnectTimer = nil
}

func (s *server) finalizeReconnectLocked(player *player, now int64) {
	if player == nil {
		return
	}
	if player.DisconnectedAtMS != nil {
		delta := maxInt64(0, now-*player.DisconnectedAtMS)
		player.TotalAbsentMS += delta
	}
	player.DisconnectedAtMS = nil
	player.DisconnectVersion++
	s.clearDisconnectTimerLocked(player)
}

func (s *server) computeKickRemainingMSLocked(player *player, now int64) int64 {
	if player == nil {
		return 0
	}
	currentOffline := int64(0)
	if !player.Connected && player.DisconnectedAtMS != nil {
		currentOffline = maxInt64(0, now-*player.DisconnectedAtMS)
	}
	return maxInt64(0, disconnectGraceMS-(player.TotalAbsentMS+currentOffline))
}

func (s *server) playerShouldBeKickedOnDisconnectLocked(room *room, player *player) bool {
	if room == nil || player == nil || room.Phase != phaseGame || room.Game == nil {
		return false
	}
	status, ok := room.Game.playerStatus(player.ID)
	return ok && status == playerAlive
}

func (s *server) markLeftBunkerByDisconnectLocked(room *room, targetID string) {
	if room == nil || room.Game == nil || targetID == "" {
		return
	}

	actorID := room.ControlID
	if room.Players[actorID] == nil {
		actorID = room.HostID
	}
	if room.Players[actorID] == nil {
		for _, id := range room.JoinOrder {
			if room.Players[id] != nil {
				actorID = id
				break
			}
		}
	}
	if actorID == "" {
		return
	}

	result := room.Game.handleAction(actorID, "markLeftBunker", map[string]any{
		"targetPlayerId": targetID,
	})
	if result.Error != "" {
		log.Printf("[disconnect] markLeftBunker failed room=%s player=%s err=%s", room.Code, targetID, result.Error)
		return
	}
	if result.StateChanged {
		s.broadcastGameViewsLocked(room)
		for _, event := range result.Events {
			s.broadcastEventLocked(room, event)
		}
		s.rescheduleRoomGameTimerLocked(room)
	}
}

func (s *server) scheduleDisconnectKickLocked(room *room, player *player) {
	if !s.playerShouldBeKickedOnDisconnectLocked(room, player) {
		return
	}
	remaining := s.computeKickRemainingMSLocked(player, time.Now().UnixMilli())
	if remaining <= 0 {
		s.markLeftBunkerByDisconnectLocked(room, player.ID)
		return
	}

	s.clearDisconnectTimerLocked(player)
	player.DisconnectVersion++
	version := player.DisconnectVersion
	roomCode := room.Code
	playerID := player.ID
	player.DisconnectTimer = time.AfterFunc(time.Duration(remaining)*time.Millisecond, func() {
		s.mu.Lock()
		defer s.mu.Unlock()

		room := s.rooms[roomCode]
		if room == nil {
			return
		}
		player := room.Players[playerID]
		if player == nil || player.Connected {
			return
		}
		if player.DisconnectVersion != version {
			return
		}
		if !s.playerShouldBeKickedOnDisconnectLocked(room, player) {
			return
		}
		if s.computeKickRemainingMSLocked(player, time.Now().UnixMilli()) > 0 {
			return
		}
		s.markLeftBunkerByDisconnectLocked(room, playerID)
		s.broadcastRoomStateLocked(room)
	})
}

func (s *server) handleDisconnect(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, ok := s.connToID[conn]
	if !ok {
		return
	}
	delete(s.connToID, conn)

	room := s.rooms[info.RoomCode]
	if room == nil {
		return
	}
	player := room.Players[info.PlayerID]
	if player == nil {
		return
	}
	s.handlePlayerConnectionLostLocked(room, player, conn)
}

func (s *server) handlePlayerConnectionLostLocked(room *room, player *player, conn *websocket.Conn) {
	if room == nil || player == nil || player.Connection != conn {
		return
	}

	player.Connection = nil
	player.Connected = false
	now := time.Now().UnixMilli()
	player.DisconnectedAtMS = &now
	s.clearDisconnectTimerLocked(player)
	player.DisconnectVersion++
	log.Printf("[ws] disconnect room=%s player=%s phase=%s", room.Code, player.ID, room.Phase)

	if room.Phase == phaseLobby {
		s.removeLobbyPlayerLocked(room, player.ID)
		if s.rooms[room.Code] != nil {
			s.broadcastRoomStateLocked(room)
		}
		return
	}

	s.scheduleDisconnectKickLocked(room, player)
	if room.HostID == player.ID {
		next := s.pickNextHostLocked(room, player.ID)
		if next != "" {
			s.transferHostLocked(room, next, "disconnect_timeout")
		}
	}
	s.broadcastRoomStateLocked(room)
	if room.Game != nil {
		s.broadcastGameViewsLocked(room)
	}
}

func (s *server) buildRoomStateLocked(room *room) roomState {
	players := make([]playerSummary, 0, len(room.JoinOrder))
	for _, playerID := range room.JoinOrder {
		player := room.Players[playerID]
		if player == nil {
			continue
		}
		summary := playerSummary{
			PlayerID:  player.ID,
			Name:      player.Name,
			Connected: player.Connected,
		}
		if player.DisconnectedAtMS != nil {
			value := *player.DisconnectedAtMS
			summary.DisconnectedAt = &value
		}
		totalAbsent := player.TotalAbsentMS
		summary.TotalAbsentMS = &totalAbsent
		currentOffline := int64(0)
		if !player.Connected && player.DisconnectedAtMS != nil {
			currentOffline = maxInt64(0, time.Now().UnixMilli()-*player.DisconnectedAtMS)
		}
		summary.CurrentOffMS = &currentOffline
		remaining := maxInt64(0, disconnectGraceMS-(totalAbsent+currentOffline))
		summary.KickRemainMS = &remaining
		if room.Game != nil {
			if status, ok := room.Game.playerStatus(player.ID); ok && status == playerLeftBunker {
				flag := true
				summary.LeftBunker = &flag
			}
		}
		players = append(players, summary)
	}

	state := roomState{
		RoomCode:            room.Code,
		Players:             players,
		HostID:              room.HostID,
		ControlID:           room.ControlID,
		Phase:               room.Phase,
		ScenarioMeta:        room.Scenario,
		Settings:            room.Settings,
		Ruleset:             room.Ruleset,
		RulesOverriddenHost: room.RulesOverridden,
		RulesPresetCount:    room.RulesPresetCount,
	}
	if room.IsDev {
		flag := true
		state.IsDev = &flag
	}
	return state
}

func (s *server) broadcastRoomStateLocked(room *room) {
	s.pruneLobbyDisconnectedPlayersLocked(room)
	if room == nil || s.rooms[room.Code] == nil {
		return
	}
	state := s.buildRoomStateLocked(room)
	for _, p := range room.Players {
		if p.Connection == nil {
			continue
		}
		if p.NeedsFullState || !room.LastRoomStateSent {
			s.writeWSLocked(p.Connection, wsServerMessage{Type: "roomState", Payload: state})
		} else {
			s.writeWSLocked(p.Connection, wsServerMessage{
				Type: "statePatch",
				Payload: map[string]any{
					"roomState": state,
				},
			})
		}
		p.NeedsFullState = false
	}
	room.LastRoomStateSent = true
}

func (s *server) broadcastGameViewsLocked(room *room) {
	if room.Game == nil {
		return
	}
	for _, playerID := range room.JoinOrder {
		player := room.Players[playerID]
		if player == nil || player.Connection == nil {
			continue
		}
		view := room.Game.buildGameView(room, player.ID)
		if player.NeedsFullGameView {
			s.writeWSLocked(player.Connection, wsServerMessage{Type: "gameView", Payload: view})
		} else {
			s.writeWSLocked(player.Connection, wsServerMessage{
				Type: "statePatch",
				Payload: map[string]any{
					"gameView": view,
				},
			})
		}
		player.NeedsFullGameView = false
	}
}

func (s *server) broadcastEventLocked(room *room, event gameEvent) {
	for _, player := range room.Players {
		if player.Connection == nil {
			continue
		}
		s.writeWSLocked(player.Connection, wsServerMessage{
			Type:    "gameEvent",
			Payload: event,
		})
	}
}

func (s *server) cancelRoomGameTimerLocked(room *room) {
	if room == nil || room.GameTimer == nil {
		return
	}
	room.GameTimer.Stop()
	room.GameTimer = nil
}

func (s *server) rescheduleRoomGameTimerLocked(room *room) {
	if room == nil {
		return
	}
	s.cancelRoomGameTimerLocked(room)
	if room.Phase != phaseGame || room.Game == nil {
		return
	}
	timer := room.Game.currentTimer()
	if timer == nil {
		return
	}
	delay := time.Until(time.UnixMilli(timer.EndsAt))
	if delay < 0 {
		delay = 0
	}
	room.GameTimerVersion++
	version := room.GameTimerVersion
	roomCode := room.Code
	room.GameTimer = time.AfterFunc(delay, func() {
		s.handleRoomGameTimer(roomCode, version)
	})
}

func (s *server) handleRoomGameTimer(roomCode string, version int64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room := s.rooms[roomCode]
	if room == nil {
		return
	}
	if room.GameTimerVersion != version {
		return
	}
	room.GameTimer = nil
	if room.Phase != phaseGame || room.Game == nil {
		return
	}

	result := room.Game.handleTimerExpired(time.Now().UnixMilli())
	if result.Error != "" {
		log.Printf("[timer] room=%s timer action failed: %s", room.Code, result.Error)
		s.rescheduleRoomGameTimerLocked(room)
		return
	}
	if result.StateChanged {
		s.broadcastGameViewsLocked(room)
		for _, event := range result.Events {
			s.broadcastEventLocked(room, event)
		}
	}
	s.rescheduleRoomGameTimerLocked(room)
}

func (s *server) writeWSLocked(conn *websocket.Conn, message wsServerMessage) {
	if conn == nil {
		return
	}
	_ = conn.SetWriteDeadline(time.Now().Add(writeTimeoutSeconds * time.Second))
	if err := conn.WriteJSON(message); err != nil {
		info, ok := s.connToID[conn]
		if ok {
			delete(s.connToID, conn)
			if room := s.rooms[info.RoomCode]; room != nil {
				if player := room.Players[info.PlayerID]; player != nil {
					s.handlePlayerConnectionLostLocked(room, player, conn)
				}
			}
		}
		_ = conn.Close()
	}
}

func (s *server) sendError(conn *websocket.Conn, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sendErrorLocked(conn, message)
}

func (s *server) sendErrorLocked(conn *websocket.Conn, message string) {
	safeMessage := sanitizeHumanText(message, "Server error.")
	s.writeWSLocked(conn, wsServerMessage{
		Type: "error",
		Payload: map[string]any{
			"message": safeMessage,
		},
	})
}

func decodePayload(payload any, out any) bool {
	bytes, err := json.Marshal(payload)
	if err != nil {
		return false
	}
	if err := json.Unmarshal(bytes, out); err != nil {
		return false
	}
	return true
}

func asString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case json.Number:
		return v.String()
	default:
		return ""
	}
}

func asInt(value any, fallback int) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case float32:
		return int(v)
	case int:
		return v
	case int64:
		return int(v)
	case json.Number:
		if i, err := v.Int64(); err == nil {
			return int(i)
		}
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i
		}
	}
	return fallback
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("content-type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
