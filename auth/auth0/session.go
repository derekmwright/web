package auth0

import "encoding/json"

type SessionUser struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Picture       string `json:"picture"`

	Custom json.RawMessage `json:"-"`
}

func (u *SessionUser) CustomClaims() (map[string]any, error) {
	if len(u.Custom) == 0 {
		return nil, nil
	}

	var claims map[string]any
	if err := json.Unmarshal(u.Custom, &claims); err != nil {
		return nil, err
	}

	return claims, nil
}
