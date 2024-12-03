package ingestions

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/message/router/plugin"
	"github.com/google/uuid"

	"github.com/ubix/dripper/manager/constants"
)

// TODO: use AddNoPublisherHandler since we don't need the returned messages
func NewKafkaWorkerHandler(EndpointSet *EndpointSet, subscriber message.Subscriber, publisher message.Publisher, router *message.Router) {

	router.AddPlugin(plugin.SignalsHandler)

	deadLetterMiddleware, _ := middleware.PoisonQueue(publisher, constants.BrokerTopics.DeadLetter)

	router.AddMiddleware(
		middleware.Retry{
			MaxRetries:      5,
			InitialInterval: time.Minute * 1,
		}.Middleware,
		deadLetterMiddleware,
	)

	router.AddHandler(
		constants.HandlerNames.CreateSource,
		constants.BrokerTopics.CreateSourceRequests,
		subscriber,
		constants.BrokerTopics.CreateSourceResponses,
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			req, err := decodeCreateSourceFromDefinitionRequestMessage(msg.Context(), msg)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			res, err := EndpointSet.CreateSourceFromDefinitionEndpoint(msg.Context(), req)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			payload, err := encodeCreateSourceFromDefinitionResponseMessage(msg.Context(), res)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			resMsg := message.NewMessage(msg.UUID, payload)
			resMsg.SetContext(msg.Context())

			return []*message.Message{resMsg}, nil
		},
	)

	router.AddHandler(
		constants.HandlerNames.UpdateSource,
		constants.BrokerTopics.UpdateSourceRequests,
		subscriber,
		constants.BrokerTopics.UpdateSourceResponses,
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			req, err := decodeUpdateSourceRequestMessage(msg.Context(), msg)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			res, err := EndpointSet.UpdateSourceEndpoint(msg.Context(), req)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			payload, err := encodeUpdateSourceResponseMessage(msg.Context(), res)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			resMsg := message.NewMessage(msg.UUID, payload)
			resMsg.SetContext(msg.Context())

			return []*message.Message{resMsg}, nil
		},
	)

	router.AddHandler(
		constants.HandlerNames.GetSourceTables,
		constants.BrokerTopics.GetSourceTablesRequests,
		subscriber,
		constants.BrokerTopics.GetSourceTablesResponses,
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			req, err := decodeGetSourceTablesRequestMessage(msg.Context(), msg)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			res, err := EndpointSet.GetSourceTablesEndpoint(msg.Context(), req)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			payload, err := encodeGetSourceTablesResponseMessage(msg.Context(), res)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			resMsg := message.NewMessage(msg.UUID, payload)
			resMsg.SetContext(msg.Context())

			return []*message.Message{resMsg}, nil
		},
	)

	router.AddHandler(
		constants.HandlerNames.SuccessfulSyncNotification,
		constants.BrokerTopics.SuccessfulSyncNotificationsRequests,
		subscriber,
		constants.BrokerTopics.SuccessfulSyncNotificationsResponses,
		publisher,
		func(msg *message.Message) ([]*message.Message, error) {
			req, err := decodeTriggerProfilingRequestMessage(msg.Context(), msg)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			_, err = EndpointSet.TriggerProfilingEndpoint(msg.Context(), req)
			if err != nil {
				router.Logger().Error("msg", err, nil)
				return nil, err
			}

			return []*message.Message{}, nil
		},
	)
}

func decodeCreateSourceFromDefinitionRequestMessage(_ context.Context, m *message.Message) (*createSourceFromDefinitionRequestMessage, error) {
	var createSourceBody createSourceFromDefinitionAsyncRequest
	err := json.Unmarshal(m.Payload, &createSourceBody)
	if err != nil {
		return nil, err
	}

	ticketId, err := uuid.Parse(m.UUID)
	if err != nil {
		return nil, err
	}

	return &createSourceFromDefinitionRequestMessage{
		body:     createSourceBody,
		ticketId: &ticketId,
	}, nil
}

func decodeGetSourceTablesRequestMessage(_ context.Context, m *message.Message) (*getSourceTablesRequestMessage, error) {
	var sourceId uuid.UUID
	err := json.Unmarshal(m.Payload, &sourceId)
	if err != nil {
		return nil, err
	}

	ticketId, err := uuid.Parse(m.UUID)
	if err != nil {
		return nil, err
	}

	return &getSourceTablesRequestMessage{
		sourceId: &sourceId,
		ticketId: &ticketId,
	}, nil
}

func encodeCreateSourceFromDefinitionResponseMessage(_ context.Context, response interface{}) ([]byte, error) {
	res := response.(*createSourceFromDefinitionResponseMessage)

	payload, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func encodeGetSourceTablesResponseMessage(_ context.Context, response interface{}) ([]byte, error) {
	res := response.(*getSourceTablesResponseMessage)

	payload, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return payload, nil
}

func decodeTriggerProfilingRequestMessage(_ context.Context, m *message.Message) (*triggerProfilingRequestMessage, error) {
	req := triggerProfilingRequestMessage{}
	err := json.Unmarshal(m.Payload, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func decodeUpdateSourceRequestMessage(_ context.Context, m *message.Message) (*updateSourceAsyncRequestMessage, error) {
	var updateSourceBody updateSourceAsyncRequest
	err := json.Unmarshal(m.Payload, &updateSourceBody)
	if err != nil {
		return nil, err
	}

	ticketId, err := uuid.Parse(m.UUID)
	if err != nil {
		return nil, err
	}

	return &updateSourceAsyncRequestMessage{
		source:   updateSourceBody,
		ticketId: &ticketId,
	}, nil
}

func encodeUpdateSourceResponseMessage(_ context.Context, response interface{}) ([]byte, error) {
	res := response.(*updateSourceAsyncResponseMessage)

	payload, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
