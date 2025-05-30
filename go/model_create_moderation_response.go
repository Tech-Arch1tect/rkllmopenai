/*
 * OpenAI API
 *
 * APIs for sampling from and fine-tuning language models
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package rkllmopenai

type CreateModerationResponse struct {
	Id string `json:"id"`

	Model string `json:"model"`

	Results []CreateModerationResponseResultsInner `json:"results"`
}
