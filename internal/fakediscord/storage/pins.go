package storage

import (
	"slices"
	"sync"
)

var Pins = &pins{
	ps: make(map[string][]string),
}

type pins struct {
	mx sync.RWMutex
	ps map[string][]string
}

// Store pins the message in the channel. Pinning an already pinned message has no effect
func (p *pins) Store(channel, message string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	if slices.Contains(p.ps[channel], message) {
		return
	}

	p.ps[channel] = append(p.ps[channel], message)
}

// Delete unpins the message from the channel
func (p *pins) Delete(channel, message string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	p.ps[channel] = slices.DeleteFunc(slices.Clone(p.ps[channel]), func(m string) bool {
		return m == message
	})
}

func (p *pins) Load(channel string) []string {
	p.mx.RLock()
	defer p.mx.RUnlock()

	return slices.Clone(p.ps[channel])
}
