package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestServer_Edge_RemovedOverlayAPI_ReturnsNotFound(t *testing.T) {
	srv, testServer, hostConn, _, _ := setupTestRoom(t, 18086)
	_ = srv
	defer testServer.Close()
	defer hostConn.Close()

	paths := []string{"/api/overlay-links", "/api/overlay-control/state", "/api/overlay-control/save"}
	for _, route := range paths {
		resp, err := http.Get(testServer.URL + route)
		if err != nil {
			t.Fatalf("request %s failed: %v", route, err)
		}
		_ = resp.Body.Close()
		if resp.StatusCode != http.StatusNotFound {
			t.Fatalf("expected 404 for %s, got %d", route, resp.StatusCode)
		}
	}
}

func TestServer_Edge_UnknownWSMessageReturnsError(t *testing.T) {
	_, testServer, hostConn, _, _ := setupTestRoom(t, 18087)
	defer testServer.Close()
	defer hostConn.Close()

	mustSend(t, hostConn, map[string]any{
		"type":    "notExistsAction",
		"payload": map[string]any{},
	})
	_ = mustReadType(t, hostConn, "error")
}

func TestServer_Edge_RoomLifecycle_RemoveLobbyRoomAfterLastDisconnect(t *testing.T) {
	srv, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18088)
	defer testServer.Close()

	_ = hostConn.Close()
	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		srv.mu.Lock()
		_, exists := srv.rooms[roomCode]
		srv.mu.Unlock()
		if !exists {
			break
		}
		time.Sleep(40 * time.Millisecond)
	}

	srv.mu.Lock()
	_, exists := srv.rooms[roomCode]
	srv.mu.Unlock()
	if exists {
		t.Fatalf("room should be removed after last lobby player disconnect")
	}
}

func TestServer_Edge_LobbyDisconnect_RemovesPlayerButKeepsRoom(t *testing.T) {
	srv, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18089)
	defer testServer.Close()
	defer hostConn.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	guestConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("guest dial failed: %v", err)
	}

	mustSend(t, guestConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":      "Guest",
			"roomCode":  roomCode,
			"sessionId": "edge-guest",
		},
	})
	guestAck := mustReadType(t, guestConn, "helloAck")
	_ = mustReadRoomState(t, guestConn)
	guestID := mustNestedString(t, guestAck, "payload", "playerId")
	if guestID == "" {
		t.Fatalf("guest player id must not be empty")
	}

	_ = guestConn.Close()

	deadline := time.Now().Add(4 * time.Second)
	for time.Now().Before(deadline) {
		srv.mu.Lock()
		room := srv.rooms[roomCode]
		guestExists := room != nil && room.Players[guestID] != nil
		srv.mu.Unlock()
		if !guestExists {
			break
		}
		time.Sleep(40 * time.Millisecond)
	}

	srv.mu.Lock()
	room := srv.rooms[roomCode]
	srv.mu.Unlock()
	if room == nil {
		t.Fatalf("room should stay alive when host is still connected")
	}
	if room.Players[guestID] != nil {
		t.Fatalf("guest should be removed from lobby after disconnect")
	}
	if len(room.Players) != 1 {
		t.Fatalf("expected one player left in room, got %d", len(room.Players))
	}
}

func TestServer_Edge_TransferHost_AutoAndManual(t *testing.T) {
	_, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18090)
	defer testServer.Close()
	defer hostConn.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	guestConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("guest dial failed: %v", err)
	}
	defer guestConn.Close()

	mustSend(t, guestConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":      "Guest",
			"roomCode":  roomCode,
			"sessionId": "edge-transfer-guest",
		},
	})
	guestAck := mustReadType(t, guestConn, "helloAck")
	_ = mustReadRoomState(t, guestConn)
	_ = mustReadRoomState(t, hostConn)
	guestID := mustNestedString(t, guestAck, "payload", "playerId")
	if guestID == "" {
		t.Fatalf("guest player id must not be empty")
	}

	mustSend(t, hostConn, map[string]any{
		"type":    "requestHostTransfer",
		"payload": map[string]any{},
	})
	autoTransferState := mustReadRoomState(t, hostConn)
	if autoTransferState.HostID != guestID || autoTransferState.ControlID != guestID {
		t.Fatalf("auto host transfer failed: host=%s control=%s want=%s", autoTransferState.HostID, autoTransferState.ControlID, guestID)
	}
	hostChanged := mustReadType(t, hostConn, "hostChanged")
	newHostID := mustNestedString(t, hostChanged, "payload", "newHostId")
	newControlID := mustNestedString(t, hostChanged, "payload", "newControlId")
	if newHostID != guestID {
		t.Fatalf("hostChanged newHostId mismatch: got=%s want=%s", newHostID, guestID)
	}
	if newControlID != guestID {
		t.Fatalf("hostChanged newControlId mismatch: got=%s want=%s", newControlID, guestID)
	}

	mustSend(t, guestConn, map[string]any{
		"type": "requestHostTransfer",
		"payload": map[string]any{
			"targetPlayerId": "missing-player-id",
		},
	})
	_ = mustReadType(t, guestConn, "error")
}

