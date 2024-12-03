package gormimpl

import (
	"context"
	"testing"

	"github.com/Schub-cloud/security-bulletins/api/feed"
	"github.com/Schub-cloud/security-bulletins/api/utils"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type PostDAOTestSuite struct {
	suite.Suite
	postRepo feed.PostRepository
	db       *gorm.DB
	ctx      context.Context
}

func (s *PostDAOTestSuite) SetupSuite() {
	ctx := context.Background()
	utils.SetTestEnv()
	s.db, _ = gorm.Open(postgres.Open(utils.BuildDBConnectionString()), &gorm.Config{})
	s.postRepo = NewPostsRepository(s.db)
	s.ctx = ctx
}

func (s *PostDAOTestSuite) TearDownSuite() {
	db, _ := s.db.DB()
	db.Close()
}

func (s *PostDAOTestSuite) TearDownTest() {
	s.db.Where("name != ?", "_").Delete(&feed.Post{})
}

func (s *PostDAOTestSuite) TestCreateFeed() {
	post := ValidPost()

	s.postRepo.Create(s.ctx, &post)

	s.NotNil(post.CreatedAt)
	s.NotNil(post.ID)
}

func (s *PostDAOTestSuite) TestListFeed() {
	mockFeed := ValidFeedDAO()
	post1 := ValidPostDAO()
	post2 := ValidPostDAO()

	post1.Title = "feed1"
	post2.Title = "feed2"

	var expectedPost1 feed.Post
	var expectedPost2 feed.Post

	s.db.Create(&mockFeed)
	post1.FeedID = mockFeed.ID
	post2.FeedID = mockFeed.ID
	s.db.Create(&post1)
	s.db.Create(&post2)

	post1.toModel(&expectedPost1)
	post2.toModel(&expectedPost2)

	posts, _ := s.postRepo.List(s.ctx, int(post1.FeedID))

	s.Equal(len(posts), 2)
	s.Equal(expectedPost1, posts[0])
	s.Equal(expectedPost2, posts[1])
}

func TestPostDAOTestSuite(t *testing.T) {
	suite.Run(t, new(FeedDAOTestSuite))
}
