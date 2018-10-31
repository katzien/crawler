package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/katzien/crawler/pkg"
	"log"
	"net/url"
	"time"
)

const (
	// DefaultURL is the default URL which will be crawled if no url flag has been specified.
	DefaultURL = "https://www.google.com"

	// DefaultDepth is the default max crawling depth if no depth flag has been specified.
	DefaultDepth = 2

	// DefaultTimeout defines the default max allowed crawling time (in seconds) if no timeout flag has been specified.
	DefaultTimeout = 60 * time.Second

	// DefaultGraph specifies whether the sitemap should be saved in a graph file or output to stdout.
	// By default, the program will output the pages and links found between them in a text format on the screen.
	// If the graph flag is specified, the sitemap will be rendered as a graph and saved to an .svg file instead.
	// A program called "dot" (part of Graphviz) is required to render the graph file.
	DefaultGraph = false
)

func main() {

	startURL, maxDepth, timeout, graph := parseFlags()

	u, err := url.Parse(startURL)
	if err != nil {
		log.Fatal(err)
	}

	if u.Scheme == "" || u.Host == "" {
		log.Fatal("invalid URL: a full, non-relative URL including the protocol must be specified (e.g. https://google.com)")
	}

	if maxDepth < 0 {
		log.Fatal("depth cannot be negative")
	}

	if timeout < 0 {
		log.Fatal("timeout cannot be negative")
	}

	var ctx context.Context
	var cancel context.CancelFunc
	var tInfo string

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		tInfo = fmt.Sprintf(" (timeout %s)", timeout.String())
	} else {
		ctx = context.Background()
		tInfo = " (no timeout specified)"
	}

	dInfo := ""
	if maxDepth > 0 {
		dInfo = fmt.Sprintf(" up to %d level(s) deep", maxDepth)
	}

	fmt.Printf("Crawling %s%s%s.\n", u.String(), dInfo, tInfo)

	c := crawler.NewCrawler(u, maxDepth)

	sitemap := c.Crawl(ctx)

	if ctx.Err() != nil && ctx.Err() == context.DeadlineExceeded {
		fmt.Println("Max crawling time exceeded, saving current results...")
	}

	if graph {
		err = crawler.Graph(sitemap)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Printf("Sitemap graph file saved in %s.\n", crawler.DefaultOutputFileSvg)
	} else {
		text, err := crawler.Text(sitemap)
		if err != nil {
			log.Fatal(err.Error())
		}
		fmt.Println(text)
	}

	fmt.Println("Done!")
}

func parseFlags() (string, int, time.Duration, bool) {
	u := flag.String("url", DefaultURL, fmt.Sprintf("Full URL of the website to be crawled, e.g. https://google.com (defaults to %s if not specified)", DefaultURL))
	d := flag.Int("depth", DefaultDepth, fmt.Sprintf("Number of nested levels to parse (0 for unlimited; defaults to %d)", DefaultDepth))
	t := flag.Duration("timeout", DefaultTimeout, fmt.Sprintf("Max allowed crawling time in seconds (0 for unlimited; defaults to %s)", DefaultTimeout.String()))
	g := flag.Bool("graph", DefaultGraph, fmt.Sprintf("Renders the sitemap as a graph saved to an .svg file rather than as text on the screen. Graphviz (dot) is required for this to work."))

	flag.Parse()

	return *u, *d, *t, *g
}
