package fakediscord

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bwmarrin/discordgo"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient() *Client {
	endpoint := discordgo.EndpointDiscord

	if endpoint == "https://discord.com/" {
		panic("fakediscord not configured. Call fakediscord.Configure(baseURL) before configuring client")
	}

	return &Client{
		baseURL: endpoint,
		http:    http.DefaultClient,
	}
}

func (c *Client) WithHTTPClient(httpClient *http.Client) *Client {
	c.http = httpClient

	return c
}

func (c *Client) WithBaseURL(baseURL string) *Client {
	c.baseURL = baseURL

	return c
}

func (c *Client) Interaction(i *discordgo.InteractionCreate) (*discordgo.InteractionCreate, error) {
	bs, err := json.Marshal(i)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal interaction: %w", err)
	}

	res, err := c.http.Post(c.baseURL+"api/v"+discordgo.APIVersion+"/interactions", "application/json", bytes.NewBuffer(bs))
	if err != nil {
		return nil, fmt.Errorf("failed to send interaction: %w", err)
	}

	if res.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	i2 := &discordgo.InteractionCreate{}

	err = json.NewDecoder(res.Body).Decode(i2)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal interaction: %w", err)
	}

	return i2, nil
}
