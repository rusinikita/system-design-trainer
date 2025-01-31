package db

import (
	"context"
	"database/sql"
	"github.com/rusinikita/system-design-trainer/med-care-app-cache/model"
	"time"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{
		db: db,
	}
}

// GetArticleFeed returns personalized articles based on user segments with pagination
func (r *DashboardRepository) GetArticleFeed(ctx context.Context, userID int64, limit int, publishedFrom *time.Time) ([]model.Article, error) {
	query := `
		WITH user_segments AS (
			SELECT segment_id, weight
			FROM user_segments
			WHERE user_id = $1
		)
		SELECT DISTINCT ON (a.id)
			a.id,
			a.title,
			a.content,
			a.source,
			a.type,
			a.published_at,
			ra.read_at IS NOT NULL as is_read,
			COALESCE(ra.is_saved, false) as is_saved,
			COALESCE(
				SUM(us.weight * ags.relevance_score) 
				OVER (PARTITION BY a.id),
				0
			) as relevance
		FROM articles a
		JOIN article_segments ags ON a.id = ags.article_id
		JOIN user_segments us ON ags.segment_id = us.segment_id
		LEFT JOIN read_articles ra ON a.id = ra.article_id AND ra.user_id = $1
		WHERE ($3::timestamp IS NULL OR a.published_at < $3)
		ORDER BY a.id, relevance DESC, a.published_at DESC
		LIMIT $2
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, publishedFrom)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		var a model.Article
		err := rows.Scan(
			&a.ID,
			&a.Title,
			&a.Content,
			&a.Source,
			&a.Type,
			&a.PublishedAt,
			&a.IsRead,
			&a.IsSaved,
			&a.Relevance,
		)
		if err != nil {
			return nil, err
		}
		articles = append(articles, a)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// GetLatestCarePlanSteps returns the 3 most recent care plan steps for a user
func (r *DashboardRepository) GetLatestCarePlanSteps(ctx context.Context, userID int64) ([]model.CarePlanStep, error) {
	query := `
		WITH user_care_plans AS (
			SELECT id
			FROM care_plans
			WHERE user_id = $1 AND status = 'active'
		)
		SELECT 
			cps.id,
			cps.type,
			cps.title,
			cps.description,
			cps.status,
			cps.due_date,
			cps.completed_at,
			cps.metadata,
			cps.order_number
		FROM care_plan_steps cps
		JOIN user_care_plans ucp ON cps.care_plan_id = ucp.id
		WHERE cps.status != 'completed'
		ORDER BY 
			CASE 
				WHEN cps.status = 'overdue' THEN 1
				WHEN cps.status = 'pending' THEN 2
				ELSE 3
			END,
			cps.due_date ASC
		LIMIT 3
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []model.CarePlanStep
	for rows.Next() {
		var s model.CarePlanStep
		err := rows.Scan(
			&s.ID,
			&s.Type,
			&s.Title,
			&s.Description,
			&s.Status,
			&s.DueDate,
			&s.CompletedAt,
			&s.Metadata,
			&s.OrderNumber,
		)
		if err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return steps, nil
}
