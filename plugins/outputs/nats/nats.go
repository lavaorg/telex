package nats

import (
	"fmt"

	nats_client "github.com/nats-io/go-nats"

	"github.com/lavaorg/telex"
	"github.com/lavaorg/telex/internal/tls"
	"github.com/lavaorg/telex/plugins/outputs"
	"github.com/lavaorg/telex/plugins/serializers"
)

type NATS struct {
	// Servers is the NATS server pool to connect to
	Servers []string
	// Credentials
	Username string
	Password string
	// NATS subject to publish metrics to
	Subject string
	tls.ClientConfig

	conn       *nats_client.Conn
	serializer serializers.Serializer
}

var sampleConfig = `
  ## URLs of NATS servers
  servers = ["nats://localhost:4222"]
  ## Optional credentials
  # username = ""
  # password = ""
  ## NATS subject for producer messages
  subject = "telex"

  ## Optional TLS Config
  # tls_ca = "/etc/telex/ca.pem"
  # tls_cert = "/etc/telex/cert.pem"
  # tls_key = "/etc/telex/key.pem"
  ## Use TLS but skip chain & host verification
  # insecure_skip_verify = false

  ## Data format to output.
  ## Each data format has its own unique set of configuration options, read
  ## more about them here:
  ## https:/github.com/lavaorg/telex/blob/master/docs/DATA_FORMATS_OUTPUT.md
  data_format = "influx"
`

func (n *NATS) SetSerializer(serializer serializers.Serializer) {
	n.serializer = serializer
}

func (n *NATS) Connect() error {
	var err error

	// set default NATS connection options
	opts := nats_client.DefaultOptions

	// override max reconnection tries
	opts.MaxReconnect = -1

	// override servers, if any were specified
	opts.Servers = n.Servers

	// override authentication, if any was specified
	if n.Username != "" {
		opts.User = n.Username
		opts.Password = n.Password
	}

	// override TLS, if it was specified
	tlsConfig, err := n.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}
	if tlsConfig != nil {
		// set NATS connection TLS options
		opts.Secure = true
		opts.TLSConfig = tlsConfig
	}

	// try and connect
	n.conn, err = opts.Connect()

	return err
}

func (n *NATS) Close() error {
	n.conn.Close()
	return nil
}

func (n *NATS) SampleConfig() string {
	return sampleConfig
}

func (n *NATS) Description() string {
	return "Send telex measurements to NATS"
}

func (n *NATS) Write(metrics []telex.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	for _, metric := range metrics {
		buf, err := n.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		err = n.conn.Publish(n.Subject, buf)
		if err != nil {
			return fmt.Errorf("FAILED to send NATS message: %s", err)
		}
	}
	return nil
}

func init() {
	outputs.Add("nats", func() telex.Output {
		return &NATS{}
	})
}
