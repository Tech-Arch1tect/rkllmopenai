/*
 * OpenAI API
 *
 * APIs for sampling from and fine-tuning language models
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package rkllmopenai

type CreateEmbeddingResponse struct {
	Object string `json:"object"`

	Model string `json:"model"`

	Data []CreateEmbeddingResponseDataInner `json:"data"`

	Usage CreateEmbeddingResponseUsage `json:"usage"`
}
