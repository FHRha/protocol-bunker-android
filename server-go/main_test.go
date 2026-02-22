package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestServer_HealthAndScenarios(t *testing.T) {
	assetsRoot, clientRoot := makeTestRoots(t)
	srv, err := newServer(config{
		Host:               "127.0.0.1",
		Port:               18080,
		AssetsRoot:         assetsRoot,
		ClientDistRoot:     clientRoot,
		EnableDevScenarios: false,
	})
	if err != nil {
		t.Fatalf("newServer failed: %v", err)
	}
	testServer := httptest.NewServer(srv.routes())
	defer testServer.Close()

	healthResp, err := http.Get(testServer.URL + "/health")
	if err != nil {
		t.Fatalf("health request failed: %v", err)
	}
	defer healthResp.Body.Close()
	if healthResp.StatusCode != http.StatusOK {
		t.Fatalf("health status %d", healthResp.StatusCode)
	}

	var health map[string]any
	if err := json.NewDecoder(healthResp.Body).Decode(&health); err != nil {
		t.Fatalf("health decode failed: %v", err)
	}
	if health["status"] != "ok" {
		t.Fatalf("unexpected health payload: %+v", health)
	}

	scenariosResp, err := http.Get(testServer.URL + "/api/scenarios")
	if err != nil {
		t.Fatalf("scenarios request failed: %v", err)
	}
	defer scenariosResp.Body.Close()
	if scenariosResp.StatusCode != http.StatusOK {
		t.Fatalf("scenarios status %d", scenariosResp.StatusCode)
	}

	var scenarios []scenarioMeta
	if err := json.NewDecoder(scenariosResp.Body).Decode(&scenarios); err != nil {
		t.Fatalf("decode scenarios failed: %v", err)
	}
	if len(scenarios) == 0 || scenarios[0].ID != scenarioClassic {
		t.Fatalf("unexpected scenarios: %+v", scenarios)
	}
}

func TestServer_ClientIndexInjectsIdentityModeInDevTab(t *testing.T) {
	assetsRoot, clientRoot := makeTestRoots(t)
	srv, err := newServer(config{
		Host:               "127.0.0.1",
		Port:               18082,
		AssetsRoot:         assetsRoot,
		ClientDistRoot:     clientRoot,
		IdentityMode:       "dev_tab",
		EnableDevScenarios: true,
	})
	if err != nil {
		t.Fatalf("newServer failed: %v", err)
	}
	testServer := httptest.NewServer(srv.routes())
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/")
	if err != nil {
		t.Fatalf("root request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	html := string(body)
	if !strings.Contains(html, "__BUNKER_IDENTITY_MODE__=\"dev_tab\"") {
		t.Fatalf("identity mode script not injected: %s", html)
	}
	if !strings.Contains(html, "__BUNKER_DEV_TAB_IDENTITY__=true") {
		t.Fatalf("dev tab flag not injected: %s", html)
	}
}

func TestServer_DevScenarioDefaultsToHostContinuePermission(t *testing.T) {
	assetsRoot, clientRoot := makeTestRoots(t)
	srv, err := newServer(config{
		Host:               "127.0.0.1",
		Port:               18083,
		AssetsRoot:         assetsRoot,
		ClientDistRoot:     clientRoot,
		EnableDevScenarios: true,
	})
	if err != nil {
		t.Fatalf("newServer failed: %v", err)
	}
	testServer := httptest.NewServer(srv.routes())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	conn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}
	defer conn.Close()

	mustSend(t, conn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":       "Host",
			"create":     true,
			"scenarioId": scenarioDevTest,
			"sessionId":  "sess-dev-host",
		},
	})
	_ = mustReadType(t, conn, "helloAck")
	roomStateMsg := mustReadRoomState(t, conn)
	if roomStateMsg.Settings.ContinuePermission != "host_only" {
		t.Fatalf("dev scenario continuePermission mismatch: %s", roomStateMsg.Settings.ContinuePermission)
	}
}

