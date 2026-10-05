package storage

import (
	"slices"
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

	// reacting is idempotent: a user can only react once with each emoji
	if slices.Contains(r.messages[message][reaction], user) {
		return
	}

	r.messages[message][reaction] = append(r.messages[message][reaction], user)
}

func (r *reactionStore) LoadMessageReaction(message, reaction string) (users []string, ok bool) {
	r.mx.RLock()
	defer r.mx.RUnlock()

	users, ok = r.messages[message][reaction]

	// return a copy so that callers do not share the backing array with the store
	return slices.Clone(users), ok
}

func (r *reactionStore) DeleteMessageReactions(message string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	delete(r.messages, message)
}

func (r *reactionStore) DeleteMessageReaction(message, reaction, user string) {
	r.mx.Lock()
	defer r.mx.Unlock()

	if _, ok := r.messages[message]; !ok {
		return
	}

	users, ok := r.messages[message][reaction]
	if !ok {
		return
	}

	// build a new slice rather than modifying the existing one in place, as it may still be held by a reader
	r.messages[message][reaction] = slices.DeleteFunc(slices.Clone(users), func(u string) bool {
		return u == user
	})
}
