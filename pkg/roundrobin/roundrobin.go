package roundrobin

import (
	"errors"
	"net/url"
	"sync"
)

var ErrNoURLs = errors.New("no urls available")

type RoundRobin struct {
	urls []*url.URL
	idx  int
	mtx  sync.Mutex
}

func (r *RoundRobin) Pick() (*url.URL, error) {
	if len(r.urls) == 0 {
		return nil, ErrNoURLs
	}

	item := r.urls[r.idx]

	r.mtx.Lock()
	r.idx = (r.idx + 1) % len(r.urls)
	r.mtx.Unlock()

	return item, nil
}

func (r *RoundRobin) Length() int {
	return len(r.urls)
}

func New(urls []*url.URL) *RoundRobin {
	return &RoundRobin{
		urls: urls,
	}
}
