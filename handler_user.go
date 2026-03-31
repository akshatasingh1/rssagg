package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/akshatasingh1/rssagg/internal/database"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func (apiCfg *apiConfig) handlerCreateUser(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Name string `json:"name"`
	}
	decoder := json.NewDecoder(r.Body)

	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Error parsing JSON: %v", err))
		return
	}

	user, err := apiCfg.DB.CreateUser(r.Context(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      params.Name,
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Failed to create user: %v", err))
		return
	}

	respondWithJSON(w, 201, databaseUsertoUser(user))
	log.Printf("User created with ID: %v", user.ID)
}

func (apiCfg *apiConfig) handlerGetUser(w http.ResponseWriter, r *http.Request, user database.User) {
	respondWithJSON(w, 200, databaseUsertoUser(user))
}

// UPDATED: Now with Redis Read-Through Caching!
func (apiCfg *apiConfig) handlerGetPostsForUser(w http.ResponseWriter, r *http.Request, user database.User) {
	limit := 10

	// 1. Create a unique Cache Key for this user
	cacheKey := fmt.Sprintf("posts:user:%s:limit:%d", user.ID.String(), limit)

	// 2. THE CACHE HIT: Check Redis first
	cachedData, err := apiCfg.RedisClient.Get(r.Context(), cacheKey).Result()
	if err == nil {
		// Found it! Return instantly.
		log.Println("Cache HIT! Returning posts from Redis.")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(cachedData))
		return
	} else if err != redis.Nil {
		// Log actual Redis errors, but don't break the app.
		log.Printf("Redis error: %v", err)
	}

	// 3. THE CACHE MISS: Not in Redis, go to PostgreSQL
	log.Println("Cache MISS! Fetching from PostgreSQL.")
	posts, err := apiCfg.DB.GetPostsForUser(r.Context(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		respondWithError(w, 400, fmt.Sprintf("Failed to get posts for user: %v", err))
		return
	}

	// Convert DB models to JSON response format
	postsResponse := databasePostsToPosts(posts)

	// 4. Save to Redis for next time! (60 second Time-To-Live)
	postsJSON, marshalErr := json.Marshal(postsResponse)
	if marshalErr == nil {
		apiCfg.RedisClient.Set(r.Context(), cacheKey, postsJSON, 60*time.Second)
	}

	// 5. Send response to user
	respondWithJSON(w, 200, postsResponse)
}