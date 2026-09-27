package chatglm_sdk

import (
	"os"
	"testing"
	"topic-chain/chatglm_sdk/request"
)

func TestNewClient(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("requires a live ChatGLM API; set RUN_INTEGRATION_TESTS=1 to run")
	}

	client := NewGLMClient(request.GLMPro)
	// client.SendSingleQuestion("你好")
	historyQ := []string{"你好。", "你会跳舞吗？"}
	historyA := []string{"你好， 我是你的人工助手。", "不好意思，我是虚拟的人工智能，无法跳舞。"}
	question := "我之前问过你什么问题？"
	client.SendMultiTalk(historyQ, historyA, question)

}
