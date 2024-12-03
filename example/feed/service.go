package feed

import (
	"context"
	"time"
)

// Models
type Post struct {
	ID        uint      `json:"id"`
	FeedID    uint      `json:"feedID"`
	Title     string    `json:"title"`
	Link      string    `json:"link"`
	PubDate   time.Time `json:"pubDate"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
}

type Feed struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Posts     []Post    `json:"posts"`
	Link      string    `json:"link"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	DeletedAt time.Time `json:"deletedAt"`
}

// Repository Interface
//
//go:generate mockery --name PostRepository
type PostRepository interface {
	Create(ctx context.Context, post *Post) error
	List(ctx context.Context, feedID int) ([]Post, error)
}

//go:generate mockery --name FeedRepository
type FeedRepository interface {
	Create(ctx context.Context, feed *Feed) error
	List(ctx context.Context) ([]Feed, error)
}

// Service Interface
//
//go:generate mockery --name FeedService
type FeedService interface {
	CreateFeed(ctx context.Context, feed *Feed) error
	CreatePost(ctx context.Context, post *Post) error
	ListFeeds(ctx context.Context) ([]Feed, error)
}

// Service Implementation
type feedService struct {
	feedRepo FeedRepository
	postRepo PostRepository
}

func NewFeedService(feedRepo FeedRepository, postRepo PostRepository) *feedService {
	return &feedService{
		feedRepo: feedRepo,
		postRepo: postRepo,
	}
}

func (s *feedService) CreateFeed(ctx context.Context, feed *Feed) error {
	return s.feedRepo.Create(ctx, feed)
}

func (s *feedService) CreatePost(ctx context.Context, post *Post) error {
	return s.postRepo.Create(ctx, post)
}

func (s *feedService) ListFeeds(ctx context.Context) ([]Feed, error) {
	feeds, err := s.feedRepo.List(ctx)
	return feeds, err
}
