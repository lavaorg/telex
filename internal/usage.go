// +build !windows

package internal

const Usage = `telex, The plugin-driven server agent for collecting and reporting metrics.

Usage:

  telex [commands|flags]

The commands & flags are:

  version             print the version to stdout

  --aggregator-filter <filter>   filter the aggregators to enable, separator is :
  --config <file>                configuration file to load
  --debug                        turn on debug logging
  --input-filter <filter>        filter the inputs to enable, separator is :
  --input-list                   print available input plugins.
  --output-filter <filter>       filter the outputs to enable, separator is :
  --output-list                  print available output plugins.
  --pidfile <file>               file to write our pid to
  --pprof-addr <address>         pprof address to listen on, don't activate pprof if empty
  --processor-filter <filter>    filter the processors to enable, separator is :
  --quiet                        run in quiet mode
  --test                         gather metrics, print them out, and exit;
                                 processors, aggregators, and outputs are not run
  --version                      display the version and exit

Examples:

  # generate a telex config file:
  telex config > telex.conf

  # generate config with only cpu input & influxdb output plugins defined
  telex --input-filter cpu --output-filter influxdb config

  # run a single telex collection, outputing metrics to stdout
  telex --config telex.conf --test

  # run telex with all plugins defined in config file
  telex --config telex.conf

  # run telex, enabling the cpu & memory input, and influxdb output plugins
  telex --config telex.conf --input-filter cpu:mem --output-filter influxdb

  # run telex with pprof
  telex --config telex.conf --pprof-addr localhost:6060
`
