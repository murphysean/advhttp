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

func NewClientCredentialsTokenTracker(tokenEndpoint, tokenInfoEndpoint, client_id, client_secret string, scope []string) (tokenTracker *TokenTracker, err error) {
	token, tokenExpires, err := GetClientCredentialsToken(tokenEndpoint, client_id, client_secret, scope)
	if err != nil {
		return
	}

	//TODO Should I just GetTokenInfo

	tokenTracker = new(TokenTracker)
	tokenTracker.method = "client_credentials"
	tokenTracker.tokenEndpoint = tokenEndpoint
	tokenTracker.tokenInfoEndpoint = tokenInfoEndpoint
	tokenTracker.clientId = client_id
	tokenTracker.clientSecret = client_secret
	tokenTracker.scope = scope
	tokenTracker.token = token
	tokenTracker.tokenExpires = tokenExpires

	return
}

func NewPasswordTokenTracker(tokenEndpoint, tokenInfoEndpoint, client_id, client_secret, username, password string, scope []string) (tokenTracker *TokenTracker, err error) {
	token, tokenExpires, refreshToken, err := GetPasswordToken(tokenEndpoint, client_id, client_secret, username, password, scope)
	if err != nil {
		return
	}

	//TODO Should I just GetTokenInfo

	tokenTracker = new(TokenTracker)
	tokenTracker.method = "client_credentials"
	tokenTracker.tokenEndpoint = tokenEndpoint
	tokenTracker.tokenInfoEndpoint = tokenInfoEndpoint
	tokenTracker.clientId = client_id
	tokenTracker.clientSecret = client_secret
	tokenTracker.scope = scope
	tokenTracker.token = token
	tokenTracker.tokenExpires = tokenExpires
	tokenTracker.refreshToken = refreshToken

	return
}

func (tt *TokenTracker) GetToken() (token string, err error) {
	if time.Now().Before(tt.tokenExpires) {
		token = tt.token
	} else {
		tt.tokenInfo = nil
		switch tt.method {
		case "client_credentials":
			var tokenExpires time.Time
			token, tokenExpires, err = GetClientCredentialsToken(tt.tokenEndpoint, tt.clientId, tt.clientSecret, tt.scope)
			tt.tokenExpires = tokenExpires
		default:
			err = errors.New("Unknown Method Type on TokenTracker")
		}
		return
	}

	return
}

func (tt *TokenTracker) GetTokenInformation() (tokenInfo map[string]interface{}, err error) {
	if time.Now().Before(tt.tokenExpires) {
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

func GetPasswordToken(tokenEndpoint, client_id, client_secret, username, password string, scope []string) (token string, tokenExpires time.Time, refreshToken string, err error) {
	client := &http.Client{}

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

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
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

func GetRefreshToken(tokenEndpoint, client_id, client_secret, refresh_token string, scope []string) (token string, tokenExpires time.Time, err error) {
	client := &http.Client{}

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

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
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

func GetClientCredentialsToken(tokenEndpoint, client_id, client_secret string, scope []string) (token string, tokenExpires time.Time, err error) {
	client := &http.Client{}

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

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
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

func GetTokenInformation(tokenInfoEndpoint, token string) (tokenInfo map[string]interface{}, err error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", tokenInfoEndpoint, nil)
	if err != nil {
		return
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&tokenInfo)
	if err != nil {
		return
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err = errors.New("Token Info Endpoint returned status: " + resp.Status)
	}

	return
}
