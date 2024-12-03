package feed

import (
	"context"

	"github.com/go-kit/kit/endpoint"
)

type EndpointSet struct {
	CreateFeedEndpoint endpoint.Endpoint
	ListFeedsEndpoint  endpoint.Endpoint
}

func NewEndpointSet(svc FeedService) *EndpointSet {
	return &EndpointSet{
		CreateFeedEndpoint: makeCreateFeedEndpoint(svc),
		ListFeedsEndpoint:  makeListFeedsEndpoint(svc),
	}
}

func makeCreateFeedEndpoint(svc FeedService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(createFeedRequest)
		feed := &Feed{
			Name: req.Name,
			Link: req.Link,
		}
		err := svc.CreateFeed(ctx, feed)
		if err != nil {
			return createFeedResponse{
				Err: err,
			}, nil
		}
		return createFeedResponse{
			Feed: *feed,
			Err:  nil,
		}, nil
	}
}

func makeListFeedsEndpoint(svc FeedService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		feeds, err := svc.ListFeeds(ctx)
		if err != nil {
			return listFeedsResponse{
				Err: err,
			}, nil
		}
		return listFeedsResponse{
			Feeds: feeds,
		}, nil
	}
}

// Request/Response schemas
// CreateFeed
type createFeedRequest struct {
	Name string `json:"name"`
	Link string `json:"link"`
}

type createFeedResponse struct {
	Feed Feed  `json:"feed,omitempty"`
	Err  error `json:"error,omitempty"`
}

// ListFeeds

type listFeedsRequest struct{}

type listFeedsResponse struct {
	Feeds []Feed `json:"feeds,omitempty"`
	Err   error  `json:"error,omitempty"`
}
