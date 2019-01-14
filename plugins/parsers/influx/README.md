# InfluxDB Line Protocol

There are no additional configuration options for InfluxDB [line protocol][]. The
metrics are parsed directly into Telex metrics.

[line protocol]: https://docs.influxdata.com/influxdb/latest/write_protocols/line/

### Configuration

```toml
[[inputs.file]]
  files = ["example"]

  ## Data format to consume.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ##   https://github.com/lavaorg/telex/blob/master/docs/DATA_FORMATS_INPUT.md
  data_format = "influx"
```

