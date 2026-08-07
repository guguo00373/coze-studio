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

	runModel "github.com/coze-dev/coze-studio/backend/api/model/workflow"
	appWorkflow "github.com/coze-dev/coze-studio/backend/application/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
)

type workflowRunner struct{}

func NewWorkflowRunner() TargetRunner { return &workflowRunner{} }

func (r *workflowRunner) Run(ctx context.Context, targetID string, inputJSON string) (*entity.TargetRunResult, error) {
	userID, err := getOperatorUserID(ctx)
	if err != nil {
		return nil, err
	}

	authCtx := buildAuthContext(ctx, userID)
	req := &runModel.OpenAPIRunFlowRequest{
		WorkflowID: targetID,
		Parameters: &inputJSON,
		ExecuteMode: func() *string { m := "RELEASE"; return &m }(),
	}

	start := nowMillis()
	resp, err := appWorkflow.SVC.OpenAPIRun(authCtx, req)
	if err != nil {
		return nil, err
	}

	result := &entity.TargetRunResult{LatencyMS: nowMillis() - start}
	if resp != nil && resp.Data != nil {
		result.ActualOutput = *resp.Data
	}
	if resp != nil && resp.Token != nil {
		result.InputTokens = *resp.Token
	}
	return result, nil
}
