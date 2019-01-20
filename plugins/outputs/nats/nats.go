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