func TestServer_WSDevScenario_DossierHasCards(t *testing.T) {
	assetsRoot, clientRoot := makeTestRoots(t)
	srv, err := newServer(config{
		Host:               "127.0.0.1",
		Port:               18084,
		AssetsRoot:         assetsRoot,
		ClientDistRoot:     clientRoot,
		EnableDevScenarios: true,
	})
	if err != nil {
		t.Fatalf("newServer failed: %v", err)
	}
	testServer := httptest.NewServer(srv.routes())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	clients := make([]*websocket.Conn, 0, 4)
	defer func() {
		for _, c := range clients {
			_ = c.Close()
		}
	}()

	hostConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("host dial failed: %v", err)
	}
	clients = append(clients, hostConn)

	mustSend(t, hostConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":       "Host",
			"create":     true,
			"scenarioId": scenarioDevTest,
			"sessionId":  "sess-dev-host-2",
		},
	})
	_ = mustReadType(t, hostConn, "helloAck")
	hostRoomState := mustReadRoomState(t, hostConn)
	roomCode := hostRoomState.RoomCode
	if roomCode == "" {
		t.Fatalf("room code must not be empty")
	}

	for i := 0; i < 3; i++ {
		conn, err := dialWS(wsURL)
		if err != nil {
			t.Fatalf("player dial failed: %v", err)
		}
		clients = append(clients, conn)

		mustSend(t, conn, map[string]any{
			"type": "hello",
			"payload": map[string]any{
				"name":      fmt.Sprintf("P%d", i+1),
				"roomCode":  roomCode,
				"sessionId": fmt.Sprintf("sess-dev-%d", i+1),
			},
		})
		_ = mustReadType(t, conn, "helloAck")
		_ = mustReadRoomState(t, conn)
	}

	mustSend(t, hostConn, map[string]any{
		"type":    "startGame",
		"payload": map[string]any{},
	})
	hostView := mustReadGameView(t, hostConn)

	if len(hostView.You.Categories) != len(categoryOrder)-1 {
		t.Fatalf("unexpected dossier categories count: got=%d want=%d", len(hostView.You.Categories), len(categoryOrder)-1)
	}
	if len(hostView.You.Specials) == 0 {
		t.Fatalf("expected at least one special condition in dossier")
	}
	for _, slot := range hostView.You.Categories {
		if len(slot.Cards) == 0 {
			t.Fatalf("dossier category %q has no cards", slot.Category)
		}
		for _, card := range slot.Cards {
			if card.InstanceID == "" {
				t.Fatalf("dossier category %q has card with empty instanceId", slot.Category)
			}
		}
	}
}

func TestServer_WSDevScenario_UsesDedicatedRuntimeActionSet(t *testing.T) {
	assetsRoot, clientRoot := makeTestRoots(t)
	srv, err := newServer(config{
		Host:               "127.0.0.1",
		Port:               18085,
		AssetsRoot:         assetsRoot,
		ClientDistRoot:     clientRoot,
		EnableDevScenarios: true,
	})
	if err != nil {
		t.Fatalf("newServer failed: %v", err)
	}
	testServer := httptest.NewServer(srv.routes())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	clients := make([]*websocket.Conn, 0, 4)
	defer func() {
		for _, c := range clients {
			_ = c.Close()
		}
	}()

	hostConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("host dial failed: %v", err)
	}
	clients = append(clients, hostConn)

	mustSend(t, hostConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":       "Host",
			"create":     true,
			"scenarioId": scenarioDevTest,
			"sessionId":  "sess-dev-runtime",
		},
	})
	_ = mustReadType(t, hostConn, "helloAck")
	hostRoomState := mustReadRoomState(t, hostConn)
	roomCode := hostRoomState.RoomCode

	for i := 0; i < 3; i++ {
		conn, err := dialWS(wsURL)
		if err != nil {
			t.Fatalf("player dial failed: %v", err)
		}
		clients = append(clients, conn)
		mustSend(t, conn, map[string]any{
			"type": "hello",
			"payload": map[string]any{
				"name":      fmt.Sprintf("P%d", i+1),
				"roomCode":  roomCode,
				"sessionId": fmt.Sprintf("sess-dev-runtime-%d", i+1),
			},
		})
		_ = mustReadType(t, conn, "helloAck")
		_ = mustReadRoomState(t, conn)
	}

	mustSend(t, hostConn, map[string]any{
		"type":    "startGame",
		"payload": map[string]any{},
	})
	_ = mustReadGameView(t, hostConn)

	mustSend(t, hostConn, map[string]any{
		"type": "setBunkerOutcome",
		"payload": map[string]any{
			"outcome": "survived",
		},
	})
	errMsg := mustReadType(t, hostConn, "error")
	message := mustNestedString(t, errMsg, "payload", "message")
	if message == "" {
		t.Fatalf("expected error message for unsupported dev action")
	}
}

