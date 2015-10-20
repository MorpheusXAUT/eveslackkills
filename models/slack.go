package models

// SlackPayload represents the payload to be sent to the Slack web hook
type SlackPayload struct {
	Attachments []SlackAttachment `json:"attachments"`
}

// SlackAttachment represents a single attachment as sent as a Slack payload
type SlackAttachment struct {
	Title       string       `json:"title,omitempty"`
	TitleLink   string       `json:"title_link,omitempty"`
	ThumbURL    string       `json:"thumb_url,omitempty"`
	Fallback    string       `json:"fallback,omitempty"`
	Color       string       `json:"color,omitempty"`
	Fields      []SlackField `json:"fields,omitempty"`
	UnfurlLinks bool         `json:"unfurl_links"`
	Channel     string       `json:"channel,omitempty"`
}

// SlackField represents a single field used within Slack attachments for formatting
type SlackField struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short"`
}
