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

## Configuration

Create a file at `~/.config/tesla.yml` with the following contents:

```
retries: 5
redis:
  host: "127.0.0.1"
client_id: "e4a9949fcfa04068f59abb5a658f2bac0a3428e4652315490b659d5ab3f35a9e"
client_secret: "c75f14bbadc8bee3a7594412c31416f8300256d7668ea7e6e7f06727bfb9d220"
username: "your_tesla@login.address"
password: "your_tesla_password"
```