func TestServer_WSGameFlow(t *testing.T) {
	assetsRoot, clientRoot := makeTestRoots(t)
	srv, err := newServer(config{
		Host:               "127.0.0.1",
		Port:               18081,
		AssetsRoot:         assetsRoot,
		ClientDistRoot:     clientRoot,
		EnableDevScenarios: false,
	})
	if err != nil {
		t.Fatalf("newServer failed: %v", err)
	}
	testServer := httptest.NewServer(srv.routes())
	defer testServer.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	clients := make([]*websocket.Conn, 0, 4)
	cleanup := func() {
		for _, c := range clients {
			_ = c.Close()
		}
	}
	defer cleanup()

	hostConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("host dial failed: %v", err)
	}
	clients = append(clients, hostConn)

	mustSend(t, hostConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":       "Host",
			"create":     true,
			"scenarioId": scenarioClassic,
			"sessionId":  "sess-host",
		},
	})

	hostAck := mustReadType(t, hostConn, "helloAck")
	hostPlayerID := mustNestedString(t, hostAck, "payload", "playerId")
	roomStateMsg := mustReadRoomState(t, hostConn)
	roomCode := roomStateMsg.RoomCode
	if roomCode == "" {
		t.Fatalf("room code must not be empty")
	}

	playerByID := map[string]*websocket.Conn{hostPlayerID: hostConn}
	for i := 0; i < 3; i++ {
		conn, err := dialWS(wsURL)
		if err != nil {
			t.Fatalf("player %d dial failed: %v", i, err)
		}
		clients = append(clients, conn)
		mustSend(t, conn, map[string]any{
			"type": "hello",
			"payload": map[string]any{
				"name":        fmt.Sprintf("P%d", i+1),
				"roomCode":    roomCode,
				"sessionId":   fmt.Sprintf("sess-%d", i+1),
				"playerToken": "",
			},
		})
		ack := mustReadType(t, conn, "helloAck")
		playerID := mustNestedString(t, ack, "payload", "playerId")
		playerByID[playerID] = conn
		_ = mustReadRoomState(t, conn)
	}

	// Force voting in the first round to keep the test short.
	mustSend(t, hostConn, map[string]any{
		"type": "updateRules",
		"payload": map[string]any{
			"mode": "manual",
			"manualConfig": map[string]any{
				"bunkerSlots":         2,
				"votesByRound":        []int{1},
				"targetReveals":       7,
				"seedTemplatePlayers": 4,
			},
		},
	})
	_ = mustReadRoomState(t, hostConn)

	// Start game.
	mustSend(t, hostConn, map[string]any{"type": "startGame", "payload": map[string]any{}})
	hostView := mustReadGameView(t, hostConn)
	if hostView.Phase != scenarioPhaseReveal {
		t.Fatalf("expected reveal phase, got %s", hostView.Phase)
	}
	viewsByPlayer := map[string]gameView{hostPlayerID: hostView}
	for playerID, conn := range playerByID {
		if playerID == hostPlayerID {
			continue
		}
		viewsByPlayer[playerID] = mustReadGameView(t, conn)
	}

	// Reveal cycle for first round.
	turnSafety := 0
	for hostView.Phase != scenarioPhaseVoting && turnSafety < 12 {
		turnSafety++
		currentTurn := ""
		if hostView.Public.CurrentTurnPlayerID != nil {
			currentTurn = *hostView.Public.CurrentTurnPlayerID
		}
		if currentTurn == "" {
			t.Fatalf("currentTurn is empty while phase=%s", hostView.Phase)
		}
		actorConn := playerByID[currentTurn]
		if actorConn == nil {
			t.Fatalf("no connection for currentTurn=%s", currentTurn)
		}

		actorView := viewsByPlayer[currentTurn]
		cardID := firstHiddenCardID(actorView)
		if cardID == "" {
			t.Fatalf("no hidden card available for player=%s", currentTurn)
		}

		mustSend(t, actorConn, map[string]any{
			"type":    "revealCard",
			"payload": map[string]any{"cardId": cardID},
		})
		viewsByPlayer[currentTurn] = mustReadGameView(t, actorConn)
		mustSend(t, actorConn, map[string]any{
			"type":    "continueRound",
			"payload": map[string]any{},
		})
		viewsByPlayer[currentTurn] = mustReadGameView(t, actorConn)
		hostView = viewsByPlayer[currentTurn]
	}

	if hostView.Phase != scenarioPhaseVoting {
		t.Fatalf("expected voting phase, got %s", hostView.Phase)
	}
	if hostView.Public.VotePhase == nil || *hostView.Public.VotePhase != votePhaseVoting {
		t.Fatalf("expected votePhase=voting, got %+v", hostView.Public.VotePhase)
	}

	alive := make([]string, 0, len(hostView.Public.Players))
	for _, p := range hostView.Public.Players {
		if p.Status == playerAlive {
			alive = append(alive, p.PlayerID)
		}
	}
	if len(alive) < minClassicPlayers {
		t.Fatalf("expected at least %d alive players, got %d", minClassicPlayers, len(alive))
	}

	for idx, voterID := range alive {
		targetID := alive[(idx+1)%len(alive)]
		voterConn := playerByID[voterID]
		mustSend(t, voterConn, map[string]any{
			"type":    "vote",
			"payload": map[string]any{"targetPlayerId": targetID},
		})
	}

	hostView = mustReadGameViewUntil(t, hostConn, func(v gameView) bool {
		return v.Public.VotePhase != nil && *v.Public.VotePhase == votePhaseSpecialWindow
	})

	mustSend(t, hostConn, map[string]any{"type": "finalizeVoting", "payload": map[string]any{}})
	hostView = mustReadGameViewUntil(t, hostConn, func(v gameView) bool {
		if v.Phase == scenarioPhaseResolution || v.Phase == scenarioPhaseEnded {
			return true
		}
		return v.Phase == scenarioPhaseVoting && v.Public.VotePhase != nil && *v.Public.VotePhase == votePhaseVoting
	})

	if hostView.Phase == scenarioPhaseVoting && hostView.Public.VotePhase != nil && *hostView.Public.VotePhase == votePhaseVoting {
		alive = alive[:0]
		for _, p := range hostView.Public.Players {
			if p.Status == playerAlive {
				alive = append(alive, p.PlayerID)
			}
		}
		if len(alive) < 2 {
			t.Fatalf("expected at least 2 alive players in revote, got %d", len(alive))
		}
		focusTarget := alive[0]
		for _, voterID := range alive {
			targetID := focusTarget
			if voterID == focusTarget {
				targetID = alive[1]
			}
			voterConn := playerByID[voterID]
			mustSend(t, voterConn, map[string]any{
				"type":    "vote",
				"payload": map[string]any{"targetPlayerId": targetID},
			})
		}
		hostView = mustReadGameViewUntil(t, hostConn, func(v gameView) bool {
			return v.Public.VotePhase != nil && *v.Public.VotePhase == votePhaseSpecialWindow
		})
		mustSend(t, hostConn, map[string]any{"type": "finalizeVoting", "payload": map[string]any{}})
		hostView = mustReadGameViewUntil(t, hostConn, func(v gameView) bool {
			return v.Phase == scenarioPhaseResolution || v.Phase == scenarioPhaseEnded
		})
	}
	if hostView.Phase != scenarioPhaseResolution && hostView.Phase != scenarioPhaseEnded {
		t.Fatalf("expected resolution|ended, got %s", hostView.Phase)
	}
}

