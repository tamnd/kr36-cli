package kr36

import (
	"testing"

	"github.com/tamnd/any-cli/kit"
)

// These tests are offline: they exercise the URI driver's pure string functions
// and the host wiring (mint, resolve), which need no network. The client's
// HTTP behaviour is covered in kr36_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "kr36" {
		t.Errorf("Scheme = %q, want kr36", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "kr36" {
		t.Errorf("Identity.Binary = %q, want kr36", info.Identity.Binary)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"1234567890", "article", "1234567890"},
		{"https://36kr.com/p/1234567890", "article", "1234567890"},
		{"https://36kr.com/p/1234567890?f=rss", "article", "1234567890"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestLocate(t *testing.T) {
	got, err := Domain{}.Locate("article", "1234567890")
	want := "https://36kr.com/p/1234567890"
	if err != nil || got != want {
		t.Errorf("Locate = (%q, %v), want (%q, nil)", got, err, want)
	}
}

// TestHostWiring mounts the driver in a kit Host and checks the round trip:
// a record mints to its URI and a bare id resolves back to the same URI.
func TestHostWiring(t *testing.T) {
	h, err := kit.Open()
	if err != nil {
		t.Fatal(err)
	}

	a := &Article{Rank: 1, ID: "1234567890", Title: "Test", URL: "https://36kr.com/p/1234567890"}
	u, err := h.Mint(a)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if want := "kr36://article/1234567890"; u.String() != want {
		t.Errorf("Mint = %q, want %q", u.String(), want)
	}

	got, err := h.ResolveOn("kr36", "1234567890")
	if err != nil || got.String() != "kr36://article/1234567890" {
		t.Errorf("ResolveOn = (%q, %v), want kr36://article/1234567890", got.String(), err)
	}
}
