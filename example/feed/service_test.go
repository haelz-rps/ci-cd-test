package feed

import (
	context "context"
	"testing"

	"github.com/stretchr/testify/suite"
)

type FeedServiceTestSuite struct {
	suite.Suite
}

func (s *FeedServiceTestSuite) TestCreateFeed() {
	mockFeedRepo := NewMockFeedRepository(s.T())
	mockPostRepo := NewMockPostRepository(s.T())

	feedSvc := NewFeedService(mockFeedRepo, mockPostRepo)

	testCases := []struct {
		name    string
		feed    Feed
		returns []interface{}
	}{
		{
			name:    "Create Feed",
			feed:    Feed{},
			returns: []interface{}{nil},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(subT *testing.T) {
			subT.Parallel()

			ctx := context.Background()

			mockFeedRepo.On("Create", ctx, &tc.feed).Return(tc.returns...)
			feedSvc.CreateFeed(ctx, &tc.feed)

			s.True(mockFeedRepo.AssertCalled(subT, "Create", ctx, &tc.feed))
		})
	}
}

func (s *FeedServiceTestSuite) TestCreatePost() {
	mockFeedRepo := NewMockFeedRepository(s.T())
	mockPostRepo := NewMockPostRepository(s.T())

	feedSvc := NewFeedService(mockFeedRepo, mockPostRepo)

	testCases := []struct {
		name    string
		post    Post
		returns []interface{}
	}{
		{
			name:    "Create Post",
			post:    Post{},
			returns: []interface{}{nil},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(subT *testing.T) {
			subT.Parallel()

			ctx := context.Background()

			mockPostRepo.On("Create", ctx, &tc.post).Return(tc.returns...)
			feedSvc.CreatePost(ctx, &tc.post)

			s.True(mockPostRepo.AssertCalled(subT, "Create", ctx, &tc.post))
		})
	}
}

func (s *FeedServiceTestSuite) TestListFeeds() {
	mockFeedRepo := NewMockFeedRepository(s.T())
	mockPostRepo := NewMockPostRepository(s.T())

	feedSvc := NewFeedService(mockFeedRepo, mockPostRepo)

	testCases := []struct {
		name    string
		returns []interface{}
	}{
		{
			name:    "List Feeds",
			returns: []interface{}{[]Feed{}, nil},
		},
	}

	for _, tc := range testCases {
		tc := tc
		s.T().Run(tc.name, func(subT *testing.T) {
			subT.Parallel()

			ctx := context.Background()

			mockFeedRepo.On("List", ctx).Return(tc.returns...)
			feedSvc.ListFeeds(ctx)

			s.True(mockFeedRepo.AssertCalled(subT, "List", ctx))
		})
	}
}

func TestFeedServiceTestSuite(t *testing.T) {
	suite.Run(t, new(FeedServiceTestSuite))
}
