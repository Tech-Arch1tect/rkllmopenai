package model

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"syscall"
	"unsafe"

	"github.com/tech-arch1tect/rkllmwrapper-go/generated"
)

const (
	ContextSize       = 4096
	MaxNewTokens      = 1024
	DefaultBufferSize = 16384

	RKLLMInputPrompt = 0
	RKLLMInputToken  = 1
)

type ModelRunner struct {
	currentModel string
	logger       *log.Logger
}

func NewModelRunner(logger *log.Logger) *ModelRunner {
	return &ModelRunner{logger: logger}
}

func (r *ModelRunner) Ensure(ctx context.Context, m Model) error {
	if r.currentModel == m.ModelName {
		return nil
	}
	if r.currentModel != "" {
		r.logger.Printf("Destroying previous instance: %s\n", r.currentModel)
		r.Destroy()
	}
	r.logger.Printf("Initialising %s (tokens=%d, ctx=%d)\n", m.ModelPath, MaxNewTokens, ContextSize)
	if ret := generated.Rkllm_init_simple(m.ModelPath, MaxNewTokens, ContextSize); ret != 0 {
		return fmt.Errorf("initialise %s failed: code %d", m.ModelPath, ret)
	}
	r.currentModel = m.ModelName
	return nil
}

func (r *ModelRunner) Run(ctx context.Context, modelName, fifoPath string, msgs []ChatMessage) (string, error) {
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

	if err := r.Ensure(ctx, m); err != nil {
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
