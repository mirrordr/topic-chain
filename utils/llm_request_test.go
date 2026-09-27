package utils

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"
)

func tSaveUsage() {
	usage_input_gpt := 3
	usage_output_gpt := 4
	filePath := "./record/usage.txt"
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		fmt.Println("文件打开失败", err)
	}
	//及时关闭file句柄
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			fmt.Println("文件句柄关闭失败: ", err)
		}
	}(file)
	write := bufio.NewWriter(file)
	_, err = write.WriteString("input_token: " + strconv.Itoa(usage_input_gpt) + "\toutput_token: " + strconv.Itoa(usage_output_gpt) + "\n")
	if err != nil {
		fmt.Println("写入失败: ", err)
		return
	}
	err = write.Flush()
	if err != nil {
		fmt.Println("文件flush失败: ", err)
		return
	}
}

func getInput() {
	num := 0
	fmt.Scanf("%s", &num)
	fmt.Println("num: ", num)
}

func requireLLMIntegration(t *testing.T) {
	t.Helper()
	if os.Getenv("RUN_INTEGRATION_TESTS") != "1" {
		t.Skip("requires a live LLM service; set RUN_INTEGRATION_TESTS=1 to run")
	}
}

func TestDefer(t *testing.T) {
	requireLLMIntegration(t)
	getInput()
	time.Sleep(time.Minute)
	defer fmt.Println("你好")
}

func TestSave(t *testing.T) {
	requireLLMIntegration(t)
	tSaveUsage()
}

func TestMultiQuestion(t *testing.T) {
	requireLLMIntegration(t)
	historyQ := []string{"你好。", "你会跳舞吗？"}
	historyA := []string{"你好， 我是你的人工助手。", "不好意思，我是虚拟的人工智能，无法跳舞。"}
	question := "1+1等于几？"
	SendMultiTalkToLLM(historyQ, historyA, question, "GPT", make(chan any))
}
