package main

import (
	"net/http"
	"strconv"

	"github.com/akshatasingh1/rssagg/internal/database"
)

// Notice the custom signature: it takes a database.User just like your getPosts handler
func (apiCfg *apiConfig) handlerSearchPosts(w http.ResponseWriter, r *http.Request, user database.User) {
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		respondWithError(w, http.StatusBadRequest, "Missing search query parameter 'q'")
		return
	}

	// Default to 10 items per page
	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
		limit = parsedLimit
	}

	// Default to Page 1
	pageStr := r.URL.Query().Get("page")
	page := 1
	if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
		page = parsedPage
	}

	// Calculate the OFFSET (e.g., Page 2 with limit 10 means skip the first 10)
	offset := (page - 1) * limit

	posts, err := apiCfg.DB.SearchPostsForUser(r.Context(), database.SearchPostsForUserParams{
		UserID:         user.ID,
		PlaintoTsquery: searchTerm,
		Limit:          int32(limit),
		Offset:         int32(offset), // Pass the calculated offset to the database
	})
	
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't search posts")
		return
	}

	respondWithJSON(w, http.StatusOK, databasePostsToPosts(posts))
}