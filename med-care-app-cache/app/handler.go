package main

import (
	"context"
	"fmt"
	"github.com/rusinikita/system-design-trainer/med-care-app-cache/db"
	"github.com/rusinikita/system-design-trainer/med-care-app-cache/model"
	"golang.org/x/sync/errgroup"
	"time"
)

type Handler struct {
	repo *db.DashboardRepository
}

func NewHandler(repo *db.DashboardRepository) *Handler {
	return &Handler{repo}
}

func (h *Handler) UserDashboard(ctx context.Context, userID int64, publishedFrom *time.Time, limit int) (*model.FeedResponse, error) {
	startTime := time.Now()
	var err error
	defer func() {
		status := "400"
		if err != nil {
			status = "500"
		}
		duration := time.Since(startTime).Seconds()
		httpRequestsTotal.WithLabelValues("feed", status).Inc()
		httpRequestDuration.WithLabelValues("feed", status).Observe(duration)
	}()

	// Create context with timeout
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// Initialize response struct
	response := model.FeedResponse{}

	// Create error group
	g, ctx := errgroup.WithContext(ctx)

	// Get articles
	g.Go(func() error {
		startOp := time.Now()
		articles, err := h.repo.GetArticleFeed(ctx, userID, limit, publishedFrom)
		dbOperationDuration.WithLabelValues("get_article_feed").Observe(time.Since(startOp).Seconds())
		if err != nil {
			return err
		}
		response.Articles = articles
		return nil
	})

	// Get care plan steps
	g.Go(func() error {
		startOp := time.Now()
		steps, err := h.repo.GetLatestCarePlanSteps(ctx, userID)
		dbOperationDuration.WithLabelValues("get_care_plan_steps").Observe(time.Since(startOp).Seconds())
		if err != nil {
			return err
		}
		response.Steps = steps
		return nil
	})

	// Wait for all goroutines and check for errors
	if err = g.Wait(); err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	return &response, nil
}