func makeTestRoots(t *testing.T) (string, string) {
	t.Helper()
	tmp := t.TempDir()
	assetsRoot := filepath.Join(tmp, "assets")
	clientRoot := filepath.Join(tmp, "client", "dist")
	if err := os.MkdirAll(filepath.Join(clientRoot, "assets"), 0o755); err != nil {
		t.Fatalf("mkdir client failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(clientRoot, "index.html"), []byte("<!doctype html><html><body>ok</body></html>"), 0o644); err != nil {
		t.Fatalf("write index failed: %v", err)
	}

	decks := []string{"Профессия", "Здоровье", "Хобби", "Багаж", "Биология", "Факты"}
	for _, deck := range decks {
		deckDir := filepath.Join(assetsRoot, "decks", deck)
		if err := os.MkdirAll(deckDir, 0o755); err != nil {
			t.Fatalf("mkdir deck failed: %v", err)
		}
		cardCount := 5
		if deck == "Факты" {
			cardCount = 12
		}
		for i := 0; i < cardCount; i++ {
			name := fmt.Sprintf("card_%s_%02d.jpg", strings.ToLower(deck), i+1)
			if err := os.WriteFile(filepath.Join(deckDir, name), []byte("card"), 0o644); err != nil {
				t.Fatalf("write card failed: %v", err)
			}
		}
	}
	return assetsRoot, clientRoot
}

func postJSON(
	t *testing.T,
	method string,
	url string,
	payload map[string]any,
	wantStatus int,
) map[string]any {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload failed: %v", err)
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(body))
	if err != nil {
		t.Fatalf("new request failed: %v", err)
	}
	req.Header.Set("content-type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != wantStatus {
		t.Fatalf("unexpected status %d, want %d", resp.StatusCode, wantStatus)
	}
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("decode json failed: %v", err)
	}
	return out
}

