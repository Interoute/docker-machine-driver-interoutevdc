# Docker Machine Driver for [Interoute VDC](https://cloudstore.interoute.com)

This is an Interoute VDC driver for [Docker Machine](https://docs.docker.com/machine/).
It allows for the creation of Docker hosts on the [Interoute VDC](https://cloudstore.interoute.com) cloud platform.

## Requirements

* [Docker Machine](https://docs.docker.com/machine/) 0.5.1 or later

## Usage

```bash
docker-machine create -d interoutevdc \
  --interoutevdc-apikey INTEROUTEVDC_API_KEY \
  --interoutevdc-secretkey INTEROUTEVDC_SECRET_KEY \
  --interoutevdc-vdcregion INTEROUTEVDC_REGION \
  --interoutevdc-networkid "UUID" \
  --interoutevdc-serviceofferingid "UUID" \
  --interoutevdc-templatefilter "self" \
  --interoutevdc-templateid "UUID" \
  --interoutevdc-zoneid "UUID" \
  --interoutevdc-diskofferingid "UUID" \
  --interoutevdc-disksize 25 \
  docker-machine
```

## Acknowledgements

[docker-machine-driver-cloudstack](https://github.com/atsaki/docker-machine-driver-cloudstack) written by [@atsaki](https://github.com/atsaki).

[Go-Cloudstack](https://github.com/xanzy/go-cloudstack) written by [@xanzy](https://github.com/xanzy).

## Author
Radu Stefanache ([@radu-stefanache](https://github.com/radu-stefanache)).

## Note
This driver is supposed to be used in conjunction with [Rancher](http://rancher.com/) hence the UUIDs instead of names.
