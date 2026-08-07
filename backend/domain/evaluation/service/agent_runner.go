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
	"strconv"
	"time"

	"github.com/cloudwego/eino/schema"

	"github.com/coze-dev/coze-studio/backend/api/model/conversation/run"
	"github.com/coze-dev/coze-studio/backend/application/conversation"
	msgModel "github.com/coze-dev/coze-studio/backend/crossdomain/message/model"
	runEntity "github.com/coze-dev/coze-studio/backend/domain/conversation/agentrun/entity"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
	authEntity "github.com/coze-dev/coze-studio/backend/domain/openauth/openapiauth/entity"
	"github.com/coze-dev/coze-studio/backend/pkg/ctxcache"
	"github.com/coze-dev/coze-studio/backend/types/consts"
)

type agentRunner struct{}

func NewAgentRunner() TargetRunner { return &agentRunner{} }

// buildAuthContext injects the operator identity into the context so that the
// agent execution pipeline can resolve the user id without an openapi auth
// header. connector_id=0 makes the agent pipeline read the published version
// directly (GetSingleAgent) and skip connector-based release checks.
func buildAuthContext(ctx context.Context, userID int64) context.Context {
	ctxcache.Store(ctx, consts.OpenapiAuthKeyInCtx, &authEntity.ApiKey{
		UserID:      userID,
		ConnectorID: 0,
	})
	return ctx
}

func (r *agentRunner) Run(ctx context.Context, targetID string, inputJSON string) (*entity.TargetRunResult, error) {
	agentID, err := strconv.ParseInt(targetID, 10, 64)
	if err != nil {
		return nil, err
	}

	userID, err := getOperatorUserID(ctx)
	if err != nil {
		return nil, err
	}

	authCtx := buildAuthContext(ctx, userID)
	req := &run.ChatV3Request{
		BotID: agentID,
		User:  strconv.FormatInt(userID, 10),
		AdditionalMessages: []*run.EnterMessage{
			{
				Role:        string(schema.User),
				Content:     inputJSON,
				ContentType: run.ContentTypeText,
			},
		},
	}

	start := nowMillis()
	resp, err := conversation.ConversationOpenAPISVC.OpenapiAgentRunSync(authCtx, req)
	if err != nil {
		return nil, err
	}

	var (
		runID          int64
		conversationID int64
	)
	if resp != nil && resp.ChatDetail != nil {
		runID = resp.ChatDetail.ID
		conversationID = resp.ChatDetail.ConversationID
	}

	// The sync API returns as soon as the run is created, while the model
	// streams in the background, so the answer message may not be persisted
	// yet. Poll until an answer appears (token usage is carried on the answer
	// message's ext).
	actualOutput, inputTokens, outputTokens := collectAgentRunResult(authCtx, runID, conversationID)

	result := &entity.TargetRunResult{
		ActualOutput: actualOutput,
		LatencyMS:    nowMillis() - start,
	}
	if inputTokens > 0 || outputTokens > 0 {
		result.InputTokens = inputTokens
		result.OutputTokens = outputTokens
	} else if finalRun := getFinalRunRecord(authCtx, runID); finalRun != nil && finalRun.Usage != nil {
		result.InputTokens = finalRun.Usage.LlmPromptTokens
		result.OutputTokens = finalRun.Usage.LlmCompletionTokens
	}

	return result, nil
}

const (
	agentAnswerPollInterval = 500 * time.Millisecond
	agentAnswerMaxWait      = 90 * time.Second
)

// collectAgentRunResult polls the run's messages until an answer is persisted.
func collectAgentRunResult(ctx context.Context, runID, conversationID int64) (string, int64, int64) {
	if runID == 0 {
		return "", 0, 0
	}
	deadline := time.Now().Add(agentAnswerMaxWait)
	for {
		if content, input, output := fetchAgentAnswer(ctx, conversationID, runID); content != "" {
			return content, input, output
		}
		runRec, err := conversation.ConversationSVC.AgentRunDomainSVC.GetByID(ctx, runID)
		if err == nil && runRec != nil && isTerminalRunStatus(runRec.Status) {
			break
		}
		if time.Now().After(deadline) {
			break
		}
		time.Sleep(agentAnswerPollInterval)
	}
	return fetchAgentAnswer(ctx, conversationID, runID)
}

func fetchAgentAnswer(ctx context.Context, conversationID, runID int64) (string, int64, int64) {
	messages, err := conversation.ConversationSVC.MessageDomainSVC.GetByRunIDs(ctx, conversationID, []int64{runID})
	if err != nil {
		return "", 0, 0
	}
	// Walk backwards to find the last answer message produced by the run.
	for i := len(messages) - 1; i >= 0; i-- {
		m := messages[i]
		if m != nil && m.MessageType == msgModel.MessageTypeAnswer && m.Content != "" {
			input, output := parseMessageTokens(m.Ext)
			return m.Content, input, output
		}
	}
	return "", 0, 0
}

// parseMessageTokens reads the token usage carried on an answer message's ext.
func parseMessageTokens(ext map[string]string) (input, output int64) {
	if v, ok := ext["input_tokens"]; ok {
		input, _ = strconv.ParseInt(v, 10, 64)
	}
	if v, ok := ext["output_tokens"]; ok {
		output, _ = strconv.ParseInt(v, 10, 64)
	}
	return input, output
}

func getFinalRunRecord(ctx context.Context, runID int64) *runEntity.RunRecordMeta {
	if runID == 0 {
		return nil
	}
	rec, err := conversation.ConversationSVC.AgentRunDomainSVC.GetByID(ctx, runID)
	if err != nil {
		return nil
	}
	return rec
}

func isTerminalRunStatus(status runEntity.RunStatus) bool {
	switch status {
	case runEntity.RunStatusCompleted, runEntity.RunStatusFailed,
		runEntity.RunStatusCancelled, runEntity.RunStatusExpired:
		return true
	default:
		return false
	}
}
