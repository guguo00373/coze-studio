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

package evaluation

import (
	"strconv"
	"time"

	"gorm.io/datatypes"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
)

type EvalSetCreateRequest struct {
	SpaceID     int64  `json:"space_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SchemaJSON  string `json:"schema_json"`
}

type EvalSetUpdateRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SchemaJSON  string `json:"schema_json"`
}

type EvalSetItemAddRequest struct {
	EvalSetID int64                       `json:"eval_set_id"`
	Items     []*EvalSetItemCreatePayload `json:"items"`
}

type EvalSetItemCreatePayload struct {
	DataJSON string `json:"data_json"`
}

type EvaluatorCreateRequest struct {
	SpaceID     int64   `json:"space_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        int     `json:"type"`
	ModelID     string  `json:"model_id"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
}

type EvaluatorUpdateRequest struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ModelID     string  `json:"model_id"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
}

type ExperimentCreateRequest struct {
	SpaceID      int64   `json:"space_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	EvalSetID    int64   `json:"eval_set_id"`
	TargetType   int     `json:"target_type"`
	TargetID     string  `json:"target_id"`
	TargetConfig string  `json:"target_config"`
	EvaluatorIDs []int64 `json:"evaluator_ids"`
	Concurrency  int     `json:"concurrency"`
}

type ExperimentUpdateRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type EvalSetVO struct {
	ID          int64  `json:"id"`
	SpaceID     int64  `json:"space_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SchemaJSON  string `json:"schema_json"`
	ItemCount   int64  `json:"item_count"`
	Status      int    `json:"status"`
	CreatedBy   string `json:"created_by"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type EvalSetItemVO struct {
	ID        int64  `json:"id"`
	EvalSetID int64  `json:"eval_set_id"`
	DataJSON  string `json:"data_json"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type EvaluatorVO struct {
	ID          int64   `json:"id"`
	SpaceID     int64   `json:"space_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        int     `json:"type"`
	ModelID     string  `json:"model_id"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
	Status      int     `json:"status"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

type ExperimentVO struct {
	ID           int64  `json:"id"`
	SpaceID      int64  `json:"space_id"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	EvalSetID    int64  `json:"eval_set_id"`
	TargetType   int    `json:"target_type"`
	TargetID     string `json:"target_id"`
	TargetConfig string `json:"target_config_json"`
	EvaluatorIDs string `json:"evaluator_ids"`
	Concurrency  int    `json:"concurrency"`
	Status       int    `json:"status"`
	RunStats     string `json:"run_stats_json"`
	ErrorMsg     string `json:"error_msg"`
	CreatedBy    string `json:"created_by"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type ExperimentItemResultVO struct {
	ID               int64  `json:"id"`
	ExperimentID     int64  `json:"experiment_id"`
	EvalSetItemID    int64  `json:"eval_set_item_id"`
	Status           int    `json:"status"`
	InputJSON        string `json:"input_json"`
	ActualOutput     string `json:"actual_output"`
	OutputJSON       string `json:"output_json"`
	TargetErrorMsg   string `json:"target_error_msg"`
	EvaluatorResults string `json:"evaluator_results_json"`
	TokensUsed       string `json:"tokens_used_json"`
	LatencyMS        int64  `json:"latency_ms"`
}

type ExperimentDetailVO struct {
	Experiment *ExperimentVO             `json:"experiment"`
	Results    []*ExperimentItemResultVO `json:"results"`
	Total      int64                     `json:"total"`
}

func convertEvalSet(es *entity.EvalSet) *EvalSetVO {
	if es == nil {
		return nil
	}
	return &EvalSetVO{
		ID:          es.ID,
		SpaceID:     es.SpaceID,
		Name:        es.Name,
		Description: es.Description,
		SchemaJSON:  string(es.SchemaJSON),
		ItemCount:   es.ItemCount,
		Status:      es.Status,
		CreatedBy:   es.CreatedBy,
		CreatedAt:   formatTime(es.CreatedAt),
		UpdatedAt:   formatTime(es.UpdatedAt),
	}
}

func convertEvalSetItem(item *entity.EvalSetItem) *EvalSetItemVO {
	if item == nil {
		return nil
	}
	return &EvalSetItemVO{
		ID:        item.ID,
		EvalSetID: item.EvalSetID,
		DataJSON:  string(item.DataJSON),
		CreatedAt: formatTime(item.CreatedAt),
		UpdatedAt: formatTime(item.UpdatedAt),
	}
}

func convertEvaluator(ev *entity.Evaluator) *EvaluatorVO {
	if ev == nil {
		return nil
	}
	return &EvaluatorVO{
		ID:          ev.ID,
		SpaceID:     ev.SpaceID,
		Name:        ev.Name,
		Description: ev.Description,
		Type:        ev.Type,
		ModelID:     ev.ModelID,
		Prompt:      ev.Prompt,
		Temperature: ev.Temperature,
		Status:      ev.Status,
		CreatedBy:   ev.CreatedBy,
		CreatedAt:   formatTime(ev.CreatedAt),
		UpdatedAt:   formatTime(ev.UpdatedAt),
	}
}

func convertExperiment(expt *entity.Experiment) *ExperimentVO {
	if expt == nil {
		return nil
	}
	return &ExperimentVO{
		ID:           expt.ID,
		SpaceID:      expt.SpaceID,
		Name:         expt.Name,
		Description:  expt.Description,
		EvalSetID:    expt.EvalSetID,
		TargetType:   expt.TargetType,
		TargetID:     expt.TargetID,
		TargetConfig: string(expt.TargetConfig),
		EvaluatorIDs: string(expt.EvaluatorIDs),
		Concurrency:  expt.Concurrency,
		Status:       expt.Status,
		RunStats:     string(expt.RunStats),
		ErrorMsg:     expt.ErrorMsg,
		CreatedBy:    expt.CreatedBy,
		CreatedAt:    formatTime(expt.CreatedAt),
		UpdatedAt:    formatTime(expt.UpdatedAt),
	}
}

func convertItemResult(r *entity.ExperimentItemResult) *ExperimentItemResultVO {
	if r == nil {
		return nil
	}
	return &ExperimentItemResultVO{
		ID:               r.ID,
		ExperimentID:     r.ExperimentID,
		EvalSetItemID:    r.EvalSetItemID,
		Status:           r.Status,
		InputJSON:        string(r.InputJSON),
		ActualOutput:     r.ActualOutput,
		OutputJSON:       string(r.OutputJSON),
		TargetErrorMsg:   r.TargetErrorMsg,
		EvaluatorResults: string(r.EvaluatorResults),
		TokensUsed:       string(r.TokensUsed),
		LatencyMS:        r.LatencyMS,
	}
}

func formatTime(t time.Time) string {
	return strconv.FormatInt(t.UnixMilli(), 10)
}

var _ = datatypes.JSON(nil)
