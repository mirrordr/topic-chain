package utils

import (
	"context"
	"fmt"
	"github.com/sashabaranov/go-openai"
	"log"
	"topic-chain/config"
)

// your embedding model name
var (
	targetMode = openai.SmallEmbedding3
	key        = config.LoadStrFromEnv("OPENAI_API_KEY")
	baseUrl    = config.LoadStrFromEnv("OPENAI_BASE_URL")
)

func EmbeddingText(text string) []float32 {
	config := openai.DefaultConfig(key)
	config.BaseURL = baseUrl
	client := openai.NewClientWithConfig(config)

	queryReq := openai.EmbeddingRequest{
		Input: []string{text},
		Model: targetMode,
	}

	queryResponse, err := client.CreateEmbeddings(context.Background(), queryReq)
	if err != nil {
		log.Fatal("Error creating query embedding:", err)
	}

	embedding := queryResponse.Data[0].Embedding

	fmt.Println("embedding length: ", len(embedding))

	// fmt.Println(embedding)

	return embedding

}
