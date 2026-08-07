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

//go:generate mockgen -destination ../../../internal/mock/domain/evaluation/evaluation_mock.go --package evaluation -source service.go
type Evaluation interface {
	EvalSetService
	EvaluatorService
	ExperimentService
}

type EvalSetService interface {
	CreateEvalSet(ctx context.Context, meta *entity.EvalSetCreateMeta) (*entity.EvalSet, error)
	UpdateEvalSet(ctx context.Context, id int64, name string, description string, schemaJSON string) (*entity.EvalSet, error)
	DeleteEvalSet(ctx context.Context, id int64) error
	GetEvalSet(ctx context.Context, id int64) (*entity.EvalSet, error)
	ListEvalSets(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.EvalSet, int64, error)
	AddEvalSetItems(ctx context.Context, evalSetID int64, items []*entity.EvalSetItemCreateMeta) ([]*entity.EvalSetItem, error)
	UpdateEvalSetItem(ctx context.Context, id int64, dataJSON string) error
	DeleteEvalSetItems(ctx context.Context, evalSetID int64, itemIDs []int64) error
	ListEvalSetItems(ctx context.Context, evalSetID int64, page, pageSize int) ([]*entity.EvalSetItem, int64, error)
}

type EvaluatorService interface {
	CreateEvaluator(ctx context.Context, meta *entity.EvaluatorCreateMeta) (*entity.Evaluator, error)
	UpdateEvaluator(ctx context.Context, id int64, name string, description string, modelID string, prompt string, temperature float64) (*entity.Evaluator, error)
	DeleteEvaluator(ctx context.Context, id int64) error
	GetEvaluator(ctx context.Context, id int64) (*entity.Evaluator, error)
	ListEvaluators(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.Evaluator, int64, error)
	CreateCustomTemplate(ctx context.Context, meta *entity.CustomEvaluatorTemplateCreateMeta) (*entity.CustomEvaluatorTemplate, error)
	DeleteCustomTemplate(ctx context.Context, id int64, createdBy string) error
	ListCustomTemplates(ctx context.Context, createdBy string) ([]*entity.CustomEvaluatorTemplate, error)
}

type ExperimentService interface {
	CreateExperiment(ctx context.Context, meta *entity.ExperimentCreateMeta) (*entity.Experiment, error)
	UpdateExperiment(ctx context.Context, id int64, name string, description string) (*entity.Experiment, error)
	DeleteExperiment(ctx context.Context, id int64) error
	GetExperiment(ctx context.Context, id int64) (*entity.Experiment, error)
	ListExperiments(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.Experiment, int64, error)
	StartExperiment(ctx context.Context, experimentID int64) error
	GetExperimentDetail(ctx context.Context, experimentID int64, page, pageSize int) (*entity.Experiment, []*entity.ExperimentItemResult, int64, error)
	ListTargets(ctx context.Context, targetType int, createdBy string) ([]*entity.TargetInfo, error)
}
