package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// generateSummary takes the raw RSS description and asks Gemini to summarize it
func generateSummary(text string) (string, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("GEMINI_API_KEY not found in environment")
	}

	// If the post has no description, we can't summarize it
	if text == "" {
		return "", nil
	}

	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=" + apiKey

	// We use a map to build the exact JSON structure that Google's API requires
	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": "Summarize this RSS post in exactly two short sentences. Do not include any intro or outro text. Here is the post: " + text},
				},
			},
		},
	}

	jsonReq, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	// Send the request to Gemini
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonReq))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// NEW: Check if Google sent back an HTTP error (like 429 Rate Limit)
	if resp.StatusCode != http.StatusOK {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		return "", fmt.Errorf("gemini API rejected request with status %d: %v", resp.StatusCode, errorResponse)
	}

	// Create a temporary struct to decode Gemini's nested JSON response
	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return "", err
	}

	// Extract the text if it was successfully generated
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		return geminiResp.Candidates[0].Content.Parts[0].Text, nil
	}

	return "", fmt.Errorf("no summary generated")
	
}