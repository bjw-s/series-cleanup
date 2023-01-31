// Package trakt implements the Trakt.tv API
package trakt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bjw-s/seriescleanup/internal/helpers"
)

type deviceCodePayload struct {
	ClientID string `json:"client_id"`
}

type deviceCodePollPayload struct {
	Code         string `json:"code"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
}

type deviceCode struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURL string `json:"verification_url"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

func (code *deviceCode) ExchangeForAccessToken(api *API) (*accessToken, error) {
	deviceCodePollPayload := deviceCodePollPayload{}
	deviceCodePollPayload.Code = code.DeviceCode
	deviceCodePollPayload.ClientID = api.ClientID
	deviceCodePollPayload.ClientSecret = api.ClientSecret

	result, err := api.sendRequest(http.MethodPost, "/oauth/device/token", deviceCodePollPayload)
	if err != nil {
		return nil, err
	}
	return getAccessTokenDataFromAPIResponse(result)
}

type accessTokenRefreshPayload struct {
	RefreshToken string `json:"refresh_token"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	RedirectURI  string `json:"redirect_uri"`
	GrantType    string `json:"grant_type"`
}

type accessToken struct {
	AccessToken          string `json:"access_token"`
	TokenType            string `json:"token_type"`
	ExpiresIn            int64  `json:"expires_in"`
	ExpirationWithBuffer int64
	RefreshToken         string `json:"refresh_token"`
	Scope                string `json:"scope"`
	CreatedAt            int64  `json:"created_at"`
}

func (token *accessToken) WriteToFile(path string) error {
	token.ExpirationWithBuffer = token.ExpiresIn * 3 / 4
	jsonString, err := json.Marshal(token)
	if err != nil {
		return err
	}

	err = os.WriteFile(path, jsonString, os.ModePerm)
	if err != nil {
		return err
	}

	return nil
}

func (token *accessToken) ReadFromFile(path string) error {
	if !helpers.FileExists(path) {
		return fmt.Errorf("file %v does not exist", path)
	}

	file, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	err = json.Unmarshal(file, &token)
	if err != nil {
		return err
	}

	return nil
}

func (token *accessToken) DeleteCacheFile(path string) error {
	if !helpers.FileExists(path) {
		return fmt.Errorf("file %v does not exist", path)
	}

	err := os.Remove(path)
	if err != nil {
		return err
	}

	return nil
}

func (token *accessToken) HasExpired() bool {
	accessTokenExpirationDate := time.Unix(token.CreatedAt+token.ExpiresIn, 0)
	currentTime := time.Now()
	return accessTokenExpirationDate.Sub(currentTime) < 0
}

func (token *accessToken) WillExpireSoon() bool {
	accessTokenExpirationDateWithBuffer := time.Unix(token.CreatedAt+token.ExpirationWithBuffer, 0)
	currentTime := time.Now()
	return accessTokenExpirationDateWithBuffer.Sub(currentTime) < 0
}

func (token *accessToken) Refresh(api *API) (*accessToken, error) {
	accessTokenRefreshPayload := accessTokenRefreshPayload{}
	accessTokenRefreshPayload.RefreshToken = token.RefreshToken
	accessTokenRefreshPayload.ClientID = api.ClientID
	accessTokenRefreshPayload.ClientSecret = api.ClientSecret
	accessTokenRefreshPayload.RedirectURI = "urn:ietf:wg:oauth:2.0:oob"
	accessTokenRefreshPayload.GrantType = "refresh_token"

	result, err := api.sendRequest(http.MethodPost, "/oauth/token", accessTokenRefreshPayload)
	if err != nil {
		return nil, err
	}
	return getAccessTokenDataFromAPIResponse(result)
}

func getAccessTokenDataFromAPIResponse(response *apiResponse) (*accessToken, error) {
	switch response.StatusCode {
	case 200:
		accessTokenData := accessToken{}
		err := json.Unmarshal(response.Body, &accessTokenData)
		if err != nil {
			return nil, err
		}
		return &accessTokenData, nil
	default:
		return nil, fmt.Errorf("could not get access token")
	}
}

func authenticateFromFile(file string, api *API) (*accessToken, error) {
	accessTokenData := &accessToken{}
	err := accessTokenData.ReadFromFile(file)
	if err != nil {
		return nil, err
	}

	if accessTokenData.HasExpired() {
		err = accessTokenData.DeleteCacheFile(file)
		if err != nil {
			return nil, err
		}
		accessTokenData, err = authenticateWithDeviceToken(api)
		if err != nil {
			return nil, err
		}
	} else if accessTokenData.WillExpireSoon() {
		accessTokenData, err = accessTokenData.Refresh(api)
		if err != nil {
			return nil, err
		}
	}

	return accessTokenData, nil
}

func authenticateWithDeviceToken(api *API) (*accessToken, error) {
	deviceCode, err := getDeviceToken(api)
	if err != nil {
		return nil, err
	}

	fmt.Printf("\n")
	fmt.Printf("Please authorize this script to connect to your trakt account\n")
	fmt.Printf("Go to %v\n", deviceCode.VerificationURL)
	fmt.Printf("and enter the following code: %v\n", deviceCode.UserCode)
	fmt.Printf("\n")
	fmt.Printf("The script will wait for 10 minutes for authorization before aborting\n")
	fmt.Printf("\n")

	waitTimer := 0
	for waitTimer < deviceCode.ExpiresIn {
		accessTokenData, err := deviceCode.ExchangeForAccessToken(api)
		if err != nil {
			return nil, err
		}

		if accessTokenData != nil {
			return accessTokenData, nil
		}

		waitTimer = waitTimer + deviceCode.Interval
		time.Sleep(time.Duration(deviceCode.Interval) * time.Second)
	}

	return nil, fmt.Errorf("could not exchange device code for access token")
}

func getDeviceToken(api *API) (*deviceCode, error) {
	deviceCodePayload := deviceCodePayload{}
	deviceCodePayload.ClientID = api.ClientID

	result, err := api.sendRequest(http.MethodPost, "/oauth/device/code", deviceCodePayload)
	if err != nil {
		return nil, err
	}
	deviceCodeData := deviceCode{}
	err = json.Unmarshal(result.Body, &deviceCodeData)
	if err != nil {
		return nil, err
	}
	return &deviceCodeData, nil
}

// Authenticate against the Trakt API
func (api *API) Authenticate() error {
	err := api.validate()
	if err != nil {
		return err
	}

	var authDatafile = api.DataPath + "/trakt.json"
	var accessToken *accessToken

	if helpers.FileExists(authDatafile) {
		accessToken, err = authenticateFromFile(authDatafile, api)
		if err != nil {
			return err
		}
	} else {
		accessToken, err = authenticateWithDeviceToken(api)
		if err != nil {
			return err
		}
	}

	if accessToken != nil {
		err = accessToken.WriteToFile(authDatafile)
		if err != nil {
			return err
		}
		api.accessToken = accessToken.AccessToken
		api.IsAuthenticated = true
		return nil
	}

	return nil
}
