package advhttp

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// A structure to cache oauth2 state. Has helper methods which will allow users to pull cached
// tokens from the structure when needed without utilizing the network to obtain new tokens
// constantly
type TokenTracker struct {
	method            string
	tokenEndpoint     string
	tokenInfoEndpoint string
	clientId          string
	clientSecret      string
	scope             []string
	refreshToken      string
	token             string
	tokenExpires      time.Time
	tokenInfo         map[string]interface{}
}

// This method will return a new TokenTracker set up to cache and obtain new tokens in behalf
// of a oauth2 client. It will always use the client_credentials grant type to obtain the new
// tokens from the token endpoint.
func NewClientCredentialsTokenTracker(tokenEndpoint, tokenInfoEndpoint, client_id, client_secret string, scope []string) (tokenTracker *TokenTracker, err error) {
	tokenTracker = new(TokenTracker)
	tokenTracker.method = "client_credentials"
	tokenTracker.tokenEndpoint = tokenEndpoint
	tokenTracker.tokenInfoEndpoint = tokenInfoEndpoint
	tokenTracker.clientId = client_id
	tokenTracker.clientSecret = client_secret
	tokenTracker.scope = scope

	tokenTracker.token, tokenTracker.tokenExpires, err = GetClientCredentialsToken(tokenEndpoint, client_id, client_secret, scope)

	return
}

// This method will return a new TokenTracker set up to cache and obtain tokens on behalf of a
// particular user. It will use the password grant type to obtain the first token, and a refresh
// token. Subsequent tokens will be obtained by using the refresh_token grant type. It will not
// cache the username and password for the user. So if for any reason the refresh_token is
// invalidated or revoked special logic will need to be done (outside of this lib) to reset the
// TokenTracker so it doesn't forever remain in a bad state.
func NewPasswordTokenTracker(tokenEndpoint, tokenInfoEndpoint, client_id, client_secret, username, password string, scope []string) (tokenTracker *TokenTracker, err error) {
	tokenTracker = new(TokenTracker)
	tokenTracker.method = "refresh"
	tokenTracker.tokenEndpoint = tokenEndpoint
	tokenTracker.tokenInfoEndpoint = tokenInfoEndpoint
	tokenTracker.clientId = client_id
	tokenTracker.clientSecret = client_secret
	tokenTracker.scope = scope

	tokenTracker.token, tokenTracker.tokenExpires, tokenTracker.refreshToken, err = GetPasswordToken(tokenEndpoint, client_id, client_secret, username, password, scope)

	return
}

// This method of the token tracker will verify the veracity of the token against the token info
// endpoint before it returns it to the user. If the token info endpoint returns anything other
// than a 200 status, this endpoint will then attempt to get a fresh token.
func (tt *TokenTracker) GetSafeToken() (token string, err error) {
	if time.Now().Before(tt.tokenExpires.Add(time.Second * -10)) {
		tt.tokenInfo, err = tt.GetTokenInformation()
		if err != nil {
			token, err = tt.GetNewToken()
			tt.token = token
			return
		}
	} else {
		token, err = tt.GetNewToken()
	}
	return
}

// The get token method will grab a cached token from it's store if it has not expired (using the
// expires_in value from the original token call. It is not guaranteed to be a valid token as it
// may have been invalidated or revoked before it expired. If you want to ensure the token is
// valid use the GetSafeToken() method.
func (tt *TokenTracker) GetToken() (token string, err error) {
	if time.Now().Before(tt.tokenExpires.Add(time.Second * -10)) {
		token = tt.token
	} else {
		token, err = tt.GetNewToken()
	}
	return
}

// This method will fetch a new token from the token endpoint. It will replace any cached tokens
// that the tracker has. It uses the client credentials to get a new token on behalf of a client,
// and uses the refresh token to get a token on behalf of a client and user combination.
func (tt *TokenTracker) GetNewToken() (token string, err error) {
	tt.tokenInfo = nil
	switch tt.method {
	case "client_credentials":
		var tokenExpires time.Time
		token, tokenExpires, err = GetClientCredentialsToken(tt.tokenEndpoint, tt.clientId, tt.clientSecret, tt.scope)
		tt.token = token
		tt.tokenExpires = tokenExpires
	case "refresh":
		var tokenExpires time.Time
		token, tokenExpires, err = GetRefreshToken(tt.tokenEndpoint, tt.clientId, tt.clientSecret, tt.refreshToken, tt.scope)
		tt.token = token
		tt.tokenExpires = tokenExpires
	default:
		err = errors.New("Unknown Method Type on TokenTracker")
	}
	return
}

