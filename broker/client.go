package broker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const baseURL = "https://api.buildings.ability.abb/buildings/openbos/apiproxy/v1"

type OpenBOSClient struct {
	GatewayID    string
	HTTPClient   *http.Client
	AccessToken  string
	clientID     string
	clientSecret string
}

func NewOpenBOSClient(gatewayID, clientID, clientSecret string) (*OpenBOSClient, error) {
	client := &OpenBOSClient{
		GatewayID:    gatewayID,
		HTTPClient:   &http.Client{Timeout: 10 * time.Second},
		clientID:     clientID,
		clientSecret: clientSecret,
	}

	if err := client.authenticateWithClientCredentials(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *OpenBOSClient) authenticateWithClientCredentials() error {
	tokenURL := "https://login.microsoftonline.com/372ee9e0-9ce0-4033-a64a-c07073a91ecd/oauth2/v2.0/token"

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("scope", "api://openbos/.default")

	req, err := http.NewRequest("POST", tokenURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
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

	c.AccessToken = accessToken
	return nil
}

type DataPoint struct {
	BusUnitId     *string  `json:"busUnitId"`
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Direction     string   `json:"direction"`
	TypeId        string   `json:"typeId"`
	DisplayUnitId *string  `json:"displayUnitId"`
	PublicId      string   `json:"publicId"`
}

type Property struct {
	DefaultValue  any      `json:"defaultValue"`
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	Direction     string   `json:"direction"`
	TypeId        string   `json:"typeId"`
	DisplayUnitId *string  `json:"displayUnitId"`
	PublicId      string   `json:"publicId"`
}

type AssetTemplate struct {
	Datapoints         []DataPoint   `json:"datapoints"`
	Properties         []Property    `json:"properties"`
	Usages             []interface{} `json:"usages"`
	Id                 string        `json:"id"`
	Icon               string        `json:"icon"`
	IconFillColor      *string       `json:"iconFillColor"`
	Name               string        `json:"name"`
	Tags               []string      `json:"tags"`
	ParentId           *string       `json:"parentId"`
	Version            string        `json:"version"`
	InstancesCount     int           `json:"instancesCount"`
	PublicId           string        `json:"publicId"`
	IsExternal         bool          `json:"isExternal"`
	SupportMasterSlave bool          `json:"supportMasterSlave"`
	IsDefault          bool          `json:"isDefault"`
}

func (c *OpenBOSClient) getAssetTemplates() ([]AssetTemplate, error) {
	url := fmt.Sprintf("%s/gateway/%s/api/v1/ontology/functionalblocktemplate/details", baseURL, c.GatewayID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.AccessToken)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("requesting: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading body: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d body: %s", resp.StatusCode, string(body))
	}

	var templates []AssetTemplate
	if err := json.Unmarshal(body, &templates); err != nil {
		return nil, fmt.Errorf("unmarshalling: %v", err)
	}

	return templates, nil
}
