package response

type GLMResult struct {
	Code    int     `json:"code"`
	Msg     string  `json:"msg"`
	Success bool    `json:"success"`
	Data    GLMData `json:"data"`
}

type GLMData struct {
	Choices    []GLMDataChoice `json:"choices"`
	RequestId  string          `json:"request_id"`
	TaskId     string          `json:"task_id"`
	TaskStatus string          `json:"task_status"`
	Usage      GLMDataUsage    `json:"usage"`
}

type GLMDataChoice struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GLMDataUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}
