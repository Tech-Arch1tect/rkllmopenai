package model

type Model struct {
	ModelName string
	ModelPath string
	ModelDir  string
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
