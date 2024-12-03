package gormimpl

import (
	"time"

	"github.com/Schub-cloud/security-bulletins/api/feed"
)

func ValidFeed() feed.Feed {
	return feed.Feed{
		Name: "feed",
		Link: "https://fake.com",
	}
}

func ValidFeedDAO() FeedDAO {
	return FeedDAO{
		Name: "feed",
		Link: "https://fake.com",
	}
}

func ValidPost() feed.Post {
	return feed.Post{
		FeedID:  1,
		Title:   "feed",
		PubDate: time.Now(),
		Link:    "https://fake.com",
	}
}

func ValidPostDAO() PostDAO {
	return PostDAO{
		FeedID:  1,
		Title:   "feed",
		PubDate: time.Now(),
		Link:    "https://fake.com",
	}
}
