// +build openbsd

package main

import (
	"golang.org/x/sys/unix"
	"log"
)

func MustUnveil(path, perms string) {
	if err := unix.Unveil(path, perms); err != nil {
		log.Fatalf("unable to unveil %s@%s (%v)", perms, path, err)
	}
}

func MustPledge(promises, execpromises string) {
	out := "promises='" + promises + "',execpromises='" + execpromises + "'"
	if err := unix.Pledge(promises, execpromises); err != nil {
		log.Fatalf("unable to pledge %s (%v)", out, err)
	}
}

func Lockdown() {
	// initially I thought I needed to unveil $HOME/.aws here too but of
	// course sensible unveil semantics cause EACCES to be returned, so
	// the AWS SDK doesn't bother with it. Works just fine, and so all
	// we need is a CA certificate bundle
	MustUnveil("/etc/ssl/cert.pem", "r")
	MustPledge("stdio rpath inet dns", "")
}
