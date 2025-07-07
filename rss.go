package main

import (
	"fmt"
	"encoding/xml"
	"io"
	"net/http"
	"context"
	"html"
	"time"
	"github.com/google/uuid"
	"gator/internal/database"
	"database/sql"
	"log"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	// Fetch feed time!

	// NewRequestWithContex - prepares the request to send with clientDo
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		fmt.Println("Error doing New Request With Context")
		return &RSSFeed{}, err
	}

	// something about setting the header to gator
	request.Header.Set("User-Agent", "Gator")
	// Client DO sends a HTTP request and returns a HTTP response

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println("Error sending Client Do request/response")
		return &RSSFeed{}, err
	}

	// Read the response from *http.Response 
	f, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body")
		return &RSSFeed{}, err
	}

	feed, err := xmlUnmarshall(f)
	if err != nil {
		fmt.Println("Error unmarshalling XML response")
		return &RSSFeed{}, err
	}
		
	// Unescape the titles and descriptions here
	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)
	for i := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(feed.Channel.Item[i].Title)
		feed.Channel.Item[i].Description = html.UnescapeString(feed.Channel.Item[i].Description)
	}
	
	return feed, nil
	

}


// XML unmarshalling function here

func xmlUnmarshall(xmlItem []byte) (*RSSFeed, error) {
	// Function to do the xmlUnmarshalling
	// Returns a RSSFeed pointer?
	feed := &RSSFeed{}
	err := xml.Unmarshal(xmlItem, feed)
	if err != nil {
		fmt.Println("Error unmarshalling XML feed")
		return &RSSFeed{}, err
	}
	return feed, nil
}


func scrapeFeeds(s *state) error {
	ctx := context.Background()
	
	feedID, err := s.db.GetNextFeedToFetch(ctx)
	if err !=nil {
		if err.Error() == "sql: no rows in result set" {
			// No users found create a new record
			fmt.Println("No feeds found being followed, grow your user base")
			return nil
		} else {
			fmt.Println("Error getting next Feed Details")
			return err
		}
	}



	feedInfo, err := s.db.GetFeedURLfromID(ctx, feedID)
		if err != nil {
			if err.Error() == "sql: no rows in result set" {
					// No users found create a new record
					fmt.Println("No feeds found with given ID")
					return nil
			} else {
				fmt.Println("Error getting feed information")
				return err
			}
		}



	markReturns := database.MarkFeedFetchedParams{
		UpdatedAt: time.Now(),
		LastFetchedAt: sql.NullTime{
			Time: time.Now(),
			Valid: true,
		},
		ID: feedID,
	}

	err = s.db.MarkFeedFetched(ctx, markReturns)
	if err != nil {
		fmt.Println("Error marking feed as fetched")
		return err
	}



	feed, err := s.db.GetFeedUrl(ctx, feedInfo.Url)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			// No users found create a new record
			fmt.Println("URL not found")
			return nil
		} else {
			fmt.Println("Error getting Feed Details")
			return err
		}
	}



	if feed.Url == "" {
		fmt.Print("Retrieved feed was nil or empty, skipping...")
		return nil
	}

	response, err := fetchFeed(ctx, feed.Url)
	if err != nil {
		fmt.Println("Error fetching feed")
		return err
	}
	
	for i := range response.Channel.Item {
		fmt.Printf("Title: %v : %v\n", response.Channel.Title, response.Channel.Item[i].Title)
	}

	

	for i := range response.Channel.Item {

		newPost := database.CreatePostsParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Title: response.Channel.Item[i].Title,
		Url: feed.Url,
		FeedID: feedID,
		}
	
		// Handle description conditionally
		if response.Channel.Item[i].Description != "" {
			newPost.Description = sql.NullString{String: response.Channel.Item[i].Description, Valid: true}
		} else {
			newPost.Description = sql.NullString{Valid: false}
		}
		// Check and populate the published date
		if response.Channel.Item[i].PubDate != "" {
			pubTime, err := time.Parse(time.RFC822, response.Channel.Item[i].PubDate)
			if err != nil {
			fmt.Println("Error parsing time, published time will be set to null")
			newPost.PublishedAt = sql.NullTime{Valid: false}
			} else {
    			newPost.PublishedAt = sql.NullTime{Time: pubTime, Valid: true}
			}
	} 

		err := s.db.CreatePosts(ctx, newPost)
		if err != nil {
			log.Printf("Error returned from CreatePosts: %v\n", err)
			return err
		}
		fmt.Printf("New post saved to database\n")

	}
	
	return nil

}


