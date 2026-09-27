package utils

import (
	"bufio"
	"context"
	"fmt"
	openai "github.com/sashabaranov/go-openai"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	chatglm_sdk "topic-chain/chatglm_sdk"
	"topic-chain/chatglm_sdk/request"
	"topic-chain/config"
	"topic-chain/dify_sdk"
	. "topic-chain/meta"
	"topic-chain/prompt"
)

var apiRequestCounter int
var retryTimes = 5

var (
	openAIModelName = config.LoadStrFromEnv("OPENAI_MODEL_NAME")
	openAIApiKey    = config.LoadStrFromEnv("OPENAI_API_KEY")
	openAIBaseURL   = config.LoadStrFromEnv("OPENAI_BASE_URL")
)

var usage_input_gpt = 0
var usage_output_gpt = 0

var searchOnly = true

const (
	All   = 0
	Part  = 1
	New   = 2
	Other = 3
)

func SaveUsage() {
	ts := time.Now().UnixMilli()
	filePath := fmt.Sprintf("./record/usage_%d.txt", ts)
	// 检查 filepath 是否存在，不存在的话先创建
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		_, err2 := os.Create(filePath)
		if err2 != nil {
			fmt.Println("创建文件失败：", err2)
			return
		}
	}
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

func sendMsgToLLM(msg string, llm string) string {
	content := ""
	time.Sleep(2 * time.Second)
	var startTime = time.Now()
	if llm == "GPT" {
		apiRequestCounter++
		client := newGPTClient()
		startTime = time.Now()
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model: openAIModelName,
				Messages: []openai.ChatCompletionMessage{
					{
						Role:    openai.ChatMessageRoleUser,
						Content: msg,
					},
				},
			},
		)
		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return ""
		}
		if len(resp.Choices) == 0 {
			fmt.Println("ChatCompletion returned no choices")
			return ""
		}
		usage_input_gpt += resp.Usage.PromptTokens
		usage_output_gpt += resp.Usage.CompletionTokens
		fmt.Printf("问答累计消耗情况：\n输入总量：%d\n输出总量：%d\n", usage_input_gpt, usage_output_gpt)
		content = resp.Choices[0].Message.Content
	} else if strings.HasPrefix(llm, "GLM") {
		apiRequestCounter++
		client := chatglm_sdk.NewGLMClient("")
		if llm == "GLMPro" {
			client = chatglm_sdk.NewGLMClient(request.GLMPro)
		} else if llm == "GLMTurbo" {
			client = chatglm_sdk.NewGLMClient(request.GLMTurbo)
		} else {
			return "不支持的GLM模型类型：" + llm
		}

		client.SendSingleQuestion(msg)
		_, err := client.GetPrevResp()
		if err != nil {
			panic("GLM请求出现错误")
		}
		promptTokens, resultTokens, err := client.GetUsage()
		if err != nil {
			panic("GLM请求出现错误")
		}
		usage_input_gpt += promptTokens
		usage_output_gpt += resultTokens
		content, _ = client.GetPrevResp()
		content = strings.TrimPrefix(content, "\" ")
		content = strings.TrimSuffix(content, "\"")
	} else if llm == "dify" {
		content = dify_sdk.SendMessageToDify(msg)
	} else {
		return llm + "为不支持的模型类型"
	}
	fmt.Println("本次请求内容：", msg)
	fmt.Println("回答：", content)
	fmt.Println("本次 llm 请求耗时：", time.Since(startTime))
	fmt.Println("生成内容长度：", len(content))
	// return resp.Choices[0].Message.Content
	return content
}

func SendSingleTalkToLLM(msg string, llm string) string {
	return sendMsgToLLM(msg, llm)
}

func SendSummaryMsgToLLM(msg string, llm string) string {
	return sendMsgToLLM(msg, llm)
}

func SendJudgeMsgToLLM(msg string, llm string) string {
	return sendMsgToLLM(msg, llm)
}

func SendQTypeMsgToLLM(msg string, llm string) string {
	return sendMsgToLLM(msg, llm)
}

