package fakediscord

import "github.com/bwmarrin/discordgo"

// Configure points discordgo at fakediscord by overriding its package-level endpoints.
//
// Deprecated: Configure only overrides a subset of discordgo's endpoints, and applies to every session in the process.
// Use ConfigureSession for each session instead, and NewClient(token).WithBaseURL(baseURL) for the client.
func Configure(baseURL string) {
	overrideEndPoints(baseURL)
}

// overrideEndpoints overrides the package global endpoints in bwmarrin/discordgo, in order to enable overriding the
// Discord API URL with the fakediscord base URL
func overrideEndPoints(baseURL string) {
	discordgo.EndpointDiscord = baseURL
	discordgo.EndpointAPI = discordgo.EndpointDiscord + "api/v" + discordgo.APIVersion + "/"
	discordgo.EndpointGateway = discordgo.EndpointAPI + "gateway"
	discordgo.EndpointGatewayBot = discordgo.EndpointGateway + "/bot"
	discordgo.EndpointChannels = discordgo.EndpointAPI + "channels/"
	discordgo.EndpointGuildCreate = discordgo.EndpointAPI + "guilds"
	discordgo.EndpointGuilds = discordgo.EndpointAPI + "guilds/"
	discordgo.EndpointUsers = discordgo.EndpointAPI + "users/"
	discordgo.EndpointApplications = discordgo.EndpointAPI + "applications"
	discordgo.EndpointWebhooks = discordgo.EndpointAPI + "webhooks/"
}
