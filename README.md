# Telex 

Telex is an agent for collecting, processing, aggregating, and writing metrics.

Design goals are to have a minimal memory footprint with a plugin system so
that developers in the community can easily add support for collecting
metrics.

Telex is plugin-driven and has the concept of 4 distinct plugin types:

1. [Input Plugins](#input-plugins) collect metrics from the system, services, or 3rd party APIs
2. [Processor Plugins](#processor-plugins) transform, decorate, and/or filter metrics
3. [Aggregator Plugins](#aggregator-plugins) create aggregate metrics (e.g. mean, min, max, quantiles, etc.)
4. [Output Plugins](#output-plugins) write metrics to various destinations

New plugins are designed to be easy to contribute, we'll eagerly accept pull
requests and will manage the set of plugins that Telex supports.

## Contributing

There are many ways to contribute:
- Fix and [report bugs](https://github.com/lavaorg/telex/issues/new)
- [Improve documentation](https://github.com/lavaorg/telex/issues?q=is%3Aopen+label%3Adocumentation)
- [Review code and feature proposals](https://github.com/lavaorg/telex/pulls)
- Answer questions and discuss here on github and on the [Community Site](https://community.influxdata.com/)
- [Contribute plugins](CONTRIBUTING.md)

## Installation:

You can download the binaries directly from the [downloads](https://www.influxdata.com/downloads) page
or from the [releases](https://github.com/lavaorg/telex/releases) section.

### Ansible Role:

Ansible role: https://github.com/rossmcdonald/telex

### From Source:

Telex requires golang version 1.9 or newer, the Makefile requires GNU make.

1. [Install Go](https://golang.org/doc/install) >=1.9 (1.10 recommended)
2. [Install dep](https://golang.github.io/dep/docs/installation.html) ==v0.5.0
3. Download Telex source:
   ```
   go get -d github.com/lavaorg/telex
   ```
4. Run make from the source directory
   ```
   cd "$HOME/go/src/github.com/lavaorg/telex"
   make
   ```

## How to use it:

See usage with:

```
telex --help
```

#### Generate a telex config file:

```
telex config > telex.conf
```

#### Generate config with only cpu input & influxdb output plugins defined:

```
telex --input-filter cpu --output-filter influxdb config
```

#### Run a single telex collection, outputing metrics to stdout:

```
telex --config telex.conf --test
```

#### Run telex with all plugins defined in config file:

```
telex --config telex.conf
```

#### Run telex, enabling the cpu & memory input, and influxdb output plugins:

```
telex --config telex.conf --input-filter cpu:mem --output-filter influxdb
```

## Documentation

[Latest Release Documentation][release docs].

For documentation on the latest development code see the [documentation index][devel docs].

[devel docs]: docs

## Input Plugins

For inspiration for more plugins please see the orignal https://github.com/influxdata/telegraf project that is the parent of Telex.

* [bcache](./plugins/inputs/bcache)
* [bond](./plugins/inputs/bond)
* [cgroup](./plugins/inputs/cgroup)
* [conntrack](./plugins/inputs/conntrack)
* [cpu](./plugins/inputs/cpu)
* [diskio](./plugins/inputs/diskio)
* [disk](./plugins/inputs/disk)
* [dns query time](./plugins/inputs/dns_query)
* [exec](./plugins/inputs/exec) (generic executable plugin, support JSON, influx, graphite and nagios)
* [file](./plugins/inputs/file)
* [filestat](./plugins/inputs/filestat)
* [filecount](./plugins/inputs/filecount)
* [hddtemp](./plugins/inputs/hddtemp)
* [http_listener_v2](./plugins/inputs/http_listener_v2)
* [http](./plugins/inputs/http) (generic HTTP plugin, supports using input data formats)
* [http_response](./plugins/inputs/http_response)
* [internal](./plugins/inputs/internal)
* [interrupts](./plugins/inputs/interrupts)
* [iptables](./plugins/inputs/iptables)
* [ipvs](./plugins/inputs/ipvs)
* [kernel_vmstat](./plugins/inputs/kernel_vmstat)
* [linux_sysctl_fs](./plugins/inputs/linux_sysctl_fs)
* [logparser](./plugins/inputs/logparser)
* [mem](./plugins/inputs/mem)
* [net](./plugins/inputs/net)
* [net_response](./plugins/inputs/net_response)
* [netstat](./plugins/inputs/net)
* [nstat](./plugins/inputs/nstat)
* [ping](./plugins/inputs/ping)
* [processes](./plugins/inputs/processes)
* [procstat](./plugins/inputs/procstat)
* [sensors](./plugins/inputs/sensors)
* [smart](./plugins/inputs/smart)
* [socket_listener](./plugins/inputs/socket_listener)
* [swap](./plugins/inputs/swap)
* [syslog](./plugins/inputs/syslog)
* [sysstat](./plugins/inputs/sysstat)
* [system](./plugins/inputs/system)
* [tail](./plugins/inputs/tail)
* [temp](./plugins/inputs/temp)
* [zfs](./plugins/inputs/zfs)

## Parsers

- [InfluxDB Line Protocol](/plugins/parsers/influx)
- [CSV](/plugins/parsers/csv)
- [Grok](/plugins/parsers/grok)
- [JSON](/plugins/parsers/json)
- [Value](/plugins/parsers/value), ie: 45 or "booyah"

## Serializers

- [InfluxDB Line Protocol](/plugins/serializers/influx)
- [JSON](/plugins/serializers/json)

## Processor Plugins

* [converter](./plugins/processors/converter)
* [enum](./plugins/processors/enum)
* [override](./plugins/processors/override)
* [parser](./plugins/processors/parser)
* [printer](./plugins/processors/printer)
* [regex](./plugins/processors/regex)
* [rename](./plugins/processors/rename)
* [strings](./plugins/processors/strings)
* [topk](./plugins/processors/topk)

## Aggregator Plugins

* [basicstats](./plugins/aggregators/basicstats)
* [minmax](./plugins/aggregators/minmax)
* [histogram](./plugins/aggregators/histogram)
* [valuecounter](./plugins/aggregators/valuecounter)

## Output Plugins

* [file](./plugins/outputs/file)
* [http](./plugins/outputs/http)
* [socket_writer](./plugins/outputs/socket_writer)
