package fakediscord

import "github.com/bwmarrin/discordgo"

func Configure(baseURL string) {
	overrideEndPoints(baseURL)
}

// overrideEndpoints overrides the package global endpoints in bwmarrin/discordgo, in order to enable overriding the
// Discord API URL with the fakediscord base URL
//
// As fakediscord grows we will need to add more endpoints here, or come up with a better way to override the base URL
func overrideEndPoints(baseURL string) {
	discordgo.EndpointDiscord = baseURL
	discordgo.EndpointAPI = discordgo.EndpointDiscord + "api/v" + discordgo.APIVersion + "/"
	discordgo.EndpointGateway = discordgo.EndpointAPI + "gateway"
	discordgo.EndpointChannels = discordgo.EndpointAPI + "channels/"
	discordgo.EndpointGuildCreate = discordgo.EndpointAPI + "guilds"
	discordgo.EndpointGuilds = discordgo.EndpointAPI + "guilds/"
	discordgo.EndpointUsers = discordgo.EndpointAPI + "users/"
	discordgo.EndpointApplications = discordgo.EndpointAPI + "applications"
	discordgo.EndpointWebhooks = discordgo.EndpointAPI + "webhooks/"
}
