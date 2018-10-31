# 🕸 Crawler

Crawler is a sitemap generator.

⚠️ This is a toy project, not really intended for use in production.

## Usage

Run `go run cmd/main.go` to kick off crawling.

You can specify the following options:

`-depth` Number of nested levels to parse (0 for unlimited; defaults to 2)

`-graph` Renders the sitemap as a graph saved to an .svg file rather than as text on the screen. Graphviz (dot) is required for this to work.

`-timeout` Max allowed crawling time in seconds (0 for unlimited; defaults to 1m0s)

`-url` Full URL of the website to be crawled, e.g. https://google.com (defaults to https://www.google.com if not specified)

### Example output:

```
$ go run cmd/main.go
Crawling https://www.google.com up to 2 level(s) deep (timeout 1m0s).
2018/10/31 23:18:58 parsing https://www.google.com/language_tools returned an error: Get https://translate.google.com/: URL is outside the starting domain, ignoring
2018/10/31 23:18:58 parsing https://www.google.com/intl/en/ads returned an error: Get https://ads.google.com/intl/en/home/: URL is outside the starting domain, ignoring

pages:

https://www.google.com
https://www.google.com/webhp
https://www.google.com/services
https://www.google.com/intl/en/about
https://www.google.com/intl/en/policies/terms
https://www.google.com/preferences
https://www.google.com/advanced_search
https://www.google.com/intl/en/policies/privacy

links:

https://www.google.com/intl/en/about -> https://www.google.com/.
https://www.google.com/intl/en/about -> https://www.google.com/./stories
https://www.google.com/intl/en/about -> https://www.google.com/permissions
https://www.google.com/intl/en/about -> https://www.google.com/accessibility
https://www.google.com/intl/en/about -> https://www.google.com/policies/terms
https://www.google.com/intl/en/about -> https://www.google.com/./locations
https://www.google.com/intl/en/about -> https://www.google.com/doodles
https://www.google.com/intl/en/about -> https://www.google.com/press/blog-social-directory.html
https://www.google.com/intl/en/about -> https://www.google.com/diversity
https://www.google.com/intl/en/about -> https://www.google.com/./responsible-supply-chain
https://www.google.com/intl/en/about -> https://www.google.com/policies/privacy
https://www.google.com/intl/en/about -> https://www.google.com/./products
https://www.google.com/intl/en/about -> https://www.google.com/./our-story
https://www.google.com/intl/en/about -> https://www.google.com
https://www.google.com/intl/en/about -> https://www.google.com/contact
https://www.google.com/intl/en/about -> https://www.google.com/./our-commitments
https://www.google.com/intl/en/about -> https://www.google.com/./appsecurity
https://www.google.com/intl/en/about -> https://www.google.com/./software-principles.html
https://www.google.com/intl/en/about -> https://www.google.com/./unwanted-software-policy.html
https://www.google.com/preferences -> https://www.google.com/history
https://www.google.com/preferences -> https://www.google.com/services
https://www.google.com/preferences -> https://www.google.com/intl/en/policies
https://www.google.com/preferences -> https://www.google.com/intl/en/about.html
https://www.google.com/preferences -> https://www.google.com/preferences
https://www.google.com/preferences -> https://www.google.com/support/websearch
https://www.google.com/preferences -> https://www.google.com
https://www.google.com/preferences -> https://www.google.com/webhp
https://www.google.com/preferences -> https://www.google.com/intl/en/ads
https://www.google.com/advanced_search -> https://www.google.com/preferences
https://www.google.com/advanced_search -> https://www.google.com
https://www.google.com/advanced_search -> https://www.google.com/intl/en/ads
https://www.google.com/advanced_search -> https://www.google.com/services
https://www.google.com/advanced_search -> https://www.google.com/intl/en/policies
https://www.google.com/advanced_search -> https://www.google.com/intl/en/about.html
https://www.google.com -> https://www.google.com/preferences
https://www.google.com -> https://www.google.com/language_tools
https://www.google.com -> https://www.google.com/intl/en/ads
https://www.google.com -> https://www.google.com/intl/en/about.html
https://www.google.com -> https://www.google.com/intl/en/policies/terms
https://www.google.com -> https://www.google.com/search
https://www.google.com -> https://www.google.com/advanced_search
https://www.google.com -> https://www.google.com/services
https://www.google.com -> https://www.google.com/setprefdomain
https://www.google.com -> https://www.google.com/intl/en/policies/privacy
https://www.google.com/webhp -> https://www.google.com/language_tools
https://www.google.com/webhp -> https://www.google.com/services
https://www.google.com/webhp -> https://www.google.com/search
https://www.google.com/webhp -> https://www.google.com/advanced_search
https://www.google.com/webhp -> https://www.google.com/intl/en/about.html
https://www.google.com/webhp -> https://www.google.com/setprefdomain
https://www.google.com/webhp -> https://www.google.com/intl/en/policies/privacy
https://www.google.com/webhp -> https://www.google.com/intl/en/policies/terms
https://www.google.com/webhp -> https://www.google.com/preferences
https://www.google.com/webhp -> https://www.google.com/intl/en/ads
https://www.google.com/services -> https://www.google.com/../services
https://www.google.com/services -> https://www.google.com
https://www.google.com/services -> https://www.google.com/intl/en/analytics/data-studio

Done!
```

## Testing

Run `go test ./pkg/` to run the unit tests.

Fixture files are under  `/fixtures`.