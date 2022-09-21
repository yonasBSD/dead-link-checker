package check

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/gocolly/colly/v2/extensions"
	"github.com/rs/zerolog/log"
	"golang.org/x/exp/slices"

	"github.com/JenswBE/dead-link-checker/cmd/config"
	"github.com/JenswBE/dead-link-checker/internal/record"
)

const (
	requestTimeout = 30 * time.Second
)

var ignoredSchemes = []string{
	"data:",
	"ftp:",
	"javascript:",
	"mailto:",
	"tel:",
}

var tags = map[string]tagConfig{
	"a":   {linkAttributes: []string{"href"}},          // Anchors
	"img": {linkAttributes: []string{"src", "srcset"}}, // Images
	"link": { // CSS stylesheets
		linkAttributes: []string{"href"},
		ignoreWhenAttributeMatches: map[string][]string{
			"rel": {"dns-prefetch", "pingback", "preconnect", "profile"},
		},
	},
	"script": {linkAttributes: []string{"src"}},    // JS scripts
	"source": {linkAttributes: []string{"srcset"}}, // Part of <picture>
}

// Run checks the provided site. This call blocks until the whole site is checked.
func Run(siteConfig config.SiteConfig, recorder *record.Recorder) error {
	// Create collector
	collector := colly.NewCollector(
		colly.Async(true),
		colly.DisallowedURLFilters(siteConfig.IgnoredLinks...),
		colly.IgnoreRobotsTxt(),
		extensions.RandomUserAgent,
	)
	collector.SetRequestTimeout(requestTimeout)

	// Define OnHTML callback
	for linkTag, config := range tags {
		for _, linkAttr := range config.linkAttributes {
			query := fmt.Sprintf("%s[%s]", linkTag, linkAttr)
			collector.OnHTML(query, handleHTML(collector, recorder, linkTag, linkAttr))
		}
	}

	// Define OnRequest callback
	collector.OnRequest(func(r *colly.Request) {
		log.Debug().Msgf("Visiting %v", r.URL)
	})

	// Define OnError callback
	siteURL := siteConfig.URL.String()
	collector.OnError(func(r *colly.Response, err error) {
		report := record.BrokenLink{
			AbsoluteURL: r.Request.URL.String(),
			BrokenLinkDetails: record.BrokenLinkDetails{
				StatusCode:        r.StatusCode,
				StatusDescription: err.Error(),
			},
		}
		recorder.RecordBrokenLink(report)
		log.Warn().Err(err).
			Int("status_code", report.StatusCode).
			Str("link_value", r.Request.Ctx.Get("link_value")).
			Str("page_url", r.Request.Ctx.Get("page_url")).
			Str("tag", r.Request.Ctx.Get("tag")).
			Str("attribute", r.Request.Ctx.Get("attribute")).
			Str("site_url", siteURL).
			Msg("Following link returned error")
	})

	// Start initial request
	ctx := colly.NewContext()
	ctx.Put("site_url", siteURL)
	if err := collector.Request(http.MethodGet, siteConfig.URL.String(), nil, ctx, nil); err != nil {
		return fmt.Errorf("failed to start collector for site %s: %w", siteConfig.URL, err)
	}

	// Wait for collector to finish
	collector.Wait()
	return nil
}

func handleHTML(collector *colly.Collector, recorder *record.Recorder, tag, attr string) colly.HTMLCallback {
	return func(e *colly.HTMLElement) {
		// Set context
		e.Request.Ctx.Put("page_url", e.Request.URL.String())
		e.Request.Ctx.Put("tag", tag)
		e.Request.Ctx.Put("attribute", attr)
		linkReport := record.Link{
			PageURL:   e.Request.URL.String(),
			Tag:       tag,
			Attribute: attr,
		}

		// Check if tag should be ignored
		for attr, attrValues := range tags[tag].ignoreWhenAttributeMatches {
			attrValue := strings.TrimSpace(e.Attr(attr))
			if slices.Contains(attrValues, attrValue) {
				log.Debug().Str("tag", tag).Str("attribute", attr).Str("attribute_value", attrValue).
					Msg("Link ignored because attribute value is in list to ignore")
				return
			}
		}

		// Process attribute
		switch attr {
		case "srcset":
			items := strings.Split(e.Attr("srcset"), ",")
			for _, item := range items {
				// item is e.g. "/images/example4x.jpg 4x"
				itemParts := strings.Split(strings.TrimSpace(item), " ")
				handleLinkValue(collector, recorder, e, itemParts[0], linkReport)
			}
		default:
			handleLinkValue(collector, recorder, e, e.Attr(attr), linkReport)
		}
	}
}

func handleLinkValue(
	collector *colly.Collector,
	recorder *record.Recorder,
	e *colly.HTMLElement,
	linkValue string,
	linkReport record.Link,
) {
	site := e.Request.Ctx.Get("site_url")
	logger := log.With().Str("site_url", site).Str("link_value", linkValue).Logger()
	if strings.HasPrefix(linkValue, "#") {
		// Skip link as it's a hash link to the current page
		logger.Debug().Str("page_url", e.Request.URL.String()).
			Msg("Link ignored because it is a hash link to the current page")
		return
	}
	if !strings.HasPrefix(e.Request.URL.String(), site) {
		// Skip link as we are already on an external site
		logger.Debug().Str("page_url", e.Request.URL.String()).
			Msg("Link ignored because we are on an external site")
		return
	}
	if hasIgnoredScheme(linkValue) {
		// Skip link as it has an ignored scheme
		logger.Debug().Msg("Link ignored because it has an ignored scheme")
		return
	}

	// Extract tag content
	switch e.Name {
	case "a":
		linkReport.TagText = record.NewTagTextContent(e.Text)
	case "img":
		linkReport.TagText = record.NewTagTextAttribute("alt", e.Attr("alt"))
	case "link":
		linkReport.TagText = record.NewTagTextAttribute("rel", e.Attr("rel"))
	default:
		linkReport.TagText = record.NewTagTextNone()
	}

	// Process link
	linkReport.LinkValue = linkValue
	linkReport.AbsoluteURL = e.Request.AbsoluteURL(linkValue)
	recorder.RecordLink(linkReport)

	// Clone context
	ctxClone := colly.NewContext()
	e.Request.Ctx.ForEach(func(k string, v interface{}) interface{} {
		ctxClone.Put(k, v)
		return nil
	})

	// Visit link
	ctxClone.Put("link_value", linkValue)
	err := collector.Request(http.MethodGet, linkReport.AbsoluteURL, nil, ctxClone, nil)
	if err != nil && !errors.Is(err, colly.ErrAlreadyVisited) && !errors.Is(err, colly.ErrForbiddenURL) {
		log.Error().Err(err).Str("url", linkReport.AbsoluteURL).Str("method", http.MethodGet).
			Msg("Failed to send request. Will mark as broken link.")
		recorder.RecordBrokenLink(record.BrokenLink{
			AbsoluteURL: linkReport.AbsoluteURL,
			BrokenLinkDetails: record.BrokenLinkDetails{
				StatusCode:        0,
				StatusDescription: "Failed to create request: " + err.Error(),
			},
		})
	}
}

func hasIgnoredScheme(linkValue string) bool {
	for _, scheme := range ignoredSchemes {
		if strings.HasPrefix(linkValue, scheme) {
			return true
		}
	}
	return false
}
