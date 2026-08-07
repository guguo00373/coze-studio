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

package entity

import "github.com/coze-dev/coze-studio/backend/domain/evaluation/internal/dal/model"

type Experiment = model.Experiment
type ExperimentItemResult = model.ExperimentItemResult
type ExperimentAggrResult = model.ExperimentAggrResult

type TargetType int

const (
	TargetTypeAgent    TargetType = 1
	TargetTypeWorkflow TargetType = 2
	TargetTypeChatflow TargetType = 3
)

// TargetInfo describes a selectable evaluation target (e.g. a published agent).
type TargetInfo struct {
	TargetID string     `json:"target_id"`
	Name     string     `json:"name"`
	Type     TargetType `json:"type"`
}

type ExperimentStatus int

const (
	ExperimentStatusPending    ExperimentStatus = 0
	ExperimentStatusRunning    ExperimentStatus = 1
	ExperimentStatusSuccess    ExperimentStatus = 2
	ExperimentStatusFailed     ExperimentStatus = 3
	ExperimentStatusPartial    ExperimentStatus = 4
)

type ItemResultStatus int

const (
	ItemResultStatusPending ItemResultStatus = 0
	ItemResultStatusSuccess ItemResultStatus = 1
	ItemResultStatusFailed  ItemResultStatus = 2
)

type ExperimentCreateMeta struct {
	SpaceID        int64
	Name           string
	Description    string
	EvalSetID      int64
	TargetType     int
	TargetID       string
	TargetConfig   string
	EvaluatorIDs   []int64
	Concurrency    int
	CreatedBy      string
}

type TargetRunResult struct {
	Input        string
	Expected     string
	ActualOutput string
	ExtOutput    map[string]string
	InputTokens  int64
	OutputTokens int64
	LatencyMS    int64
}

type EvaluatorResult struct {
	EvaluatorID int64   `json:"evaluator_id"`
	Score       float64 `json:"score"`
	Reasoning   string  `json:"reasoning"`
}
