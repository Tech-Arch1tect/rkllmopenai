package model

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/tech-arch1tect/tokenizers-cpp-go/manual"
)

type TokenizerConfig struct {
	TokenType      string      `json:"tokenizer_class"`
	ModelMaxLength interface{} `json:"model_max_length"`
}

type Tokenizer struct {
	Model  Model
	Config TokenizerConfig
}

func NewTokenizer(model Model) (*Tokenizer, error) {
	cfg, err := loadTokenizerConfig(model)
	if err != nil {
		return nil, err
	}
	return &Tokenizer{Model: model, Config: cfg}, nil
}

func loadTokenizerConfig(model Model) (TokenizerConfig, error) {
	path := filepath.Join(model.ModelDir, "tokenizer_config.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return TokenizerConfig{}, fmt.Errorf("read config %s: %w", path, err)
	}
	var cfg TokenizerConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return TokenizerConfig{}, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func (t *Tokenizer) Tokenize(messages []ChatMessage) ([]int32, string, error) {
	simple := GetSimplifiedModelName(filepath.Base(t.Model.ModelDir))
	toks, ok := GetSpecialTokens(simple)
	if !ok {
		return nil, "", fmt.Errorf("no special tokens for '%s'", simple)
	}

	var b strings.Builder
	if tok := toks["bos"]; tok != "<nil>" {
		b.WriteString(tok)
	}

	for _, m := range messages {
		rtok := toks[m.Role]
		b.WriteString(rtok)
		b.WriteString(" ")
		b.WriteString(m.Content)

		if eot := toks["eot"]; eot != "<nil>" {
			if !(m.Role == "assistant" && strings.HasPrefix(eot, "<|Assistant|><think>")) {
				b.WriteString(eot)
			}
		}

		b.WriteString("\n")
	}

	if toks["addGenerationPrompt"] == "true" {
		if eot := toks["eot"]; eot != "<nil>" {
			b.WriteString(eot)
		}
	}

	if toks["postGenerationPrompt"] != "<nil>" {
		b.WriteString(toks["postGenerationPrompt"])
	}

	prompt := b.String()

	tokFile := filepath.Join(t.Model.ModelDir, "tokenizer.json")
	data, err := os.ReadFile(tokFile)
	if err != nil {
		return nil, "", fmt.Errorf("read tokenizer.json: %w", err)
	}
	tknzr, err := manual.NewFromJSON(string(data))
	if err != nil {
		return nil, "", err
	}
	defer tknzr.Free()

	ids, err := tknzr.Encode(prompt, false)
	return ids, prompt, err
}

var SpecialTokens = map[string]map[string]string{
	"deepseek_r1": {
		"unk":                 "<nil>",
		"eos":                 "<|end_of_sentence|>",
		"pad":                 "<|pad|>",
		"bos":                 "<|begin_of_sentence|>",
		"bot":                 "<|begin_of_thinking|>",
		"eot":                 "<|Assistant|><think>\n",
		"user":                "<|User|>",
		"assistant":           "<|Assistant|>",
		"system":              "<|System|>",
		"addGenerationPrompt": "true",
	},
	"gemma_3": {
		"unk":                  "<unk>",
		"eos":                  "<eos>",
		"pad":                  "<pad>",
		"bos":                  "<bos>\n",
		"bot":                  "<nil>",
		"eot":                  "\n<end_of_turn>",
		"user":                 "<start_of_turn>user\n",
		"assistant":            "<start_of_turn>model\n",
		"system":               "<start_of_turn>system\n",
		"addGenerationPrompt":  "false",
		"postGenerationPrompt": "<start_of_turn>model\n",
	},
}

func GetSpecialTokens(key string) (map[string]string, bool) {
	if i := strings.Index(key, ":"); i != -1 {
		key = key[:i]
	}
	m, ok := SpecialTokens[key]
	return m, ok
}

func GetSimplifiedModelName(full string) string {
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(full), filepath.Ext(full)))

	qwenRe := regexp.MustCompile(`(?i)(qwen\d*)`)
	qwenMatch := qwenRe.FindStringSubmatch(name)

	switch {
	case regexp.MustCompile(`(?i)deepseek`).MatchString(name):
		if strings.Contains(name, "r1") {
			return "deepseek_r1"
		}
		return "deepseek"
	case len(qwenMatch) > 1:
		family := strings.ToLower(qwenMatch[1])
		if strings.Contains(family, "2") {
			family = "qwen2.5"
		}
		return family
	case regexp.MustCompile(`(?i)mistral`).MatchString(name):
		return "mistral"
	case regexp.MustCompile(`(?i)gemma`).MatchString(name):
		parts := regexp.MustCompile(`[^a-zA-Z0-9]+`).Split(name, -1)
		if len(parts) > 1 && parts[0] == "gemma" {
			return parts[0] + "_" + parts[1]
		}
		return "gemma"
	default:
		parts := regexp.MustCompile(`[^a-zA-Z]+`).Split(name, -1)
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
		return sanitise(name)
	}
}

func sanitise(s string) string {
	re := regexp.MustCompile(`[^a-z0-9]`)
	return re.ReplaceAllString(strings.ToLower(s), "-")
}
