package main

import (
	"bytes"
	"encoding/json"
	"errors"
	log "github.com/Sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

// HashType is an encryption method for basic auth keys
type HashType string
type Endpoint string // A type for endpoints

// AccessSpecs define what URLS a user has access to an what methods are enabled
type AccessSpec struct {
	URL     string   `json:"url"`
	Methods []string `json:"methods"`
}

// AccessDefinition defines which versions of an API a key has access to
type AccessDefinition struct {
	APIName     string       `json:"api_name"`
	APIID       string       `json:"api_id"`
	Versions    []string     `json:"versions"`
	AllowedURLs []AccessSpec `bson:"allowed_urls"  json:"allowed_urls"` // mapped string MUST be a valid regex
}

// SessionState objects represent a current API session, mainly used for rate limiting.
type SessionState struct {
	LastCheck        int64                       `json:"last_check"`
	Allowance        float64                     `json:"allowance"`
	Rate             float64                     `json:"rate"`
	Per              float64                     `json:"per"`
	Expires          int64                       `json:"expires"`
	QuotaMax         int64                       `json:"quota_max"`
	QuotaRenews      int64                       `json:"quota_renews"`
	QuotaRemaining   int64                       `json:"quota_remaining"`
	QuotaRenewalRate int64                       `json:"quota_renewal_rate"`
	AccessRights     map[string]AccessDefinition `json:"access_rights"`
	OrgID            string                      `json:"org_id"`
	OauthClientID    string                      `json:"oauth_client_id"`
	OauthKeys        map[string]string           `json:"oauth_keys"`
	BasicAuthData    struct {
		Password string   `json:"password"`
		Hash     HashType `json:"hash_type"`
	} `json:"basic_auth_data"`
	JWTData struct {
		Secret string `json:"secret"`
	} `json:"jwt_data"`
	HMACEnabled   bool   `json:"hmac_enabled"`
	HmacSecret    string `json:"hmac_string"`
	IsInactive    bool   `json:"is_inactive"`
	ApplyPolicyID string `json:"apply_policy_id"`
	DataExpires   int64  `json:"data_expires"`
	Monitor       struct {
		TriggerLimits []float64 `json:"trigger_limits"`
	} `json:"monitor"`
	MetaData interface{} `json:"meta_data"`
	Tags     []string    `json:"tags"`
}

type OAuthResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
	RedirectTo  string `json:"redirect_to"`
	TokenType   string `json:"token_type"`
}

func generateBasicTykSesion(baseAPIID, baseVersion, policyID, orgID string) SessionState {
	// Create a generic access token withour policy
	basicSessionState := SessionState{
		Allowance:        1,
		Rate:             1,
		Per:              1,
		Expires:          -1,
		QuotaMax:         1,
		QuotaRenews:      60,
		QuotaRemaining:   1,
		QuotaRenewalRate: 1,
		AccessRights:     map[string]AccessDefinition{},
		OrgID:            orgID,
		ApplyPolicyID:    policyID,
		MetaData:         map[string]interface{}{"Origin": "TAP"},
		Tags:             []string{"TykOrigin-TAP"},
	}

	accessEntry := AccessDefinition{
		APIName:  "Base",
		APIID:    baseAPIID,
		Versions: []string{baseVersion},
	}
	basicSessionState.AccessRights[baseAPIID] = accessEntry

	return basicSessionState
}

const (
	HASH_PlainText  HashType = ""
	HASH_BCrypt     HashType = "bcrypt"
	OAUTH_AUTHORIZE Endpoint = "tyk/oauth/authorize-client/"
)

func readBody(response *http.Response) ([]byte, error) {
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return []byte(""), err
	}

	return contents, nil

}

// DispatchGateway will dispatch a request to the gateway API
func DispatchGateway(target Endpoint, method string, body io.Reader, ctype string) ([]byte, error) {
	preparedEndpoint := GatewayHost + string(target)

	log.Debug("Calling: ", preparedEndpoint)
	newRequest, err := http.NewRequest(method, preparedEndpoint, body)
	if err != nil {
		log.Error("Failed to create request")
		log.Error(err)
	}

	if ctype == "" {
		ctype = "application/json"
	}

	newRequest.Header.Add("x-tyk-authorization", AdminSecret)
	newRequest.Header.Add("content-type", ctype)
	c := &http.Client{}
	response, reqErr := c.Do(newRequest)

	if reqErr != nil {
		return []byte{}, reqErr
	}

	retBody, bErr := readBody(response)
	if bErr != nil {
		return []byte{}, bErr
	}

	if response.StatusCode > 201 {
		log.Warning("Response code was: ", response.StatusCode)
		return retBody, errors.New("Response code was not 200!")
	}

	log.Debug("API Response: ", string(retBody))

	return retBody, nil
}

func RequestOAuthToken(APIlistenPath, redirect_uri, responseType, clientId, secret, orgID, policyID, BaseAPIID string) (*OAuthResponse, error) {
	// Create a generic access token withour policy
	basicSessionState := generateBasicTykSesion(BaseAPIID, "Default", policyID, orgID)
	basicSessionState.OauthClientID = clientId

	keyDataJSON, err := json.Marshal(basicSessionState)

	if err != nil {
		return nil, err
	}

	if clientId == "" {
		return nil, errors.New("Requires client ID")
	}

	// Make the Auth request
	response := &OAuthResponse{}
	target := "/" + strings.Join([]string{APIlistenPath, string(OAUTH_AUTHORIZE)}, "/")
	data := "response_type=" + responseType
	data += "&client_id=" + clientId
	data += "&redirect_uri=" + redirect_uri
	data += "&key_rules=" + url.QueryEscape(string(keyDataJSON))

	log.Debug("Request data sent: ", data)

	body := bytes.NewBuffer([]byte(data))
	rawResp, dErr := DispatchGateway(Endpoint(target), "POST", body, "application/x-www-form-urlencoded")
	if dErr != nil {
		return nil, err
	}

	log.Debug("Returned: ", rawResp)
	decErr := json.Unmarshal(rawResp, &response)
	if decErr != nil {
		return nil, decErr
	}

	return response, nil
}
