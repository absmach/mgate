# Mainflux

[![build][ci-badge]][ci-url]
[![go report card][grc-badge]][grc-url]
[![coverage][cov-badge]][cov-url]
[![license][license]](LICENSE)
[![chat][gitter-badge]][gitter]

![banner][banner]

Mainflux is modern, scalable, secure, open-source, and patent-free IoT cloud platform written in Go.

It accepts user and thing connections over various network protocols (i.e. HTTP,
MQTT, WebSocket, CoAP), thus making a seamless bridge between them. It is used as the IoT middleware
for building complex IoT solutions.

For more details, check out the [official documentation][docs].

Mainflux is member of the [Linux Foundation][lf] and an active contributor
to the [EdgeX Foundry][edgex] project. It has been made with :heart: by [Mainflux Labs][company],
which maintains the project and offers professional services around it.

## Features

- Multi-protocol connectivity and bridging (HTTP, MQTT, WebSocket and CoAP)
- Device management and provisioning (Zero Touch provisioning)
- Mutual TLS Authentication (mTLS) using X.509 Certificates
- Fine-grained access control
- Message persistence (Cassandra, InfluxDB, MongoDB and PostgresSQL)
- Platform logging and instrumentation support (Grafana, Prometheus and OpenTracing)
- Event sourcing
- Container-based deployment using [Docker][docker] and [Kubernetes][kubernetes]
- [LoRaWAN][lora] network integration
- SDK
- CLI
- Small memory footprint and fast execution
- Domain-driven design architecture, high-quality code and test coverage

## Prerequisites

The following are needed to run Mainflux:

- [Docker](https://docs.docker.com/install/) (version 18.09)
- [Docker compose](https://docs.docker.com/compose/install/) (version 1.24.1)

Developing Mainflux will also require:

- [Go](https://golang.org/doc/install) (version 1.13.3)
- [Protobuf](https://github.com/protocolbuffers/protobuf#protocol-compiler-installation) (version 3.6.1)

## Install

Once the prerequisites are installed, execute the following commands from the project's root:

```bash
docker-compose -f docker/docker-compose.yml up
```

This will bring up the Mainflux docker services and interconnect them. This command can also be executed using the project's included Makefile:

```bash
make run
```

## Usage

The quickest way to start using Mainflux is via the CLI. The latest version can be downloaded from the [official releases page][rel]. 

It can also be built and used from the project's root directory:

```bash
make cli
./build/mainflux-cli version
```

Additional details on using the CLI can be found in the [CLI documentation](https://mainflux.readthedocs.io/en/latest/cli/).

## Documentation

Official documentation is hosted at [Mainflux Read The Docs page][docs]. Documentation is auto-generated, checkout the instructions on [official docs repository](https://github.com/mainflux/docs):

If you spot an error or a need for corrections, please let us know - or even better: send us a PR.

Additional practical information, news and tutorials can be found on the [Mainflux blog][blog].

## Authors

Main architect and BDFL of Mainflux project is [@drasko][drasko].

Additionally, [@nmarcetic][nikola] and [@janko-isidorovic][janko] assured
overall architecture and design, while [@manuio][manu] and [@darkodraskovic][darko]
helped with crafting initial implementation and continuously worked on the project evolutions.

Besides them, Mainflux is constantly improved and actively
developed by [@anovakovic01][alex], [@dusanb94][dusan], [@srados][sava],
[@gsaleh][george], [@blokovi][iva], [@chombium][kole], [@mteodor][mirko] and a large set of contributors.

Maintainers are listed in [MAINTAINERS](MAINTAINERS) file.

The Mainflux team would like to give special thanks to [@mijicd][dejan] for his monumental work
on designing and implementing a highly improved and optimized version of the platform,
and [@malidukica][dusanm] for his effort on implementing the initial user interface.

## Contributing

Thank you for your interest in Mainflux and the desire to contribute!

1. Take a look at our [open issues](https://github.com/mainflux/mainflux/issues). The [good-first-issue](https://github.com/mainflux/mainflux/labels/good-first-issue) label is specifically for issues that are great for getting started.
2. Checkout the [contribution guide](CONTRIBUTING.md) to learn more about our style and conventions.
3. Make your changes compatible to our workflow.

### We're Hiring

If you are interested in working professionally on Mainflux,
please head to company's [careers page][careers] or shoot us an e-mail at <careers@mainflux.com>.

>The best way to grab our attention is by sending PRs :sunglasses:.

## Community

- [Google group][forum]
- [Gitter][gitter]
- [Twitter][twitter]

## License

[Apache-2.0](LICENSE)

[banner]: https://github.com/mainflux/docs/blob/master/docs/img/gopherBanner.jpg
[ci-badge]: https://semaphoreci.com/api/v1/mainflux/mainflux/branches/master/badge.svg
[ci-url]: https://semaphoreci.com/mainflux/mainflux
[docs]: http://mainflux.readthedocs.io
[docker]: https://www.docker.com
[forum]: https://groups.google.com/forum/#!forum/mainflux
[gitter]: https://gitter.im/mainflux/mainflux?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge
[gitter-badge]: https://badges.gitter.im/Join%20Chat.svg
[grc-badge]: https://goreportcard.com/badge/github.com/mainflux/mainflux
[grc-url]: https://goreportcard.com/report/github.com/mainflux/mainflux
[cov-badge]: https://codecov.io/gh/mainflux/mainflux/branch/master/graph/badge.svg
[cov-url]: https://codecov.io/gh/mainflux/mainflux
[license]: https://img.shields.io/badge/license-Apache%20v2.0-blue.svg
[twitter]: https://twitter.com/mainflux
[lora]: https://lora-alliance.org/
[kubernetes]: https://kubernetes.io/
[rel]: https://github.com/mainflux/mainflux/releases
[careers]: https://www.mainflux.com/careers.html
[lf]: https://www.linuxfoundation.org/
[edgex]: https://www.edgexfoundry.org/
[company]: https://www.mainflux.com/
[blog]: https://medium.com/mainflux-iot-platform
[drasko]: https://github.com/drasko
[nikola]: https://github.com/nmarcetic
[dejan]: https://github.com/mijicd
[manu]: https://github.com/manuIO
[darko]: https://github.com/darkodraskovic
[janko]: https://github.com/janko-isidorovic
[alex]: https://github.com/anovakovic01
[dusan]: https://github.com/dusanb94
[sava]: https://github.com/srados
[george]: https://github.com/gesaleh
[iva]: https://github.com/blokovi
[kole]: https://github.com/chombium
[dusanm]: https://github.com/malidukica
[mirko]: https://github.com/mteodor
