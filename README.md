# ru1

TLDR just a dynamic DNS updater.

`ru1`, or _[AWS] Route53 update 1 [record]_, is a tool that keeps an AWS
Route53 `A` or `AAAA` record up to date with the Internet-routable IP addresses
configured on a network interface.

## configuration

Config is uncomplicated YAML and looks like:

```
interface: em0
type: A
targets:
- zone: myzone.com
  names:
  - moist
  - ''
- zone: myzone.net
  names:
  - ''
```

The above configuration would watch interface `em0` and generate/maintain the
below DNS A records:

* `moist.myzone.com`
* `myzone.com` (zone apex is represented by an empty string in the names list)
* `myzone.net` (as above)

## demo

Normal usage looks like this:

```
$ ./ru1 -config=ru1.yaml
2019/08/25 14:55:17 em0 added 1.2.3.4
2019/08/25 14:55:17 em0 current 1.2.3.4
2019/08/25 14:55:19 found zone 'myzone1' ID '/hostedzone/ABCDEFGHIJKL'
2019/08/25 14:55:19 g4.pe: upsert appeared to succeed
2019/08/25 14:55:35 g4.pe: upsert sync status: PENDING
2019/08/25 14:55:51 g4.pe: upsert sync status: INSYNC
2019/08/25 14:55:52 found zone 'myzone2' ID '/hostedzone/MNOPQRSTUVWX'
2019/08/25 14:55:52 l0.pe: upsert appeared to succeed
2019/08/25 14:56:08 l0.pe: upsert sync status: PENDING
2019/08/25 14:56:24 l0.pe: upsert sync status: INSYNC
```

## options

```
Usage of ./ru1:
  -config string
    	path to configuration file (default "ru1.yaml")
```

## notes

* updating with multiple IP addresses should work but is untested
* healthchecks would be nice
* tested on OpenBSD and macOS
* on OpenBSD, now uses [pledge](https://man.openbsd.org/pledge.2)
  and [unveil](https://man.openbsd.org/unveil.2)
* investigate making it daemonize (tricky to do properly with Go...)