func SendMultiTalkToLLM(historyQuestions []string, historyAnswers []string, curQuestion string, llm string, res chan any) {
	if len(historyQuestions) != len(historyAnswers) {
		panic("SendMultiTalkToLLM-error 历史对话数据长度不一致")
	}
	startTime := time.Now()
	content := ""
	if llm == "GPT" {
		apiRequestCounter++
		client := newGPTClient()
		messages := make([]openai.ChatCompletionMessage, 0, 2*len(historyQuestions)+1)
		for index, _ := range historyQuestions {
			chatMessage := openai.ChatCompletionMessage{
				Role:    "",
				Content: "",
			}
			chatMessage.Role = openai.ChatMessageRoleUser
			chatMessage.Content = historyQuestions[index]
			messages = append(messages, chatMessage)
			chatMessage.Role = openai.ChatMessageRoleAssistant
			chatMessage.Content = historyAnswers[index]
			messages = append(messages, chatMessage)
		}
		messages = append(messages, openai.ChatCompletionMessage{Role: openai.ChatMessageRoleUser, Content: curQuestion})
		startTime = time.Now()
		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:     openAIModelName,
				Messages:  messages,
				MaxTokens: 4096,
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			res <- ""
			return
		}
		if len(resp.Choices) == 0 {
			fmt.Println("ChatCompletion returned no choices")
			res <- ""
			return
		}
		usage_input_gpt += resp.Usage.PromptTokens
		usage_output_gpt += resp.Usage.CompletionTokens
		fmt.Printf("问答累计消耗情况：\n输入总量：%d\n输出总量：%d\n", usage_input_gpt, usage_output_gpt)
		content = resp.Choices[0].Message.Content
	} else if strings.HasPrefix(llm, "GLM") {
		apiRequestCounter++
		client := chatglm_sdk.NewGLMClient("")
		if llm == "GLMPro" {
			client = chatglm_sdk.NewGLMClient(request.GLMPro)
		} else if llm == "GLMTurbo" {
			client = chatglm_sdk.NewGLMClient(request.GLMTurbo)
		} else if llm == "GLM40520" {
			client = chatglm_sdk.NewGLMClient(request.GLM40520)
		} else {
			res <- "不支持的GLM模型类型：" + llm
			return
		}
		if len(historyQuestions) == 0 {
			client.SendSingleQuestion(curQuestion)
		} else {
			client.SendMultiTalk(historyQuestions, historyAnswers, curQuestion)
		}
		_, err := client.GetPrevResp()
		if err != nil {
			panic("GLM请求出现错误")
		}
		content, _ = client.GetPrevResp()
		inputToken, outputToken, err := client.GetUsage()
		fmt.Println("本次token耗费为：", inputToken, outputToken)
		content = strings.TrimPrefix(content, "\" ")
		content = strings.TrimSuffix(content, "\"")
	} else {
		res <- llm + "为不支持的模型类型"
	}

	fmt.Println("本次 llm 请求耗时：", time.Since(startTime))
	if len(content) == 0 {
		fmt.Println("stop!")
	}
	fmt.Println("多轮对话生成内容：", content)
	fmt.Println("生成内容长度：", len(content))
	// return resp.Choices[0].Message.Content
	res <- content
}

