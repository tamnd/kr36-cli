package kr36

// rssRoot is the top-level RSS envelope.
type rssRoot struct {
	Channel struct {
		Items []rssItem `xml:"item"`
	} `xml:"channel"`
}

// rssItem is one article entry from the RSS feed.
type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// Article is the public record type returned by Client.News.
type Article struct {
	Rank    int    `json:"rank"               table:"rank"`
	ID      string `json:"id"      kit:"id"   table:"id"`
	Title   string `json:"title"              table:"title"`
	Summary string `json:"summary,omitempty"  table:"-"`
	PubDate string `json:"pub_date,omitempty" table:"pub_date"`
	URL     string `json:"url"                table:"url,url"`
}