func TestServer_Edge_TransferHost_NoCandidates_ErrorMessage(t *testing.T) {
	_, testServer, hostConn, _, _ := setupTestRoom(t, 18095)
	defer testServer.Close()
	defer hostConn.Close()

	mustSend(t, hostConn, map[string]any{
		"type":    "requestHostTransfer",
		"payload": map[string]any{},
	})
	msg := mustReadType(t, hostConn, "error")
	message := mustNestedString(t, msg, "payload", "message")
	if !strings.Contains(message, "No online players available") {
		t.Fatalf("unexpected no-candidates message: %q", message)
	}
}

func TestServer_Edge_TransferHost_OfflineTarget_ErrorMessage(t *testing.T) {
	srv, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18096)
	defer testServer.Close()
	defer hostConn.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	guestConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("guest dial failed: %v", err)
	}
	defer guestConn.Close()

	mustSend(t, guestConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":      "Guest",
			"roomCode":  roomCode,
			"sessionId": "edge-transfer-offline",
		},
	})
	guestAck := mustReadType(t, guestConn, "helloAck")
	_ = mustReadRoomState(t, guestConn)
	_ = mustReadRoomState(t, hostConn)
	guestID := mustNestedString(t, guestAck, "payload", "playerId")

	srv.mu.Lock()
	room := srv.rooms[roomCode]
	if room == nil {
		srv.mu.Unlock()
		t.Fatalf("room not found")
	}
	target := room.Players[guestID]
	if target == nil {
		srv.mu.Unlock()
		t.Fatalf("guest player not found")
	}
	target.Connected = false
	target.Connection = nil
	srv.mu.Unlock()

	mustSend(t, hostConn, map[string]any{
		"type": "requestHostTransfer",
		"payload": map[string]any{
			"targetPlayerId": guestID,
		},
	})
	msg := mustReadType(t, hostConn, "error")
	message := mustNestedString(t, msg, "payload", "message")
	if !strings.Contains(message, "Target player is offline") {
		t.Fatalf("unexpected offline-target message: %q", message)
	}
}

func TestServer_Edge_TransferHost_DevMode_Works(t *testing.T) {
	_, testServer, hostConn, roomCode, _ := setupTestRoomWithConfig(t, 18097, config{
		IdentityMode: "dev_tab",
	})
	defer testServer.Close()
	defer hostConn.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	guestConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("guest dial failed: %v", err)
	}
	defer guestConn.Close()

	mustSend(t, guestConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":      "Guest",
			"roomCode":  roomCode,
			"sessionId": "edge-transfer-dev",
		},
	})
	guestAck := mustReadType(t, guestConn, "helloAck")
	_ = mustReadRoomState(t, guestConn)
	_ = mustReadRoomState(t, hostConn)
	guestID := mustNestedString(t, guestAck, "payload", "playerId")

	mustSend(t, hostConn, map[string]any{
		"type": "requestHostTransfer",
		"payload": map[string]any{
			"targetPlayerId": guestID,
		},
	})
	state := mustReadRoomState(t, hostConn)
	if state.ControlID != guestID || state.HostID != guestID {
		t.Fatalf("dev transfer failed: host=%s control=%s want=%s", state.HostID, state.ControlID, guestID)
	}
}

func TestServer_Edge_TransferHost_RejectSameHostTarget(t *testing.T) {
	srv, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18091)
	defer testServer.Close()
	defer hostConn.Close()

	srv.mu.Lock()
	room := srv.rooms[roomCode]
	if room == nil {
		srv.mu.Unlock()
		t.Fatalf("room not found")
	}
	hostID := room.HostID
	srv.mu.Unlock()

	if hostID == "" {
		t.Fatalf("host id must not be empty")
	}

	mustSend(t, hostConn, map[string]any{
		"type": "requestHostTransfer",
		"payload": map[string]any{
			"targetPlayerId": hostID,
		},
	})
	_ = mustReadType(t, hostConn, "error")
}

