// +build windows

package internal

const Usage = `Telegraf, The plugin-driven server agent for collecting and reporting metrics.

Usage:

  telegraf [commands|flags]

The commands & flags are:

  config              print out full sample configuration to stdout
  version             print the version to stdout

  --aggregator-filter <filter>   filter the aggregators to enable, separator is :
  --config <file>                configuration file to load
  --config-directory <directory> directory containing additional *.conf files
  --debug                        turn on debug logging
  --input-filter <filter>        filter the inputs to enable, separator is :
  --input-list                   print available input plugins.
  --output-filter <filter>       filter the outputs to enable, separator is :
  --output-list                  print available output plugins.
  --pidfile <file>               file to write our pid to
  --pprof-addr <address>         pprof address to listen on, don't activate pprof if empty
  --processor-filter <filter>    filter the processors to enable, separator is :
  --quiet                        run in quiet mode
  --sample-config                print out full sample configuration
  --test                         gather metrics, print them out, and exit;
                                 processors, aggregators, and outputs are not run
  --usage <plugin>               print usage for a plugin, ie, 'telegraf --usage mysql'
  --version                      display the version and exit

  --console                      run as console application (windows only)
  --service <service>            operate on the service (windows only)
  --service-name                 service name (windows only)

Examples:

  # generate a telegraf config file:
  telegraf config > telex.conf

  # generate config with only cpu input & influxdb output plugins defined
  telegraf --input-filter cpu --output-filter influxdb config

  # run a single telegraf collection, outputing metrics to stdout
  telegraf --config telex.conf --test

  # run telegraf with all plugins defined in config file
  telegraf --config telex.conf

  # run telegraf, enabling the cpu & memory input, and influxdb output plugins
  telegraf --config telex.conf --input-filter cpu:mem --output-filter influxdb

  # run telegraf with pprof
  telegraf --config telex.conf --pprof-addr localhost:6060

  # run telegraf without service controller
  telegraf --console install --config "C:\Program Files\Telegraf\telex.conf"

  # install telegraf service
  telegraf --service install --config "C:\Program Files\Telegraf\telex.conf"

  # install telegraf service with custom name
  telegraf --service install --service-name=my-telegraf
`
