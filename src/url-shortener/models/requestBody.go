package models

type GenerateRequestBody struct {
	URL         string `json:"url"`
	Shortcode   string `json:"shortcode,omitempty"`
	Description string `json:"description"`
	AndroidURL  string `json:"androidUrl,omitempty"`
	IOSURL      string `json:"iosUrl,omitempty"`
	LinuxURL    string `json:"linuxUrl,omitempty"`
	MacURL      string `json:"macUrl,omitempty"`
	WindowsURL  string `json:"windowsUrl,omitempty"`
}
