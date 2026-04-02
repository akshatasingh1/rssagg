package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	//"github.com/akshatasingh1/rssagg/internal/auth"
	"github.com/akshatasingh1/rssagg/internal/database"
	"github.com/go-chi/chi"
	"github.com/google/uuid"
)

func (apiCfg *apiConfig) handlerCreateFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {

	type parameters struct {
		FeedID uuid.UUID `json:"feed_id"`
	}
	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	feedFollow, err := apiCfg.DB.CreateFeedFollow(r.Context(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    params.FeedID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Failed to create feed follow: %v", err))
		return
	}

	respondWithJSON(w, 201, databaseFeedFollowToFeedFollow(feedFollow))
}

func (apiCfg *apiConfig) handlerGetFeedFollows(w http.ResponseWriter, r *http.Request, user database.User) {

	feedFollows, err := apiCfg.DB.GetFeedFollows(r.Context(), user.ID)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Failed to get feed follows: %v", err))
		return
	}

	respondWithJSON(w, 200, databaseFeedFollowsToFeedFollows(feedFollows))
}

func (apiCfg *apiConfig) handlerDeleteFeedFollow(w http.ResponseWriter, r *http.Request, user database.User) {
	// 1. Read 'feedID' from the URL parameter
	feedIDStr := chi.URLParam(r, "feedID")
	feedID, err := uuid.Parse(feedIDStr)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Could not parse feed ID: %v", err))
		return
	}

	// 2. Use the new SQL function to delete by FeedID
	err = apiCfg.DB.DeleteFeedFollowByFeedId(r.Context(), database.DeleteFeedFollowByFeedIdParams{
		FeedID: feedID,
		UserID: user.ID,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Failed to delete feed follow: %v", err))
		return
	}

	// 3. CACHE INVALIDATION: Nuke the stale dashboard cache!
	// This forces React to fetch a fresh, empty list on the next load.
	cacheKey := fmt.Sprintf("posts:user:%s:page:1:v2", user.ID.String())
	apiCfg.RedisClient.Del(r.Context(), cacheKey)

	respondWithJSON(w, 200, struct{}{})
}
