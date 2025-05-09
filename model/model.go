package model

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"

	"github.com/Tech-Arch1tect/rkllmopenai/config"
	"github.com/tech-arch1tect/rkllmwrapper-go/generated"
)

type ModelSettings struct {
	Temperature      float32
	MaxNewTokens     int32
	TopP             float32
	FrequencyPenalty float32
	PresencePenalty  float32
}

const (
	ContextSize       = 4096
	DefaultBufferSize = 16384

	RKLLMInputPrompt int32 = 0
	RKLLMInputToken  int32 = 1
)

type ModelRunner struct {
	currentSettings ModelSettings
	currentModel    string
	logger          *log.Logger
}

func NewModelRunner(logger *log.Logger) *ModelRunner {
	return &ModelRunner{logger: logger}
}

func (r *ModelRunner) Ensure(ctx context.Context, m Model, settings ModelSettings) error {
	if r.currentModel == m.ModelName && r.currentSettings == settings {
		return nil
	}
	if r.currentModel != "" {
		r.logger.Printf("Destroying previous instance: %s\n", r.currentModel)
		r.Destroy()
	}
	opts := generated.RkllmOptions{
		Max_new_tokens:     settings.MaxNewTokens,
		Max_context_len:    ContextSize,
		Num_cpus:           int32(config.C.NumCPUs),
		Temperature:        settings.Temperature,
		Top_p:              settings.TopP,
		Frequency_penalty:  settings.FrequencyPenalty,
		Presence_penalty:   settings.PresencePenalty,
		Skip_special_token: 1,
	}
	r.currentSettings = settings
	r.logger.Printf("Initialising %s with opts %+v\n", m.ModelPath, opts)
	ret := generated.Rkllmwrapper_init(m.ModelPath, []generated.RkllmOptions{opts})
	if ret != 0 {
		return fmt.Errorf("initialise %s failed: code %d", m.ModelPath, ret)
	}
	r.currentModel = m.ModelName
	return nil
}

func (r *ModelRunner) Run(ctx context.Context, modelName, fifoPath string, msgs []ChatMessage, settings ModelSettings) (string, error) {
	r.logger.Printf("Running model %s with %d messages", modelName, len(msgs))

	RefreshModelList()
	var m Model
	for _, mm := range ModelList {
		if mm.ModelName == modelName {
			m = mm
			break
		}
	}
	if m.ModelPath == "" {
		return "", fmt.Errorf("model %q not found", modelName)
	}

	if err := r.Ensure(ctx, m, settings); err != nil {
		return "", err
	}

	if fifoPath != "" {
		if err := EnsureFifo(fifoPath); err != nil {
			return "", fmt.Errorf("failed to setup FIFO: %w", err)
		}
	}

	tokenizer, err := NewTokenizer(m)
	if err != nil {
		return "", fmt.Errorf("failed to create tokenizer: %w", err)
	}
	tokenised, prompt, err := tokenizer.Tokenize(msgs)
	if err != nil {
		return "", fmt.Errorf("failed to tokenize messages: %w", err)
	}
	r.logger.Println("Prompt:", prompt)

	buf := make([]byte, DefaultBufferSize)
	ret := generated.Rkllm_run_ex(
		unsafe.Pointer(&tokenised[0]),
		RKLLMInputToken,
		buf,
		int32(len(buf)),
		uint64(len(tokenised)),
		fifoPath,
	)
	if ret != 0 {
		return "", fmt.Errorf("inference error: code %d", ret)
	}
	return string(bytes.TrimRight(buf, "\x00")), nil
}

func EnsureFifo(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return syscall.Mkfifo(path, 0o666)
	}
	return nil
}

func (r *ModelRunner) Destroy() {
	r.logger.Println("Destroying instance:", r.currentModel)
	generated.Rkllm_destroy_simple()
	r.currentModel = ""
}
