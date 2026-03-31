package main

import (
	"context"
	"database/sql"
	"log"
	"time"

	"github.com/akshatasingh1/rssagg/internal/database"
)

func startAIWorker(db *database.Queries, timeBetweenRequests time.Duration) {
	log.Printf("AI Worker started! Checking for blank summaries every %v", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)

	// Infinite background loop
	for ; ; <-ticker.C {
		// 1. Ask the database for ONE post that has a NULL summary
		post, err := db.GetNextPostToSummarize(context.Background())
		if err != nil {
			// sql.ErrNoRows means there are no blank posts. The database is fully caught up!
			if err == sql.ErrNoRows {
				continue
			}
			log.Printf("AI Worker DB error: %v", err)
			continue
		}

		log.Printf("AI Worker: Found post '%s'. Generating summary...", post.Title)

		// 2. Make sure the post actually has a description to summarize
		var textToSummarize string
		if post.Description.Valid && post.Description.String != "" {
			textToSummarize = post.Description.String
		} else {
			// If there is no text, fill it with a placeholder so it doesn't get stuck in a loop
			db.UpdatePostSummary(context.Background(), database.UpdatePostSummaryParams{
				ID:      post.ID,
				Summary: sql.NullString{String: "No content available to summarize.", Valid: true},
			})
			continue
		}

		// 3. Call the real Gemini API
		summaryText, err := generateSummary(textToSummarize)
		if err != nil {
			log.Printf("AI Worker failed: %v", err)
			// If it fails (like a Rate Limit), we DO NOT update the database.
			// This means on the next tick, it will grab this exact same post and try again!
			continue
		}

		// 4. Save the successful AI summary to the database
		err = db.UpdatePostSummary(context.Background(), database.UpdatePostSummaryParams{
			ID:      post.ID,
			Summary: sql.NullString{String: summaryText, Valid: true},
		})

		if err != nil {
			log.Printf("AI Worker failed to save to DB: %v", err)
			continue
		}

		log.Printf("AI Worker successfully saved summary!")
	}
}