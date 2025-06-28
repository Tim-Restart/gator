package main

import (
	"fmt"
	"encoding/xml"
	"io"
	"net/http"
	"context"
	"html"
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