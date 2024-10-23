package storage

import "sync"

var Messages sync.Map // id : discordgo.Message
