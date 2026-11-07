package web

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"darbelis.eu/wshub/internal/handler"
	"darbelis.eu/wshub/internal/state"
	"darbelis.eu/wshub/pkg/util"
	"github.com/gin-gonic/gin"
)

func TestRegisterSessionTags(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		query    string
		wantCode int
		wantTags []string
	}{
		{"no tags parameter", "clientId=c1&sessionToken=t1", http.StatusOK, []string{}},
		{"empty tags parameter", "clientId=c1&sessionToken=t1&tags=", http.StatusOK, []string{}},
		{"one tag", "clientId=c1&sessionToken=t1&tags=room", http.StatusOK, []string{"room"}},
		{"two tags", "clientId=c1&sessionToken=t1&tags=room,news", http.StatusOK, []string{"room", "news"}},
		{"empty tag between tags", "clientId=c1&sessionToken=t1&tags=room,,news", http.StatusBadRequest, nil},
		{"only a separator", "clientId=c1&sessionToken=t1&tags=,", http.StatusBadRequest, nil},
		{"tag with a space", "clientId=c1&sessionToken=t1&tags=a%20b", http.StatusBadRequest, nil},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			registry := state.NewClientsRegistry()
			controller := NewApiForWsController(
				registry,
				util.NewFixedClock(tagTestNow),
				nil,
				handler.NewWaitResponseMessageHandler(),
			)
			router := gin.New()
			router.POST("/register-session", controller.RegisterSession)

			request := httptest.NewRequest(http.MethodPost, "/register-session?"+test.query, nil)
			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, request)

			if recorder.Code != test.wantCode {
				t.Fatalf("status %d, body %s", recorder.Code, recorder.Body.String())
			}

			session, found := registry.FindSessionByClientId("c1")
			if test.wantTags == nil {
				if found {
					t.Errorf("a rejected request registered a session: %+v", session)
				}
				return
			}
			if !found {
				t.Fatal("the session was not registered")
			}
			if !slices.Equal(session.Tags, test.wantTags) {
				t.Errorf("tags %v, want %v", session.Tags, test.wantTags)
			}
		})
	}
}
