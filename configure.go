package main

import (
	"flag"
	"fmt"
	"net"
	"time"
)

type Config struct {
	Zone        *string // DNS zone (actual zone, not ID)
	Name        *string // DNS record to update, not including zone
	Address     *string // flag hath no IP address type
	Type        *string // A or AAAA
	IP          net.IP
	TTL         *int64
	Patience    *time.Duration
	GiveUpAfter *int
	Interface   *string
}

func (c Config) FQDN() string {
	return fmt.Sprintf("%s.%s.", *c.Name, *c.Zone)
}

func Configure() (*Config, error) {
	c := &Config{
		Zone:        flag.String("zone", "", "DNS zone within which to manage entry"),
		Name:        flag.String("name", "z", "A record to update in the zone"),
		Type:        flag.String("type", "A", "Record type to update (A or AAAA)"),
		TTL:         flag.Int64("ttl", 300, "Time-to-live value to apply to this record"),
		Patience:    flag.Duration("patience", 15*time.Second, "How often to check for change completion"),
		GiveUpAfter: flag.Int("give-up-after", 20, "Give up after this many completion checks"),
		Interface:   flag.String("interface", "em0", "network interface to observe for changes"),
	}
	flag.Parse()
	if *c.Type != "A" && *c.Type != "AAAA" {
		return nil, fmt.Errorf("invalid record type specified: %s", *c.Type)
	}
	return c, nil
}
