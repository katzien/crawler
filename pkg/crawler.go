package crawler

import (
	"context"
	"log"
	"net/url"
	"sync"
)

// Sitemap is the data structure holding current sitemap information.
// It's a map of a page URL to the links found on that page.
type Sitemap map[CanonicalURL]Links

// CanonicalURL represents the normalised page URL (a full URL with no query params or fragments).
type CanonicalURL string

// Links is a slice containing links found on a given page.
type Links []string

// Crawler is used to crawl a given starting URL, up to a max depth.
type Crawler struct {
	startURL     string
	maxDepth     int
	parser       Parser
	sitemap      Sitemap
	sMutex       sync.Mutex
	keepCrawling bool
}

// NewCrawler returns an instance of the Crawler with all its required properties initialised.
func NewCrawler(start *url.URL, depth int) Crawler {

	p := NewParser(start.Scheme, start.Host)

	p.normalise(start)

	return Crawler{
		startURL:     start.String(),
		maxDepth:     depth,
		parser:       p,
		sitemap:      make(Sitemap),
		sMutex:       sync.Mutex{},
		keepCrawling: true,
	}
}

// Crawl will start crawling the URL given to the Crawler as the starting URL.
// Once the maximum depth is reached or no new pages are found, a Sitemap struct will be returned with the results.
// Crawl accepts a cancellable context and stops crawling when the context is cancelled, returning the current results.
func (c *Crawler) Crawl(ctx context.Context) Sitemap {
	var sitemap Sitemap

	out := make(chan Sitemap, 1)

	go func() {
		defer close(out)
		c.parsePage(c.startURL, 0)
		out <- c.sitemap
	}()

	select {
	case <-ctx.Done():
		c.keepCrawling = false
		sitemap = <-out
	case sitemap = <-out:
	}

	return sitemap
}

func (c *Crawler) parsePage(l string, lvl int) {

	if c.keepCrawling && (c.maxDepth == 0 || lvl < c.maxDepth) && !c.known(CanonicalURL(l)) {
		page, err := c.parser.parse(l)
		if err != nil {
			log.Printf("parsing %s returned an error: %s", l, err.Error())
			return
		}

		c.add(page)

		for _, link := range page.Links {
			if !c.known(CanonicalURL(link)) {
				c.parsePage(link, lvl+1)
			}
		}
	}

	return
}

func (c *Crawler) add(p Page) {
	c.sMutex.Lock()
	c.sitemap[p.Addr] = p.Links
	c.sMutex.Unlock()
}

func (c *Crawler) known(u CanonicalURL) bool {
	c.sMutex.Lock()
	_, ok := c.sitemap[u]
	c.sMutex.Unlock()

	return ok
}
