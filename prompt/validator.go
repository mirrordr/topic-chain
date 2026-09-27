package prompt

import (
	"errors"
	"strings"
)

// CheckResultValidity 用来判断 llm 的输出结果是否符合预期要求
func CheckResultValidity(res string, validator func(content string) bool) (bool, error) {
	if strings.TrimSpace(res) == "" {
		return false, errors.New("传入待校验的值为空字符串")
	}
	return validator(res), nil
}
