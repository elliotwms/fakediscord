package fakediscord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Client struct {
	baseURL string
	http    *http.Client
	token   string
}

// NewClient creates a client for fakediscord's internal endpoints. Set the fakediscord URL with WithBaseURL (or the
// deprecated Configure)
func NewClient(token string) *Client {
	return &Client{
		baseURL: discordgo.EndpointDiscord,
		http:    http.DefaultClient,
		token:   token,
	}
}

func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	c.http = httpClient

	return c
}

// WithBaseURL sets the URL of the fakediscord instance, e.g. "http://localhost:8080/"
func (c *Client) WithBaseURL(baseURL string) *Client {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	c.baseURL = baseURL

	return c
}

func (c *Client) Interaction(i *discordgo.InteractionCreate) (*discordgo.InteractionCreate, error) {
	bs, err := json.Marshal(i)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal interaction: %w", err)
	}

	res, err := c.do(http.MethodPost, "interactions", bs)
	if err != nil {
		return nil, fmt.Errorf("failed to send interaction: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusCreated {
		if res.StatusCode == http.StatusBadRequest || res.StatusCode == http.StatusNotFound {
			s := &struct {
				Error string `json:"error"`
			}{}

			err = json.NewDecoder(res.Body).Decode(s)
			if err != nil {
				return nil, fmt.Errorf("failed to parse interaction response: %w", err)
			}

			return nil, errors.New(s.Error)
		}
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	i2 := &discordgo.InteractionCreate{}

	err = json.NewDecoder(res.Body).Decode(i2)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal interaction: %w", err)
	}

	return i2, nil
}

func (c *Client) do(method string, path string, bs []byte) (*http.Response, error) {
	if c.baseURL == "https://"+discordHost+"/" {
		return nil, errors.New("fakediscord base URL not set: use WithBaseURL")
	}

	u := c.baseURL + "api/v" + discordgo.APIVersion + "/" + path
	req, err := http.NewRequest(method, u, bytes.NewBuffer(bs))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bot "+c.token)

	return c.http.Do(req)
}
