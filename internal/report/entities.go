package report

import (
	"fmt"
	"strings"
)

type SiteReport struct {
	SiteURL string
	Report
}

// NewEmptySiteReport returns a new empty site report.
// All maps are initialized with an empty map.
func NewEmptySiteReport(siteURL string) SiteReport {
	return SiteReport{
		SiteURL: siteURL,
		Report: Report{
			Statistics: Statistics{
				LinksCountTotal:     0,
				LinksCountByPageURL: map[string]int{},
			},
			BrokenLinksByPageURL: map[string][]BrokenLink{},
		},
	}
}

type Report struct {
	Statistics           Statistics
	BrokenLinksByPageURL map[string][]BrokenLink
}

type Statistics struct {
	LinksCountTotal     int
	LinksCountByPageURL map[string]int
}

type BrokenLink struct {
	LinkValue              string // Value of the link, most likely a URL. Can be relative or absolute.
	AbsoluteURL            string // Absolute URL based on the link value. Empty if not a valid URL.
	Tag                    string // HTML tag, e.g. "img".
	IsTagTextTypeAttribute bool   // The text type for this tag is "ATTRIBUTE".
	IsTagTextTypeContent   bool   // The text type for this tag is "CONTENT".
	IsTagTextTypeNone      bool   // The text type for this tag is "NONE".
	TagTextKey             string // Key of the text for the tag. E.g. "alt" for an image.
	TagTextValue           string // Content of the HTML tag or relevant attribute. E.g. text content for anchor, "alt" attribute value for image image, ...
	Attribute              string // HTML attribute of tag, e.g. "src".
	StatusCode             int    // HTTP status code
	StatusDescription      string // Human readable status or error description
	CountOnPage            int    // Count of this exact tag/attribute/tag content combo
}

func (l *BrokenLink) String() string {
	if l == nil {
		return ""
	}
	// "link_value=xxx;absolute_url=yyy;tag=zzz;..."
	return fmt.Sprintf(strings.Join([]string{
		"link_value=%s",
		"absolute_url=%s",
		"tag=%s",
		"is_tag_text_type_attribute=%t",
		"is_tag_text_type_content=%t",
		"is_tag_text_type_none=%t",
		"tag_text_key=%s",
		"attribute=%s",
		"status_code=%d",
		"status_description=%s",
		"count_on_page=%s",
	}, ";"),
		l.LinkValue,
		l.AbsoluteURL,
		l.Tag,
		l.IsTagTextTypeAttribute,
		l.IsTagTextTypeContent,
		l.IsTagTextTypeNone,
		l.TagTextKey,
		l.TagTextValue,
		l.Attribute,
		l.StatusCode,
		l.StatusDescription,
		l.CountOnPage,
	)
}
