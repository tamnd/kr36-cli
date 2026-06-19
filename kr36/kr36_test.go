package kr36_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tamnd/kr36-cli/kr36"
)

const mockRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>36Kr</title>
    <description>36氪最新文章</description>
    <item>
      <title>OpenAI raises $6.6 billion</title>
      <link><![CDATA[https://36kr.com/p/1234567890]]></link>
      <guid>https://36kr.com/p/1234567890</guid>
      <pubDate>2024-03-15 10:30:00  +0800</pubDate>
      <description><![CDATA[<p><b>OpenAI</b> announced a <em>record</em> funding round today.</p>]]></description>
    </item>
    <item>
      <title>BYD surpasses Tesla in Q1 deliveries</title>
      <link><![CDATA[https://36kr.com/p/9876543210]]></link>
      <guid>https://36kr.com/p/9876543210</guid>
      <pubDate>2024-03-14 08:00:00  +0800</pubDate>
      <description><![CDATA[<p>BYD delivered 300,000 vehicles in Q1 2024.</p>]]></description>
    </item>
    <item>
      <title>Third article title</title>
      <link><![CDATA[https://36kr.com/p/1111111111]]></link>
      <guid>https://36kr.com/p/1111111111</guid>
      <pubDate>2024-03-13 12:00:00  +0800</pubDate>
      <description><![CDATA[Third article summary text.]]></description>
    </item>
  </channel>
</rss>`

const emptyRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>36Kr</title>
    <description>36氪最新文章</description>
  </channel>
</rss>`

func newTestClient(ts *httptest.Server) *kr36.Client {
	return kr36.NewClient(kr36.Config{
		BaseURL:   ts.URL,
		UserAgent: "test-agent/1.0",
		Rate:      0,
		Timeout:   5 * time.Second,
		Retries:   0,
	})
}

func TestNewsParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer ts.Close()
	c := newTestClient(ts)
	articles, err := c.News(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 3 {
		t.Fatalf("want 3 articles, got %d", len(articles))
	}
	a := articles[0]
	if a.Rank != 1 {
		t.Errorf("rank: want 1, got %d", a.Rank)
	}
	if a.Title != "OpenAI raises $6.6 billion" {
		t.Errorf("title: got %q", a.Title)
	}
	if a.URL != "https://36kr.com/p/1234567890" {
		t.Errorf("url: got %q", a.URL)
	}
	if a.Published != "2024-03-15" {
		t.Errorf("published: got %q, want 2024-03-15", a.Published)
	}
}

func TestNewsLimitRespected(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer ts.Close()
	c := newTestClient(ts)
	articles, err := c.News(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 2 {
		t.Fatalf("want 2 articles, got %d", len(articles))
	}
	if articles[1].Rank != 2 {
		t.Errorf("articles[1].Rank = %d, want 2", articles[1].Rank)
	}
}

func TestNewsHTMLStripped(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer ts.Close()
	c := newTestClient(ts)
	articles, err := c.News(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	summary := articles[0].Summary
	if strings.Contains(summary, "<") || strings.Contains(summary, ">") {
		t.Errorf("HTML not stripped from summary: %q", summary)
	}
	if !strings.Contains(summary, "OpenAI") {
		t.Errorf("expected text content in summary, got: %q", summary)
	}
}

func TestNewsSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer ts.Close()
	c := newTestClient(ts)
	_, err := c.News(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent header not sent")
	}
}

func TestNewsRetriesOn503(t *testing.T) {
	var calls int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(503)
			return
		}
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer ts.Close()
	c := kr36.NewClient(kr36.Config{
		BaseURL: ts.URL, UserAgent: "test", Rate: 0, Timeout: 5 * time.Second, Retries: 3,
	})
	articles, err := c.News(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) == 0 {
		t.Error("expected articles after retries")
	}
	if atomic.LoadInt32(&calls) < 3 {
		t.Errorf("expected at least 3 calls, got %d", calls)
	}
}

func TestNewsRanksAssigned(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer ts.Close()
	c := newTestClient(ts)
	articles, err := c.News(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	for i, a := range articles {
		if a.Rank != i+1 {
			t.Errorf("articles[%d].Rank = %d, want %d", i, a.Rank, i+1)
		}
	}
}

func TestNewsEmptyFeed(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(emptyRSS))
	}))
	defer ts.Close()
	c := newTestClient(ts)
	articles, err := c.News(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 0 {
		t.Fatalf("want 0 articles, got %d", len(articles))
	}
}

func TestNewsContextCancelled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer ts.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	c := newTestClient(ts)
	_, err := c.News(ctx, 0)
	if err == nil {
		t.Fatal("expected error when context is cancelled, got nil")
	}
}

func TestNewsBadXML(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"not":"xml"}`))
	}))
	defer ts.Close()
	c := newTestClient(ts)
	_, err := c.News(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error for bad XML, got nil")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("error should mention parse failure, got: %v", err)
	}
}

func TestNewsLimitZero(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer ts.Close()
	c := newTestClient(ts)
	articles, err := c.News(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(articles) != 3 {
		t.Fatalf("limit 0 should return all 3 articles, got %d", len(articles))
	}
}
