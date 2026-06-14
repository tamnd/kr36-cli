// Package kr36 is the library behind the kr36 command: the HTTP client,
// request shaping, and the typed data models for 36kr (36氪).
//
// The client fetches the public RSS feed at https://36kr.com/feed.
// No authentication is required. It sets a real User-Agent, paces requests,
// and retries transient 429/5xx errors with exponential back-off.
package kr36

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

// DefaultUserAgent identifies the client to 36kr.
const DefaultUserAgent = "Mozilla/5.0 (compatible; kr36/dev; +https://github.com/tamnd/kr36-cli)"

// Config holds constructor parameters.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://36kr.com",
		UserAgent: DefaultUserAgent,
		Rate:      500 * time.Millisecond,
		Timeout:   30 * time.Second,
		Retries:   3,
	}
}

// Client talks to the 36kr RSS feed.
type Client struct {
	cfg        Config
	httpClient *http.Client
	mu         sync.Mutex
	last       time.Time
}

// NewClient returns a Client with the given config.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:        cfg,
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}
}

// htmlTagRe strips HTML tags from description fields.
var htmlTagRe = regexp.MustCompile(`<[^>]+>`)

// stripHTML removes HTML tags and collapses whitespace.
func stripHTML(s string) string {
	s = htmlTagRe.ReplaceAllString(s, " ")
	// collapse runs of whitespace
	parts := strings.Fields(s)
	return strings.Join(parts, " ")
}

// parsePubDate parses the 36kr pubDate format and returns "YYYY-MM-DD".
// Falls back to the raw string if parsing fails.
func parsePubDate(s string) string {
	t, err := time.Parse("2006-01-02 15:04:05  -0700", s)
	if err != nil {
		// try with single space
		t, err = time.Parse("2006-01-02 15:04:05 -0700", s)
		if err != nil {
			return strings.TrimSpace(s)
		}
	}
	return t.Format("2006-01-02")
}

// News fetches the latest articles from the 36kr RSS feed.
// It returns at most limit items (all items if limit <= 0).
func (c *Client) News(ctx context.Context, limit int) ([]Article, error) {
	url := c.cfg.BaseURL + "/feed"
	raw, err := c.get(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("news: %w", err)
	}

	var root rssRoot
	if err := xml.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("news: parse rss: %w", err)
	}

	items := root.Channel.Items
	if limit > 0 && limit < len(items) {
		items = items[:limit]
	}

	out := make([]Article, 0, len(items))
	for i, it := range items {
		out = append(out, Article{
			Rank:    i + 1,
			Title:   strings.TrimSpace(it.Title),
			Summary: stripHTML(it.Description),
			PubDate: parsePubDate(it.PubDate),
			URL:     strings.TrimSpace(it.Link),
		})
	}
	return out, nil
}

// get issues a GET request with retry logic.
func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		b, retry, err := c.do(ctx, url)
		if err == nil {
			return b, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, url string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}

	b, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, true, err
	}
	return b, false, nil
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		d = 5 * time.Second
	}
	return d
}
