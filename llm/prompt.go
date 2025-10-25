package llm

import (
	"fmt"
	"strings"
)

func GeneratePrompt(msgs []string) string {
	return fmt.Sprintf("你是一个智能助手，你的任务是根据用户的问题生成符合要求的文本。问题：%s", strings.Join(msgs, " "))
}
