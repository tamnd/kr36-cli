package kr36

import "encoding/xml"

// rssRoot is the top-level RSS envelope.
type rssRoot struct {
	XMLName xml.Name   `xml:"rss"`
	Channel rssChannel `xml:"channel"`
}

// rssChannel holds the feed metadata and item list.
type rssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Language    string    `xml:"language"`
	Items       []rssItem `xml:"item"`
}

// rssItem is one article entry from the RSS feed.
type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// Article is the public record type returned by Client.News.
type Article struct {
	Rank      int    `json:"rank"      table:"RANK"`
	Title     string `json:"title"     table:"TITLE"`
	Summary   string `json:"summary"   table:"SUMMARY"`
	Published string `json:"published" table:"DATE"`
	URL       string `json:"url"       table:"URL"`
}
