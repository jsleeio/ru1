package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func generateChangeBatch(target Target, config Config, addrs []string) *route53.ChangeBatch {
	var rrs []*route53.ResourceRecord
	for _, addr := range addrs {
		rrs = append(rrs, &route53.ResourceRecord{Value: aws.String(addr)})
	}
	var changes []*route53.Change
	for _, name := range target.FQDNs() {
		changes = append(changes, &route53.Change{
			Action: aws.String("UPSERT"),
			ResourceRecordSet: &route53.ResourceRecordSet{
				Name:            aws.String(name),
				TTL:             aws.Int64(config.TTL),
				Type:            aws.String(config.Type),
				ResourceRecords: rrs,
			},
		})
	}
	return &route53.ChangeBatch{Changes: changes}
}

func findZoneID(client *route53.Route53, zone string) (string, error) {
	lhzbni := &route53.ListHostedZonesByNameInput{DNSName: aws.String(zone)}
	lhzbno, err := client.ListHostedZonesByName(lhzbni)
	if err != nil {
		return "", fmt.Errorf("unable to find zone: %v", err)
	}
	if *lhzbno.IsTruncated {
		return "", fmt.Errorf("unsupported: zone search for '%s' returned more than one result", zone)
	}
	id := lhzbno.HostedZones[0].Id
	log.Printf("found zone '%s' ID '%s'", zone, *id)
	return *id, nil
}

func UpdateTarget(client *route53.Route53, target Target, config Config, addrs []string) error {
	id, err := findZoneID(client, target.Zone)
	if err != nil {
		return err
	}
	// I understand why it is like this, but: die in a fire anyway
	crrsi := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: aws.String(id),
		ChangeBatch:  generateChangeBatch(target, config, addrs),
	}
	crrso, err := client.ChangeResourceRecordSets(crrsi)
	if err != nil {
		return fmt.Errorf("%s: unable to upsert: %v", target.Zone, err)
	}
	log.Printf("%s: upsert appeared to succeed", target.Zone)
	if *crrso.ChangeInfo.Status != "INSYNC" {
		gci := &route53.GetChangeInput{Id: crrso.ChangeInfo.Id}
		for i := 0; i < config.Retries; i++ {
			time.Sleep(config.Patience) // sleep first, we already got one status
			gco, err := client.GetChange(gci)
			if err != nil {
				return fmt.Errorf("error waiting for change to be in sync: %v", err)
			}
			log.Printf("%s: upsert sync status: %s", target.Zone, *gco.ChangeInfo.Status)
			if *gco.ChangeInfo.Status == "INSYNC" {
				break
			}
		}
	}
	return nil
}

func WatchInterface(config Config, receiver func([]string)) error {
	var mutex sync.Mutex
	ticker := time.NewTicker(config.Interval)
	ipas := NewIPAddressSet(config.AF, config.Interface)
	for range ticker.C {
		mutex.Lock()
		ipas.Update()
		for removed, _ := range ipas.LastRemoved {
			log.Printf("%s removed %s", config.Interface, removed)
		}
		for added, _ := range ipas.LastAdded {
			log.Printf("%s added %s", config.Interface, added)
		}
		if len(ipas.LastRemoved) > 0 || len(ipas.LastAdded) > 0 {
			var clist []string
			for current, _ := range ipas.Current {
				log.Printf("%s current %s", config.Interface, current)
				clist = append(clist, current)
			}
			if receiver != nil && len(clist) > 0 {
				receiver(clist)
			}
		}
		mutex.Unlock()
	}
	return nil
}

func main() {
	configFile := flag.String("config", "ru1.yaml", "path to configuration file")
	flag.Parse()
	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("error loading configuration: %v", err)
	}
	Lockdown()
	client := route53.New(session.New(aws.NewConfig().WithMaxRetries(config.Retries)))
	receiver := func(addrs []string) {
		for _, target := range config.Targets {
			if err := UpdateTarget(client, target, *config, addrs); err != nil {
				log.Printf("error updating DNS: %v", err)
			}
		}
	}
	if err := WatchInterface(*config, receiver); err != nil {
		log.Fatalf("error watching interface: %v", err)
	}
}
