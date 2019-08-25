package main

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/route53"
)

func addrsToResourceRecords(addrs []string) []*route53.ResourceRecord {
	var rrs []*route53.ResourceRecord
	for _, addr := range addrs {
		rrs = append(rrs, &route53.ResourceRecord{Value: aws.String(addr)})
	}
	return rrs
}

func UpdateRecord(config Config, addrs []string) error {
	awsconfig := aws.NewConfig().WithMaxRetries(15)
	client := route53.New(session.New(awsconfig))
	lhzbni := &route53.ListHostedZonesByNameInput{DNSName: config.Zone}
	lhzbno, err := client.ListHostedZonesByName(lhzbni)
	if err != nil {
		return fmt.Errorf("unable to find zone: %v", err)
	}
	if *lhzbno.IsTruncated {
		return fmt.Errorf("unsupported: zone search for '%s' returned more than one result", *config.Zone)
	}
	id := lhzbno.HostedZones[0].Id
	log.Printf("found zone '%s' ID '%s'", *config.Zone, *id)
	// I understand why it is like this, but: die in a fire anyway
	crrsi := &route53.ChangeResourceRecordSetsInput{
		HostedZoneId: id,
		ChangeBatch: &route53.ChangeBatch{
			Changes: []*route53.Change{
				&route53.Change{
					Action: aws.String("UPSERT"),
					ResourceRecordSet: &route53.ResourceRecordSet{
						Name:            aws.String(config.FQDN()),
						TTL:             config.TTL,
						Type:            config.Type,
						ResourceRecords: addrsToResourceRecords(addrs),
					},
				},
			},
		},
	}
	crrso, err := client.ChangeResourceRecordSets(crrsi)
	if err != nil {
		return fmt.Errorf("unable to upsert %s: %v", config.FQDN(), err)
	}
	log.Printf("upsert of '%s' appeared to succeed", config.FQDN())
	if *crrso.ChangeInfo.Status != "INSYNC" {
		gci := &route53.GetChangeInput{Id: crrso.ChangeInfo.Id}
		for i := 0; i < *config.GiveUpAfter; i++ {
			time.Sleep(*config.Patience) // sleep first, we already got one status
			gco, err := client.GetChange(gci)
			if err != nil {
				return fmt.Errorf("error waiting for change to be in sync: %v", err)
			}
			log.Printf("upsert sync status: %s", *gco.ChangeInfo.Status)
			if *gco.ChangeInfo.Status == "INSYNC" {
				break
			}
		}
	}
	return nil
}

func WatchInterface(name string, interval time.Duration, receiver func([]string)) error {
	var mutex sync.Mutex
	ticker := time.NewTicker(interval)
	ipas := NewIPAddressSet("inet", name)
	for range ticker.C {
		mutex.Lock()
		ipas.Update()
		for removed, _ := range ipas.LastRemoved {
			log.Printf("%s removed %s", name, removed)
		}
		for added, _ := range ipas.LastAdded {
			log.Printf("%s added %s", name, added)
		}
		if len(ipas.LastRemoved) > 0 || len(ipas.LastAdded) > 0 {
			var clist []string
			for current, _ := range ipas.Current {
				log.Printf("%s current %s", name, current)
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
	Lockdown()
	config, err := Configure()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	receiver := func(addrs []string) {
		if err := UpdateRecord(*config, addrs); err != nil {
			log.Printf("error updating DNS: %v", err)
		}
	}
	if err := WatchInterface(*config.Interface, 2*time.Second, receiver); err != nil {
		log.Fatalf("error watching interface: %v", err)
	}
}
