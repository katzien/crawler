package crawler

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCrawl(t *testing.T) {
	rawhtml, err := ioutil.ReadFile("../fixtures/simple.html")
	if err != nil {
		t.Fatalf("failed to parse the %s fixture file: %s", "simple.html", err.Error())
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(rawhtml))
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Crawl(): failed to parse test server addr %s as URL", ts.URL)
	}

	c := NewCrawler(tsURL, 1)

	actual := c.Crawl(context.TODO())

	if len(actual) != 1 {
		t.Errorf("Crawl(): expected 1 page in sitemap, got %d", len(actual))
		t.FailNow()
	}

	for key, links := range actual {
		if key != CanonicalURL(ts.URL) {
			t.Errorf("Crawl(): expected page in sitemap to be %s, got %s", CanonicalURL(ts.URL), key)
		}

		for _, ll := range links {
			if ll != ts.URL+"/foo/bar" {
				t.Errorf("Crawl(): expected page link in sitemap to be %s, got %s", ts.URL+"/foo/bar", ll)
			}
		}
	}
}

func TestCrawlRespectsContextCancellation(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("Crawl(cancelled ctx): didn't expect the server to be called")
		t.FailNow()
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("Crawl(cancelled ctx): failed to parse test server addr %s as URL", ts.URL)
	}

	c := NewCrawler(tsURL, 1)

	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	actual := c.Crawl(ctx)

	if c.keepCrawling {
		t.Error("Crawl(cancelled ctx): expected c.keepCrawling to be false, got true")
		t.FailNow()
	}

	if len(actual) != 0 {
		t.Errorf("Crawl(): expected the sitemap to be empty, got %v", actual)
		t.FailNow()
	}
}

func TestParsePage(t *testing.T) {
	rawhtml, err := ioutil.ReadFile("../fixtures/simple.html")
	if err != nil {
		t.Fatalf("failed to parse the %s fixture file: %s", "simple.html", err.Error())
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(rawhtml))
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parsePage(): failed to parse test server addr %s as URL", ts.URL)
	}

	c := NewCrawler(tsURL, 1)

	c.parsePage(ts.URL, 0)

	expected := Links{
		"https://local.com/foo/bar",
	}

	actual := c.sitemap[CanonicalURL(ts.URL)]

	if len(actual) != len(expected) {
		t.Errorf("parsePage(): expected %d links in sitemap, got %d", len(expected), len(actual))
		t.Errorf("expected: %v", expected)
		t.Errorf("actual: %v", actual)
	}
}

func TestParsePageRecursive(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp string
		if r.URL.Path == "/foo" {
			resp = "<p>This is foo, head over to <a href=\"/foo/bar\">bar</a>."
		} else {
			resp = "<p>Go to <a href=\"/foo\">foo</a>.</p><p><a href=\"https://google.com\">To the Googles!</a></p>"
		}
		fmt.Fprintln(w, resp)
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parsePage(): failed to parse test server addr %s as URL", ts.URL)
	}

	c := NewCrawler(tsURL, 2)

	c.parsePage(ts.URL, 0)

	if len(c.sitemap) != 2 {
		t.Errorf("parsePage(): expected 2 links in sitemap, got %d", len(c.sitemap))
		t.Errorf("actual: %v", c.sitemap)
	}

	expected1 := Page{Addr: CanonicalURL(ts.URL), Links: Links{ts.URL + "/foo"}}
	expected2 := Page{Addr: CanonicalURL(ts.URL + "/foo"), Links: Links{ts.URL + "/foo/bar"}}

	actual1, found := c.sitemap[expected1.Addr]
	if !found {
		t.Errorf("parsePage(): expected sitemap %v to contain page %s", c.sitemap, expected1.Addr)
	}

	if len(actual1) != 1 {
		t.Errorf("parsePage(): expected page %s to have one link, got %v", expected1.Addr, actual1)
	}

	for i := 0; i < 1; i++ {
		if actual1[i] != ts.URL+"/foo" {
			t.Errorf("parsePage(): expected page %s to have link %s, got %v", expected1.Addr, ts.URL+"/foo", actual1)
		}
	}

	actual2, found := c.sitemap[expected2.Addr]
	if !found {
		t.Errorf("parsePage(): expected sitemap %v to contain page %s", c.sitemap, expected2.Addr)
	}

	if len(actual2) != 1 {
		t.Errorf("parsePage(): expected page %s to have one link, got %v", expected2.Addr, actual2)
	}

	for i := 0; i < 1; i++ {
		if actual2[i] != ts.URL+"/foo/bar" {
			t.Errorf("parsePage(): expected page %s to have link %s, got %v", expected2.Addr, ts.URL+"/foo/bar", actual2)
		}
	}
}

