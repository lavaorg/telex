package kafka

import (
	"crypto/tls"
	"fmt"
	"log"
	"strings"

	"github.com/lavaorg/telex"
	tlsint "github.com/lavaorg/telex/internal/tls"
	"github.com/lavaorg/telex/plugins/outputs"
	"github.com/lavaorg/telex/plugins/serializers"
	uuid "github.com/satori/go.uuid"

	"github.com/Shopify/sarama"
)

var ValidTopicSuffixMethods = []string{
	"",
	"measurement",
	"tags",
}

type (
	Kafka struct {
		Brokers          []string
		Topic            string
		ClientID         string      `toml:"client_id"`
		TopicSuffix      TopicSuffix `toml:"topic_suffix"`
		RoutingTag       string      `toml:"routing_tag"`
		RoutingKey       string      `toml:"routing_key"`
		CompressionCodec int
		RequiredAcks     int
		MaxRetry         int
		MaxMessageBytes  int `toml:"max_message_bytes"`

		Version string `toml:"version"`

		// Legacy TLS config options
		// TLS client certificate
		Certificate string
		// TLS client key
		Key string
		// TLS certificate authority
		CA string

		tlsint.ClientConfig

		// SASL Username
		SASLUsername string `toml:"sasl_username"`
		// SASL Password
		SASLPassword string `toml:"sasl_password"`

		tlsConfig tls.Config
		producer  sarama.SyncProducer

		serializer serializers.Serializer
	}
	TopicSuffix struct {
		Method    string   `toml:"method"`
		Keys      []string `toml:"keys"`
		Separator string   `toml:"separator"`
	}
)

func ValidateTopicSuffixMethod(method string) error {
	for _, validMethod := range ValidTopicSuffixMethods {
		if method == validMethod {
			return nil
		}
	}
	return fmt.Errorf("Unknown topic suffix method provided: %s", method)
}

func (k *Kafka) GetTopicName(metric telex.Metric) string {
	var topicName string
	switch k.TopicSuffix.Method {
	case "measurement":
		topicName = k.Topic + k.TopicSuffix.Separator + metric.Name()
	case "tags":
		var topicNameComponents []string
		topicNameComponents = append(topicNameComponents, k.Topic)
		for _, tag := range k.TopicSuffix.Keys {
			tagValue := metric.Tags()[tag]
			if tagValue != "" {
				topicNameComponents = append(topicNameComponents, tagValue)
			}
		}
		topicName = strings.Join(topicNameComponents, k.TopicSuffix.Separator)
	default:
		topicName = k.Topic
	}
	return topicName
}

func (k *Kafka) SetSerializer(serializer serializers.Serializer) {
	k.serializer = serializer
}

func (k *Kafka) Connect() error {
	err := ValidateTopicSuffixMethod(k.TopicSuffix.Method)
	if err != nil {
		return err
	}
	config := sarama.NewConfig()

	if k.Version != "" {
		version, err := sarama.ParseKafkaVersion(k.Version)
		if err != nil {
			return err
		}
		config.Version = version
	}

	if k.ClientID != "" {
		config.ClientID = k.ClientID
	} else {
		config.ClientID = "telex"
	}

	config.Producer.RequiredAcks = sarama.RequiredAcks(k.RequiredAcks)
	config.Producer.Compression = sarama.CompressionCodec(k.CompressionCodec)
	config.Producer.Retry.Max = k.MaxRetry
	config.Producer.Return.Successes = true

	if k.MaxMessageBytes > 0 {
		config.Producer.MaxMessageBytes = k.MaxMessageBytes
	}

	// Legacy support ssl config
	if k.Certificate != "" {
		k.TLSCert = k.Certificate
		k.TLSCA = k.CA
		k.TLSKey = k.Key
	}

	tlsConfig, err := k.ClientConfig.TLSConfig()
	if err != nil {
		return err
	}

	if tlsConfig != nil {
		config.Net.TLS.Config = tlsConfig
		config.Net.TLS.Enable = true
	}

	if k.SASLUsername != "" && k.SASLPassword != "" {
		config.Net.SASL.User = k.SASLUsername
		config.Net.SASL.Password = k.SASLPassword
		config.Net.SASL.Enable = true
	}

	producer, err := sarama.NewSyncProducer(k.Brokers, config)
	if err != nil {
		return err
	}
	k.producer = producer
	return nil
}

func (k *Kafka) Close() error {
	return k.producer.Close()
}

func (k *Kafka) routingKey(metric telex.Metric) string {
	if k.RoutingTag != "" {
		key, ok := metric.GetTag(k.RoutingTag)
		if ok {
			return key
		}
	}

	if k.RoutingKey == "random" {
		u := uuid.NewV4()
		return u.String()
	}

	return k.RoutingKey
}

func (k *Kafka) Write(metrics []telex.Metric) error {
	msgs := make([]*sarama.ProducerMessage, 0, len(metrics))
	for _, metric := range metrics {
		buf, err := k.serializer.Serialize(metric)
		if err != nil {
			return err
		}

		m := &sarama.ProducerMessage{
			Topic: k.GetTopicName(metric),
			Value: sarama.ByteEncoder(buf),
		}
		key := k.routingKey(metric)
		if key != "" {
			m.Key = sarama.StringEncoder(key)
		}
		msgs = append(msgs, m)
	}

	err := k.producer.SendMessages(msgs)
	if err != nil {
		// We could have many errors, return only the first encountered.
		if errs, ok := err.(sarama.ProducerErrors); ok {
			for _, prodErr := range errs {
				if prodErr.Err == sarama.ErrMessageSizeTooLarge {
					log.Printf("E! Error writing to output [kafka]: Message too large, consider increasing `max_message_bytes`; dropping batch")
					return nil
				}
				return prodErr
			}
		}
		return err
	}

	return nil
}

func init() {
	outputs.Add("kafka", func() telex.Output {
		return &Kafka{
			MaxRetry:     3,
			RequiredAcks: -1,
		}
	})
}
