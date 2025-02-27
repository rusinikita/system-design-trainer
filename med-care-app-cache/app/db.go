package main

import (
	"context"
	"github.com/rusinikita/system-design-trainer/med-care-app-cache/model"
	"time"
)

type DashboardRepository interface {
	GetArticleFeed(ctx context.Context, userID int64, limit int, publishedFrom *time.Time) ([]model.Article, error)
	GetLatestCarePlanSteps(ctx context.Context, userID int64) ([]model.CarePlanStep, error)
}
