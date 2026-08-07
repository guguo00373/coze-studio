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
	"time"

	"github.com/coze-dev/coze-studio/backend/application/base/ctxutil"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

func nowMillis() int64 {
	return time.Now().UnixMilli()
}

// getOperatorUserID resolves the currently logged-in user from the request
// session. This is the operator who owns the evaluation resources.
func getOperatorUserID(ctx context.Context) (int64, error) {
	uid := ctxutil.GetUIDFromCtx(ctx)
	if uid == nil || *uid == 0 {
		return 0, newErrUserRequired()
	}
	return *uid, nil
}

func newErrUserRequired() error {
	return errorx.New(errno.ErrUserAuthenticationFailed, errorx.KV("msg", "user id is required to run evaluation"))
}

// parseCaseData extracts the "input" and expected/reference fields from an
// eval set item's data_json so the evaluator always has the case data to
// evaluate, regardless of how the target runner built the run result.
func parseCaseData(dataJSON string) (input, expected string) {
	var m map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &m); err != nil {
		return "", ""
	}
	input, _ = m["input"].(string)
	if v, ok := m["reference_output"].(string); ok {
		expected = v
	} else if v, ok := m["expected"].(string); ok {
		expected = v
	}
	return input, expected
}

// buildTargetInput builds the payload delivered to the evaluation target. The
// "reference_output" / "expected" columns are the reference answer for the
// evaluator and must never be sent to the target, otherwise the target can
// cheat by echoing them back. Agents receive the plain "input" text;
// workflows/chatflows receive the remaining parameters as JSON.
func buildTargetInput(dataJSON string, targetType entity.TargetType) string {
	stripped := stripReferenceOutput(dataJSON)
	if targetType == entity.TargetTypeWorkflow || targetType == entity.TargetTypeChatflow {
		return stripped
	}
	input, _ := parseCaseData(dataJSON)
	if input != "" {
		return input
	}
	return stripped
}

func stripReferenceOutput(dataJSON string) string {
	var m map[string]any
	if err := json.Unmarshal([]byte(dataJSON), &m); err != nil {
		return dataJSON
	}
	delete(m, "reference_output")
	delete(m, "expected")
	b, _ := json.Marshal(m)
	return string(b)
}
