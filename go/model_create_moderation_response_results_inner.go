/*
 * OpenAI API
 *
 * APIs for sampling from and fine-tuning language models
 *
 * API version: 2.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package rkllmopenai

type CreateModerationResponseResultsInner struct {
	Flagged bool `json:"flagged"`

	Categories CreateModerationResponseResultsInnerCategories `json:"categories"`

	CategoryScores CreateModerationResponseResultsInnerCategoryScores `json:"category_scores"`
}
