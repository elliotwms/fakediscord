package storage

import (
	"sync"
)

var Pins = &pins{
	ps: make(map[string][]string),
}

type pins struct {
	mx sync.RWMutex
	ps map[string][]string
}

func (p *pins) Store(channel, message string) {
	p.mx.Lock()
	defer p.mx.Unlock()

	if _, ok := p.ps[channel]; !ok {
		p.ps[channel] = []string{}
	}

	p.ps[channel] = append(p.ps[channel], message)
}

func (p *pins) Load(channel string) []string {
	p.mx.RLock()
	defer p.mx.RUnlock()

	if v, ok := p.ps[channel]; ok {
		return v
	}

	return []string{}
}
