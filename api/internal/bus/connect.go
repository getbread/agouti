package bus

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

func Connect(url string, capabilities map[string]interface{}, httpClient *http.Client) (*Client, bool, error) {
	requestBody, err := capabilitiesToJSON(capabilities)
	if err != nil {
		return nil, false, err
	}

	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	sessionID, hasStatus, err := openSession(url, requestBody, httpClient)
	if err != nil {
		return nil, false, err
	}

	sessionURL := fmt.Sprintf("%s/session/%s", url, sessionID)
	return &Client{sessionURL, httpClient}, hasStatus, nil
}

func capabilitiesToJSON(capabilities map[string]interface{}) (io.Reader, error) {
	if capabilities == nil {
		capabilities = map[string]interface{}{}
	}
	desiredCapabilities := struct {
		DesiredCapabilities map[string]interface{} `json:"desiredCapabilities"`
	}{capabilities}

	capabiltiesJSON, err := json.Marshal(desiredCapabilities)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(capabiltiesJSON), err
}

func openSession(url string, body io.Reader, httpClient *http.Client) (sessionID string, hasStatus bool, err error) {
	request, err := http.NewRequest("POST", fmt.Sprintf("%s/session", url), body)
	if err != nil {
		return "", hasStatus, err
	}

	request.Header.Add("Content-Type", "application/json")

	response, err := httpClient.Do(request)
	if err != nil {
		return "", hasStatus, err
	}
	defer response.Body.Close()

	var sessionResponse struct {
		SessionID string
		// fallback for GeckoDriver
		Value struct {
			SessionID string
		}
		Status *int `json:"status,omitempty"`
	}
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", hasStatus, err
	}

	if err := json.Unmarshal(responseBody, &sessionResponse); err != nil {
		return "", hasStatus, err
	}

	hasStatus = sessionResponse.Status != nil

	if sessionResponse.SessionID == "" {
		// fallback for GeckoDriver
		if sessionResponse.Value.SessionID != "" {
			return sessionResponse.Value.SessionID, hasStatus, nil
		}
		return "", hasStatus, errors.New("failed to retrieve a session ID")
	}

	return sessionResponse.SessionID, hasStatus, nil
}
