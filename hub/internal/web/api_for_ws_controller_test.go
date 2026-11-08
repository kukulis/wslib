package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"darbelis.eu/wshub/internal/handler"
	"darbelis.eu/wshub/internal/messages"
	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/util"
	"darbelis.eu/wshub/pkg/ws"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var tagTestNow = time.Date(2026, time.November, 2, 12, 0, 0, 0, time.UTC)

type tagFixture struct {
	t        *testing.T
	registry *state.ClientsRegistry
	hub      *ws.Hub
	server   *httptest.Server
	router   *gin.Engine
}

// newTagFixture starts a real hub behind a test server, so the tests can read
// what a connected client receives.
func newTagFixture(t *testing.T) *tagFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)

	fixture := &tagFixture{t: t, registry: state.NewClientsRegistry()}
	fixture.hub = ws.NewHub("test", messages.Encode, messages.Decode)
	go fixture.hub.Run()

	fixture.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fixture.hub.RegisterWebSocketConnection(r.URL.Query().Get("connectionId"), w, r)
	}))
	t.Cleanup(fixture.server.Close)

	controller := NewApiForWsController(
		fixture.registry,
		util.NewFixedClock(tagTestNow),
		fixture.hub,
		handler.NewWaitResponseMessageHandler(),
	)
	fixture.router = gin.New()
	fixture.router.POST("/tag-message", controller.SendTagMessage)

	return fixture
}

// connectClient registers a session with the tags, connects a WebSocket client
// and logs the session in on that connection.
func (f *tagFixture) connectClient(clientId string, tags ...string) *websocket.Conn {
	f.t.Helper()

	if err := f.registry.AddSession(state.ClientSession{ClientId: clientId, Token: "t", Tags: tags}, tagTestNow); err != nil {
		f.t.Fatalf("add session %s: %v", clientId, err)
	}

	connectionId := "conn-" + clientId
	url := "ws" + strings.TrimPrefix(f.server.URL, "http") + "/?connectionId=" + connectionId
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		f.t.Fatalf("dial %s: %v", clientId, err)
	}
	f.t.Cleanup(func() { _ = conn.Close() })

	// the hub registers the connection asynchronously
	deadline := time.Now().Add(2 * time.Second)
	for !f.hub.HasConnection(connectionId) {
		if time.Now().After(deadline) {
			f.t.Fatalf("the hub did not register %s", connectionId)
		}
		time.Sleep(5 * time.Millisecond)
	}

	if _, err := f.registry.LoginSession(clientId, "t", connectionId, tagTestNow); err != nil {
		f.t.Fatalf("login %s: %v", clientId, err)
	}

	return conn
}

func (f *tagFixture) post(query string, body string) *httptest.ResponseRecorder {
	f.t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/tag-message?"+query, strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, request)

	return recorder
}

func readEnvelope(t *testing.T, conn *websocket.Conn) (envelopeType string, data map[string]any) {
	t.Helper()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var envelope struct {
		Type string         `json:"type"`
		Data map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("decode %s: %v", raw, err)
	}

	return envelope.Type, envelope.Data
}

func expectNothing(t *testing.T, conn *websocket.Conn) {
	t.Helper()

	_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	if _, raw, err := conn.ReadMessage(); err == nil {
		t.Errorf("a client without the tag received %s", raw)
	}
}

func deliveredCount(t *testing.T, recorder *httptest.ResponseRecorder) int {
	t.Helper()

	var body struct {
		Delivered int `json:"delivered"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("response %q: %v", recorder.Body.String(), err)
	}

	return body.Delivered
}

func TestSendTagMessageReachesEveryClientWithTheTag(t *testing.T) {
	fixture := newTagFixture(t)
	first := fixture.connectClient("client-1", "room", "user-42")
	second := fixture.connectClient("client-2", "room")
	other := fixture.connectClient("client-3", "lobby")

	recorder := fixture.post("tag=room&messageId=m1", `{"type":"news","data":{"text":"hello"}}`)

	if recorder.Code != http.StatusOK || deliveredCount(t, recorder) != 2 {
		t.Fatalf("status %d, body %s", recorder.Code, recorder.Body.String())
	}
	for name, conn := range map[string]*websocket.Conn{"first": first, "second": second} {
		envelopeType, data := readEnvelope(t, conn)
		if envelopeType != "news" || data["text"] != "hello" {
			t.Errorf("%s received %s %v", name, envelopeType, data)
		}
	}
	expectNothing(t, other)
}

func TestSendTagMessageToATagWithoutClientsDeliversNothing(t *testing.T) {
	fixture := newTagFixture(t)
	_ = fixture.registry.AddSession(state.ClientSession{ClientId: "never-logged-in", Token: "t", Tags: []string{"room"}}, tagTestNow)

	for _, tag := range []string{"room", "unknown"} {
		recorder := fixture.post("tag="+tag, `{"type":"news"}`)
		if recorder.Code != http.StatusOK || deliveredCount(t, recorder) != 0 {
			t.Errorf("tag %s: status %d, body %s", tag, recorder.Code, recorder.Body.String())
		}
	}
}

func TestSendTagMessageSkipsASessionWhoseConnectionIsGone(t *testing.T) {
	fixture := newTagFixture(t)
	_ = fixture.registry.AddSession(state.ClientSession{ClientId: "ghost", Token: "t", Tags: []string{"room"}}, tagTestNow)
	// logged in on a connection the hub does not know
	_, _ = fixture.registry.LoginSession("ghost", "t", "conn-ghost", tagTestNow)
	live := fixture.connectClient("live", "room")

	recorder := fixture.post("tag=room", `{"type":"news"}`)

	if deliveredCount(t, recorder) != 1 {
		t.Errorf("body %s", recorder.Body.String())
	}
	if envelopeType, _ := readEnvelope(t, live); envelopeType != "news" {
		t.Errorf("live client received %q", envelopeType)
	}
}

func TestSendTagMessageRejectsBadRequests(t *testing.T) {
	fixture := newTagFixture(t)

	tests := []struct{ name, query, body string }{
		{"missing tag", "", `{"type":"news"}`},
		{"tag with a space", "tag=a%20b", `{"type":"news"}`},
		{"bad message id", "tag=room&messageId=a/b", `{"type":"news"}`},
		{"body is not json", "tag=room", `not json`},
		{"wait for response", "tag=room", `{"type":"news","waitResponse":1000}`},
		{"data is not an object", "tag=room", `{"type":"news","data":[1,2]}`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if recorder := fixture.post(test.query, test.body); recorder.Code != http.StatusBadRequest {
				t.Errorf("status %d, body %s", recorder.Code, recorder.Body.String())
			}
		})
	}
}
