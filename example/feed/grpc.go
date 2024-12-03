package feed

import (
	"context"

	pb "github.com/Schub-cloud/security-bulletins/api/pb/feed"
	grpctransport "github.com/go-kit/kit/transport/grpc"
	kitlog "github.com/go-kit/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type grpcServer struct {
	pb.UnimplementedFeedServiceServer
	createFeed grpctransport.Handler
	listFeeds  grpctransport.Handler
}

func NewGRPCServer(endpoints EndpointSet, logger kitlog.Logger) pb.FeedServiceServer {
	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
	}

	return &grpcServer{
		createFeed: grpctransport.NewServer(
			endpoints.CreateFeedEndpoint,
			decodeGRPCCreateFeedRequest,
			encodeGRPCCreateFeedResponse,
			options...,
		),
		listFeeds: grpctransport.NewServer(
			endpoints.ListFeedsEndpoint,
			decodeGRPCListFeedsRequest,
			encodeGRPCListFeedsResponse,
			options...,
		),
	}
}

func (s *grpcServer) CreateFeed(ctx context.Context, req *pb.CreateFeedRequest) (*pb.CreateFeedResponse, error) {
	_, resp, err := s.createFeed.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.CreateFeedResponse), nil
}

func (s *grpcServer) ListFeeds(ctx context.Context, req *pb.ListFeedsRequest) (*pb.ListFeedsResponse, error) {
	_, resp, err := s.listFeeds.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*pb.ListFeedsResponse), nil
}

func encodeGRPCCreateFeedResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(createFeedResponse)
	return &pb.CreateFeedResponse{
		Feed: encodeGRPCFeed(resp.Feed),
		Err:  encodeGRPCError(resp.Err),
	}, nil
}

func encodeGRPCListFeedsResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(listFeedsResponse)
	return &pb.ListFeedsResponse{
		Feeds: encodeGRPCFeeds(resp.Feeds),
		Err:   encodeGRPCError(resp.Err),
	}, nil
}

func encodeGRPCFeeds(feeds []Feed) []*pb.Feed {
	var pbFeeds []*pb.Feed
	for _, f := range feeds {
		pbFeeds = append(pbFeeds, encodeGRPCFeed(f))
	}
	return pbFeeds
}

func encodeGRPCFeed(feed Feed) *pb.Feed {
	return &pb.Feed{
		Id:        uint64(feed.ID),
		Name:      feed.Name,
		Link:      feed.Link,
		Posts:     encodeGRPCPosts(feed.Posts),
		CreatedAt: timestamppb.New(feed.CreatedAt),
		UpdatedAt: timestamppb.New(feed.UpdatedAt),
		DeletedAt: timestamppb.New(feed.DeletedAt),
	}
}

func encodeGRPCPosts(posts []Post) []*pb.Post {
	var pbPosts []*pb.Post
	for _, p := range posts {
		pbPosts = append(pbPosts, encodeGRPCPost(p))
	}
	return pbPosts
}

func encodeGRPCPost(post Post) *pb.Post {
	return &pb.Post{
		Id:        uint64(post.ID),
		FeedID:    uint64(post.FeedID),
		Title:     post.Title,
		Link:      post.Link,
		PubDate:   timestamppb.New(post.PubDate),
		CreatedAt: timestamppb.New(post.CreatedAt),
		UpdatedAt: timestamppb.New(post.UpdatedAt),
		DeletedAt: timestamppb.New(post.DeletedAt),
	}
}

func encodeGRPCError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func decodeGRPCCreateFeedRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.CreateFeedRequest)
	return createFeedRequest{
		Name: req.Name,
		Link: req.Link,
	}, nil
}

func decodeGRPCListFeedsRequest(_ context.Context, _ interface{}) (interface{}, error) {
	return listFeedsRequest{}, nil
}
