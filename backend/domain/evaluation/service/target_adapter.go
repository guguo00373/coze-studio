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

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
)

// TargetRunner executes a single evaluation set item against an evaluation target
// (agent / workflow / chatflow) and captures the actual output for scoring.
type TargetRunner interface {
	Run(ctx context.Context, targetID string, inputJSON string) (*entity.TargetRunResult, error)
}

type targetRunnerFactory struct {
	agentRunner    TargetRunner
	workflowRunner TargetRunner
	chatflowRunner TargetRunner
}

func NewTargetRunnerFactory() *targetRunnerFactory {
	return &targetRunnerFactory{
		agentRunner:    &agentRunner{},
		workflowRunner: &workflowRunner{},
		chatflowRunner: &chatflowRunner{},
	}
}

func (f *targetRunnerFactory) GetRunner(targetType int) TargetRunner {
	switch entity.TargetType(targetType) {
	case entity.TargetTypeAgent:
		return f.agentRunner
	case entity.TargetTypeWorkflow:
		return f.workflowRunner
	case entity.TargetTypeChatflow:
		return f.chatflowRunner
	default:
		return nil
	}
}
