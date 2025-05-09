/*
 * OpenAI API
 *
 * APIs for sampling from and fine-tuning language models
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package rkllmopenai

type ChatCompletionResponseMessage struct {

	// The role of the author of this message.
	Role string `json:"role"`

	// The contents of the message.
	Content *string `json:"content,omitempty"`

	FunctionCall ChatCompletionRequestMessageFunctionCall `json:"function_call,omitempty"`
}
