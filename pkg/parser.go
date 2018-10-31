package crawler

import (
	"errors"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// FetchTimeout defines the max amount of time the parser will try to fetch a given page for.
const FetchTimeout = 5 * time.Second

var (
	// ErrExternalDomain is returned when the given URL redirects to a domain outside the starting domain
	ErrExternalDomain = errors.New("URL is outside the starting domain, ignoring")

	// ErrTooManyRedirects is returned after 10 consecutive redirects from a given URL
	ErrTooManyRedirects = errors.New("stopped after 10 redirects")
)

// Page defines the data structure representing a single web page.
// Addr is the full URL of the page with no query params or fragments.
// Links is a collection of links found on the page.
type Page struct {
	Addr  CanonicalURL
	Links Links
}

// Parser parses the DOM of a single web page.
type Parser struct {
	domainScheme string
	domainHost   string
}

// NewParser returns an instance of the Parser with all its required properties initialised.
// The given domain scheme and host values are used as the scheme and host values of any relative URLs found on the page.
func NewParser(domainScheme string, domainHost string) Parser {
	return Parser{domainScheme: domainScheme, domainHost: domainHost}
}

func (p *Parser) parse(u string) (Page, error) {
	var page Page
	var links []string
	var key CanonicalURL
	mLinks := make(map[string]bool)

	key = CanonicalURL(u)

	client := &http.Client{
		Timeout: FetchTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Host != p.domainHost {
				return ErrExternalDomain
			}

			if len(via) >= 10 {
				return ErrTooManyRedirects
			}

			newKey := *req.URL
			p.normalise(&newKey)
			key = CanonicalURL(newKey.String())

			return nil
		},
	}

	resp, err := client.Get(u)
	if err != nil {
		return page, err
	}
	defer resp.Body.Close()

	z := html.NewTokenizer(resp.Body)

	for {
		tt := z.Next()

		switch {
		case tt == html.ErrorToken:
			// End of the document, return results
			for link := range mLinks {
				links = append(links, link)
			}

			page = Page{Addr: key, Links: links}
			return page, nil
		case tt == html.StartTagToken:
			t := z.Token()

			isAnchor := t.Data == "a"
			if isAnchor {
				for _, a := range t.Attr {
					if a.Key == "href" {

						l, err := url.Parse(a.Val)
						if err != nil {
							log.Printf("failed to parse URL %s found on page %s, it will be ignored\n", a.Val, u)
							continue
						}

						p.normalise(l)

						if l.Host == p.domainHost && (l.Scheme == "http" || l.Scheme == "https") {
							key := l.String()

							if _, ok := mLinks[key]; !ok {
								mLinks[key] = true
							}
						}
						break
					}
				}
			}
		}
	}
}

// Normalise turns relative URLs into absolute by adding the starting page's scheme and domain.
// It also removes the trailing slash and any query params or fragments from the given URL.
func (p *Parser) normalise(u *url.URL) {
	if u.Host == "" {
		u.Host = p.domainHost
	}

	if u.Scheme == "" {
		u.Scheme = p.domainScheme
	}

	u.Path = strings.TrimRight(u.Path, "/")

	u.RawQuery = ""
	u.Fragment = ""
}