func TestParsePageRespectsMaxDepth(t *testing.T) {
	c := NewCrawler(&url.URL{}, 2)
	c.parsePage("", 3)

	if len(c.sitemap) > 0 {
		t.Errorf("sitemap was expected to be empty, got %v", c.sitemap)
	}
}

func TestParsePageIgnoresMaxDepthIfSetToUnlimited(t *testing.T) {
	rawhtml, err := ioutil.ReadFile("../fixtures/simple.html")
	if err != nil {
		t.Fatalf("failed to parse the %s fixture file: %s", "simple.html", err.Error())
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(rawhtml))
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("parsePage(): failed to parse test server addr %s as URL", ts.URL)
	}

	c := NewCrawler(tsURL, 0)

	c.parsePage(ts.URL, -1)

	expected := Links{
		"https://local.com/foo/bar",
	}

	actual := c.sitemap[CanonicalURL(ts.URL)]

	if len(actual) != len(expected) {
		t.Errorf("parsePage(): expected %d links in sitemap, got %d", len(expected), len(actual))
		t.Errorf("expected: %v", expected)
		t.Errorf("actual: %v", actual)
	}
}

func TestParsePageRespectsCancellationFlag(t *testing.T) {
	c := NewCrawler(&url.URL{}, 0)
	c.keepCrawling = false
	c.parsePage("", 1)

	if len(c.sitemap) > 0 {
		t.Errorf("sitemap was expected to be empty, got %v", c.sitemap)
	}
}

func TestAdd(t *testing.T) {
	addr := CanonicalURL("https://foo.com")
	p := Page{Addr: addr, Links: Links{"https://bar.com", "http://baz.com"}}

	c := NewCrawler(&url.URL{}, 0)

	if _, ok := c.sitemap[addr]; ok {
		t.Errorf("add(): sitemap %v was not expected to contain %s", c.sitemap, addr)
	}

	c.add(p)

	if _, ok := c.sitemap[addr]; !ok {
		t.Errorf("add(): expected sitemap %v to contain %s", c.sitemap, addr)
	}

	if len(c.sitemap[addr]) != len(p.Links) {
		t.Errorf("add(): expected %d links saved, got %d", len(c.sitemap[addr]), len(p.Links))
		t.Errorf("expected: %v", p.Links)
		t.Errorf("actual: %v", c.sitemap[addr])
	}

	for _, ll := range p.Links {
		found := false
		for _, kk := range c.sitemap[addr] {
			if kk == ll {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("add(): expected sitemap %v to contain %s", c.sitemap, ll)
		}
	}
}

func TestKnown(t *testing.T) {
	addr := CanonicalURL("https://foo.com")

	c := NewCrawler(&url.URL{}, 0)

	if c.known(addr) {
		t.Errorf("known(): sitemap %v was not expected to contain %s", c.sitemap, addr)
	}

	c.sitemap[addr] = Links{"https://bar.com", "http://baz.com"}

	if !c.known(addr) {
		t.Errorf("known(): expected sitemap %v to contain %s", c.sitemap, addr)
	}
}
