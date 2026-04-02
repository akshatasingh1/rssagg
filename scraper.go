package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"sync"

	"github.com/akshatasingh1/rssagg/internal/database"
	"github.com/google/uuid"
)

func startScraping(
	db *database.Queries,
	concurrency int,
	timeBetweenRequest time.Duration,
) {
	log.Printf("Scraping on %v goroutines every %s duration", concurrency, timeBetweenRequest)
	ticker := time.NewTicker(timeBetweenRequest)
	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(
			context.Background(),
			int32(concurrency),
		)
		if err != nil {
			log.Println("Error fetching feeds:", err)
			continue
		}
		wg := &sync.WaitGroup{}
		for _, feed := range feeds {
			wg.Add(1)
			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()

	}
}

func parseDate(raw string) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}, fmt.Errorf("empty publication date")
	}

	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC850,
		time.ANSIC,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		time.RFC3339Nano,
		"Mon, 02 Jan 2006 15:04:05 -0700",
		"02 Jan 2006 15:04:05 MST",
	}

	for _, layout := range layouts {
		if t, err := time.Parse(layout, raw); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unrecognized date format: %q", raw)
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	// 1. Fetch the feed, passing in our saved headers from the database
	rssFeed, newLastModified, newEtag, err := urlToFeed(
		feed.Url,
		feed.LastModified.String,
		feed.Etag.String,
	)

	// 2. Error Handling & The ETag Magic
	if err != nil {
		if err == ErrNotModified {
			// THE MAGIC: It hasn't changed!
			log.Printf("Feed %s has not changed since last fetch. Skipping.", feed.Name)

			// We still mark it as fetched with its OLD headers so it goes to the back of the queue
			db.MarkFeedAsFetched(context.Background(), database.MarkFeedAsFetchedParams{
				ID:           feed.ID,
				LastModified: feed.LastModified,
				Etag:         feed.Etag,
			})
			return
		}

		// If it's a real error (like the website being down), log it and exit
		log.Printf("Error fetching feed: %v", err)
		return
	}

	// 3. If we reach here, the feed DID change. Let's mark it as fetched and save the NEW headers.
	_, err = db.MarkFeedAsFetched(context.Background(), database.MarkFeedAsFetchedParams{
		ID: feed.ID,
		LastModified: sql.NullString{
			String: newLastModified,
			Valid:  newLastModified != "",
		},
		Etag: sql.NullString{
			String: newEtag,
			Valid:  newEtag != "",
		},
	})
	if err != nil {
		log.Printf("Error marking feed as fetched: %v\n", err)
		return
	}

	// 4. Parse and save the actual posts (This part remains exactly the same!)
	for _, item := range rssFeed.Channel.Item {

		validURL := item.Link
		if validURL == "" {
			validURL = item.GUID // If link is broken, use GUID!
		}

		// If it's STILL blank, safely skip it so we don't poison the database
		if validURL == "" {
			log.Printf("Skipping article with missing URL: %s", item.Title)
			continue
		}

		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}

		// Use a flexible date parser to handle different blog formats
		pubAt, err := parseDate(item.PubDate)
		if err != nil {
			log.Printf("Couldn't parse time %v: %v", item.PubDate, err)
			continue
		}

		emptySummary := sql.NullString{
			String: "",
			Valid:  false, // This tells PostgreSQL to save it as NULL
		}

		_, err = db.CreatePost(context.Background(),
			database.CreatePostParams{
				ID:          uuid.New(),
				CreatedAt:   time.Now().UTC(),
				UpdatedAt:   time.Now().UTC(),
				Title:       item.Title,
				Description: description,
				Summary:     emptySummary,
				PublishedAt: pubAt,
				Url:         validURL, // <--- FIXED: Now using the correct variable!
				FeedID:      feed.ID,
			})

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Printf("Failed to create post:%v", err)
			continue
		}
	}
	log.Printf("Feed %s collected, %v posts found", feed.Name, len(rssFeed.Channel.Item))
}
