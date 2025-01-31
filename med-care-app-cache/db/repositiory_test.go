package db

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetArticleFeed(t *testing.T) {
	// 1. Prepare test database
	db, err := sql.Open("postgres", "postgres://localhost:5432/testdb?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Clear test data
	cleanup := []string{
		"DELETE FROM read_articles",
		"DELETE FROM article_segments",
		"DELETE FROM articles",
		"DELETE FROM user_segments",
	}
	for _, query := range cleanup {
		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err)
	}

	// Insert test data
	testData := []string{
		`INSERT INTO articles (id, title, content, source, type, published_at) VALUES
			(1, 'Article 1', 'Content 1', 'Source 1', 'news', NOW() - INTERVAL '1 day'),
			(2, 'Article 2', 'Content 2', 'Source 1', 'scientific', NOW() - INTERVAL '2 days'),
			(3, 'Article 3', 'Content 3', 'Source 2', 'news', NOW() - INTERVAL '3 days')`,

		`INSERT INTO user_segments (user_id, segment_id, weight) VALUES
			(1, 1, 0.8),
			(1, 2, 0.6)`,

		`INSERT INTO article_segments (article_id, segment_id, relevance_score) VALUES
			(1, 1, 0.9),
			(2, 1, 0.5),
			(2, 2, 0.7),
			(3, 2, 0.3)`,

		`INSERT INTO read_articles (user_id, article_id, read_at, is_saved) VALUES
			(1, 1, NOW(), true)`,
	}

	for _, query := range testData {
		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err)
	}

	// 2. Create repository
	repo := NewDashboardRepository(db)

	// 3. Call method
	articles, err := repo.GetArticleFeed(ctx, 1, 10, nil)
	require.NoError(t, err)

	// 4. Verify results
	assert.Len(t, articles, 3)

	// Check first article (should be most relevant)
	assert.Equal(t, int64(1), articles[0].ID)
	assert.True(t, articles[0].IsRead)
	assert.True(t, articles[0].IsSaved)
	assert.InDelta(t, 0.72, articles[0].Relevance, 0.01) // 0.8 * 0.9

	// Check second article
	assert.Equal(t, int64(2), articles[1].ID)
	assert.False(t, articles[1].IsRead)
	assert.False(t, articles[1].IsSaved)
	// Should combine both segment scores: (0.8 * 0.5) + (0.6 * 0.7)
	assert.InDelta(t, 0.82, articles[1].Relevance, 0.01)

	// Test pagination
	publishedFrom := articles[1].PublishedAt
	articlesPage2, err := repo.GetArticleFeed(ctx, 1, 2, &publishedFrom)
	require.NoError(t, err)
	assert.Len(t, articlesPage2, 1)
	assert.Equal(t, int64(3), articlesPage2[0].ID)
}

func TestGetLatestCarePlanSteps(t *testing.T) {
	// 1. Prepare test database
	db, err := sql.Open("postgres", "postgres://localhost:5432/testdb?sslmode=disable")
	require.NoError(t, err)
	defer db.Close()

	ctx := context.Background()

	// Clear test data
	cleanup := []string{
		"DELETE FROM care_plan_steps",
		"DELETE FROM care_plans",
	}
	for _, query := range cleanup {
		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err)
	}

	// Insert test data
	now := time.Now()
	testData := []string{
		`INSERT INTO care_plans (id, user_id, title, status, start_date, end_date) VALUES
			(1, 1, 'Care Plan 1', 'active', NOW(), NOW() + INTERVAL '30 days')`,

		fmt.Sprintf(`INSERT INTO care_plan_steps (id, care_plan_id, type, title, description, status, due_date, metadata, order_number) VALUES
			(1, 1, 'appointment', 'Doctor Visit', 'Regular checkup', 'overdue', '%s', '{"doctor": "Dr. Smith"}', 1),
			(2, 1, 'test', 'Blood Test', 'Annual test', 'pending', '%s', '{"test_type": "blood"}', 2),
			(3, 1, 'checkup', 'Weight Check', 'Monthly check', 'pending', '%s', '{}', 3),
			(4, 1, 'upload', 'Upload Results', 'Test results', 'completed', '%s', '{}', 4)`,
			now.Add(-24*time.Hour).Format(time.RFC3339),  // overdue
			now.Add(24*time.Hour).Format(time.RFC3339),   // tomorrow
			now.Add(48*time.Hour).Format(time.RFC3339),   // day after tomorrow
			now.Add(-48*time.Hour).Format(time.RFC3339)), // completed
	}

	for _, query := range testData {
		_, err := db.ExecContext(ctx, query)
		require.NoError(t, err)
	}

	// 2. Create repository
	repo := NewDashboardRepository(db)

	// 3. Call method
	steps, err := repo.GetLatestCarePlanSteps(ctx, 1)
	require.NoError(t, err)

	// 4. Verify results
	assert.Len(t, steps, 3)

	// First step should be overdue
	assert.Equal(t, "s1", steps[0].ID)
	assert.Equal(t, "overdue", steps[0].Status)
	assert.Equal(t, "appointment", steps[0].Type)

	// Second and third steps should be pending, ordered by due date
	assert.Equal(t, "s2", steps[1].ID)
	assert.Equal(t, "pending", steps[1].Status)
	assert.Equal(t, "test", steps[1].Type)

	assert.Equal(t, "s3", steps[2].ID)
	assert.Equal(t, "pending", steps[2].Status)
	assert.Equal(t, "checkup", steps[2].Type)

	// Verify order by due date
	assert.True(t, steps[0].DueDate.Before(steps[1].DueDate))
	assert.True(t, steps[1].DueDate.Before(steps[2].DueDate))
}
