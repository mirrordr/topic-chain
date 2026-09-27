package request

type GLMParameter struct {
	Prompt      []GLMPrompt `json:"prompt"`
	Temperature float32     `json:"temperature"`
	TopP        float32     `json:"top_p"`
	RequestId   string      `json:"request_id"`
	ReturnType  string      `json:"return_type"`
	Ref         GLMRef      `json:"ref"`
}

type GLMPrompt struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GLMRef struct {
	Enable      bool   `json:"enable"`
	SearchQuery string `json:"search_query"`
}
