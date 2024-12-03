package feed

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-kit/kit/transport"
	kithttp "github.com/go-kit/kit/transport/http"
	kitlog "github.com/go-kit/log"
)

// HTTP Handler
func NewHTTPHandler(endpointSet *EndpointSet, logger kitlog.Logger) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		kithttp.ServerErrorEncoder(encodeError),
	}

	createFeedHandler := kithttp.NewServer(
		endpointSet.CreateFeedEndpoint,
		decodeCreateFeedRequest,
		encodeResponse,
		opts...,
	)

	listFeedsHandler := kithttp.NewServer(
		endpointSet.ListFeedsEndpoint,
		decodeListFeedRequest,
		encodeResponse,
		opts...,
	)
	r := chi.NewRouter()

	r.Route("/feeds", func(r chi.Router) {
		r.Method(http.MethodPost, "/", createFeedHandler)
		r.Method(http.MethodGet, "/", listFeedsHandler)
	})

	return r
}

// Encoding
func decodeCreateFeedRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var body createFeedRequest

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, err
	}

	return body, nil
}

func decodeListFeedRequest(_ context.Context, r *http.Request) (interface{}, error) {
	return nil, nil
}

func encodeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if e, ok := response.(errorer); ok && e.error() != nil {
		encodeError(ctx, e.error(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

type errorer interface {
	error() error
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	switch err {
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": err.Error(),
	})
}
