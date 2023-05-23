package pkg

import (
	"net"
)

// Resolver : struct for resolving hostnames
type Resolver struct{}

// NewResolver : returns new Resolver
func NewResolver() *Resolver {
	return &Resolver{}
}

// Resolve : resolves a hostname and returns a slice of ips
func (r Resolver) Resolve(hostname string) []string {
	ips, err := net.LookupHost(hostname)
	if err != nil {
		return make([]string, 0)
	}
	return ips
}

// Run : resolves a list of hosts by calling Resolve repeatedly
func (r Resolver) Run(hosts []string) map[string][]string {
	ret := make(map[string][]string)
	for i := range hosts {
		result := r.Resolve(hosts[i])
		ret[hosts[i]] = result
	}
	return ret
}
