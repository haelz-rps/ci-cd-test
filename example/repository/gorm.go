package repository

import (
	"context"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type gormRepo struct {
	DB *gorm.DB
}

func NewGormRepo() *gormRepo {
	return &gormRepo{}
}

func (r *gormRepo) Dial(ctx context.Context, host string) error {
	db, err := gorm.Open(postgres.Open(host), &gorm.Config{})
	if err != nil {
		return err
	}
	r.DB = db

	return nil
}

func (r *gormRepo) Migrate(ctx context.Context, models ...interface{}) error {
	err := r.DB.AutoMigrate(models...)
	return err
}

func (r *gormRepo) Close(ctx context.Context) error {
	db, err := r.DB.DB()
	if err != nil {
		return err
	}
	db.Close()

	return nil
}
