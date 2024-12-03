package feed

import (
	"context"
	"time"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

type loggingMiddleware struct {
	logger log.Logger
	fs     FeedService
}

func NewLoggingMiddleware(logger log.Logger, s FeedService) FeedService {
	return &loggingMiddleware{
		logger: logger,
		fs:     s,
	}
}

func (s *loggingMiddleware) CreateFeed(ctx context.Context, feed *Feed) (err error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "CreateFeed",
			"feed", *feed,
			"took", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return s.fs.CreateFeed(ctx, feed)
}

func (s *loggingMiddleware) CreatePost(ctx context.Context, post *Post) (err error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "CreatePost",
			"feed", *post,
			"took", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return s.fs.CreatePost(ctx, post)
}

func (s *loggingMiddleware) ListFeeds(ctx context.Context) (f []Feed, err error) {
	defer func(begin time.Time) {
		level.Debug(s.logger).Log(
			"method", "ListFeeds",
			"feeds", f,
			"took", time.Since(begin),
			"error", err,
		)
	}(time.Now())
	return s.fs.ListFeeds(ctx)
}
