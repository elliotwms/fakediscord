package storage

import "sync"

var Interactions sync.Map // token : discordgo.Interaction

var InteractionResponses sync.Map // token : Message ID
