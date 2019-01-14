# socket_writer Plugin

The socket_writer plugin can write to a UDP, TCP, or unix socket.

It can output data in any of the [supported output formats](https://github.com/lavaorg/telex/blob/master/docs/DATA_FORMATS_OUTPUT.md).

```toml
# Generic socket writer capable of handling multiple socket types.
[[outputs.socket_writer]]
  ## URL to connect to
  # address = "tcp://127.0.0.1:8094"
  # address = "tcp://example.com:http"
  # address = "tcp4://127.0.0.1:8094"
  # address = "tcp6://127.0.0.1:8094"
  # address = "tcp6://[2001:db8::1]:8094"
  # address = "udp://127.0.0.1:8094"
  # address = "udp4://127.0.0.1:8094"
  # address = "udp6://127.0.0.1:8094"
  # address = "unix:///tmp/telex.sock"
  # address = "unixgram:///tmp/telex.sock"

  ## Optional TLS Config
  # tls_ca = "/etc/telex/ca.pem"
  # tls_cert = "/etc/telex/cert.pem"
  # tls_key = "/etc/telex/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Period between keep alive probes.
  ## Only applies to TCP sockets.
  ## 0 disables keep alive probes.
  ## Defaults to the OS configuration.
  # keep_alive_period = "5m"

  ## Data format to generate.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https://github.com/lavaorg/telex/blob/master/docs/DATA_FORMATS_INPUT.md
  # data_format = "influx"
```
