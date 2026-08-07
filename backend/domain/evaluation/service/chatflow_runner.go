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
	"errors"
	"io"

	"github.com/coze-dev/coze-studio/backend/domain/workflow/entity/vo"

	runModel 	"github.com/coze-dev/coze-studio/backend/api/model/workflow"
	appWorkflow "github.com/coze-dev/coze-studio/backend/application/workflow"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
)

type chatflowRunner struct{}

func NewChatflowRunner() TargetRunner { return &chatflowRunner{} }

func (r *chatflowRunner) Run(ctx context.Context, targetID string, inputJSON string) (*entity.TargetRunResult, error) {
	userID, err := getOperatorUserID(ctx)
	if err != nil {
		return nil, err
	}

	authCtx := buildAuthContext(ctx, userID)
	req := &runModel.ChatFlowRunRequest{
		WorkflowID: targetID,
		AdditionalMessages: []*runModel.EnterMessage{
			{
				Role:        "user",
				Content:     inputJSON,
				ContentType: "text",
			},
		},
		ExecuteMode: func() *string { m := "RELEASE"; return &m }(),
	}

	start := nowMillis()
	streamer, err := appWorkflow.SVC.OpenAPIChatFlowRun(authCtx, req)
	if err != nil {
		return nil, err
	}

	result := &entity.TargetRunResult{LatencyMS: nowMillis() - start}
	for {
		chunks, recvErr := streamer.Recv()
		if recvErr != nil {
			if errors.Is(recvErr, io.EOF) {
				break
			}
			return result, nil
		}
		for _, chunk := range chunks {
			if chunk == nil || chunk.Event != string(vo.ChatFlowMessageCompleted) {
				continue
			}
			var detail struct {
				Content     string `json:"content"`
				ContentType string `json:"content_type"`
			}
			if err := json.Unmarshal([]byte(chunk.Data), &detail); err == nil {
				result.ActualOutput = detail.Content
			}
		}
	}

	return result, nil
}
