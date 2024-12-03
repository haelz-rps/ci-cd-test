package gormimpl

import (
	"context"
	"time"

	"github.com/Schub-cloud/security-bulletins/api/feed"
	"gorm.io/gorm"
)

type postsRepository struct {
	db *gorm.DB
}

func NewPostsRepository(db *gorm.DB) postsRepository {
	return postsRepository{
		db: db,
	}
}

type PostDAO struct {
	gorm.Model
	FeedID  uint
	Title   string
	Link    string
	PubDate time.Time
}

func (PostDAO) TableName() string {
	return "posts"
}

type postsDAO []PostDAO

func (p PostDAO) toModel(post *feed.Post) {
	post.ID = p.ID
	post.FeedID = p.FeedID
	post.Title = p.Title
	post.Link = p.Link
	post.PubDate = p.PubDate
	post.CreatedAt = p.CreatedAt
	post.UpdatedAt = p.UpdatedAt
	post.DeletedAt = p.DeletedAt.Time
}

func (pr postsRepository) Create(ctx context.Context, post *feed.Post) error {
	db := pr.db.WithContext(ctx)

	postDAO := &PostDAO{
		Title: post.Title,
		Link:  post.Link,
	}
	result := db.Create(postDAO)

	if result.Error != nil {
		return result.Error
	}

	postDAO.toModel(post)

	return nil
}

func (pr postsRepository) List(ctx context.Context, feedID int) ([]feed.Post, error) {
	var pDAO postsDAO

	db := pr.db.WithContext(ctx)

	result := db.Where("feedID = ?", feedID).Find(&pDAO)

	if result.Error != nil {
		return nil, result.Error
	}

	posts := make([]feed.Post, result.RowsAffected)
	for i, f := range pDAO {
		f.toModel(&posts[i])
	}
	return posts, nil
}
