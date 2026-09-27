package prompt

import (
	"bytes"
	"log"
	"text/template"
)

func BuildSummaryPrompt(content string) string {
	tmpl, _ := template.New("总结对话内容").Parse(Summary)
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, content)
	if err != nil {
		log.Fatal(err)
		return "tmpl exec error"
	}
	return buf.String()
}

func BuildHistoryAppendSummaryPrompt(history string, curQuestion, curAnswer string) string {
	tmpl, _ := template.New("追加总结对话内容").Parse(AppendSummary)
	content := "- 先前的对话记录摘要为:\n" + history + "\n\n- 本轮新增的QA对为:\nuser: " + curQuestion + "\nassistant: " + curAnswer

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, content)
	if err != nil {
		log.Fatal(err)
		return "tmpl exec error"
	}
	return buf.String()
}

func BuildAssociationPrompt(content string) string {
	tmpl, _ := template.New("判断对话关联性").Parse(AssociationCOT)

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, content)
	if err != nil {
		log.Fatal(err)
		return "tmpl exec error"
	}
	// fmt.Printf(buf.String())
	return buf.String()
}

func BuildQuestionTypePrompt(content string) string {
	tmpl, _ := template.New("判断提问类型").Parse(QuestionTypeCOT)
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, content)
	if err != nil {
		log.Fatal(err)
		return "tmpl exec error"
	}
	// fmt.Printf(buf.String())
	return buf.String()
}
