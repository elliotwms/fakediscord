package storage

import (
	"math"
	"sync"

	"github.com/bwmarrin/discordgo"
)

func init() {
	State = discordgo.NewState()
	State.MaxMessageCount = math.MaxInt
}

var (
	State *discordgo.State

	Commands             sync.Map // Command ID : discordgo.ApplicationCommand
	CommandNames         sync.Map // type:name : Command ID
	Interactions         sync.Map // token : discordgo.Interaction
	InteractionResponses sync.Map // token : Message ID
	InteractionCallbacks sync.Map // Interaction ID : {}
	Users                sync.Map // User ID : discordgo.User
)
