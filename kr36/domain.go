package kr36

import (
	"context"
	"regexp"
	"strings"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes kr36 as a kit Domain so a multi-domain host can
// blank-import it:
//
//	import _ "github.com/tamnd/kr36-cli/kr36"
//
// The same Domain also builds the standalone kr36 binary (cmd/kr36).
func init() { kit.Register(Domain{}) }

// Domain is the kr36 driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "kr36",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "kr36",
			Short:  "A command line for 36Kr tech news.",
			Long: `A command line for 36Kr tech news.

kr36 reads public 36Kr data over plain HTTPS, shapes it into clean records,
and prints output that pipes into the rest of your tools. No API key, nothing
to run alongside it.`,
			Site: Host,
			Repo: "https://github.com/tamnd/kr36-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// article: resolver op so that kit can mint Article records and answer
	// `kr36 article <id>` and `ant get kr36://article/<id>`.
	kit.Handle(app, kit.OpMeta{Name: "article", Group: "read", Single: true,
		URIType: "article", Resolver: true,
		Summary: "Resolve a 36Kr article ID to its URL",
		Args:    []kit.Arg{{Name: "id", Help: "numeric article ID"}}}, getArticle)

	// news: latest articles from the 36Kr RSS feed.
	kit.Handle(app, kit.OpMeta{Name: "news", Group: "read", List: true,
		URIType: "article",
		Summary: "List the latest 36Kr articles from the RSS feed"}, getNews)
}

// newClient builds the HTTP client from the kit host config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type articleInput struct {
	ID     string  `kit:"arg"   help:"numeric article ID"`
	Client *Client `kit:"inject"`
}

type newsInput struct {
	Limit  int     `kit:"flag,inherit" help:"max results"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func getArticle(_ context.Context, in articleInput, emit func(*Article) error) error {
	id := strings.TrimSpace(in.ID)
	if id == "" {
		return errs.Usage("article id is required")
	}
	return emit(&Article{
		ID:  id,
		URL: "https://" + Host + "/p/" + id,
	})
}

func getNews(ctx context.Context, in newsInput, emit func(*Article) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 20
	}
	articles, err := in.Client.News(ctx, limit)
	if err != nil {
		return err
	}
	for i := range articles {
		if err := emit(&articles[i]); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver ---

// articleURLRE matches a 36kr article URL and captures the numeric ID.
var articleURLRE = regexp.MustCompile(`(?:https?://)?36kr\.com/p/(\d+)`)

// Classify turns a 36kr URL or numeric article ID into (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", "", errs.Usage("empty 36kr reference")
	}
	if m := articleURLRE.FindStringSubmatch(input); len(m) >= 2 {
		return "article", m[1], nil
	}
	if isDigits(input) {
		return "article", input, nil
	}
	return "", "", errs.Usage("unrecognized 36kr reference: %q", input)
}

// Locate is the inverse: the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	if uriType != "article" {
		return "", errs.Usage("kr36 has no resource type %q", uriType)
	}
	return "https://" + Host + "/p/" + id, nil
}

// --- helpers ---

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}
