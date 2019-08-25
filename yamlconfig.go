package main

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type Target struct {
	Zone  string   `yaml:"zone"`
	Names []string `yaml:"names"`
}

func (t Target) FQDNs() []string {
	var fqdns []string
	for _, name := range t.Names {
		if name == "" {
			fqdns = append(fqdns, t.Zone+".")
		} else {
			fqdns = append(fqdns, fmt.Sprintf("%s.%s.", name, t.Zone))
		}
	}
	return fqdns
}

type Config struct {
	Interface string        `yaml:"interface"`
	TTL       int64         `yaml:"ttl"`
	Type      string        `yaml:"type"`
	Targets   []Target      `yaml:"targets"`
	Retries   int           `yaml:"retries"`
	Patience  time.Duration `yaml:"patience"`
	Interval  time.Duration `yaml:"interval"`
	// don't make the user specify this twice (A==inet, AAAA=inet6)
	AF string
}

func LoadConfig(filename string) (*Config, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("unable to read config file: %v", err)
	}
	c := Config{
		Retries:  8,
		Patience: 15 * time.Second,
		Interval: 5 * time.Second,
		Type:     "A",
		TTL:      60,
		AF:       "inet",
	}
	if err := yaml.Unmarshal(content, &c); err != nil {
		return nil, fmt.Errorf("unable to unmarshal config file: %v", err)
	}
	if c.Type == "AAAA" {
		c.AF = "inet6"
	}
	return &c, nil
}
