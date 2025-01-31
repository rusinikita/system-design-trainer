package main

import (
	"context"
	_ "embed"
	"fmt"
	dbTool "github.com/rusinikita/system-design-trainer/tooling/db"
	"log"
	"math/rand"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

//go:embed schema.sql
var schemaSQL string

const (
	// Input parameters
	usersCount        = 1_000_000
	articlesCount     = 10_000
	segmentTypesCount = 50 // Total number of possible segments

	// Ranges
	minUserSegments    = 2
	maxUserSegments    = 7
	minArticleSegments = 1
	maxArticleSegments = 5
	minSteps           = 10
	maxSteps           = 20

	// Batch sizes
	batchSize = 1000
)

func main() {
	db := dbTool.Conn()
	defer db.Close()

	ctx := context.Background()

	_, err := db.Exec(schemaSQL)
	if err != nil {
		log.Fatalf("Failed to load schema: %v", err)
	}

	// Clear existing data
	log.Println("Clearing existing data...")
	tables := []string{
		"read_articles",
		"article_segments",
		"articles",
		"care_plan_steps",
		"care_plans",
		"user_segments",
		"segment_types",
	}
	for _, table := range tables {
		if _, err := db.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			log.Fatalf("Failed to truncate %s: %v", table, err)
		}
	}

	// Initialize segment types
	segmentTypes := make([]string, segmentTypesCount)
	values := make([]string, segmentTypesCount)
	args := make([]interface{}, 0, segmentTypesCount*3)

	for i := 0; i < segmentTypesCount; i++ {
		segmentTypes[i] = fmt.Sprintf("medical_condition_%d", i)
		values[i] = fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args = append(args,
			i,
			segmentTypes[i],
			fmt.Sprintf("Description for %s", segmentTypes[i]),
		)
	}

	// Insert segment types
	log.Println("Inserting segment types...")
	query := fmt.Sprintf("INSERT INTO segment_types (id, name, description) VALUES %s",
		strings.Join(values, ","))
	if _, err := db.ExecContext(ctx, query, args...); err != nil {
		log.Fatalf("Failed to insert segment types: %v", err)
	}

	// Insert articles and their segments
	log.Println("Inserting articles and segments...")
	for i := 0; i < articlesCount; i += batchSize {
		// Insert articles batch
		articleBatch := make([]interface{}, 0, batchSize*6)
		articleValueStrings := make([]string, 0, batchSize)
		articleIDs := make([]int64, batchSize)

		end := i + batchSize
		if end > articlesCount {
			end = articlesCount
		}

		for j := i; j < end; j++ {
			articleID := int64(j)
			articleIDs[j-i] = articleID
			valueIdx := (j - i) * 6

			articleValueStrings = append(articleValueStrings,
				fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d)",
					valueIdx+1, valueIdx+2, valueIdx+3, valueIdx+4, valueIdx+5, valueIdx+6))

			_type := "scientific"
			if rand.Intn(2) == 0 {
				_type = "news"
			}

			articleBatch = append(articleBatch,
				articleID,
				fmt.Sprintf("Article Title %d", j),
				fmt.Sprintf("Content for article %d", j),
				"Source",
				_type,
				time.Now().Add(-time.Duration(rand.Intn(365))*24*time.Hour),
			)
		}

		query := fmt.Sprintf("INSERT INTO articles (id, title, content, source, type, published_at) VALUES %s",
			strings.Join(articleValueStrings, ","))
		if _, err := db.ExecContext(ctx, query, articleBatch...); err != nil {
			log.Fatalf("Failed to insert articles: %v", err)
		}

		// Insert article segments
		for _, articleID := range articleIDs[:end-i] {
			segmentCount := minArticleSegments + rand.Intn(maxArticleSegments-minArticleSegments+1)
			segments := rand.Perm(len(segmentTypes))[:segmentCount]

			segmentBatch := make([]interface{}, 0, segmentCount*3)
			segmentValueStrings := make([]string, 0, segmentCount)

			for j, segIdx := range segments {
				valueIdx := j * 3
				segmentValueStrings = append(segmentValueStrings,
					fmt.Sprintf("($%d, $%d, $%d)",
						valueIdx+1, valueIdx+2, valueIdx+3))

				segmentBatch = append(segmentBatch,
					articleID,
					segIdx,
					0.1+rand.Float64()*0.9, // relevance score between 0.1 and 1.0
				)
			}

			query := fmt.Sprintf("INSERT INTO article_segments (article_id, segment_id, relevance_score) VALUES %s",
				strings.Join(segmentValueStrings, ","))
			if _, err := db.ExecContext(ctx, query, segmentBatch...); err != nil {
				log.Fatalf("Failed to insert article segments: %v", err)
			}
		}
	}

	// Insert users, their segments, care plans and steps
	log.Println("Inserting users and related data...")
	for i := 0; i < usersCount; i += batchSize {
		end := i + batchSize
		if end > usersCount {
			end = usersCount
		}

		// Insert users batch
		userBatch := make([]interface{}, 0, batchSize*2)
		userValueStrings := make([]string, 0, batchSize)

		for j := i; j < end; j++ {
			valueIdx := (j - i) * 2
			userValueStrings = append(userValueStrings,
				fmt.Sprintf("($%d, $%d)",
					valueIdx+1, valueIdx+2))

			userBatch = append(userBatch,
				j,                         // user id
				fmt.Sprintf("User %d", j), // user name
			)
		}

		query := fmt.Sprintf("INSERT INTO users (id, name) VALUES %s",
			strings.Join(userValueStrings, ","))
		if _, err := db.ExecContext(ctx, query, userBatch...); err != nil {
			log.Fatalf("Failed to insert users: %v", err)
		}

		// Process other user-related data in batches
		for j := i; j < end; j++ {
			userID := j

			// Insert user segments
			segmentCount := minUserSegments + rand.Intn(maxUserSegments-minUserSegments+1)
			segments := rand.Perm(len(segmentTypes))[:segmentCount]

			segmentBatch := make([]interface{}, 0, segmentCount*3)
			segmentValueStrings := make([]string, 0, segmentCount)

			for k, segIdx := range segments {
				valueIdx := k * 3
				segmentValueStrings = append(segmentValueStrings,
					fmt.Sprintf("($%d, $%d, $%d)",
						valueIdx+1, valueIdx+2, valueIdx+3))

				segmentBatch = append(segmentBatch,
					userID,
					segIdx,
					0.1+rand.Float64()*0.9, // weight between 0.1 and 1.0
				)
			}

			query := fmt.Sprintf("INSERT INTO user_segments (user_id, segment_id, weight) VALUES %s",
				strings.Join(segmentValueStrings, ","))
			if _, err := db.ExecContext(ctx, query, segmentBatch...); err != nil {
				log.Fatalf("Failed to insert user segments: %v", err)
			}

			// Create care plan
			carePlanID := j
			_, err := db.ExecContext(ctx,
				`INSERT INTO care_plans (id, user_id, title, status, start_date, end_date)
				VALUES ($1, $2, $3, $4, $5, $6)`,
				carePlanID,
				userID,
				fmt.Sprintf("Care Plan for User %d", j),
				"active",
				time.Now(),
				time.Now().AddDate(0, 6, 0),
			)
			if err != nil {
				log.Fatalf("Failed to insert care plan: %v", err)
			}

			// Insert care plan steps
			stepCount := minSteps + rand.Intn(maxSteps-minSteps+1)
			stepTypes := []string{"appointment", "test", "upload", "checkup"}
			stepStatuses := []string{"pending", "completed", "overdue"}

			stepBatch := make([]interface{}, 0, stepCount*9)
			stepValueStrings := make([]string, 0, stepCount)

			for k := 0; k < stepCount; k++ {
				valueIdx := k * 9
				stepValueStrings = append(stepValueStrings,
					fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)",
						valueIdx+1, valueIdx+2, valueIdx+3, valueIdx+4, valueIdx+5,
						valueIdx+6, valueIdx+7, valueIdx+8, valueIdx+9))

				stepType := stepTypes[rand.Intn(len(stepTypes))]
				dueDate := time.Now().AddDate(0, 0, rand.Intn(180))
				var completedAt *time.Time
				status := stepStatuses[rand.Intn(len(stepStatuses))]
				if status == "completed" {
					t := dueDate.Add(-time.Duration(rand.Intn(48)) * time.Hour)
					completedAt = &t
				}

				stepBatch = append(stepBatch,
					carePlanID*100+k,
					carePlanID,
					stepType,
					fmt.Sprintf("%s %d", stepType, k+1),
					fmt.Sprintf("Description for %s %d", stepType, k+1),
					status,
					dueDate,
					completedAt,
					k+1,
				)
			}

			query = fmt.Sprintf(`INSERT INTO care_plan_steps 
				(id, care_plan_id, type, title, description, status, due_date, completed_at, order_number) 
				VALUES %s`, strings.Join(stepValueStrings, ","))
			if _, err := db.ExecContext(ctx, query, stepBatch...); err != nil {
				log.Fatalf("Failed to insert care plan steps: %v", err)
			}
		}

		if (i+batchSize)%10000 == 0 {
			log.Printf("Processed %d users...", i+batchSize)
		}
	}

	// Generate read_articles
	log.Println("Generating read articles...")

	// Get all article IDs
	articleIDs := make([]int64, 0, articlesCount)
	rows, err := db.QueryContext(ctx, "SELECT id FROM articles")
	if err != nil {
		log.Fatalf("Failed to fetch article IDs: %v", err)
	}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			log.Fatalf("Failed to scan article ID: %v", err)
		}
		articleIDs = append(articleIDs, id)
	}
	rows.Close()

	// Each user reads between 5 and 20 random articles
	for i := 0; i < usersCount; i += batchSize {
		end := i + batchSize
		if end > usersCount {
			end = usersCount
		}

		// Process users in batches
		for j := i; j < end; j++ {
			userID := j // Using integer user ID

			// Randomly select number of articles this user has read
			readCount := 5 + rand.Intn(16) // 5 to 20 articles

			// Randomly select articles
			selectedArticles := make(map[int]bool)
			readBatch := make([]interface{}, 0, readCount*4) // user_id, article_id, read_at, is_saved
			readValueStrings := make([]string, 0, readCount)

			for k := 0; k < readCount; k++ {
				// Ensure unique articles for this user
				var articleIndex int
				for {
					articleIndex = rand.Intn(len(articleIDs))
					if !selectedArticles[articleIndex] {
						selectedArticles[articleIndex] = true
						break
					}
				}

				valueIdx := k * 4
				readValueStrings = append(readValueStrings,
					fmt.Sprintf("($%d, $%d, $%d, $%d)",
						valueIdx+1, valueIdx+2, valueIdx+3, valueIdx+4))

				readAt := time.Now().Add(-time.Duration(rand.Intn(90)) * 24 * time.Hour)
				isSaved := rand.Float32() < 0.3 // 30% chance of saving the article

				readBatch = append(readBatch,
					userID,
					articleIDs[articleIndex],
					readAt,
					isSaved,
				)
			}

			query := fmt.Sprintf("INSERT INTO read_articles (user_id, article_id, read_at, is_saved) VALUES %s",
				strings.Join(readValueStrings, ","))
			if _, err := db.ExecContext(ctx, query, readBatch...); err != nil {
				log.Fatalf("Failed to insert read articles: %v", err)
			}
		}

		if (i+batchSize)%10000 == 0 {
			log.Printf("Processed read_articles for %d users...", i+batchSize)
		}
	}

	log.Println("Data generation completed successfully")
}
