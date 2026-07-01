package headscale

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (c *Client) BaseURL() string {
	return c.baseURL
}

func (c *Client) do(method, path string, body any) (*http.Response, error) {
	endpoint, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("build URL: %w", err)
	}

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, endpoint, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP request: %w", err)
	}

	if resp.StatusCode >= 400 {
		raw, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, fmt.Errorf("headscale %s %s: HTTP %d: %s", method, path, resp.StatusCode, string(raw))
	}

	return resp, nil
}

type User struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"createdAt,omitempty"`
}

type createUserRequest struct {
	Name string `json:"name"`
}

type createUserResponse struct {
	User User `json:"user"`
}

func (c *Client) CreateUser(name string) (User, error) {
	resp, err := c.do("POST", "/api/v1/user", createUserRequest{Name: name})
	if err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result createUserResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return User{}, fmt.Errorf("decode user: %w", err)
	}
	return result.User, nil
}

type listUsersResponse struct {
	Users []User `json:"users"`
}

func (c *Client) ListUsers() ([]User, error) {
	resp, err := c.do("GET", "/api/v1/user", nil)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result listUsersResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}
	return result.Users, nil
}

func (c *Client) DeleteUser(id uint64) error {
	path := fmt.Sprintf("/api/v1/user/%d", id)
	resp, err := c.do("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

type PreAuthKey struct {
	ID        uint64 `json:"id"`
	Key       string `json:"key"`
	Reusable  bool   `json:"reusable"`
	Ephemeral bool   `json:"ephemeral"`
	Used      bool   `json:"used"`
}

type createPreAuthKeyRequest struct {
	User       uint64   `json:"user"`
	Reusable   bool     `json:"reusable"`
	Ephemeral  bool     `json:"ephemeral"`
	Expiration string   `json:"expiration,omitempty"`
	AclTags    []string `json:"aclTags,omitempty"`
}

type createPreAuthKeyResponse struct {
	PreAuthKey PreAuthKey `json:"preAuthKey"`
}

func (c *Client) CreatePreAuthKey(userID uint64, reusable, ephemeral bool, expiration time.Time) (PreAuthKey, error) {
	req := createPreAuthKeyRequest{
		User:       userID,
		Reusable:   reusable,
		Ephemeral:  ephemeral,
		Expiration: expiration.UTC().Format(time.RFC3339),
	}
	resp, err := c.do("POST", "/api/v1/preauthkey", req)
	if err != nil {
		return PreAuthKey{}, fmt.Errorf("create preauthkey: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result createPreAuthKeyResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PreAuthKey{}, fmt.Errorf("decode preauthkey: %w", err)
	}
	return result.PreAuthKey, nil
}

type listPreAuthKeysResponse struct {
	PreAuthKeys []PreAuthKey `json:"preAuthKeys"`
}

func (c *Client) ListPreAuthKeys() ([]PreAuthKey, error) {
	resp, err := c.do("GET", "/api/v1/preauthkey", nil)
	if err != nil {
		return nil, fmt.Errorf("list preauthkeys: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result listPreAuthKeysResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode preauthkeys: %w", err)
	}
	return result.PreAuthKeys, nil
}

func (c *Client) ExpirePreAuthKey(id uint64) error {
	resp, err := c.do("POST", "/api/v1/preauthkey/expire", map[string]uint64{"id": id})
	if err != nil {
		return fmt.Errorf("expire preauthkey: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

func (c *Client) DeletePreAuthKey(id uint64) error {
	resp, err := c.do("DELETE", "/api/v1/preauthkey", map[string]uint64{"id": id})
	if err != nil {
		return fmt.Errorf("delete preauthkey: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}

type Node struct {
	ID           uint64   `json:"id"`
	MachineKey    string   `json:"machineKey"`
	NodeKey      string   `json:"nodeKey"`
	IPAddresses   []string `json:"ipAddresses"`
	Name         string   `json:"name"`
	GivenName    string   `json:"givenName"`
	User         User     `json:"user"`
	Online       bool     `json:"online"`
	Tags         []string `json:"tags"`
}

type listNodesResponse struct {
	Nodes []Node `json:"nodes"`
}

func (c *Client) ListNodes() ([]Node, error) {
	resp, err := c.do("GET", "/api/v1/node", nil)
	if err != nil {
		return nil, fmt.Errorf("list nodes: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result listNodesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode nodes: %w", err)
	}
	return result.Nodes, nil
}

func (c *Client) DeleteNode(nodeID uint64) error {
	path := fmt.Sprintf("/api/v1/node/%d", nodeID)
	resp, err := c.do("DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("delete node: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return nil
}
