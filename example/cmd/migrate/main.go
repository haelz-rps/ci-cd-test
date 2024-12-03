package main

import (
	"context"

	"github.com/Schub-cloud/security-bulletins/api/repository"
	"github.com/Schub-cloud/security-bulletins/api/repository/gormimpl"
	"github.com/Schub-cloud/security-bulletins/api/utils"
)

func main() {
	ctx := context.Background()

	dsn := utils.BuildDBConnectionString()

	repo := repository.NewGormRepo()
	err := repo.Dial(ctx, dsn)
	if err != nil {
		panic(err)
	}

	err = repo.Migrate(ctx, &gormimpl.FeedDAO{}, &gormimpl.PostDAO{})
	if err != nil {
		panic(err)
	}

	repo.Close(ctx)
}
