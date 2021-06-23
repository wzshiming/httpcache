package httpcache

import (
	"net/http"
	"sync"
)

type Option func(c *option)

type option struct {
	filterer Filterer
	keyer    Keyer
	storer   Storer

	muts sync.Map
}

func (o *option) init(options []Option) {
	for _, option := range options {
		option(o)
	}
	if o.keyer == nil {
		o.keyer = JointKeyer(HostKeyer(), PathKeyer())
	}
	if o.storer == nil {
		o.storer = MemoryStorer()
	}
	if o.filterer == nil {
		o.filterer = MethodFilterer(http.MethodGet)
	}
}

func WithStorer(storer Storer) func(c *option) {
	return func(c *option) {
		c.storer = storer
	}
}

func WithKeyer(keyer Keyer) func(c *option) {
	return func(c *option) {
		c.keyer = keyer
	}
}

func WithFilterer(filterer Filterer) func(c *option) {
	return func(c *option) {
		c.filterer = filterer
	}
}
