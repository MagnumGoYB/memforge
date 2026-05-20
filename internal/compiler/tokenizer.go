package compiler

import (
	"strings"

	tiktoken "github.com/pkoukk/tiktoken-go"
)

var tokenizer *tiktoken.Tiktoken

func init() {
	enc, err := tiktoken.GetEncoding("cl100k_base")
	if err == nil {
		tokenizer = enc
	}
}

type CountResult struct {
	Tokens         int
	UsedFallback   bool
	WarningMessage string
}

func CountTokens(content string) CountResult {
	content = strings.TrimSpace(content)
	if content == "" {
		return CountResult{}
	}
	if tokenizer != nil {
		return CountResult{Tokens: len(tokenizer.Encode(content, nil, nil))}
	}
	runes := len([]rune(content))
	tokens := runes/4 + 8
	if tokens < 1 {
		tokens = 1
	}
	return CountResult{Tokens: tokens, UsedFallback: true, WarningMessage: "tokenizer fallback in use (chars/4 heuristic)"}
}
