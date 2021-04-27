package xclient

import (
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"
)

type SelectMode int

const (
	RandomSelect     SelectMode = iota // select randomly
	RoundRobinSelect                   // select using Robbin algorithm
)

type Discovery interface {
	Refresh() error // refresh from remote registry
	Update(servers []string) error
	Get(mode SelectMode) (string, error)
	GetAll() ([]string, error)
}

// MultiServerDiscovery is a discovery for multi servers without a registry center
// user provides the server addresses explicitly instead
type MultiServerDiscovery struct {
	r       *rand.Rand
	mu      sync.RWMutex
	servers []string
	index   int
}

// NewMultiServerDiscovery creates a MultiServerDiscovery instance
func NewMultiServerDiscovery(servers []string) *MultiServerDiscovery {
	d := &MultiServerDiscovery{
		servers: servers,
		r:       rand.New(rand.NewSource(time.Now().UnixNano())),
	}
	d.index = d.r.Intn(math.MaxInt32 - 1)
	return d
}

var _ Discovery = (*MultiServerDiscovery)(nil)

// Refresh doesn't make sense for MultiServerDiscovery, so ignore it
func (d *MultiServerDiscovery) Refresh() error {
	return nil
}

// Update the servers of discovery dynamically if needed
func (d *MultiServerDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.servers = servers
	return nil
}

// Get a server according to mode
func (d *MultiServerDiscovery) Get(mode SelectMode) (string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	n := len(d.servers)
	if n == 0 {
		return "", fmt.Errorf("rpc discovery: no available servers")
	}

	switch mode {
	case RandomSelect:
		return d.servers[d.r.Intn(n)], nil
	case RoundRobinSelect:
		s := d.servers[d.index%n]
		d.index = (d.index + 1) % n
		return s, nil
	default:
		return "", fmt.Errorf("rpc discovery: not supported select mode")
	}
}

// returns all servers in discovery
func (d *MultiServerDiscovery) GetAll() ([]string, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	servers := make([]string, len(d.servers))
	copy(servers, d.servers)
	return servers, nil
}
