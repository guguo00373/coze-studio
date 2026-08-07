/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/bizpkg/llm/modelbuilder"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
)

const (
	defaultEvalPrompt = `You are an expert evaluator. Score the following assistant response
on a scale from 0 to 100, considering correctness, completeness and faithfulness
to the expected answer. Return ONLY a JSON object in the form:
{"score": <number 0-100>, "reasoning": "<brief explanation>"}

Input:
{{input}}

Expected answer:
{{expected}}

Actual output:
{{output}}
`
)

type evaluatorExecutor struct{}

func NewEvaluatorExecutor() *evaluatorExecutor { return &evaluatorExecutor{} }

// Evaluate runs a single evaluator against one target run result and returns
// the parsed score plus reasoning.
func (e *evaluatorExecutor) Evaluate(ctx context.Context, evaluator *entity.Evaluator, runResult *entity.TargetRunResult) (*entity.EvaluatorResult, error) {
	if evaluator == nil {
		return nil, fmt.Errorf("evaluator is nil")
	}
	prompt := evaluator.Prompt
	if prompt == "" {
		prompt = defaultEvalPrompt
	}
	prompt = renderPrompt(prompt, runResult)

	modelID, err := strconv.ParseInt(evaluator.ModelID, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid model_id %q: %w", evaluator.ModelID, err)
	}

	params := &modelbuilder.LLMParams{}
	if evaluator.Temperature != 0 {
		t := float32(evaluator.Temperature)
		params.Temperature = &t
	}

	bcm, _, err := modelbuilder.BuildModelByID(ctx, modelID, params)
	if err != nil {
		return nil, fmt.Errorf("build evaluator model failed: %w", err)
	}

	resp, err := bcm.Generate(ctx, []*schema.Message{
		schema.UserMessage(prompt),
	})
	if err != nil {
		return nil, fmt.Errorf("evaluator generate failed: %w", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("evaluator returned empty response")
	}

	return parseEvaluatorOutput(resp.Content)
}

func renderPrompt(template string, result *entity.TargetRunResult) string {
	replacer := strings.NewReplacer(
		"{{input}}", result.Input,
		"{{expected}}", result.Expected,
		"{{output}}", result.ActualOutput,
	)
	rendered := replacer.Replace(template)
	if !strings.Contains(template, "{{input}}") &&
		!strings.Contains(template, "{{expected}}") &&
		!strings.Contains(template, "{{output}}") {
		// The custom prompt does not reference the case data, so the model has
		// nothing to evaluate. Append the standard context block plus an
		// explicit JSON output instruction.
		rendered += fmt.Sprintf(`
### Case
Input:
%s

Expected answer:
%s

Actual output:
%s

Score the assistant response on a scale from 0 to 100 considering correctness, completeness and faithfulness to the expected answer. Return ONLY a JSON object in the form: {"score": <number 0-100>, "reasoning": "<brief explanation>"}`, result.Input, result.Expected, result.ActualOutput)
	}
	return rendered
}

func parseEvaluatorOutput(content string) (*entity.EvaluatorResult, error) {
	content = strings.TrimSpace(content)
	// Strip markdown code fences if present.
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	parsed, err := unmarshalEvalJSON(content)
	if err != nil {
		// The model may wrap the JSON object with prose or reasoning text;
		// fall back to extracting the first balanced object.
		if extracted := extractJSONObject(content); extracted != "" {
			parsed, err = unmarshalEvalJSON(extracted)
		}
		if err != nil {
			return nil, fmt.Errorf("evaluator output is not valid JSON: %w", err)
		}
	}

	score, err := parseScore(parsed.Score)
	if err != nil {
		return nil, err
	}
	return &entity.EvaluatorResult{
		Score:     score,
		Reasoning: parsed.Reasoning,
	}, nil
}

func unmarshalEvalJSON(content string) (*struct {
	Score     json.RawMessage `json:"score"`
	Reasoning string          `json:"reasoning"`
}, error) {
	var parsed struct {
		Score     json.RawMessage `json:"score"`
		Reasoning string          `json:"reasoning"`
	}
	if err := json.Unmarshal([]byte(content), &parsed); err != nil {
		return nil, err
	}
	return &parsed, nil
}

// extractJSONObject returns the first balanced {...} object in s, ignoring
// braces inside quoted string values. It returns "" if no object is found.
func extractJSONObject(s string) string {
	start := -1
	depth := 0
	inString := false
	escaped := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case inString:
			switch {
			case escaped:
				escaped = false
			case c == '\\':
				escaped = true
			case c == '"':
				inString = false
			}
		case c == '"':
			inString = true
		case c == '{':
			if depth == 0 {
				start = i
			}
			depth++
		case c == '}':
			depth--
			if depth == 0 && start >= 0 {
				return s[start : i+1]
			}
		}
	}
	return ""
}

func parseScore(raw json.RawMessage) (float64, error) {
	var score float64
	if err := json.Unmarshal(raw, &score); err != nil {
		return 0, fmt.Errorf("evaluator score is not a number: %w", err)
	}
	if score < 0 || score > 100 {
		return 0, fmt.Errorf("evaluator score %v out of range [0,100]", score)
	}
	return score, nil
}
