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
	TokenType      string `json:"tokenizer_class"`
	ModelMaxLength int    `json:"model_max_length"`
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
		if m.Role != "assistant" {
			if eot := toks["eot"]; eot != "<nil>" && !strings.HasPrefix(eot, "<|Assistant|><think>") {
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

	variants := []string{}
	variantPatterns := map[string]*regexp.Regexp{
		"coder":    regexp.MustCompile(`(?i)(?:^|[-_\s])coder(?:$|[-_\s])`),
		"math":     regexp.MustCompile(`(?i)(?:^|[-_\s])math(?:$|[-_\s])`),
		"chat":     regexp.MustCompile(`(?i)(?:^|[-_\s])chat(?:$|[-_\s])`),
		"instruct": regexp.MustCompile(`(?i)(?:^|[-_\s])instruct(?:$|[-_\s])`),
	}
	for tag, re := range variantPatterns {
		if re.MatchString(name) {
			variants = append(variants, tag)
		}
	}

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
		return appendSuffix(family, name)
	case regexp.MustCompile(`(?i)mistral`).MatchString(name):
		return appendSuffix("mistral", name)
	default:
		parts := regexp.MustCompile(`[^a-zA-Z]+`).Split(name, -1)
		if len(parts) > 0 && parts[0] != "" {
			return appendSuffix(strings.ToLower(parts[0]), name)
		}
		return sanitise(name)
	}
}

func appendSuffix(base, name string) string {
	if p := regexp.MustCompile(`(?i)(\d+\.?\d*)B`).FindStringSubmatch(name); len(p) > 1 {
		base += ":" + strings.ToLower(p[1]) + "b"
	}
	return base
}

func sanitise(s string) string {
	re := regexp.MustCompile(`[^a-z0-9]`)
	return re.ReplaceAllString(strings.ToLower(s), "-")
}