// 初始化本次对话的链数据
// initContent输入样例：
// user: q1
// assistant: a1
// user: q2
// assistant: a2
// ......
func InitSessionChain(initContent string, talkLLM, judgeLLM, qTypeLLM, summaryLLM string, indexerMode string) *Session {
	globalId := 0
	startTime := time.Now()
	initSession := NewSession()
	userTurns := strings.Split(initContent, "user: ")
	for index, val := range userTurns {
		if val != "" {
			val = strings.Split(val, "\nassistant")[0]
			userTurns[index] = val
		}
	}
	userTurns = userTurns[1:]
	aiTurns := strings.Split(initContent, "assistant: ")
	for index, val := range aiTurns {
		if !strings.HasPrefix(val, "user: ") {
			val = strings.Split(val, "\nuser")[0]
			aiTurns[index] = val
		}
	}
	aiTurns = aiTurns[1:]
	// fmt.Println(userTurns, aiTurns)
	if len(userTurns) != len(aiTurns) {
		panic("初始对话格式有误")
	}
	for index, _ := range userTurns {
		// 判断该对话是否属于某一个已存在的对话主题
		flagIndex := searchMatchedTopicsIndex(initSession, userTurns[index], judgeLLM, indexerMode)
		// 若本次提问不属于已有的某个链，则为该提问创建新链
		if flagIndex == -1 {
			curChain := NewTalkChain()
			// 总结对话主题
			content := buildQAPair(userTurns[index], aiTurns[index])
			input := prompt.BuildSummaryPrompt(content)
			topic := SendSummaryMsgToLLM(input, summaryLLM)
			if strings.TrimSpace(topic) == "" {
				fmt.Println("failed to initialize topic chain: the LLM returned no topic")
				return initSession
			}

			curChain.Topic = topic
			// 将对话纳入进新链中
			curChain.Question = append(curChain.Question, userTurns[index])
			curChain.Answer = append(curChain.Answer, aiTurns[index])
			curChain.Total++
			curChain.Id = globalId

			if indexerMode == "vector-mode" {
				UpsertVectorToQDrant(globalId, EmbeddingText(topic), topic)
			}
			globalId++
			// 更新对话链索引顺序（最新的索引在链表的最前面）
			node := &SingLinkList{
				Val:  initSession.Total,
				Next: nil,
			}
			initSession.OrderedIndexList.HeadInsert(node)
			initSession.Chains = append(initSession.Chains, curChain)
			initSession.Topics = append(initSession.Topics, topic)
			initSession.Total++

		} else {
			// 如果当前命中了已经存在的话题链则进行相关信息提取
			curChain := initSession.Chains[flagIndex]
			curChain.Question = append(curChain.Question, userTurns[index])
			curChain.Answer = append(curChain.Answer, aiTurns[index])
			// 更新对话链主题
			// content := buildMultiQAPair(curChain.Question, curChain.Answer)
			history := curChain.Topic
			curQuestion := userTurns[index]
			curAnswer := aiTurns[index]
			input := prompt.BuildHistoryAppendSummaryPrompt(history, curQuestion, curAnswer)
			newTopic := SendSummaryMsgToLLM(input, summaryLLM)
			if strings.TrimSpace(newTopic) == "" {
				fmt.Println("failed to update topic chain: the LLM returned no topic")
				return initSession
			}
			curChain.Topic = newTopic
			if indexerMode == "vector-mode" {
				UpsertVectorToQDrant(flagIndex, EmbeddingText(newTopic), newTopic)
			}
			curChain.Total++
			initSession.OrderedIndexList.ChangeNodeToHead(initSession.OrderedIndexList.SearchNodeByVal(flagIndex))
			initSession.Topics[flagIndex] = newTopic
		}
	}
	fmt.Println("初始化链花费时间为：", time.Since(startTime))
	fmt.Println("总请求次数为：", apiRequestCounter)
	fmt.Println("链初始化状态：")
	fmt.Println("已有对话链主题数：", initSession.Total)
	fmt.Println("对话链主题：")
	for _, val := range initSession.Topics {
		fmt.Println(val)
	}
	return initSession
}

func buildQAPair(question, answer string) string {
	builder := strings.Builder{}
	builder.WriteString("user: ")
	builder.WriteString(question)
	builder.WriteString("\n")
	builder.WriteString("assistant: ")
	builder.WriteString(answer)
	builder.WriteString("\n")
	return builder.String()
}

func buildMultiQAPair(question, answer []string) string {
	builder := strings.Builder{}
	if len(question) != len(answer) {
		panic("buildMultiQAPair-error QA数组长度不一致")
	}
	for index, _ := range question {
		builder.WriteString("user: ")
		builder.WriteString(question[index])
		builder.WriteString("\n")
		builder.WriteString("assistant: ")
		builder.WriteString(answer[index])
		builder.WriteString("\n")
	}
	return builder.String()
}

