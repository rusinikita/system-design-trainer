package main

import (
	"context"
	"fmt"
	"github.com/rusinikita/system-design-trainer/med-care-app-cache/model"
	"github.com/rusinikita/system-design-trainer/tooling/metrics"
	"golang.org/x/sync/errgroup"
	"time"
)

type Handler struct {
	repo DashboardRepository
	obs  metrics.Obs
}

func NewHandler(repo DashboardRepository, obs metrics.Obs) *Handler {
	return &Handler{
		repo: repo,
		obs:  obs,
	}
}

func (h *Handler) UserDashboard(ctx context.Context, userID int64, publishedFrom *time.Time, limit int) (*model.FeedResponse, error) {
	mainSpan := h.obs.StartSpan("UserDashboard")
	var err error
	defer func() {
		mainSpan.Done(err)
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
		op := h.obs.StartSpan("GetArticleFeed")
		articles, err := h.repo.GetArticleFeed(ctx, userID, limit, publishedFrom)
		op.Done(err)
		if err != nil {
			return err
		}
		response.Articles = articles
		return nil
	})

	// Get care plan steps
	g.Go(func() error {
		op := h.obs.StartSpan("GetLatestCarePlanSteps")
		steps, err := h.repo.GetLatestCarePlanSteps(ctx, userID)
		op.Done(err)
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