func TestServer_Edge_TransferHost_RejectNonControl(t *testing.T) {
	_, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18098)
	defer testServer.Close()
	defer hostConn.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	guestConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("guest dial failed: %v", err)
	}
	defer guestConn.Close()

	mustSend(t, guestConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":      "Guest",
			"roomCode":  roomCode,
			"sessionId": "edge-transfer-non-control",
		},
	})
	_ = mustReadType(t, guestConn, "helloAck")
	_ = mustReadRoomState(t, guestConn)
	_ = mustReadRoomState(t, hostConn)

	mustSend(t, guestConn, map[string]any{
		"type":    "requestHostTransfer",
		"payload": map[string]any{},
	})
	msg := mustReadType(t, guestConn, "error")
	message := mustNestedString(t, msg, "payload", "message")
	if !strings.Contains(message, "Only current control player can transfer host role") {
		t.Fatalf("unexpected non-control message: %q", message)
	}
}

func TestServer_Edge_TransferHost_TargetNotFound_ErrorMessage(t *testing.T) {
	_, testServer, hostConn, _, _ := setupTestRoom(t, 18099)
	defer testServer.Close()
	defer hostConn.Close()

	mustSend(t, hostConn, map[string]any{
		"type": "requestHostTransfer",
		"payload": map[string]any{
			"targetPlayerId": "missing-player-id",
		},
	})
	msg := mustReadType(t, hostConn, "error")
	message := mustNestedString(t, msg, "payload", "message")
	if !strings.Contains(message, "Target player was not found") {
		t.Fatalf("unexpected target-not-found message: %q", message)
	}
}

func TestServer_Edge_ControlCompanion_HelloControlToken_NoNewPlayer(t *testing.T) {
	srv, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18092)
	defer testServer.Close()
	defer hostConn.Close()

	srv.mu.Lock()
	room := srv.rooms[roomCode]
	if room == nil {
		srv.mu.Unlock()
		t.Fatalf("room not found")
	}
	controlToken := room.ControlToken
	hostID := room.HostID
	playersBefore := len(room.Players)
	srv.mu.Unlock()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	companionConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("companion dial failed: %v", err)
	}
	defer companionConn.Close()

	mustSend(t, companionConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"roomCode":     roomCode,
			"controlToken": controlToken,
		},
	})
	ack := mustReadType(t, companionConn, "helloAck")
	_ = mustReadRoomState(t, companionConn)
	if got := mustNestedString(t, ack, "payload", "playerId"); got != hostID {
		t.Fatalf("companion must bind to current control id: got=%s want=%s", got, hostID)
	}

	mustSend(t, companionConn, map[string]any{
		"type": "updateSettings",
		"payload": map[string]any{
			"maxPlayers": 5,
		},
	})

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		srv.mu.Lock()
		room = srv.rooms[roomCode]
		ok := room != nil && room.Settings.MaxPlayers == 5 && len(room.Players) == playersBefore
		srv.mu.Unlock()
		if ok {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}

	srv.mu.Lock()
	room = srv.rooms[roomCode]
	gotPlayers := 0
	gotMax := 0
	if room != nil {
		gotPlayers = len(room.Players)
		gotMax = room.Settings.MaxPlayers
	}
	srv.mu.Unlock()
	t.Fatalf("companion action not applied or created player: players=%d want=%d maxPlayers=%d want=5", gotPlayers, playersBefore, gotMax)
}

func TestServer_Edge_ControlCompanion_HelloEditToken_NoNewPlayer(t *testing.T) {
	srv, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18093)
	defer testServer.Close()
	defer hostConn.Close()

	srv.mu.Lock()
	room := srv.rooms[roomCode]
	if room == nil {
		srv.mu.Unlock()
		t.Fatalf("room not found")
	}
	editToken := room.EditToken
	hostID := room.HostID
	playersBefore := len(room.Players)
	srv.mu.Unlock()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	companionConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("companion dial failed: %v", err)
	}
	defer companionConn.Close()

	mustSend(t, companionConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"roomCode":  roomCode,
			"editToken": editToken,
		},
	})
	ack := mustReadType(t, companionConn, "helloAck")
	_ = mustReadRoomState(t, companionConn)
	if got := mustNestedString(t, ack, "payload", "playerId"); got != hostID {
		t.Fatalf("companion must bind to current control id: got=%s want=%s", got, hostID)
	}

	srv.mu.Lock()
	room = srv.rooms[roomCode]
	gotPlayers := 0
	if room != nil {
		gotPlayers = len(room.Players)
	}
	srv.mu.Unlock()
	if gotPlayers != playersBefore {
		t.Fatalf("edit token companion must not create new player: got=%d want=%d", gotPlayers, playersBefore)
	}
}

