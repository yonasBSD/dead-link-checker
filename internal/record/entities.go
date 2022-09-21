package record

type Recording struct {
	Links              []Link
	BrokenLinkByAbsURL map[string]BrokenLinkDetails
}

type Link struct {
	// Value of the link, most likely a URL.
	// Can be relative or absolute.
	LinkValue string
	// Absolute URL based on the link value.
	// Empty if not a valid URL.
	AbsoluteURL string
	// URL of the page on which the link was found.
	PageURL string
	// HTML tag, e.g. "img".
	Tag string
	// Text content for the tag.
	TagText TagText
	// HTML attribute of tag, e.g. "src".
	Attribute string
}

type TagText struct {
	// Type of text for the tag.
	// E.g. "ATTRIBUTE" for an image.
	Type TagTextType
	// Key of the text for the tag.
	// E.g. "alt" for an image.
	Key string
	// Content of the HTML tag or relevant attribute.
	// E.g. text content for anchor, "alt" attribute value for image image, ...
	Value string
}

type BrokenLink struct {
	AbsoluteURL string
	BrokenLinkDetails
}

type BrokenLinkDetails struct {
	StatusCode        int
	StatusDescription string
}

type TagTextType string

const (
	// Relevant text is in an attribute.
	TagTextTypeAttribute = "ATTRIBUTE"
	// Relevant text is in the content of the tag.
	TagTextTypeContent = "CONTENT"
	// Tag has no relevant text.
	TagTextTypeNone = "NONE"
)

func NewTagTextAttribute(key, value string) TagText {
	return TagText{
		Type:  TagTextTypeAttribute,
		Key:   key,
		Value: value,
	}
}

func NewTagTextContent(value string) TagText {
	return TagText{
		Type:  TagTextTypeContent,
		Key:   "",
		Value: value,
	}
}

func NewTagTextNone() TagText {
	return TagText{
		Type:  TagTextTypeNone,
		Key:   "",
		Value: "",
	}
}
