package types

type ContextTalkLink struct {
	CaseId           int              `json:"case_id"`
	InitialTopics    []InitialTopic   `json:"initial_topics"`
	QuestionClassify []ClassifyResult `json:"question_classify"`
}

type InitialTopic struct {
	ChainId int    `json:"chain_id"`
	Topic   string `json:"topic"`
	TurnIds string `json:"turn_ids"`
}

type ClassifyResult struct {
	Question    string `json:"question"`
	TargetChain string `json:"target_chain"`
}

type DataSetCase struct {
	CaseId        int                `json:"case_id"`
	Meta          MetaInfo           `json:"meta"`
	Records       []DataSetRecord    `json:"records"`
	OpenQuestions []OpenQuestionInfo `json:"open_questions"`
	RecordsLink   []string           `json:"records_link"`
}

type MetaInfo struct {
	User1              RoleInfo `json:"user_1"`
	User2              RoleInfo `json:"user_2"`
	ConversationDomain string   `json:"conversation_domain"`
}

type RoleInfo struct {
	Name        string `json:"name"`
	Personality string `json:"personality"`
}

type DataSetRecord struct {
	TurnId  int    `json:"turn_id"`
	User    string `json:"user"`
	Content string `json:"content"`
}

type OpenQuestionInfo struct {
	Question       string `json:"question"`
	ExpectedAnswer string `json:"expected_answer"`
	ChosenTurn     string `json:"chosen_turn"`
}
