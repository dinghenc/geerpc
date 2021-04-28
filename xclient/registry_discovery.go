package xclient

import (
	"log"
	"net/http"
	"strings"
	"time"
)

type RegistryDiscovery struct {
	*MultiServerDiscovery
	registry   string
	timeout    time.Duration
	lastUpdate time.Time
}

const defaultUpdateTimeout = time.Second * 10

func NewRegistryDiscovery(registryAddr string, timeout time.Duration) *RegistryDiscovery {
	if timeout == 0 {
		timeout = defaultUpdateTimeout
	}

	return &RegistryDiscovery{
		MultiServerDiscovery: NewMultiServerDiscovery(make([]string, 0)),
		registry:             registryAddr,
		timeout:              timeout,
	}
}

func (d *RegistryDiscovery) Update(servers []string) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.servers = servers
	d.lastUpdate = time.Now()
	return nil
}

func (d *RegistryDiscovery) Refresh() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.lastUpdate.Add(d.timeout).After(time.Now()) {
		return nil
	}
	log.Printf("rpc registry: refresh servers from registry %s", d.registry)

	resp, err := http.Get(d.registry)
	if err != nil {
		log.Printf("rpc registry refresh err: %v", err)
		return err
	}

	servers := strings.Split(resp.Header.Get("X-Geerpc-Servers"), ",")
	d.servers = make([]string, 0, len(servers))
	for _, server := range servers {
		if strings.TrimSpace(server) != "" {
			d.servers = append(d.servers, strings.TrimSpace(server))
		}
	}

	d.lastUpdate = time.Now()
	return nil
}

func (d *RegistryDiscovery) Get(mode SelectMode) (string, error) {
	if err := d.Refresh(); err != nil {
		return "", err
	}
	return d.MultiServerDiscovery.Get(mode)
}

func (d *RegistryDiscovery) GetAll() ([]string, error) {
	if err := d.Refresh(); err != nil {
		return nil, err
	}
	return d.MultiServerDiscovery.GetAll()
}
