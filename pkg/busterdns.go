package pkg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	"github.com/OJ/gobuster/v3/libgobuster"
	"github.com/google/uuid"
)

// OptionsDNS holds all options for the dns plugin
type OptionsDNS struct {
	Domain         string
	ShowIPs        bool
	ShowCNAME      bool
	WildcardForced bool
	Resolver       string
	Timeout        time.Duration
}

// NewOptionsDNS returns a new initialized OptionsDNS
func NewOptionsDNS(domain string, isWildcard bool, resolver string) *OptionsDNS {
	return &OptionsDNS{Domain: domain, WildcardForced: isWildcard, Resolver: resolver}
}

// ErrWildcard is returned if a wildcard response is found
var ErrWildcard = errors.New("wildcard found")

// GobusterDNS is the main type to implement the interface
type GobusterDNS struct {
	resolver    *net.Resolver
	globalopts  *libgobuster.Options
	options     *OptionsDNS
	isWildcard  bool
	wildcardIps libgobuster.StringSet
}

func newCustomDialer(server string) func(ctx context.Context, network, address string) (net.Conn, error) {
	return func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		if !strings.Contains(server, ":") {
			server = fmt.Sprintf("%s:53", server)
		}
		return d.DialContext(ctx, "udp", server)
	}
}

// NewGobusterDNS creates a new initialized GobusterDNS
func NewGobusterDNS(opts *OptionsDNS) (*GobusterDNS, error) {
	if opts == nil {
		return nil, fmt.Errorf("please provide valid plugin options")
	}

	globalopts := &libgobuster.Options{}
	globalopts.Threads = 5
	globalopts.Quiet = true

	resolver := net.DefaultResolver
	if opts.Resolver != "" {
		resolver = &net.Resolver{
			PreferGo: true,
			Dial:     newCustomDialer(opts.Resolver),
		}
	}

	g := GobusterDNS{
		options:     opts,
		globalopts:  globalopts,
		wildcardIps: libgobuster.NewStringSet(),
		resolver:    resolver,
	}
	return &g, nil
}

// PreRun is the pre run implementation of gobusterdns
func (d *GobusterDNS) PreRun() error {
	// Resolve a subdomain sthat probably shouldn't exist
	guid := uuid.New()
	wildcardIps, err := d.dnsLookup(fmt.Sprintf("%s.%s", guid, d.options.Domain))
	if err == nil {
		d.isWildcard = true
		d.wildcardIps.AddRange(wildcardIps)
		if !d.options.WildcardForced {
			return ErrWildcard
		}
	}

	if !d.globalopts.Quiet {
		// Provide a warning if the base domain doesn't resolve (in case of typo)
		_, err = d.dnsLookup(d.options.Domain)
		if err != nil {
			// Not an error, just a warning. Eg. `yp.to` doesn't resolve, but `cr.yp.to` does!
			log.Printf("[-] Unable to validate base domain: %s (%v)", d.options.Domain, err)
		}
	}

	return nil
}

// RunWord is the process implementation of gobusterdns
func (d *GobusterDNS) RunWord(word string) error {
	subdomain := fmt.Sprintf("%s.%s", word, d.options.Domain)
	ips, err := d.dnsLookup(subdomain)
	if err == nil {
		if d.wildcardIps.ContainsAny(ips) {
			return fmt.Errorf("wildcard domain")
		}
		return nil
	}
	return err
}

// Run is the process implementation of wordlist gobusterdns
func (d *GobusterDNS) Run(wordlist []string) []string {
	var err error
	domains := make([]string, 0)
	for _, word := range wordlist {
		err = d.RunWord(word)
		if err == nil {
			domains = append(domains, word)
		}
	}
	return domains
}

func (d *GobusterDNS) dnsLookup(domain string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.options.Timeout)
	defer cancel()
	return d.resolver.LookupHost(ctx, domain)
}

func (d *GobusterDNS) dnsLookupCname(domain string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), d.options.Timeout)
	defer cancel()
	time.Sleep(time.Second)
	return d.resolver.LookupCNAME(ctx, domain)
}
