package record

type Recording struct {
	Links              []Link
	BrokenLinkByAbsURL map[string]BrokenLinkDetails
}

type Link struct {
	LinkValue   string  // Value of the link, most likely a URL. Can be relative or absolute.
	AbsoluteURL string  // Absolute URL based on the link value. Empty if not a valid URL.
	PageURL     string  // URL of the page on which the link was found.
	Tag         string  // HTML tag, e.g. "img".
	TagText     TagText // Text content for the tag
	Attribute   string  // HTML attribute of tag, e.g. "src".
}

type TagText struct {
	Type  TagTextType // Type of text for the tag. E.g. "ATTRIBUTE" for an image.
	Key   string      // Key of the text for the tag. E.g. "alt" for an image.
	Value string      // Content of the HTML tag or relevant attribute. E.g. text content for anchor, "alt" attribute value for image image, ...
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
	// Relevant text is in an attribute
	TagTextTypeAttribute = "ATTRIBUTE"
	// Relevant text is in the content of the tag
	TagTextTypeContent = "CONTENT"
	// Tag has no relevant text
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
