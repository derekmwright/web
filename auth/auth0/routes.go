package auth0

import (
	"github.com/go-chi/chi/v5"
)

func Register(r chi.Router, deps *deps) {
	r.Get("/login", HandleLogin(deps))
	r.Get("/callback", HandleCallback(deps))
	r.Get("/logout", HandleLogout(deps))
}
