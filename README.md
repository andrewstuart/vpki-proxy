vpki-proxy
==========

vpki-proxy is a quick & dirty configurable TLS proxy which uses Let's Encrypt
(via [rsc.io/letsencrypt](//rsc.io/letsencrypt)) to retrieve certificates for
the services. The usage of the [vpki](/andrewstuart/vpki) library means that
it should be trivial to set up vault as a cert provider as well. I intend to do
this in the near term.

This proxy will also export [Prometheus](/prometheus/prometheus) metrics if an
IP prefix is given. This should be extended in the future to allow more
flexibility.

## Setup

```bash
go get -u astuart.co/vpki-proxy
```

See [example config](blob/master/config-example.yml) for an example
configuration.

## Usage

```bash
vpki-proxy -metric-ip="192.168.1.2" config.yml
```