func TalkWithSession(question string, session *Session, talkLLM, judgeLLM, qTypeLLM, summaryLLM string, indexerMode string) (answerType int, answer string, chainId int, chainQuestions []string, err error) {
	// 限制对话生成内容长度
	// limitContentLengthPrompt := "\n回答上述问题时注意生成的字数，务必限制在200字内完成回答！注意不要用列举的方式回答问题！简洁与凝练的回答比长篇大论更有效！"
	limitContentLengthPrompt := ""
	// 判断提问是否需要结合上下文 - all（全部）- part（部分）- one（单独）
	questionType := getQuestionType(session, question, qTypeLLM)
	flagIndex := -999
	// 确定构造上下文的方式
	if questionType == "one" {
		flagIndex = -1
	} else if questionType == "part" {
		// 查找新提出的问题是否属于某一条已存在的链
		flagIndex = searchMatchedTopicsIndex(session, question, judgeLLM, indexerMode)
	} else if questionType == "all" {
		flagIndex = -2
	}
	if flagIndex == -1 {
		if searchOnly {
			fmt.Printf("当前不选择上下文进行回答")
			return New, "", -1, []string{}, nil
		}
		// 该问题不属于已有的对话链，为其新建链
		curChain := NewTalkChain()
		answer := SendSingleTalkToLLM(question+limitContentLengthPrompt, talkLLM)

		// 总结对话主题
		content := buildQAPair(question, answer)
		input := prompt.BuildSummaryPrompt(content)
		topic := SendSummaryMsgToLLM(input, summaryLLM)
		curChain.Topic = topic
		// 将对话纳入进新链中
		curChain.Question = append(curChain.Question, question)
		curChain.Answer = append(curChain.Answer, answer)
		curChain.Total++
		session.Chains = append(session.Chains, curChain)
		session.Topics = append(session.Topics, topic)
		// 更新对话链索引顺序（最新的索引在链表的最前面）
		node := &SingLinkList{
			Val:  session.Total,
			Next: nil,
		}
		session.OrderedIndexList.HeadInsert(node)
		session.Total++
	} else if flagIndex != -2 && flagIndex != -999 {
		// 该问题属于已有的某条链，为其构造富集上下文内容的多轮对话 prompt
		curChain := session.Chains[flagIndex]
		if searchOnly {
			fmt.Printf("当前选中的链为：%v", curChain)
			return Part, "", curChain.Id, curChain.Question, nil
		}
		res := make(chan any)
		defer close(res)
		go SendMultiTalkToLLM(curChain.Question, curChain.Answer, question+limitContentLengthPrompt, talkLLM, res)
		answer := (<-res).(string)
		fmt.Println("llm 回答为：", answer)
		curChain.Question = append(curChain.Question, question)
		curChain.Answer = append(curChain.Answer, answer)
		// 更新对话链主题
		content := buildMultiQAPair(curChain.Question, curChain.Answer)
		input := prompt.BuildSummaryPrompt(content)
		newTopic := SendSummaryMsgToLLM(input, summaryLLM)
		curChain.Topic = newTopic
		curChain.Total++
		session.OrderedIndexList.ChangeNodeToHead(session.OrderedIndexList.SearchNodeByVal(flagIndex))
		session.Topics[flagIndex] = newTopic
		return Part, "", curChain.Id, curChain.Question, nil
	} else if flagIndex == -2 {
		if searchOnly {
			fmt.Printf("当前需要进行全文回答")
			return All, "", -1, []string{}, nil
		}
		historyQuestion := make([]string, 0, 100)
		historyAnswer := make([]string, 0, 100)
		newChain := NewTalkChain()
		for _, curChain := range session.Chains {
			historyQuestion = append(historyQuestion, curChain.Question...)
			historyAnswer = append(historyAnswer, curChain.Answer...)
		}
		res := make(chan any)
		defer close(res)
		go SendMultiTalkToLLM(historyQuestion, historyAnswer, question+limitContentLengthPrompt, talkLLM, res)
		answer := (<-res).(string)
		fmt.Println("llm 回答为：", answer)
		// 总结对话主题
		content := buildQAPair(question, answer)
		input := prompt.BuildSummaryPrompt(content)
		topic := SendSummaryMsgToLLM(input, summaryLLM)
		newChain.Topic = topic
		// 将对话纳入进新链中
		newChain.Question = append(newChain.Question, question)
		newChain.Answer = append(newChain.Answer, answer)
		newChain.Total++
		session.Chains = append(session.Chains, newChain)
		session.Topics = append(session.Topics, topic)
		// 更新对话链索引顺序（最新的索引在链表的最前面）
		node := &SingLinkList{
			Val:  session.Total,
			Next: nil,
		}
		session.OrderedIndexList.HeadInsert(node)
		session.Total++
	}

	return Other, "", -1, []string{}, nil
}

