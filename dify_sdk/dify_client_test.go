package dify_sdk

import (
	"os"
	"testing"
)

func TestSendMessageToDify(t *testing.T) {
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("requires a live Dify service; set RUN_INTEGRATION_TESTS=1 to run")
	}
	t.Log(SendMessageToDify("你好"))

}