// This method will return the cached token information if it is available, or call the token
// information endpoint if it doesn't have any.
func (tt *TokenTracker) GetTokenInformation() (tokenInfo map[string]interface{}, err error) {
	if time.Now().Before(tt.tokenExpires.Add(time.Second * -10)) {
		if tt.tokenInfo != nil {
			tokenInfo = tt.tokenInfo
		} else {
			tokenInfo, err = GetTokenInformation(tt.tokenInfoEndpoint, tt.token)
			if err == nil {
				tt.tokenInfo = tokenInfo
			}
		}
	} else {
		err = errors.New("The token has expired")
	}

	return
}

// This method uses the password grant type of oauth2 to get a token from the token endpoint.
func GetPasswordToken(tokenEndpoint, client_id, client_secret, username, password string, scope []string) (token string, tokenExpires time.Time, refreshToken string, err error) {
	toSend := &url.Values{}
	toSend.Add("grant_type", "password")
	toSend.Add("access_type", "offline")
	toSend.Add("username", username)
	toSend.Add("password", password)
	toSend.Add("scope", strings.Join(scope, " "))

	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBuffer([]byte(toSend.Encode())))
	if err != nil {
		return
	}
	req.SetBasicAuth(client_id, client_secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		err = errors.New("Token Endpoint returned status " + resp.Status + " and Content-Type " + resp.Header.Get("Content-Type"))
		return
	}

	var tokenObj map[string]interface{}
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&tokenObj)
	if err != nil {
		return
	}

	if _, ok := tokenObj["error"]; ok {
		err = errors.New(tokenObj["error"].(string))
		return
	}

	token, _ = tokenObj["access_token"].(string)
	ei, _ := tokenObj["expires_in"].(float64)
	tokenExpires = time.Now().Add(time.Second * time.Duration(ei))
	refreshToken, _ = tokenObj["refresh_token"].(string)
	return
}

// This method uses the refresh_token grant type of oauth2 to obtain a token from the token endpoint
func GetRefreshToken(tokenEndpoint, client_id, client_secret, refresh_token string, scope []string) (token string, tokenExpires time.Time, err error) {
	toSend := &url.Values{}
	toSend.Add("grant_type", "refresh_token")
	toSend.Add("refresh_token", refresh_token)
	toSend.Add("scope", strings.Join(scope, " "))

	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBuffer([]byte(toSend.Encode())))
	if err != nil {
		return
	}
	req.SetBasicAuth(client_id, client_secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		err = errors.New("Token Endpoint returned status " + resp.Status + " and Content-Type " + resp.Header.Get("Content-Type"))
		return
	}

	var tokenObj map[string]interface{}
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&tokenObj)
	if err != nil {
		return
	}

	if _, ok := tokenObj["error"]; ok {
		err = errors.New(tokenObj["error"].(string))
		return
	}

	token = tokenObj["access_token"].(string)
	tokenExpires = time.Now().Add(time.Second * time.Duration(tokenObj["expires_in"].(float64)))
	return
}

// This method uses the client_credentials grant_type of oauth2 to obtain a token from the token
// endpoint.
func GetClientCredentialsToken(tokenEndpoint, client_id, client_secret string, scope []string) (token string, tokenExpires time.Time, err error) {
	toSend := &url.Values{}
	toSend.Add("grant_type", "client_credentials")
	toSend.Add("scope", strings.Join(scope, " "))

	req, err := http.NewRequest("POST", tokenEndpoint, bytes.NewBuffer([]byte(toSend.Encode())))
	if err != nil {
		return
	}
	req.SetBasicAuth(client_id, client_secret)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		err = errors.New("Token Endpoint returned status " + resp.Status + " and Content-Type " + resp.Header.Get("Content-Type"))
		return
	}

	var tokenObj map[string]interface{}
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&tokenObj)
	if err != nil {
		return
	}

	if _, ok := tokenObj["error"]; ok {
		err = errors.New(tokenObj["error"].(string))
		return
	}

	token = tokenObj["access_token"].(string)
	tokenExpires = time.Now().Add(time.Second * time.Duration(tokenObj["expires_in"].(float64)))
	return
}

// This method calls the token information endpoint and returns the json response as a map of
// string to interface{} values.
func GetTokenInformation(tokenInfoEndpoint, token string) (tokenInfo map[string]interface{}, err error) {
	req, err := http.NewRequest("GET", tokenInfoEndpoint, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 || !strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		err = errors.New("Token Info Endpoint returned status " + resp.Status + " and Content-Type " + resp.Header.Get("Content-Type"))
		return
	}

	d := json.NewDecoder(resp.Body)
	err = d.Decode(&tokenInfo)
	if err != nil {
		return
	}

	return
}
