package feed

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	kitlog "github.com/go-kit/log"
	"github.com/go-kit/log/level"
	kafka "github.com/segmentio/kafka-go"
)

func NewKafkaMessageHandler(endpointSet *EndpointSet, logger kitlog.Logger) func(ctx context.Context, m kafka.Message, r *kafka.Message) {
	return func(ctx context.Context, m kafka.Message, r *kafka.Message) {
		level.Debug(logger).Log("topic", m.Topic, "partition", m.Partition, "offset", m.Offset, "message", m.Value)
		headersMap := headersToMap(m.Headers)
		switch headersMap["function"] {
		case "create-feed":
			level.Debug(logger).Log("CREATE FEED")
			body, _ := decodeCreateFeedKafkaMessage(ctx, m)
			res, _ := endpointSet.CreateFeedEndpoint(ctx, body)
			encodeBrokerResponse(ctx, r, res)
			r.Headers = append(
				r.Headers,
				kafka.Header{Key: ":status", Value: []byte(fmt.Sprintf("%d", http.StatusCreated))},
			)
		case "list-feeds":
			level.Debug(logger).Log("LIST FEEDS")
			res, _ := endpointSet.ListFeedsEndpoint(ctx, nil)
			encodeBrokerResponse(ctx, r, res)
			r.Headers = append(
				r.Headers,
				kafka.Header{Key: ":status", Value: []byte(fmt.Sprintf("%d", http.StatusOK))},
			)
		default:
			level.Debug(logger).Log("NO HANDLER FOR TOPIC")
		}
		r.Topic = headersMap["zilla:reply-to"]
		r.Headers = append(
			r.Headers,
			kafka.Header{Key: "zilla:correlation-id", Value: []byte(headersMap["zilla:correlation-id"])},
		)
		r.Key = m.Key
	}
}

func headersToMap(headers []kafka.Header) map[string]string {
	headersMap := make(map[string]string)
	for _, h := range headers {
		headersMap[h.Key] = string(h.Value)
	}
	return headersMap
}

func decodeCreateFeedKafkaMessage(_ context.Context, m kafka.Message) (interface{}, error) {
	var body createFeedRequest

	if err := json.Unmarshal(m.Value, &body); err != nil {
		return nil, err
	}

	return body, nil
}

func encodeBrokerResponse(_ context.Context, r *kafka.Message, response interface{}) error {
	var err error
	r.Headers = append(r.Headers, kafka.Header{Key: "Content-Type", Value: []byte("application/json; charset=utf-8")})
	r.Value, err = json.Marshal(response)
	if err != nil {
		return err
	}
	return nil
}
