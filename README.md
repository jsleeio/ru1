# ru1

TLDR just a dynamic DNS updater.

`ru1`, or _[AWS] Route53 update 1 [record]_, is a tool that keeps an AWS
Route53 `A` or `AAAA` record up to date with the Internet-routable IP addresses
configured on a network interface.

## demo

Normal usage looks like this:

```
$ ./ru1 -interface=em0 -name=myhost -zone=example.com -ttl=60
2019/08/22 20:55:02 em0 added 1.2.3.4
2019/08/22 20:55:02 em0 current 1.2.3.4
2019/08/22 20:55:04 found zone 'example.com' ID '/hostedzone/ABCDEFGHIJKL'
2019/08/22 20:55:06 upsert of 'myhost.example.com.' appeared to succeed
2019/08/22 20:55:22 upsert sync status: PENDING
2019/08/22 20:55:38 upsert sync status: INSYNC
```

## options

```
Usage of ./ru1:
  -give-up-after int
    	Give up after this many completion checks (default 20)
  -interface string
    	network interface to observe for changes (default "em0")
  -name string
    	A record to update in the zone (default "z")
  -patience duration
    	How often to check for change completion (default 15s)
  -ttl int
    	Time-to-live value to apply to this record (default 300)
  -type string
    	Record type to update (A or AAAA) (default "A")
  -zone string
    	DNS zone within which to manage entry
```

## notes

* updating apex records is completely untested
* updating with multiple IP addresses should work but is untested
* support for multiple A records sure would be nice
* healthchecks would be nice too
* tested on OpenBSD and macOS
* on OpenBSD, now uses [pledge](https://man.openbsd.org/pledge.2)
  and [unveil](https://man.openbsd.org/unveil.2)