func searchMatchedTopicsIndex(session *Session, question string, judgeLLM string, indexerMode string) int {
	if session.Total == 0 {
		return -1
	}
	if indexerMode == "vector-mode" {
		query_vector := EmbeddingText(question)
		top1Id, topic := TopicSimilaritySearchFromQDrant(query_vector)
		log.Println("选中了主题：", topic)
		return top1Id
	}
	indexList := session.OrderedIndexList.Next

	for indexList != nil {
		index := indexList.Val.(int)
		content := "主题: " + session.Topics[index] + "\n提问: " + question
		input := prompt.BuildAssociationPrompt(content)
		for i := 0; i < retryTimes; i++ {
			flag := SendJudgeMsgToLLM(input, judgeLLM)
			flag = strings.TrimSpace(flag)

			valid, _ := prompt.CheckResultValidity(flag, func(content string) bool {
				// 检查输入是否包含[Reasoning]标记
				reasoningStart := strings.Index(content, "[Reasoning]")
				if reasoningStart == -1 {
					return false
				}
				// 检查输入是否包含[Judge]标记
				judgeStart := strings.Index(content, "[Judge]")
				if judgeStart == -1 {
					return false
				}
				// 检查关键内容是否属于期望字符串
				judgeContent := strings.TrimSpace(content[judgeStart+len("[Judge]"):])
				if judgeContent != "yes" && judgeContent != "no" {
					return false
				}
				return true
			})
			if valid {
				judgeStart := strings.Index(flag, "[Judge]")
				judgeContent := strings.TrimSpace(flag[judgeStart+len("[Judge]"):])
				if judgeContent == "yes" {
					return index
				} else {
					break
				}
			} else {
				fmt.Println("生成格式有误，尝试重新生成 - ", i)
			}
		}
		indexList = indexList.Next
	}
	return -1
}

func getQuestionType(session *Session, question string, qTypeLLM string) string {
	if session.Total == 0 {
		return "one"
	}
	content := "提问: " + question
	input := prompt.BuildQuestionTypePrompt(content)
	for i := 0; i < retryTimes; i++ {
		flag := SendQTypeMsgToLLM(input, qTypeLLM)
		flag = strings.TrimSpace(flag)
		valid, _ := prompt.CheckResultValidity(flag, func(content string) bool {
			// 检查输入是否包含[Reasoning]标记
			reasoningStart := strings.Index(content, "[Reasoning]")
			if reasoningStart == -1 {
				return false
			}
			// 检查输入是否包含[Judge]标记
			judgeStart := strings.Index(content, "[Judge]")
			if judgeStart == -1 {
				return false
			}
			// 检查关键内容是否属于期望字符串
			judgeContent := strings.TrimSpace(content[judgeStart+len("[Judge]"):])
			if judgeContent != "one" && judgeContent != "part" && judgeContent != "all" {
				return false
			}
			return true
		})
		if valid {
			judgeStart := strings.Index(flag, "[Judge]")
			return strings.TrimSpace(flag[judgeStart+len("[Judge]"):])
		} else {
			fmt.Println("生成格式有误，尝试重新生成 - ", i)
		}
	}
	personalResult := "one"
	fmt.Println("重生成次数达上限，尝试直接返回预设结果: ", personalResult)
	return personalResult
}

func newGPTClient() *openai.Client {
	openAIConf := openai.DefaultConfig(openAIApiKey)
	if openAIBaseURL != "" {
		openAIConf.BaseURL = openAIBaseURL
	}
	client := openai.NewClientWithConfig(openAIConf)
	return client
}
