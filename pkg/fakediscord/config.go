package fakediscord

import "github.com/bwmarrin/discordgo"

func Configure(baseURL string) {
	overrideEndPoints(baseURL)
}

// overrideEndpoints overrides the package global endpoints in bwmarrin/discordgo, in order to enable overriding the
// Discord API URL with the fakediscord base URL
func overrideEndPoints(baseURL string) {
	discordgo.EndpointDiscord = baseURL
	discordgo.EndpointAPI = discordgo.EndpointDiscord + "api/v" + discordgo.APIVersion + "/"
	discordgo.EndpointGateway = discordgo.EndpointAPI + "gateway"

	// as fakediscord grows we may need to add more endpoints here
}
