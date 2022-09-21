// Package report implements code to convert the recorded broken links into a usable report.
// It's also responsible for notifying the user.
package report

import (
	"golang.org/x/exp/slices"

	"github.com/JenswBE/dead-link-checker/cmd/config"
	"github.com/JenswBE/dead-link-checker/internal/record"
)

func GenerateReport(siteConfig config.SiteConfig, recording record.Recording) SiteReport {
	countOnPageByPageURLByBrokenLink := make(map[string]map[BrokenLink]int, len(recording.BrokenLinkByAbsURL))
	report := NewEmptySiteReport(siteConfig.URL.String())
	for _, link := range recording.Links {
		report.Statistics.LinksCountTotal++
		report.Statistics.LinksCountByPageURL[link.PageURL]++
		// Check if provided link is broken
		if brokenLink, ok := recording.BrokenLinkByAbsURL[link.AbsoluteURL]; ok {
			// Link is broken
			brokenLink := BrokenLink{
				LinkValue:              link.LinkValue,
				AbsoluteURL:            link.AbsoluteURL,
				Tag:                    link.Tag,
				IsTagTextTypeAttribute: link.TagText.Type == record.TagTextTypeAttribute,
				IsTagTextTypeContent:   link.TagText.Type == record.TagTextTypeContent,
				IsTagTextTypeNone:      link.TagText.Type == record.TagTextTypeNone,
				TagTextKey:             link.TagText.Key,
				TagTextValue:           link.TagText.Value,
				Attribute:              link.Attribute,
				StatusCode:             brokenLink.StatusCode,
				StatusDescription:      brokenLink.StatusDescription,
				CountOnPage:            1,
			}
			if countMap, ok := countOnPageByPageURLByBrokenLink[link.PageURL]; ok {
				countMap[brokenLink]++
			} else {
				countOnPageByPageURLByBrokenLink[link.PageURL] = map[BrokenLink]int{brokenLink: 1}
			}
		}
	}

	// Collect broken links
	report.BrokenLinksByPageURL = make(map[string][]BrokenLink, len(countOnPageByPageURLByBrokenLink))
	for pageURL, countOnPageByBrokenLink := range countOnPageByPageURLByBrokenLink {
		brokenLinks := make([]BrokenLink, 0, len(countOnPageByBrokenLink))
		for brokenLink, count := range countOnPageByBrokenLink {
			brokenLink.CountOnPage = count
			brokenLinks = append(brokenLinks, brokenLink)
		}
		slices.SortFunc(brokenLinks, func(a, b BrokenLink) bool { return a.String() < b.String() }) // Ensure consistent order
		report.BrokenLinksByPageURL[pageURL] = brokenLinks
	}
	return report
}
