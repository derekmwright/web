package auth0

import (
	"context"
	"net/http"
)

type userContextKey struct{}

type Middleware func(http.Handler) http.Handler

func authenticatedMiddleware(deps *deps, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sessionUser := deps.sessions.Get(r.Context(), "user")
		if sessionUser == nil {
			state, err := generateRandomState()
			if err != nil {
				deps.log.Error("unable to generate random state", "error", err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			deps.sessions.Put(r.Context(), StateKey, state)

			loginURL := deps.auth.AuthCodeURL(state)
			http.Redirect(w, r, loginURL, http.StatusFound)

			return
		}

		ctx := context.WithValue(r.Context(), userContextKey{}, sessionUser)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func CurrentUser(r *http.Request) SessionUser {
	return r.Context().Value(userContextKey{}).(SessionUser)
}
