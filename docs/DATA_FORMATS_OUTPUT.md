# Output Data Formats

In addition to output specific data formats, Telex supports a set of
standard data formats that may be selected from when configuring many output
plugins.

1. `influx` - [InfluxDB Line Protocol](/plugins/serializers/influx)
1. `json`   - [JSON](/plugins/serializers/json)

You will be able to identify the plugins with support by the presence of a
`data_format` config option, for example, in the `file` output plugin:

```toml
[[outputs.file]]
  ## Files to write to, "stdout" is a specially handled file.
  files = ["stdout"]

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/lavaorg/telex/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
```
