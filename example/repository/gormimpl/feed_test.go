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

type FeedDAOTestSuite struct {
	suite.Suite
	feedRepo feed.FeedRepository
	db       *gorm.DB
	ctx      context.Context
}

func (s *FeedDAOTestSuite) SetupSuite() {
	ctx := context.Background()
	utils.SetTestEnv()
	s.db, _ = gorm.Open(postgres.Open(utils.BuildDBConnectionString()), &gorm.Config{})
	s.feedRepo = NewFeedRepository(s.db)
	s.ctx = ctx
}

func (s *FeedDAOTestSuite) TearDownSuite() {
	db, _ := s.db.DB()
	db.Close()
}

func (s *FeedDAOTestSuite) TearDownTest() {
	s.db.Where("name != ?", "_").Delete(&feed.Feed{})
}

func (s *FeedDAOTestSuite) TestCreateFeed() {
	feed := ValidFeed()

	s.feedRepo.Create(s.ctx, &feed)

	s.NotNil(feed.CreatedAt)
	s.NotNil(feed.ID)
}

func (s *FeedDAOTestSuite) TestListFeed() {
	feed1 := ValidFeedDAO()
	feed2 := ValidFeedDAO()

	feed1.Name = "feed1"
	feed2.Name = "feed2"

	var expectedFeed1 feed.Feed
	var expectedFeed2 feed.Feed

	s.db.Create(&feed1)
	s.db.Create(&feed2)

	feed1.toModel(&expectedFeed1)
	feed2.toModel(&expectedFeed2)

	feeds, _ := s.feedRepo.List(s.ctx)

	s.Equal(len(feeds), 2)
	s.Equal(expectedFeed1, feeds[0])
	s.Equal(expectedFeed2, feeds[1])
}

func TestFeedDAOTestSuite(t *testing.T) {
	suite.Run(t, new(FeedDAOTestSuite))
}
