package main

import (
	"context"
	"database/sql"
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
		feeds,err := db.GetNextFeedsToFetch(
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
			go scrapeFeed(db,wg,feed)
		}
		wg.Wait()

	}
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
		description := sql.NullString{}
		if item.Description != "" {
			description.String = item.Description
			description.Valid = true
		}
		
		pubAt, err := time.Parse(time.RFC1123Z, item.PubDate)
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
				Url:         item.Link,
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