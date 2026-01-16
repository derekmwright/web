package auth0

import (
	"context"
	"log/slog"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"golang.org/x/oauth2"

	"github.com/derekmwright/web/auth/auth0/authenticator"
)

type mockSessionManager struct {
	store map[string]any
	mu    sync.RWMutex
}

func (m *mockSessionManager) Get(ctx context.Context, key string) any {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store[key]
}

func (m *mockSessionManager) Put(ctx context.Context, key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if value == nil {
		delete(m.store, key)
	} else {
		m.store[key] = value
	}
}

func TestHandleLogic(t *testing.T) {
	tests := []struct {
		name             string
		existingState    any
		wantRedirect     string
		wantSessionKey   string
		wantSessionValue string
	}{
		{
			name:             "generates and stores state",
			wantRedirect:     "https://",
			wantSessionKey:   StateKey,
			wantSessionValue: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSessions := &mockSessionManager{
				store: make(map[string]any),
			}
			if tt.existingState != nil {
				mockSessions.store[StateKey] = tt.existingState
			}

			d := &deps{
				log:      slog.Default(),
				sessions: mockSessions,
				auth: &authenticator.Authenticator{
					Config: oauth2.Config{
						Endpoint: oauth2.Endpoint{
							AuthURL: "https://test.auth0.com/authorize",
						},
					},
				},
			}

			rr := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/login", nil)

			HandleLogin(d)(rr, req)

			if !strings.HasPrefix(rr.Header().Get("Location"), "https://test.auth0.com/authorize") {
				t.Errorf("wrong redirect: %s", rr.Header().Get("Location"))
			}

			state := mockSessions.store[StateKey].(string)
			if len(state) == 0 {
				t.Fatal("state not stored or empty")
			}
		})
	}
}
