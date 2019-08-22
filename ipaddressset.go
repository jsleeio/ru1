package main

import (
	"fmt"
	"net"
	"strings"
)

type IPAddressSet struct {
	name, family                    string
	LastAdded, LastRemoved, Current map[string]struct{}
}

func NewIPAddressSet(family, name string) *IPAddressSet {
	ipas := &IPAddressSet{
		name:        name,
		family:      family,
		LastAdded:   make(map[string]struct{}),
		LastRemoved: make(map[string]struct{}),
		Current:     make(map[string]struct{}),
	}
	return ipas
}

func (i *IPAddressSet) Update() error {
	newips, err := i.interfaceAddresses()
	if err != nil {
		return err
	}
	newadded := make(map[string]struct{})
	newremoved := make(map[string]struct{})
	for k, _ := range newips {
		if _, found := i.Current[k]; !found {
			// new IP added
			newadded[k] = struct{}{}
		} else {
			// IP present in old and new sets, no change
		}
	}
	for k, _ := range i.Current {
		if _, found := newips[k]; !found {
			// IP was removed
			newremoved[k] = struct{}{}
		}
	}
	i.LastAdded = newadded
	i.LastRemoved = newremoved
	i.Current = newips
	return nil
}

func (i *IPAddressSet) shouldIgnoreAddr(ip net.IP) bool {
	blacklist := []string{
		"10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16", // rfc1918
		"100.64.0.0/10",      // CGN
		"127.0.0.0/8",        // loopback
		"169.254.0.0/16",     // link-local
		"192.0.0.0/24",       // IETF private
		"192.88.99.0/24",     // ip6-to-ip4 relay
		"192.0.2.0/24",       // TEST-NET-1 / documentation
		"198.51.100.0/24",    // TEST-NET-2 / documentation
		"203.0.113.0/24",     // TEST-NET-3 / documentation
		"198.18.0.0/15",      // benchmarking
		"224.0.0.0/4",        // multicast
		"240.0.0.0/4",        // reserved
		"255.255.255.255/32", // broadcast
	}
	for _, cidr := range blacklist {
		_, block, _ := net.ParseCIDR(cidr)
		if block.Contains(ip) {
			return true
		}
	}
	return false
}

func (i *IPAddressSet) interfaceAddresses() (map[string]struct{}, error) {
	addrs := make(map[string]struct{})
	iface, err := net.InterfaceByName(i.name)
	if err != nil {
		return nil, fmt.Errorf("unable to observe interface '%s': %v", i.name, err)
	}
	ifaddrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("unable to observe interface addresses for '%s': %v", i.name, err)
	}
	for _, addr := range ifaddrs {
		ip, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			return nil, fmt.Errorf("unable to parse address '%s' on interface '%s: %v", addr.String(), i.name, err)
		}
		if ipAddressFamily(ip.String()) == i.family && !i.shouldIgnoreAddr(ip) {
			addrs[ip.String()] = struct{}{}
		}
	}
	return addrs, nil
}

func ipAddressFamily(address string) string {
	if strings.Contains(address, ":") {
		return "inet6"
	} else {
		return "inet"
	}
}
