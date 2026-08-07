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
	"strings"
	"testing"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
)

func TestParseEvaluatorOutput_PlainJSON(t *testing.T) {
	res, err := parseEvaluatorOutput(`{"score": 88, "reasoning": "good"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 88 {
		t.Fatalf("expected score 88, got %v", res.Score)
	}
	if res.Reasoning != "good" {
		t.Fatalf("expected reasoning good, got %q", res.Reasoning)
	}
}

func TestParseEvaluatorOutput_CodeFence(t *testing.T) {
	res, err := parseEvaluatorOutput("```json\n{\"score\": 90, \"reasoning\": \"ok\"}\n```")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 90 {
		t.Fatalf("expected score 90, got %v", res.Score)
	}
}

func TestParseEvaluatorOutput_ProseWrappedJSON(t *testing.T) {
	res, err := parseEvaluatorOutput("评估结果如下：\n```json\n{\"score\": 75, \"reasoning\": \"大部分正确\"}\n```\n完。")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 75 {
		t.Fatalf("expected score 75, got %v", res.Score)
	}
}

func TestParseEvaluatorOutput_Invalid(t *testing.T) {
	if _, err := parseEvaluatorOutput("no json here"); err == nil {
		t.Fatalf("expected error for non-JSON output")
	}
}

func TestParseEvaluatorOutput_BracesInString(t *testing.T) {
	res, err := parseEvaluatorOutput(`{"score": 60, "reasoning": "contains {braces} and }"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Score != 60 {
		t.Fatalf("expected score 60, got %v", res.Score)
	}
}

func TestExtractJSONObject(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{`{"a":1}`, `{"a":1}`},
		{`prefix {"a":1} suffix`, `{"a":1}`},
		{`{"a":{"b":2}}`, `{"a":{"b":2}}`},
		{`no braces here`, ``},
		{`{"a":"has } and {"}`, `{"a":"has } and {"}`},
	}
	for _, c := range cases {
		if got := extractJSONObject(c.in); got != c.want {
			t.Errorf("extractJSONObject(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestRenderPrompt_WithPlaceholders(t *testing.T) {
	out := renderPrompt("Input: {{input}}\nOutput: {{output}}", &entity.TargetRunResult{
		Input:        "q",
		Expected:     "a",
		ActualOutput: "b",
	})
	want := "Input: q\nOutput: b"
	if out != want {
		t.Fatalf("renderPrompt = %q, want %q", out, want)
	}
}

func TestRenderPrompt_WithoutPlaceholdersAppendsContext(t *testing.T) {
	out := renderPrompt("请评估回答是否正确", &entity.TargetRunResult{
		Input:        "q1",
		Expected:     "a1",
		ActualOutput: "o1",
	})
	for _, want := range []string{"q1", "a1", "o1", `{"score": <number 0-100>`} {
		if !strings.Contains(out, want) {
			t.Fatalf("renderPrompt missing %q, got: %s", want, out)
		}
	}
}
