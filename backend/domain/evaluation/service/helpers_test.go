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

func TestParseCaseData(t *testing.T) {
	cases := []struct {
		in      string
		wantIn  string
		wantExp string
	}{
		{`{"input": "你好", "reference_output": "你好"}`, "你好", "你好"},
		{`{"input": "1+1=?", "expected": "2"}`, "1+1=?", "2"},
		{`{"input": "q", "reference_output": "a", "extra": 1}`, "q", "a"},
		{`not json`, "", ""},
		{`{"input": 123}`, "", ""},
	}
	for _, c := range cases {
		in, exp := parseCaseData(c.in)
		if in != c.wantIn || exp != c.wantExp {
			t.Errorf("parseCaseData(%q) = (%q, %q), want (%q, %q)", c.in, in, exp, c.wantIn, c.wantExp)
		}
	}
}

func TestBuildTargetInput_AgentNeverLeaksReference(t *testing.T) {
	// Agent receives only the plain "input" text; the reference answer must not
	// appear in the payload.
	payload := buildTargetInput(`{"input": "小明的公司在哪里？", "reference_output": "成华区经济开发区"}`, entity.TargetTypeAgent)
	if payload != "小明的公司在哪里？" {
		t.Fatalf("agent payload = %q, want the plain input only", payload)
	}
	if strings.Contains(payload, "成华区") || strings.Contains(payload, "reference_output") {
		t.Fatalf("reference leaked into agent payload: %q", payload)
	}
}

func TestBuildTargetInput_WorkflowStripsReference(t *testing.T) {
	// Workflows keep other parameters as JSON but drop the reference output.
	payload := buildTargetInput(`{"parameter": "你好", "reference_output": "你好"}`, entity.TargetTypeWorkflow)
	if strings.Contains(payload, "reference_output") || strings.Contains(payload, "你好") == false {
		t.Fatalf("workflow payload = %q", payload)
	}
	if !strings.Contains(payload, `"parameter":"你好"`) {
		t.Fatalf("workflow payload lost parameters: %q", payload)
	}
}

func TestBuildTargetInput_AgentWithEmptyInputFallsBack(t *testing.T) {
	payload := buildTargetInput(`{"reference_output": "你好"}`, entity.TargetTypeAgent)
	if strings.Contains(payload, "reference_output") {
		t.Fatalf("reference leaked into agent payload: %q", payload)
	}
}