func TestServer_Edge_ControlCompanion_DevMode_AllowsControlPlayerToken(t *testing.T) {
	srv, testServer, hostConn, roomCode, hostToken := setupTestRoomWithConfig(t, 18094, config{
		IdentityMode: "dev_tab",
	})
	defer testServer.Close()
	defer hostConn.Close()

	srv.mu.Lock()
	room := srv.rooms[roomCode]
	if room == nil {
		srv.mu.Unlock()
		t.Fatalf("room not found")
	}
	hostID := room.HostID
	playersBefore := len(room.Players)
	srv.mu.Unlock()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	companionConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("companion dial failed: %v", err)
	}
	defer companionConn.Close()

	mustSend(t, companionConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"roomCode":     roomCode,
			"controlToken": hostToken,
		},
	})
	ack := mustReadType(t, companionConn, "helloAck")
	_ = mustReadRoomState(t, companionConn)
	if got := mustNestedString(t, ack, "payload", "playerId"); got != hostID {
		t.Fatalf("dev companion must bind to control id: got=%s want=%s", got, hostID)
	}

	srv.mu.Lock()
	room = srv.rooms[roomCode]
	gotPlayers := 0
	if room != nil {
		gotPlayers = len(room.Players)
	}
	srv.mu.Unlock()
	if gotPlayers != playersBefore {
		t.Fatalf("dev companion must not create new player: got=%d want=%d", gotPlayers, playersBefore)
	}
}

func TestServer_Edge_ControlCompanion_HostChangedContainsNewControlId(t *testing.T) {
	srv, testServer, hostConn, roomCode, _ := setupTestRoom(t, 18100)
	defer testServer.Close()
	defer hostConn.Close()

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")

	srv.mu.Lock()
	room := srv.rooms[roomCode]
	if room == nil {
		srv.mu.Unlock()
		t.Fatalf("room not found")
	}
	controlToken := room.ControlToken
	srv.mu.Unlock()

	guestConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("guest dial failed: %v", err)
	}
	defer guestConn.Close()

	mustSend(t, guestConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":      "Guest",
			"roomCode":  roomCode,
			"sessionId": "edge-companion-host-changed",
		},
	})
	guestAck := mustReadType(t, guestConn, "helloAck")
	_ = mustReadRoomState(t, guestConn)
	_ = mustReadRoomState(t, hostConn)
	guestID := mustNestedString(t, guestAck, "payload", "playerId")

	companionConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("companion dial failed: %v", err)
	}
	defer companionConn.Close()

	mustSend(t, companionConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"roomCode":     roomCode,
			"controlToken": controlToken,
		},
	})
	_ = mustReadType(t, companionConn, "helloAck")
	_ = mustReadRoomState(t, companionConn)

	mustSend(t, hostConn, map[string]any{
		"type": "requestHostTransfer",
		"payload": map[string]any{
			"targetPlayerId": guestID,
		},
	})
	_ = mustReadRoomState(t, hostConn)           // host broadcast
	_ = mustReadType(t, hostConn, "hostChanged") // host event
	_ = mustReadRoomState(t, companionConn)      // companion broadcast
	_ = mustReadType(t, companionConn, "helloAck")
	hostChangedCompanion := mustReadType(t, companionConn, "hostChanged")
	if got := mustNestedString(t, hostChangedCompanion, "payload", "newControlId"); got != guestID {
		t.Fatalf("companion hostChanged newControlId mismatch: got=%s want=%s", got, guestID)
	}
}

