package all

import (
	_ "github.com/influxdata/telegraf/plugins/outputs/file"
	_ "github.com/influxdata/telegraf/plugins/outputs/http"
	_ "github.com/influxdata/telegraf/plugins/outputs/kafka"
	_ "github.com/influxdata/telegraf/plugins/outputs/nats"
	_ "github.com/influxdata/telegraf/plugins/outputs/socket_writer"
)
