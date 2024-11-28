//  This file is part of the Eliona project.
//  Copyright © 2024 IoTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package broker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/log"
)

// const baseURL = "https://api.buildings.ability.abb/buildings/openbos/apiproxy/v1"
// const baseURL = "http://localhost:5000"
const baseURL = "https://dev.api.buildings.ability.abb/buildings/openbos/apiproxy/v1"

const tokenURL = "https://login.microsoftonline.com/372ee9e0-9ce0-4033-a64a-c07073a91ecd/oauth2/v2.0/token"

type openBOSClient struct {
	gatewayID    string
	httpClient   *http.Client
	accessToken  string
	clientID     string
	clientSecret string
	webhookURL   string
	baseURL      string
	tokenURL     string
}

// Defined as variable function to allow overriding in tests
var newOpenBOSClient = func(gatewayID, clientID, clientSecret, webhookURL, baseURL, tokenURL string) (*openBOSClient, error) {
	client := &openBOSClient{
		gatewayID:    gatewayID,
		httpClient:   &http.Client{Timeout: 10 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
		webhookURL:   webhookURL,
		baseURL:      baseURL,
		tokenURL:     tokenURL,
	}

	if err := client.authenticateWithClientCredentials(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *openBOSClient) authenticateWithClientCredentials() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("scope", "api://openbos/.default")

	req, err := http.NewRequest("POST", c.tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to obtain access token, status code: %d body: %s", resp.StatusCode, string(body))
	}

	var tokenResponse map[string]interface{}
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return fmt.Errorf("unmarshalling: %v", err)
	}

	accessToken, ok := tokenResponse["access_token"].(string)
	if !ok {
		return errors.New("invalid token response")
	}

	c.accessToken = accessToken
	return nil
}

func (c *openBOSClient) doRequest(method, endpoint string, queryParams url.Values, body interface{}, result interface{}) error {
	url := fmt.Sprintf("%s/gateway/%s/api/v1/%s", c.baseURL, c.gatewayID, endpoint)
	if queryParams != nil && len(queryParams) > 0 {
		url += "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d body: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %v", err)
		}
	}

	return nil
}

func (c *openBOSClient) doMockRequest(method, endpoint string, queryParams url.Values, body interface{}, result interface{}) error {
	url := fmt.Sprintf("%s/api/v1/%s", c.baseURL, endpoint)
	if queryParams != nil && len(queryParams) > 0 {
		url += "?" + queryParams.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyBytes, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("encoding request body: %v", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// TODO: This part may be removed later
	// {
	dump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return fmt.Errorf("dumping request: %v", err)
	}
	log.Debug("client", "HTTP Request:\n%s\n", string(dump))

	// DumpRequest may consume the body, so we need to reset it afterward
	if bodyReader != nil {
		if seeker, ok := bodyReader.(io.Seeker); ok {
			seeker.Seek(0, io.SeekStart)
		}
		req.Body = io.NopCloser(bodyReader)
	}
	// }

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("requesting: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d body: %s", resp.StatusCode, string(bodyBytes))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("decoding response: %v", err)
		}
	}

	return nil
}
