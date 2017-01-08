# Tesla

## Intro

Simple command I use to collect and cache the state of my
Tesla in a local Redis instance to be used by various home
automation functions (e.g. automatically command Tesla
to charge when I have excess solar power and stop when
the sun starts to go down or getting lots of cloud cover).
This reduced the number of calls to the owners-api.

## Install

```
$ go get -d github.com/scottweston/tesla
$ go install github.com/scottweston/tesla
```
