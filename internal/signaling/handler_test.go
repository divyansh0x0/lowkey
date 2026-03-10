package signaling_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ayush/lowkey/internal/session"
	"github.com/ayush/lowkey/internal/signaling"
	"github.com/coder/websocket"
)

// helper: connect a WebSocket client to the test server with the given username.
func connect(t *testing.T, srv *httptest.Server, username string) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	url := "ws" + srv.URL[4:] + "/ws?username=" + username
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		t.Fatalf("failed to connect as %s: %v", username, err)
	}
	return conn
}

// helper: send a JSON message over the WebSocket.
func sendMsg(t *testing.T, conn *websocket.Conn, msg signaling.Message) {
	t.Helper()
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := conn.Write(ctx, websocket.MessageText, data); err != nil {
		t.Fatalf("write error: %v", err)
	}
}

// helper: read and parse a JSON message from the WebSocket.
func readMsg(t *testing.T, conn *websocket.Conn) signaling.Message {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read error: %v", err)
	}
	var msg signaling.Message
	if err := json.Unmarshal(data, &msg); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	return msg
}

// TestFullSignalingFlow tests the complete lifecycle:
// connect → create session → join → relay offer → relay answer → relay ICE
func TestFullSignalingFlow(t *testing.T) {
	store := session.NewMemoryStore(5 * time.Minute)
	defer store.Stop()

	hub := signaling.NewHub(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", hub.HandleWebSocket)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// --- Step 1: Both users connect ---
	alice := connect(t, srv, "alice")
	defer alice.CloseNow()

	bob := connect(t, srv, "bob")
	defer bob.CloseNow()

	// --- Step 2: Alice creates a session ---
	sendMsg(t, alice, signaling.Message{Type: signaling.TypeSessionCreate})

	created := readMsg(t, alice)
	if created.Type != signaling.TypeSessionCreated {
		t.Fatalf("expected session:created, got %s", created.Type)
	}

	var createdPayload signaling.SessionCreatedPayload
	json.Unmarshal(created.Payload, &createdPayload)
	if createdPayload.SessionID == "" {
		t.Fatal("session ID is empty")
	}
	if createdPayload.Key == "" {
		t.Fatal("encryption key is empty")
	}
	t.Logf("session created: %s", createdPayload.SessionID)

	// --- Step 3: Bob joins the session ---
	sendMsg(t, bob, signaling.Message{
		Type:      signaling.TypeSessionJoin,
		SessionID: createdPayload.SessionID,
	})

	// Bob receives session:joined
	bobJoined := readMsg(t, bob)
	if bobJoined.Type != signaling.TypeSessionJoined {
		t.Fatalf("bob expected session:joined, got %s", bobJoined.Type)
	}
	var bobPayload signaling.SessionJoinedPayload
	json.Unmarshal(bobJoined.Payload, &bobPayload)
	if bobPayload.Peer != "alice" {
		t.Fatalf("bob's peer should be alice, got %s", bobPayload.Peer)
	}

	// Alice also receives session:joined
	aliceJoined := readMsg(t, alice)
	if aliceJoined.Type != signaling.TypeSessionJoined {
		t.Fatalf("alice expected session:joined, got %s", aliceJoined.Type)
	}
	var alicePayload signaling.SessionJoinedPayload
	json.Unmarshal(aliceJoined.Payload, &alicePayload)
	if alicePayload.Peer != "bob" {
		t.Fatalf("alice's peer should be bob, got %s", alicePayload.Peer)
	}

	// Keys should match
	if bobPayload.Key != alicePayload.Key {
		t.Fatal("encryption keys don't match between peers")
	}
	t.Logf("both peers received matching keys ✓")

	// --- Step 4: Alice sends an SDP offer to Bob ---
	offerPayload, _ := json.Marshal(map[string]string{"sdp": "v=0\r\n..."})
	sendMsg(t, alice, signaling.Message{
		Type:      signaling.TypeSignalOffer,
		SessionID: createdPayload.SessionID,
		Target:    "bob",
		Payload:   offerPayload,
	})

	relayedOffer := readMsg(t, bob)
	if relayedOffer.Type != signaling.TypeSignalOffer {
		t.Fatalf("bob expected signal:offer, got %s", relayedOffer.Type)
	}
	if relayedOffer.Sender != "alice" {
		t.Fatalf("offer sender should be alice, got %s", relayedOffer.Sender)
	}
	t.Log("SDP offer relayed alice → bob ✓")

	// --- Step 5: Bob sends an SDP answer to Alice ---
	answerPayload, _ := json.Marshal(map[string]string{"sdp": "v=0\r\nanswer..."})
	sendMsg(t, bob, signaling.Message{
		Type:      signaling.TypeSignalAnswer,
		SessionID: createdPayload.SessionID,
		Target:    "alice",
		Payload:   answerPayload,
	})

	relayedAnswer := readMsg(t, alice)
	if relayedAnswer.Type != signaling.TypeSignalAnswer {
		t.Fatalf("alice expected signal:answer, got %s", relayedAnswer.Type)
	}
	if relayedAnswer.Sender != "bob" {
		t.Fatalf("answer sender should be bob, got %s", relayedAnswer.Sender)
	}
	t.Log("SDP answer relayed bob → alice ✓")

	// --- Step 6: Both exchange ICE candidates ---
	icePayload, _ := json.Marshal(map[string]string{"candidate": "candidate:..."})

	sendMsg(t, alice, signaling.Message{
		Type:   signaling.TypeSignalICE,
		Target: "bob",
		Payload: icePayload,
	})
	bobICE := readMsg(t, bob)
	if bobICE.Type != signaling.TypeSignalICE || bobICE.Sender != "alice" {
		t.Fatal("ICE relay alice → bob failed")
	}

	sendMsg(t, bob, signaling.Message{
		Type:   signaling.TypeSignalICE,
		Target: "alice",
		Payload: icePayload,
	})
	aliceICE := readMsg(t, alice)
	if aliceICE.Type != signaling.TypeSignalICE || aliceICE.Sender != "bob" {
		t.Fatal("ICE relay bob → alice failed")
	}
	t.Log("ICE candidates relayed bidirectionally ✓")

	t.Log("✅ Full signaling flow passed!")
}

// TestSessionErrors tests error conditions.
func TestSessionErrors(t *testing.T) {
	store := session.NewMemoryStore(5 * time.Minute)
	defer store.Stop()

	hub := signaling.NewHub(store)

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", hub.HandleWebSocket)
	srv := httptest.NewServer(mux)
	defer srv.Close()

	conn := connect(t, srv, "charlie")
	defer conn.CloseNow()

	// Try to join a non-existent session
	sendMsg(t, conn, signaling.Message{
		Type:      signaling.TypeSessionJoin,
		SessionID: "00000000-0000-0000-0000-000000000000",
	})

	errMsg := readMsg(t, conn)
	if errMsg.Type != signaling.TypeError {
		t.Fatalf("expected error, got %s", errMsg.Type)
	}
	t.Log("non-existent session join correctly rejected ✓")

	// Send unknown message type
	sendMsg(t, conn, signaling.Message{Type: "unknown:type"})

	errMsg2 := readMsg(t, conn)
	if errMsg2.Type != signaling.TypeError {
		t.Fatalf("expected error for unknown type, got %s", errMsg2.Type)
	}
	t.Log("unknown message type correctly rejected ✓")
}
