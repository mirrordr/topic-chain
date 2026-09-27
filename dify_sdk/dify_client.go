package dify_sdk

// if you want to use Dify to be your LLM endpoint
// you should config your dify service info in here

import (
	"log"
	"topic-chain/config"
)
import dify "github.com/bswaterb/dify-go-sdk"

var (
	difyApiKey      = config.LoadStrFromEnv("DIFY_API_KEY")
	difyApiHost     = config.LoadStrFromEnv("DIFY_API_HOST")
	difyConsoleHost = config.LoadStrFromEnv("DIFY_CONSOLE_HOST")
)

var totalTokenUsage = 0

func SendMessageToDify(message string) string {
	apiKey := difyApiKey
	apiHost := difyApiHost
	consoleHost := difyConsoleHost
	client, err := dify.CreateDifyClient(dify.DifyClientConfig{
		Key:         apiKey,
		Host:        apiHost,
		ConsoleHost: consoleHost,
		// timeout - 180s
		Timeout: 180,
		User:    "topic-chain",
	})
	if err != nil {
		log.Printf("failed to create DifyClient: %v\n", err)
		return ""
	}
	inputs := `{"user": "1"}`
	res, err := client.ChatMessages(message, inputs, "", []any{})
	if err != nil {
		log.Printf("failed to resp: %v\n%v", err, res)
		return ""
	}
	metadata := res.Metadata.(map[string]any)
	usage := metadata["usage"].(map[string]any)
	totalTokens := int(usage["total_tokens"].(float64))

	totalTokenUsage += totalTokens
	log.Printf("累计请求 tokens 用量：%d\n", totalTokenUsage)

	return res.Answer
}
