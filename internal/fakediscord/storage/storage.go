package storage

import "sync"

var (
	Channels             sync.Map // Channel ID : discordgo.Channel
	Commands             sync.Map // Command ID : discordgo.ApplicationCommand
	CommandNames         sync.Map // type:name : Command ID
	Guilds               sync.Map // Guild ID : discordgo.Guild
	Interactions         sync.Map // token : discordgo.Interaction
	InteractionResponses sync.Map // token : Message ID
	InteractionCallbacks sync.Map // Interaction ID : {}
	Messages             sync.Map // id : discordgo.Message
	Users                sync.Map // User ID : discordgo.User
)
