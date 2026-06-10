package networks

import (
	"sync"
	"time"
)

type knownHosts struct {
	mu    sync.Mutex
	hosts map[string]time.Time
}

func newKnownHosts() knownHosts {
	return knownHosts{
		hosts: map[string]time.Time{},
	}
}

func (kh *knownHosts) seen(host string, timestamp time.Time) {
	kh.mu.Lock()
	defer kh.mu.Unlock()

	if lastSeen, ok := kh.hosts[host]; ok {
		if lastSeen.After(timestamp) {
			return
		}
	}

	kh.hosts[host] = timestamp
}

func (kh *knownHosts) getKnownHostAddresses() []string {
	kh.mu.Lock()
	defer kh.mu.Unlock()
	hosts := make([]string, 0, len(kh.hosts))
	for host := range kh.hosts {
		// skip old hosts
		if kh.hosts[host].Before(time.Now().Add(-5 * time.Minute)) {
			continue
		}

		hosts = append(hosts, host)
	}
	return hosts
}
