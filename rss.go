package rss

import (
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strings"

	// "strconv"

	// "strings"
	// "time"

	postgres "project/pkg/dtbs"
)

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Chan    Channel  `xml:"channel"`
}

type Channel struct {
	XMLName     xml.Name `xml:"channel"`
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	Description string   `xml:"description"`
	Items       []Item   `xml:"item"`
}

type Item struct {
	Title       string `xml:"title"`
	URL         string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// ParseURL function reads an RSS feed and returns an array of decoded news.
func ParseURL(url string) ([]postgres.NewsItem, error) {

	var data []postgres.NewsItem

	//we call the GET method of HTTP on each url received from our configuration file
	//responses are read and the contents are placed into fields of news items
	response, err := http.Get(url)
	if err != nil {
		log.Fatalf("Could not get response because of %v./n", err)
	}
	respBody, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Could not read the response because of %v./n", err)
	}

	var feed RSS
	err = xml.Unmarshal(respBody, &feed)
	if err != nil {
		log.Fatalf("Could not process the feed because of %v./n", err)
	}

	var n postgres.NewsItem
	for _, item := range feed.Chan.Items {
		n.Title = item.Title
		n.Contents = item.Description
		n.URL = item.URL
		item.PubDate = strings.ReplaceAll(item.PubDate, ",", "")
		n.PublishedOn = string(item.PubDate)
		// fmt.Println(n.PublishedOn)
		data = append(data, n)
	}
	return data, nil
}
