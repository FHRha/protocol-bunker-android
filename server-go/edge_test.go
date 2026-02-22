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

func setupTestRoom(
	t *testing.T,
	port int,
) (*server, *httptest.Server, *websocket.Conn, string, string) {
	t.Helper()
	assetsRoot, clientRoot := makeTestRoots(t)
	srv, err := newServer(config{
		Host:               "127.0.0.1",
		Port:               port,
		AssetsRoot:         assetsRoot,
		ClientDistRoot:     clientRoot,
		EnableDevScenarios: false,
	})
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
