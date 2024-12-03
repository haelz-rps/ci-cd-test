package feed

import (
	"context"
	"time"

	"github.com/go-kit/kit/metrics"
)

type metricsMiddleware struct {
	requestCount   metrics.Counter
	requestLatency metrics.Histogram
	fs             FeedService
}

func NewMetricsMiddleware(counter metrics.Counter, latency metrics.Histogram, s FeedService) FeedService {
	return &metricsMiddleware{
		requestCount:   counter,
		requestLatency: latency,
		fs:             s,
	}
}

func (s *metricsMiddleware) CreateFeed(ctx context.Context, feed *Feed) error {
	labels := []string{"method", "create_feed"}
	defer func(begin time.Time) {
		s.requestCount.With(labels...).Add(1)
		s.requestLatency.With(labels...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.fs.CreateFeed(ctx, feed)
}

func (s *metricsMiddleware) CreatePost(ctx context.Context, post *Post) error {
	labels := []string{"method", "create_post"}
	defer func(begin time.Time) {
		s.requestCount.With(labels...).Add(1)
		s.requestLatency.With(labels...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.fs.CreatePost(ctx, post)
}

func (s *metricsMiddleware) ListFeeds(ctx context.Context) ([]Feed, error) {
	labels := []string{"method", "list_feeds"}
	defer func(begin time.Time) {
		s.requestCount.With(labels...).Add(1)
		s.requestLatency.With(labels...).Observe(time.Since(begin).Seconds())
	}(time.Now())
	return s.fs.ListFeeds(ctx)
}