func dialWS(wsURL string) (*websocket.Conn, error) {
	dialer := websocket.Dialer{HandshakeTimeout: 5 * time.Second}
	conn, _, err := dialer.Dial(wsURL, nil)
	return conn, err
}

func mustSend(t *testing.T, conn *websocket.Conn, payload map[string]any) {
	t.Helper()
	_ = conn.SetWriteDeadline(time.Now().Add(3 * time.Second))
	if err := conn.WriteJSON(payload); err != nil {
		t.Fatalf("write ws failed: %v", err)
	}
}

func mustReadType(t *testing.T, conn *websocket.Conn, want string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		msg := mustReadRaw(t, conn)
		if msgType, _ := msg["type"].(string); msgType == want {
			return msg
		}
	}
	t.Fatalf("message type %s not received", want)
	return nil
}

func mustReadRaw(t *testing.T, conn *websocket.Conn) map[string]any {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	var msg map[string]any
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read ws failed: %v", err)
	}
	return msg
}

func mustReadRoomState(t *testing.T, conn *websocket.Conn) roomState {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		msg := mustReadRaw(t, conn)
		msgType, _ := msg["type"].(string)
		switch msgType {
		case "roomState":
			return decodeRoomStatePayload(t, msg["payload"])
		case "statePatch":
			payload, _ := msg["payload"].(map[string]any)
			if payload == nil {
				continue
			}
			if roomRaw, ok := payload["roomState"]; ok {
				return decodeRoomStatePayload(t, roomRaw)
			}
		}
	}
	t.Fatalf("roomState not received")
	return roomState{}
}

func mustReadGameView(t *testing.T, conn *websocket.Conn) gameView {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		msg := mustReadRaw(t, conn)
		msgType, _ := msg["type"].(string)
		switch msgType {
		case "gameView":
			return decodeGameViewPayload(t, msg["payload"])
		case "statePatch":
			payload, _ := msg["payload"].(map[string]any)
			if payload == nil {
				continue
			}
			if viewRaw, ok := payload["gameView"]; ok {
				return decodeGameViewPayload(t, viewRaw)
			}
		}
	}
	t.Fatalf("gameView not received")
	return gameView{}
}

func mustReadGameViewUntil(t *testing.T, conn *websocket.Conn, match func(gameView) bool) gameView {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var last gameView
	hasLast := false
	for time.Now().Before(deadline) {
		view := mustReadGameView(t, conn)
		last = view
		hasLast = true
		if match(view) {
			return view
		}
	}
	if hasLast {
		return last
	}
	t.Fatalf("gameView condition not reached")
	return gameView{}
}

func decodeRoomStatePayload(t *testing.T, raw any) roomState {
	t.Helper()
	bytes, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal room payload failed: %v", err)
	}
	var state roomState
	if err := json.Unmarshal(bytes, &state); err != nil {
		t.Fatalf("unmarshal room payload failed: %v", err)
	}
	return state
}

func decodeGameViewPayload(t *testing.T, raw any) gameView {
	t.Helper()
	bytes, err := json.Marshal(raw)
	if err != nil {
		t.Fatalf("marshal game payload failed: %v", err)
	}
	var view gameView
	if err := json.Unmarshal(bytes, &view); err != nil {
		t.Fatalf("unmarshal game payload failed: %v", err)
	}
	return view
}

func mustNestedString(t *testing.T, msg map[string]any, keys ...string) string {
	t.Helper()
	var current any = msg
	for _, key := range keys {
		obj, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("expected object for key %s", key)
		}
		current = obj[key]
	}
	value, ok := current.(string)
	if !ok {
		t.Fatalf("expected string at %v", keys)
	}
	return value
}

func firstHiddenCardID(view gameView) string {
	for _, card := range view.You.Hand {
		if !card.Revealed {
			return card.Instance
		}
	}
	return ""
}
