package main

import (
	"encoding/json"
	"fmt"	
	"net/http"
	"time"
	"strings"
	//"github.com/akshatasingh1/rssagg/internal/auth"
	"github.com/akshatasingh1/rssagg/internal/database"
	"github.com/google/uuid"
	"log"


)

func (apiCfg *apiConfig) handlerCreateFeed(w http.ResponseWriter, r *http.Request, user database.User) {
	type parameters struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	// 1. ATTEMPT TO CREATE THE FEED
	feed, err := apiCfg.DB.CreateFeed(r.Context(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
		Url:       params.URL,
		UserID:    user.ID,
	})

	// 2. THE SMART FALLBACK: If creation fails, check if it already exists
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") || strings.Contains(err.Error(), "unique constraint") {
			log.Printf("Feed URL %s already exists. Subscribing user to existing feed...", params.URL)

			// Fetch the existing feed from the global directory
			existingFeed, dbErr := apiCfg.DB.GetFeedByURL(r.Context(), params.URL)
			if dbErr != nil {
				respondWithError(w, 400, fmt.Sprintf("Feed exists but couldn't fetch it: %v", dbErr))
				return
			}
			// Swap out the failed 'feed' variable with the successful existing one!
			feed = existingFeed
		} else {
			// If it's a real error (not a duplicate), fail normally
			respondWithError(w, 400, fmt.Sprintf("Failed to create feed: %v", err))
			return
		}
	}

	// 3. CREATE THE SUBSCRIPTION (Feed Follow)
	_, err = apiCfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID, // This now safely points to either the new OR existing feed
	})
	
	// Safety net: If they try to add a feed they already follow, just ignore the error
	if err != nil && !strings.Contains(err.Error(), "duplicate key") {
		respondWithError(w, 500, fmt.Sprintf("Failed to create feed follow: %v", err))
		return
	}
	cacheKey := fmt.Sprintf("posts:user:%s:page:1:v2", user.ID.String())
	err = apiCfg.RedisClient.Del(r.Context(), cacheKey).Err()
	if err != nil {
		log.Printf("Failed to invalidate cache for user %s: %v", user.ID.String(), err)
		// We don't return an error here, because the feed was still successfully added!
	}

	// 4. Respond with the successfully created (or fetched) feed
	respondWithJSON(w, 201, databaseFeedToFeed(feed))
}

func (apiCfg *apiConfig) handlerGetFeeds(w http.ResponseWriter, r *http.Request) {
	feeds, err := apiCfg.DB.GetFeeds(r.Context())
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Failed to get feeds: %v", err))
		return
	}

	respondWithJSON(w, 200, databaseFeedsToFeeds(feeds))
}

	