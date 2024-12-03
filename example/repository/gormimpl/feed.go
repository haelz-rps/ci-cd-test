package gormimpl

import (
	"context"

	"github.com/Schub-cloud/security-bulletins/api/feed"
	"gorm.io/gorm"
)

type feedRepository struct {
	db *gorm.DB
}

func NewFeedRepository(db *gorm.DB) feedRepository {
	return feedRepository{
		db: db,
	}
}

type FeedDAO struct {
	gorm.Model
	Name  string
	Posts []PostDAO `gorm:"foreignKey:FeedID"`
	Link  string
}

func (FeedDAO) TableName() string {
	return "feeds"
}

type feedsDAO []FeedDAO

func (f FeedDAO) toModel(feed *feed.Feed) {
	feed.ID = f.ID
	feed.Name = f.Name
	feed.Link = f.Link
	feed.CreatedAt = f.CreatedAt
	feed.UpdatedAt = f.UpdatedAt
	feed.DeletedAt = f.DeletedAt.Time
}

func (fr feedRepository) Create(ctx context.Context, feed *feed.Feed) error {
	db := fr.db.WithContext(ctx)

	feedDAO := FeedDAO{
		Name: feed.Name,
		Link: feed.Link,
	}

	result := db.Create(&feedDAO)
	if result.Error != nil {
		return result.Error
	}

	feedDAO.toModel(feed)

	return nil
}

func (fr feedRepository) List(ctx context.Context) ([]feed.Feed, error) {
	var fDAO feedsDAO
	result := fr.db.Find(&fDAO)

	if result.Error != nil {
		return nil, result.Error
	}

	feeds := make([]feed.Feed, result.RowsAffected)
	for i, f := range fDAO {
		f.toModel(&feeds[i])
	}

	return feeds, nil
}
