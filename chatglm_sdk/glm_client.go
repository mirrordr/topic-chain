package chatglm_sdk

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"topic-chain/chatglm_sdk/request"
	"topic-chain/chatglm_sdk/response"
	"topic-chain/chatglm_sdk/utils"
)

type GLMClient struct {
	modelName        string
	requestParameter *request.GLMParameter
	result           *response.GLMResult
}

func NewGLMClient(modelName string) *GLMClient {
	client := &GLMClient{
		modelName: modelName,
		requestParameter: &request.GLMParameter{
			Prompt:      make([]request.GLMPrompt, 0, 5),
			Temperature: 0.95,
			TopP:        0.7,
			RequestId:   "",
			ReturnType:  "json_string",
			Ref: request.GLMRef{
				Enable: true,
			},
		},
		result: &response.GLMResult{
			Code:    0,
			Msg:     "未初始化",
			Success: false,
			Data:    response.GLMData{},
		},
	}
	return client
}

func (client GLMClient) SendSingleQuestion(question string) response.GLMResult {
	prompt := request.GLMPrompt{
		Role:    "user",
		Content: question,
	}
	client.requestParameter.Prompt = []request.GLMPrompt{prompt}
	bytesData, _ := json.Marshal(client.requestParameter)
	httpClient := &http.Client{}
	httpRequest, err := http.NewRequest("POST", client.modelName, bytes.NewReader(bytesData))
	httpRequest.Header.Add("Content-Type", "application/json")
	httpRequest.Header.Add("Authorization", utils.NewGLMToken())
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// resp, err := http.Post(client.modelName, "application/json", bytes.NewReader(bytesData))
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		err := json.Unmarshal(body, client.result)
		if err != nil {
			return response.GLMResult{}
		}
	}
	return *client.result
}

func (client GLMClient) SendMultiTalk(historyQuestion []string, historyAnswer []string, curQuestion string) response.GLMResult {
	if len(historyQuestion) != len(historyAnswer) {
		panic("GLM-多轮对话长度不符！")
	}
	prompts := make([]request.GLMPrompt, 0, 10)
	if len(historyQuestion) > 0 {
		for index, _ := range historyQuestion {
			questionPrompt := request.GLMPrompt{
				Role:    "user",
				Content: historyQuestion[index],
			}
			answerPrompt := request.GLMPrompt{
				Role:    "assistant",
				Content: historyAnswer[index],
			}
			prompts = append(prompts, questionPrompt)
			prompts = append(prompts, answerPrompt)
		}
		prompts = append(prompts, request.GLMPrompt{
			Role:    "user",
			Content: curQuestion,
		})
	}
	client.requestParameter.Prompt = prompts
	bytesData, _ := json.Marshal(client.requestParameter)
	httpClient := &http.Client{}
	httpRequest, err := http.NewRequest("POST", client.modelName, bytes.NewReader(bytesData))
	httpRequest.Header.Add("Content-Type", "application/json")
	httpRequest.Header.Add("Authorization", utils.NewGLMToken())
	resp, err := httpClient.Do(httpRequest)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// resp, err := http.Post(client.modelName, "application/json", bytes.NewReader(bytesData))
	if err == nil {
		body, _ := io.ReadAll(resp.Body)
		err := json.Unmarshal(body, client.result)
		if err != nil {
			return response.GLMResult{}
		}
	}
	return *client.result

}

func (client GLMClient) GetPrevResp() (string, error) {
	if client.result.Success == true {
		return client.result.Data.Choices[0].Content, nil
	} else {
		return "", errors.New("上次请求出现错误：" + client.result.Msg)
	}
}

func (client GLMClient) GetUsage() (int, int, error) {
	if client.result.Success == true {
		return client.result.Data.Usage.PromptTokens, client.result.Data.Usage.CompletionTokens, nil
	} else {
		return -1, -1, errors.New("上次请求出现错误：" + client.result.Msg)
	}
}
