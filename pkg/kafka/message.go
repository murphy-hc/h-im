package kafka

import "github.com/segmentio/kafka-go"

// Header is a key-value pair for Kafka message headers. Wraps kafka-go's
// native Header to avoid allocation during produce/consume round-trips.
type Header = kafka.Header

// Message is a Kafka message with optional key, value, and headers.
type Message = kafka.Message
