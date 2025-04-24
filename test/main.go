package main

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

func main() {
	cfg := openai.DefaultConfig("")
	cfg.BaseURL = "http://localhost:8080/v1"

	client := openai.NewClientWithConfig(cfg)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: "DeepSeek-R1-Distill-Qwen-7B-rk3588-w8a8-opt-1-hybrid-ratio-0.0.rkllm",
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    "system",
					Content: "You are a helpful assistant that can answer questions and help with tasks.",
				},
				{
					Role:    "user",
					Content: "Hello, how are you? Tell me a joke. Keep it short and simple.",
				},
			},
			MaxTokens: 1024,
		},
	)
	if err != nil {
		panic(err)
	}
	fmt.Println(resp)
}
