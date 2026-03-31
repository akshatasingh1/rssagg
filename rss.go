package main

import (
	"encoding/xml"
	"errors" // NEW: We need this to create our custom 304 error
	"io"
	"net/http"
	"time"
)

// NEW: A custom error so the scraper knows when to safely skip the database insertion
var ErrNotModified = errors.New("feed not modified")

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// UPDATED: Now takes the old headers as arguments, and returns the new ones!
func urlToFeed(url string, lastModified string, etag string) (RSSFeed, string, string, error) {
	// 1. Upgrade to http.NewRequest so we can attach custom headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return RSSFeed{}, "", "", err
	}

	// 2. Be polite: send the headers if we have them saved from a previous scrape
	if lastModified != "" {
		req.Header.Set("If-Modified-Since", lastModified)
	}
	if etag != "" {
		req.Header.Set("If-None-Match", etag)
	}

	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}

	// 3. Execute the request using .Do() instead of .Get()
	resp, err := httpClient.Do(req)
	if err != nil {
		return RSSFeed{}, "", "", err
	}
	defer resp.Body.Close()
	// 4. Check for a 304 Not Modified status code. If we get this, it means the feed hasn't changed since our last scrape!
	if resp.StatusCode == http.StatusNotModified {
		// Stop immediately! Don't read the body, don't parse XML.
		return RSSFeed{}, "", "", ErrNotModified 
	}

	// 5. If we get here, the feed HAS changed (or it's our first time fetching it)
	dat, err := io.ReadAll(resp.Body)
	if err != nil {
		return RSSFeed{}, "", "", err
	}
	
	rssFeed := RSSFeed{}
	err = xml.Unmarshal(dat, &rssFeed)
	if err != nil {
		return RSSFeed{}, "", "", err
	}

	// 6. Grab the NEW headers from the server's response to save to our database
	newLastModified := resp.Header.Get("Last-Modified")
	newEtag := resp.Header.Get("ETag")

	return rssFeed, newLastModified, newEtag, nil
}