package authenticator

import "fmt"

var ErrEmptyDomain = fmt.Errorf("domain cannot be empty")
var ErrEmptyClientID = fmt.Errorf("client id cannot be empty")
var ErrEmptyClientSecret = fmt.Errorf("client secret cannot be empty")
var ErrEmptyRedirectURI = fmt.Errorf("redirect uri cannot be empty")
var ErrNoIDToken = fmt.Errorf("no id_token field in oauth2 token")
