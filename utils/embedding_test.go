package utils

import (
	"os"
	"testing"
)

func requireIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("requires OpenAI-compatible embeddings and Qdrant; set RUN_INTEGRATION_TESTS=1 to run")
	}
}

func TestEmbeddingText(t *testing.T) {
	requireIntegration(t)
	EmbeddingText("你好")
}

func TestInitQDrantClient(t *testing.T) {
	requireIntegration(t)
	InitQDrantClient()
}

func TestEmbeddingWorkflow(t *testing.T) {
	requireIntegration(t)
	InitQDrantClient()
	topic := "你好"
	vec := EmbeddingText(topic)
	UpsertVectorToQDrant(1, vec, topic)

	topic = "你叫什么名字"
	vec = EmbeddingText(topic)
	UpsertVectorToQDrant(2, vec, topic)

	topic = "我刚吃了黄焖鸡"
	vec = EmbeddingText(topic)
	UpsertVectorToQDrant(3, vec, topic)

	topic = "你饿了吗"
	vec = EmbeddingText(topic)
	UpsertVectorToQDrant(4, vec, topic)

	topic = "你晚饭解决了没"
	vec = EmbeddingText(topic)
	UpsertVectorToQDrant(5, vec, topic)

	query := "你吃饭了没"
	vec = EmbeddingText(query)
	TopicSimilaritySearchFromQDrant(vec)

}
