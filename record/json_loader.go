package record

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
)

type Record struct {
	TurnID  int    `json:"turn_id"`
	User    string `json:"user"`
	Content string `json:"content"`
}

type Case struct {
	CaseID        int        `json:"case_id"`
	Meta          Meta       `json:"meta"`
	Records       []Record   `json:"records"`
	OpenQuestions []Question `json:"open_questions"`
	// RecordsLink   []string   `json:"records_link"`
}

type Meta struct {
	User1              User   `json:"user_1"`
	User2              User   `json:"user_2"`
	ConversationDomain string `json:"conversation_domain"`
}

type User struct {
	Name        string `json:"name"`
	Personality string `json:"personality"`
}

type Question struct {
	Question       string `json:"question"`
	ExpectedAnswer string `json:"expected_answer"`
}

type RoleMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	TurnId  int    `json:"turn_id"`
}

func LoadData(filePath string) ([]Case, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	bytes, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var data []Case
	if err := json.Unmarshal(bytes, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func LoadTalkingRecordsFromLocalJsonFile(filePath string, caseId int) ([]RoleMessage, []RoleMessage, error) {
	data, err := LoadData(filePath)
	if err != nil {
		return nil, nil, err
	}

	var userMessages []RoleMessage
	var aiMessages []RoleMessage

	if len(data) == 0 {
		return userMessages, aiMessages, nil
	}
	talkingRecords := data[caseId].Records

	for i := 0; i < len(talkingRecords); i += 2 {
		if i+1 >= len(talkingRecords) {
			break
		}
		messageQ := RoleMessage{Role: "user", Content: talkingRecords[i].Content, TurnId: talkingRecords[i].TurnID}
		userMessages = append(userMessages, messageQ)
		messageA := RoleMessage{Role: "assistant", Content: talkingRecords[i+1].Content, TurnId: talkingRecords[i].TurnID}
		aiMessages = append(aiMessages, messageA)
	}

	return userMessages, aiMessages, nil
}

// Sentence2TurnIds
// sentences 代表一条链下的所有 Question
// allRecords 代表数据集的原始记录
func Sentence2TurnIds(sentences []string, allRecords []RoleMessage) string {
	turnIds := strings.Builder{}
	for i := 0; i < len(sentences); i++ {
		sentence := sentences[i]
		for j := 0; j < len(allRecords); j++ {
			if allRecords[j].Content == sentence {
				turnIds.WriteString(strconv.Itoa(allRecords[j].TurnId))
				break
			}
		}
		if i != len(sentences)-1 {
			turnIds.WriteString(",")
		}
	}
	return turnIds.String()
}