func TestServer_Edge_WSFastPath_HelloTransferHostStartGame(t *testing.T) {
	assetsRoot, clientRoot := makeTestRoots(t)
	srv, err := newServer(config{
		Host:               "127.0.0.1",
		Port:               18109,
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
	hostConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("host dial failed: %v", err)
	}
	defer hostConn.Close()

	mustSend(t, hostConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":       "Host",
			"create":     true,
			"scenarioId": scenarioDevTest,
			"sessionId":  "edge-fast-host",
		},
	})
	hostAck := mustReadType(t, hostConn, "helloAck")
	hostID := mustNestedString(t, hostAck, "payload", "playerId")
	editToken := mustNestedString(t, hostAck, "payload", "editToken")
	roomState := mustReadRoomState(t, hostConn)
	if hostID == "" || editToken == "" || roomState.RoomCode == "" {
		t.Fatalf("invalid host init: host=%q edit=%q room=%q", hostID, editToken, roomState.RoomCode)
	}

	companionConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("companion dial failed: %v", err)
	}
	defer companionConn.Close()
	mustSend(t, companionConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"roomCode":  roomState.RoomCode,
			"editToken": editToken,
		},
	})
	companionAck := mustReadType(t, companionConn, "helloAck")
	if got := mustNestedString(t, companionAck, "payload", "playerId"); got != hostID {
		t.Fatalf("companion must bind to host control player: got=%s want=%s", got, hostID)
	}
	_ = mustReadRoomState(t, companionConn)

	guestConn, err := dialWS(wsURL)
	if err != nil {
		t.Fatalf("guest dial failed: %v", err)
	}
	defer guestConn.Close()
	mustSend(t, guestConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":      "Guest",
			"roomCode":  roomState.RoomCode,
			"sessionId": "edge-fast-guest",
		},
	})
	guestAck := mustReadType(t, guestConn, "helloAck")
	guestID := mustNestedString(t, guestAck, "payload", "playerId")
	if guestID == "" || guestID == hostID {
		t.Fatalf("invalid guest id: %q", guestID)
	}
	_ = mustReadRoomState(t, guestConn)
	_ = mustReadRoomState(t, hostConn)
	_ = mustReadRoomState(t, companionConn)

	mustSend(t, companionConn, map[string]any{
		"type": "requestHostTransfer",
		"payload": map[string]any{
			"targetPlayerId": guestID,
		},
	})
	_ = mustReadRoomState(t, companionConn)
	hostChanged := mustReadType(t, companionConn, "hostChanged")
	if got := mustNestedString(t, hostChanged, "payload", "newHostId"); got != guestID {
		t.Fatalf("transfer failed, new host id: got=%s want=%s", got, guestID)
	}

	mustSend(t, companionConn, map[string]any{
		"type":    "startGame",
		"payload": map[string]any{},
	})
	view := mustReadGameView(t, companionConn)
	if view.Phase != scenarioPhaseReveal && view.Phase != scenarioPhaseRevealDiscussion {
		t.Fatalf("unexpected game phase after start: %s", view.Phase)
	}
}

func setupTestRoom(
	t *testing.T,
	port int,
) (*server, *httptest.Server, *websocket.Conn, string, string) {
	return setupTestRoomWithConfig(t, port, config{})
}

func setupTestRoomWithConfig(
	t *testing.T,
	port int,
	override config,
) (*server, *httptest.Server, *websocket.Conn, string, string) {
	t.Helper()
	assetsRoot, clientRoot := makeTestRoots(t)
	cfg := config{
		Host:               "127.0.0.1",
		Port:               port,
		AssetsRoot:         assetsRoot,
		ClientDistRoot:     clientRoot,
		EnableDevScenarios: false,
	}
	if override.Host != "" {
		cfg.Host = override.Host
	}
	if override.Port != 0 {
		cfg.Port = override.Port
	}
	if override.IdentityMode != "" {
		cfg.IdentityMode = override.IdentityMode
	}
	if override.EnableDevScenarios {
		cfg.EnableDevScenarios = true
	}
	srv, err := newServer(cfg)
	if err != nil {
		t.Fatalf("newServer failed: %v", err)
	}
	testServer := httptest.NewServer(srv.routes())

	wsURL := "ws" + strings.TrimPrefix(testServer.URL, "http")
	hostConn, err := dialWS(wsURL)
	if err != nil {
		testServer.Close()
		t.Fatalf("host dial failed: %v", err)
	}
	mustSend(t, hostConn, map[string]any{
		"type": "hello",
		"payload": map[string]any{
			"name":       "Host",
			"create":     true,
			"scenarioId": scenarioClassic,
			"sessionId":  "edge-session",
		},
	})
	helloAck := mustReadType(t, hostConn, "helloAck")
	roomState := mustReadRoomState(t, hostConn)

	token := mustNestedString(t, helloAck, "payload", "playerToken")
	if token == "" || roomState.RoomCode == "" {
		hostConn.Close()
		testServer.Close()
		t.Fatalf("room init failed: room=%s token=%s", roomState.RoomCode, token)
	}
	return srv, testServer, hostConn, roomState.RoomCode, token
}
