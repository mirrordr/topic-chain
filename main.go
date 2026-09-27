package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"topic-chain/config"
	"topic-chain/record"
	"topic-chain/types"
	"topic-chain/utils"
)

var (
	talkLinkInitContext = &types.ContextTalkLink{
		InitialTopics:    make([]types.InitialTopic, 0, 20),
		QuestionClassify: make([]types.ClassifyResult, 0, 20),
	}
)

const (
	LLMMode = "llm-mode"
	VecMode = "vector-mode"
)

// indexer 选用智能体或使用向量相似度匹配
var indexer = LLMMode

func main() {
	if message := validateOpenAIConfig(); message != "" {
		fmt.Println(message)
		return
	}
	caseId := 0

	if indexer == VecMode {
		utils.InitQDrantClient()
	}

	talkLinkInitContext.CaseId = caseId
	startWithCache(caseId)
	// startNewSession()
	// pureStart("GPT")

	defer utils.SaveUsage()
}

func validateOpenAIConfig() string {
	for _, key := range []string{"OPENAI_API_KEY", "OPENAI_MODEL_NAME"} {
		value := strings.TrimSpace(config.LoadStrFromEnv(key))
		if value == "" || strings.HasPrefix(strings.ToLower(value), "replace-with-") {
			return fmt.Sprintf("missing required configuration: %s", key)
		}
	}
	return ""
}

// 指定 CLCBench 中的某个 case 作为对话上下文进行初始化
func startWithCache(caseId int) {
	// userMessages 可以视为用户发言，也可以视为首先发言一方进行的发言
	// aiMessages 可以视为LLM的回复，也可以视为被动回复一方进行的发言
	userMessages, aiMessages, err := record.LoadTalkingRecordsFromLocalJsonFile("./bench/main.json", caseId)
	if err != nil {
		fmt.Printf("failed to load bench/main.json: %v\n", err)
		return
	}

	initContentBuilder := strings.Builder{}
	for i := range userMessages {
		q := userMessages[i]
		a := aiMessages[i]
		initContentBuilder.WriteString("user: ")
		initContentBuilder.WriteString(q.Content + "\n")
		initContentBuilder.WriteString("assistant: ")
		initContentBuilder.WriteString(a.Content + "\n")
	}
	initContent := initContentBuilder.String()

	talkSession := utils.InitSessionChain(initContent, "GPT", "GPT", "GPT", "GPT", indexer)

	for i := 0; i < len(talkSession.Chains); i++ {
		chain := talkSession.Chains[i]
		topic := types.InitialTopic{}
		topic.Topic = chain.Topic
		topic.ChainId = chain.Id
		topic.TurnIds = record.Sentence2TurnIds(chain.Question, userMessages)
		talkLinkInitContext.InitialTopics = append(talkLinkInitContext.InitialTopics, topic)
	}

	// saveObjToJsonFile[types.ContextTalkLink](caseId, *talkLinkInitContext)

inLoop:
	for {

		fmt.Println("请输入新问题：")
		var question string
		input, _, err := bufio.NewReader(os.Stdin).ReadLine()
		if string(input) == "cmd-quit" {
			break inLoop
		}
		if err != nil {
			panic(err)
		}
		question = string(input)
		startTime := time.Now()
		useType, _, chainId, _, _ := utils.TalkWithSession(question, talkSession, "GPT", "GPT", "GPT", "GPT", indexer)
		classify := types.ClassifyResult{Question: question}
		if useType == utils.All {
			classify.TargetChain = "[ALL]"
		} else if useType == utils.Part {
			classify.TargetChain = strconv.Itoa(chainId)
		} else if useType == utils.New {
			classify.TargetChain = "[EMPTY]"
		} else {
			classify.TargetChain = "[UNKNOWN]"
		}
		talkLinkInitContext.QuestionClassify = append(talkLinkInitContext.QuestionClassify, classify)

		fmt.Println("结合对话链回答该问题耗时：", time.Since(startTime))
	}

	saveObjToJsonFile[types.ContextTalkLink](caseId, *talkLinkInitContext)
}

func startNewSession() {

	talkSession := utils.InitSessionChain("", "GPT", "GPT", "GPT", "GPT", indexer)
inLoop:
	for {
		fmt.Println("请输入新问题：")
		var question string
		input, _, err := bufio.NewReader(os.Stdin).ReadLine()
		if err != nil {
			panic(err)
		}
		question = string(input)
		if question == "exit" {
			break inLoop
		}
		if err != nil {
			return
		}
		startTime := time.Now()
		utils.TalkWithSession(question, talkSession, "dify", "dify", "dify", "dify", indexer)
		fmt.Println("结合对话链回答该问题耗时：", time.Since(startTime))
	}
	saveObjToJsonFile[types.ContextTalkLink](-1, *talkLinkInitContext)
}

// 不采用对话链，直接与模型进行交互
func pureStart(llm string) {
	historyQ := make([]string, 0, 100)
	historyA := make([]string, 0, 100)
	reader := bufio.NewReader(os.Stdin)
	res := make(chan any)
	defer close(res)
	for true {
		fmt.Println("请输入新问题：")
		var question string
		input, _, err := reader.ReadLine()
		question = string(input)
		if question == "cmd-quit" {
			return
		}
		if err != nil {
			return
		}
		startTime := time.Now()
		if question == "cmd-regen" {
			question = historyQ[len(historyQ)-1]
			historyQ = historyQ[:len(historyQ)-1]
			historyA = historyA[:len(historyA)-1]
		}
		go utils.SendMultiTalkToLLM(historyQ, historyA, question, llm, res)
		answer := <-res
		fmt.Println("回答: ", answer)
		historyQ = append(historyQ, question)
		historyA = append(historyA, answer.(string))
		fmt.Println("结合对话链回答该问题耗时：", time.Since(startTime))
	}
}

func initTalk() []string {
	return []string{
		"u1-01-01:你好",
		"answer:你好，我是你的智能助手",

		"u1-01-02:世界上最高的山是什么山？",
		"answer:世界上最高的山峰是珠穆朗玛峰（Mount Everest），海拔高度是8,848.86米（29,031.7英尺）",

		"u1-01-03:咖喱是什么？",
		"answer:咖喱是一种多种口味和变化的烹饪调味料或菜肴，通常由香料、植物油、肉类或蔬菜、以及椰奶或番茄酱等成分制成。",

		"u1-01-04:我要怎么爬那座山？",
		"answer:如果你打算攀登珠穆朗玛峰或任何其他高海拔山峰，首先需要做充分的准备和计划，因为这是一项极具挑战性和危险的任务。",

		"u1-01-05:苹果怎么吃？",
		"answer:苹果可以直接洗净后生吃，也可以削皮、切片或切块，作为健康的零食。你还可以用苹果制作果汁、沙拉、派或苹果酱。记得去掉果核和果核周围的部分，因为它们不可食用。",
	}
}

func saveObjToJsonFile[T any](caseId int, obj T) {
	rand.NewSource(time.Now().UnixMilli())
	suffix := time.Now().Format("0102150405")
	// 创建目录
	path := fmt.Sprintf("llama/records_%d_save_%s.json", caseId, suffix)
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		fmt.Printf("创建目录失败: %s", err)
		return
	}

	file, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()
	// 将对象编码为格式化的 JSON
	jsonData, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return
	}

	// 写入格式化的 JSON 数据到文件
	if _, err := file.Write(jsonData); err != nil {
		fmt.Println("Error writing JSON to file:", err)
	} else {
		fmt.Println("Formatted JSON data successfully written to file.")
	}
}
