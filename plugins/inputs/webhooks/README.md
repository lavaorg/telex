# Webhooks

This is a Telex service plugin that start an http server and register multiple webhook listeners.

```sh
$ telex config -input-filter webhooks -output-filter influxdb > config.conf.new
```

Change the config file to point to the InfluxDB server you are using and adjust the settings to match your environment. Once that is complete:

```sh
$ cp config.conf.new /etc/telex/telex.conf
$ sudo service telex start
```

## Available webhooks

- [Filestack](filestack/)
- [Github](github/)
- [Mandrill](mandrill/)
- [Rollbar](rollbar/)
- [Papertrail](papertrail/)
- [Particle](particle/)


## Adding new webhooks plugin

1. Add your webhook plugin inside the `webhooks` folder
1. Your plugin must implement the `Webhook` interface
1. Import your plugin in the `webhooks.go` file and add it to the `Webhooks` struct

Both [Github](github/) and [Rollbar](rollbar/) are good example to follow.
