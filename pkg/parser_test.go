package crawler

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

var parserTests = []struct {
	scheme   string
	host     string
	file     string
	expected Page
}{
	{
		"http",
		"localhost",
		"simple.html",
		Page{Addr: CanonicalURL("http://simple.com"), Links: Links{"http://localhost/foo/bar"}},
	},
	{
		"http",
		"localhost",
		"nolinks.html",
		Page{Addr: CanonicalURL("http://simple.com"), Links: Links{}},
	},
	{
		"http",
		"localhost",
		"unparseable.html",
		Page{Addr: CanonicalURL("http://simple.com"), Links: Links{"http://localhost/foo/bar"}},
	},
}

func TestParse(t *testing.T) {

	for _, tt := range parserTests {

		p := NewParser(tt.scheme, tt.host)

		rawhtml, err := ioutil.ReadFile("../fixtures/" + tt.file)
		if err != nil {
			t.Fatalf("failed to parse the %s fixture file: %s", tt.file, err.Error())
		}

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, string(rawhtml))
		}))
		defer ts.Close()

		page, err := p.parse(ts.URL)

		if err != nil {
			t.Errorf("parse() returned an error: %s", err.Error())
			t.FailNow()
		}

		if page.Addr != CanonicalURL(ts.URL) {
			t.Errorf("parse() page.Addr: expected %s, actual %s", ts.URL, page.Addr)
		}

		if len(page.Links) != len(tt.expected.Links) {
			t.Errorf("parse() page.Links: expected %d link(s), got %d", len(tt.expected.Links), len(page.Links))
			t.Errorf("expected: %v", tt.expected.Links)
			t.Errorf("actual: %v", page.Links)
		}

		for _, ll := range tt.expected.Links {
			found := false
			for _, kk := range page.Links {
				if ll == kk {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("parse(%s): expected page.Links %v to contain %s", tt.file, page.Links, ll)
			}
		}
	}
}

func TestParseReturnsErrorIfPageInaccessible(t *testing.T) {
	p := NewParser("", "")
	_, err := p.parse("")

	if err == nil {
		t.Error("parse(): expected an an error to be returned, got none")
		t.Fail()
	}
}

func TestParseReturnsErrorIfPageReturnsMoreThan10Redirects(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/endless-redirect", http.StatusTemporaryRedirect)
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf(
			"couldn't parse the test server URL %s: %s", ts.URL, err.Error(),
		)
	}

	p := NewParser(tsURL.Scheme, tsURL.Host)

	_, err = p.parse(ts.URL)

	if err == nil {
		t.Errorf("parse(endless redirect): expected to get an error, got nil")
		t.FailNow()
	}

	if !strings.Contains(err.Error(), ErrTooManyRedirects.Error()) {
		t.Errorf("parse(endless redirect): didn't return the expected error")
		t.Errorf("expected: %s", ErrTooManyRedirects.Error())
		t.Errorf("got: %s", err.Error())
		t.FailNow()
	}
}

func TestParseReturnsErrorIfPageRedirectsToExternalDomain(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://google.com/foo", http.StatusTemporaryRedirect)
	}))
	defer ts.Close()

	p := NewParser("https", "notgoogle.com")

	_, err := p.parse(ts.URL)

	if err == nil {
		t.Errorf("parse(external): expected to get an error, got nil")
		t.FailNow()
	}

	if !strings.Contains(err.Error(), ErrExternalDomain.Error()) {
		t.Errorf("parse(external): didn't return the expected error")
		t.Errorf("expected: %s", ErrExternalDomain.Error())
		t.Errorf("got: %s", err.Error())
		t.FailNow()
	}
}

func TestParseFollowsRedirectsWithinDomain(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var resp string
		if r.URL.Path == "/bar" {
			resp = "<p>No more redirects, head over to <a href=\"/baz\">baz</a>."
		} else if r.URL.Path == "/foo" {
			http.Redirect(w, r, "/bar", http.StatusPermanentRedirect)
		} else {
			http.Redirect(w, r, "/foo", http.StatusTemporaryRedirect)
		}
		fmt.Fprintln(w, resp)
	}))
	defer ts.Close()

	tsURL, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatalf("couldn't parse the test server URL %s: %s", ts.URL, err.Error())
	}

	p := NewParser(tsURL.Scheme, tsURL.Host)

	page, err := p.parse(ts.URL)
	if err != nil {
		t.Errorf("parse(redirect) returned an error: %s", err.Error())
		t.FailNow()
	}

	expected := Page{
		Addr:  CanonicalURL(ts.URL + "/bar"), // making sure the parsed page's url got updated to the resulting redirect url
		Links: Links{ts.URL + "/baz"},
	}

	if page.Addr != expected.Addr {
		t.Errorf("parse(redirect) page.Addr: expected %s, actual %s", expected.Addr, page.Addr)
	}

	if len(page.Links) != len(expected.Links) {
		t.Errorf("parse(redirect) expected page.Links to contain %d links, got %d", len(expected.Links), len(page.Links))
		t.FailNow()
	}

	if page.Links[0] != expected.Links[0] {
		t.Errorf("parse(redirect) expected page.Links to be %v, got %v", expected.Links, page.Links)
		t.Fail()
	}
}

var normaliseTests = []struct {
	rawURL      string
	expectedURL string
}{
	{"/foo/bar?baz=true#476573465748", "https://google.com/foo/bar"},
	{"/foo/bar?baz=true&foo=bar", "https://google.com/foo/bar"},
	{"/foo/bar#8585utig8fug8fmgu", "https://google.com/foo/bar"},
	{"foo/bar", "https://google.com/foo/bar"},
	{"/foo/bar", "https://google.com/foo/bar"},
	{"/foo/", "https://google.com/foo"},
	{"google.com/foo", "https://google.com/google.com/foo"},
	{"//google.com", "https://google.com"},
	{"//google.com/?fruit=pear", "https://google.com"},
	{"//google.com/foo", "https://google.com/foo"},
	{"https://google.com", "https://google.com"},
	{"https://google.com/", "https://google.com"},
	{"http://google.com/foo", "http://google.com/foo"},
	{"https://google.com/?foo=bar#7dfy7dyf7dfyd", "https://google.com"},
	{"https://google.com/foo#7dfy7dyf7dfyd", "https://google.com/foo"},
	{"https://google.com/foo?bar=baz", "https://google.com/foo"},
}

func TestNormalise(t *testing.T) {

	p := NewParser("https", "google.com")

	for _, tt := range normaliseTests {

		actual, err := url.Parse(tt.rawURL)
		if err != nil {
			t.Fatalf("couldn't parse the input URL %s: %s", tt.rawURL, err.Error())
		}

		expected, err := url.Parse(tt.expectedURL)
		if err != nil {
			t.Fatalf("couldn't parse the expected URL: %s", err.Error())
		}

		p.normalise(actual)

		actualStr := actual.String()
		expectedStr := expected.String()

		if actualStr != expectedStr {
			t.Errorf("normalise(%s): expected %s, actual %s", tt.rawURL, expectedStr, actualStr)
		}
	}
}
