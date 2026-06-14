package kr36_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/tamnd/kr36-cli/kr36"
)

const mockRSS = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
<channel>
<title>36氪</title>
<link>http://36kr.com</link>
<item>
<title>曼联，要被卖了</title>
<link><![CDATA[https://36kr.com/p/3200000001]]></link>
<pubDate>2026-06-13 16:37:21  +0800</pubDate>
<description><![CDATA[<p>曼联俱乐部<b>正式宣布</b>出售，估值超过50亿英镑。</p>]]></description>
</item>
<item>
<title>OpenAI再融资</title>
<link><![CDATA[https://36kr.com/p/3200000002]]></link>
<pubDate>2026-06-12 10:00:00  +0800</pubDate>
<description><![CDATA[OpenAI完成新一轮融资，估值超过3000亿美元。]]></description>
</item>
<item>
<title>字节跳动出海</title>
<link><![CDATA[https://36kr.com/p/3200000003]]></link>
<pubDate>2026-06-11 09:00:00  +0800</pubDate>
<description><![CDATA[字节跳动加速<em>全球化</em>布局。]]></description>
</item>
<item>
<title>华为最新发布</title>
<link><![CDATA[https://36kr.com/p/3200000004]]></link>
<pubDate>2026-06-10 08:00:00  +0800</pubDate>
<description><![CDATA[华为发布最新旗舰手机。]]></description>
</item>
<item>
<title>比亚迪销量新高</title>
<link><![CDATA[https://36kr.com/p/3200000005]]></link>
<pubDate>2026-06-09 07:00:00  +0800</pubDate>
<description><![CDATA[比亚迪创下单月销量新纪录。]]></description>
</item>
</channel>
</rss>`

func newTestClient(ts *httptest.Server) *kr36.Client {
	cfg := kr36.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return kr36.NewClient(cfg)
}

func TestNewsSendsUserAgent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("User-Agent") == "" {
			t.Error("request carried no User-Agent")
		}
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	_, err := c.News(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
}

func TestNewsParsesItems(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	arts, err := c.News(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 5 {
		t.Fatalf("got %d articles, want 5", len(arts))
	}

	a := arts[0]
	if a.Rank != 1 {
		t.Errorf("rank = %d, want 1", a.Rank)
	}
	if a.Title != "曼联，要被卖了" {
		t.Errorf("title = %q", a.Title)
	}
	if a.URL != "https://36kr.com/p/3200000001" {
		t.Errorf("url = %q", a.URL)
	}
	if a.PubDate != "2026-06-13" {
		t.Errorf("pub_date = %q, want 2026-06-13", a.PubDate)
	}
}

func TestNewsLimitRespected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	arts, err := c.News(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 3 {
		t.Fatalf("got %d articles, want 3", len(arts))
	}
}

func TestNewsRetriesOn503(t *testing.T) {
	var hits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte(mockRSS))
	}))
	defer srv.Close()

	cfg := kr36.DefaultConfig()
	cfg.BaseURL = srv.URL
	cfg.Rate = 0
	cfg.Retries = 5
	c := kr36.NewClient(cfg)

	start := time.Now()
	_, err := c.News(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
	if time.Since(start) < 500*time.Millisecond {
		t.Error("retries did not back off")
	}
}

func TestNewsHTMLStripped(t *testing.T) {
	body := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0"><channel><title>36氪</title><link>http://36kr.com</link>
<item>
<title>Test</title>
<link><![CDATA[https://36kr.com/p/test]]></link>
<pubDate>2026-06-14 10:00:00  +0800</pubDate>
<description><![CDATA[<b>bold</b> text]]></description>
</item>
</channel></rss>`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	c := newTestClient(srv)
	arts, err := c.News(context.Background(), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 1 {
		t.Fatalf("got %d articles, want 1", len(arts))
	}
	if arts[0].Summary != "bold text" {
		t.Errorf("summary = %q, want %q", arts[0].Summary, "bold text")
	}
}
