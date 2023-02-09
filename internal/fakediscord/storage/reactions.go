package storage

import (
	"sync"
)

var Reactions = &reactionStore{
	messages: make(map[string]ReactionsUsers),
}

type ReactionsUsers map[string][]string

type reactionStore struct {
	mx       sync.RWMutex
	messages map[string]ReactionsUsers
}

func (r *reactionStore) Store(message, reaction, user string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.messages[message]; !ok {
		r.messages[message] = map[string][]string{}
	}

	if _, ok := r.messages[message][reaction]; !ok {
		r.messages[message][reaction] = []string{}
	}

	r.messages[message][reaction] = append(r.messages[message][reaction], user)
}

func (r *reactionStore) LoadMessageReaction(message, reaction string) (users []string, ok bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	users, ok = r.messages[message][reaction]

	return
}

func (r *reactionStore) DeleteMessageReactions(message string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	delete(r.messages, message)
}
