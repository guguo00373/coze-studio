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

	"gorm.io/datatypes"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/repository"
	"github.com/coze-dev/coze-studio/backend/pkg/errorx"
	"github.com/coze-dev/coze-studio/backend/types/errno"
)

type Components struct {
	Repo       repository.EvaluationRepository
	Runner     *targetRunnerFactory
	Evaluator  *evaluatorExecutor
	Experiment *experimentEngine
}

type evaluationImpl struct {
	Components
}

func NewEvaluationService(c *Components) Evaluation {
	return &evaluationImpl{Components: *c}
}

// ---------- Eval Set ----------

func (s *evaluationImpl) CreateEvalSet(ctx context.Context, meta *entity.EvalSetCreateMeta) (*entity.EvalSet, error) {
	es := &entity.EvalSet{
		SpaceID:     meta.SpaceID,
		Name:        meta.Name,
		Description: meta.Description,
		SchemaJSON:  datatypes.JSON(meta.SchemaJSON),
		Status:      int(entity.EvalSetStatusDraft),
		CreatedBy:   meta.CreatedBy,
	}
	if es.Name == "" {
		return nil, errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "name is required"))
	}
	return s.Repo.CreateEvalSet(ctx, es)
}

func (s *evaluationImpl) UpdateEvalSet(ctx context.Context, id int64, name string, description string, schemaJSON string) (*entity.EvalSet, error) {
	es, err := s.Repo.GetEvalSetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if es == nil {
		return nil, errorx.New(errno.ErrRecordNotFound)
	}
	if name != "" {
		es.Name = name
	}
	if description != "" {
		es.Description = description
	}
	if schemaJSON != "" {
		es.SchemaJSON = datatypes.JSON(schemaJSON)
	}
	if err := s.Repo.UpdateEvalSet(ctx, es); err != nil {
		return nil, err
	}
	return es, nil
}

func (s *evaluationImpl) DeleteEvalSet(ctx context.Context, id int64) error {
	es, err := s.Repo.GetEvalSetByID(ctx, id)
	if err != nil {
		return err
	}
	if es == nil {
		return errorx.New(errno.ErrRecordNotFound)
	}
	return s.Repo.DeleteEvalSet(ctx, id)
}

func (s *evaluationImpl) GetEvalSet(ctx context.Context, id int64) (*entity.EvalSet, error) {
	return s.Repo.GetEvalSetByID(ctx, id)
}

func (s *evaluationImpl) ListEvalSets(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.EvalSet, int64, error) {
	return s.Repo.ListEvalSets(ctx, createdBy, page, pageSize)
}

func (s *evaluationImpl) AddEvalSetItems(ctx context.Context, evalSetID int64, items []*entity.EvalSetItemCreateMeta) ([]*entity.EvalSetItem, error) {
	es, err := s.Repo.GetEvalSetByID(ctx, evalSetID)
	if err != nil {
		return nil, err
	}
	if es == nil {
		return nil, errorx.New(errno.ErrRecordNotFound)
	}
	if len(items) == 0 {
		return nil, errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "items is empty"))
	}
	entities := make([]*entity.EvalSetItem, 0, len(items))
	for _, item := range items {
		entities = append(entities, &entity.EvalSetItem{
			EvalSetID: evalSetID,
			DataJSON:  datatypes.JSON(item.DataJSON),
		})
	}
	if err := s.Repo.BatchCreateEvalSetItems(ctx, entities); err != nil {
		return nil, err
	}
	if err := s.Repo.UpdateEvalSetItemCount(ctx, evalSetID, int64(len(entities))); err != nil {
		return nil, err
	}
	return entities, nil
}

func (s *evaluationImpl) DeleteEvalSetItems(ctx context.Context, evalSetID int64, itemIDs []int64) error {
	if len(itemIDs) == 0 {
		return nil
	}
	if err := s.Repo.BatchDeleteEvalSetItems(ctx, itemIDs); err != nil {
		return err
	}
	return s.Repo.UpdateEvalSetItemCount(ctx, evalSetID, -int64(len(itemIDs)))
}

func (s *evaluationImpl) UpdateEvalSetItem(ctx context.Context, id int64, dataJSON string) error {
	if id <= 0 || dataJSON == "" {
		return errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "id and data_json are required"))
	}
	if !json.Valid([]byte(dataJSON)) {
		return errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "data_json is not valid JSON"))
	}
	return s.Repo.UpdateEvalSetItem(ctx, id, datatypes.JSON(dataJSON))
}

func (s *evaluationImpl) ListEvalSetItems(ctx context.Context, evalSetID int64, page, pageSize int) ([]*entity.EvalSetItem, int64, error) {
	return s.Repo.ListEvalSetItems(ctx, evalSetID, page, pageSize)
}

// ---------- Evaluator ----------

func (s *evaluationImpl) CreateEvaluator(ctx context.Context, meta *entity.EvaluatorCreateMeta) (*entity.Evaluator, error) {
	if meta.Name == "" || meta.ModelID == "" {
		return nil, errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "name and model_id are required"))
	}
	ev := &entity.Evaluator{
		SpaceID:     meta.SpaceID,
		Name:        meta.Name,
		Description: meta.Description,
		Type:        meta.Type,
		ModelID:     meta.ModelID,
		Prompt:      meta.Prompt,
		Temperature: meta.Temperature,
		Status:      int(entity.EvaluatorStatusDraft),
		CreatedBy:   meta.CreatedBy,
	}
	return s.Repo.CreateEvaluator(ctx, ev)
}

func (s *evaluationImpl) UpdateEvaluator(ctx context.Context, id int64, name string, description string, modelID string, prompt string, temperature float64) (*entity.Evaluator, error) {
	ev, err := s.Repo.GetEvaluatorByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if ev == nil {
		return nil, errorx.New(errno.ErrRecordNotFound)
	}
	if name != "" {
		ev.Name = name
	}
	if description != "" {
		ev.Description = description
	}
	if modelID != "" {
		ev.ModelID = modelID
	}
	if prompt != "" {
		ev.Prompt = prompt
	}
	if temperature != 0 {
		ev.Temperature = temperature
	}
	if err := s.Repo.UpdateEvaluator(ctx, ev); err != nil {
		return nil, err
	}
	return ev, nil
}

func (s *evaluationImpl) DeleteEvaluator(ctx context.Context, id int64) error {
	ev, err := s.Repo.GetEvaluatorByID(ctx, id)
	if err != nil {
		return err
	}
	if ev == nil {
		return errorx.New(errno.ErrRecordNotFound)
	}
	return s.Repo.DeleteEvaluator(ctx, id)
}

func (s *evaluationImpl) GetEvaluator(ctx context.Context, id int64) (*entity.Evaluator, error) {
	return s.Repo.GetEvaluatorByID(ctx, id)
}

func (s *evaluationImpl) ListEvaluators(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.Evaluator, int64, error) {
	return s.Repo.ListEvaluators(ctx, createdBy, page, pageSize)
}

func (s *evaluationImpl) CreateCustomTemplate(ctx context.Context, meta *entity.CustomEvaluatorTemplateCreateMeta) (*entity.CustomEvaluatorTemplate, error) {
	if meta.Name == "" || meta.Prompt == "" {
		return nil, errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "name and prompt are required"))
	}
	return s.Repo.CreateCustomTemplate(ctx, &entity.CustomEvaluatorTemplate{
		Name:        meta.Name,
		Description: meta.Description,
		Prompt:      meta.Prompt,
		CreatedBy:   meta.CreatedBy,
	})
}

func (s *evaluationImpl) DeleteCustomTemplate(ctx context.Context, id int64, createdBy string) error {
	if id <= 0 {
		return errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "id is required"))
	}
	return s.Repo.DeleteCustomTemplate(ctx, id, createdBy)
}

func (s *evaluationImpl) ListCustomTemplates(ctx context.Context, createdBy string) ([]*entity.CustomEvaluatorTemplate, error) {
	return s.Repo.ListCustomTemplates(ctx, createdBy)
}

// ---------- Experiment ----------

func (s *evaluationImpl) CreateExperiment(ctx context.Context, meta *entity.ExperimentCreateMeta) (*entity.Experiment, error) {
	if meta.Name == "" || meta.TargetID == "" {
		return nil, errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "name and target_id are required"))
	}
	if meta.EvalSetID == 0 {
		return nil, errorx.New(errno.ErrInvalidParameter, errorx.KV("msg", "eval_set_id is required"))
	}
	evaluatorIDs, _ := json.Marshal(meta.EvaluatorIDs)
	targetConfig, _ := json.Marshal(meta.TargetConfig)
	expt := &entity.Experiment{
		SpaceID:      meta.SpaceID,
		Name:         meta.Name,
		Description:  meta.Description,
		EvalSetID:    meta.EvalSetID,
		TargetType:   meta.TargetType,
		TargetID:     meta.TargetID,
		TargetConfig: datatypes.JSON(targetConfig),
		EvaluatorIDs: datatypes.JSON(evaluatorIDs),
		Concurrency:  meta.Concurrency,
		Status:       int(entity.ExperimentStatusPending),
		CreatedBy:    meta.CreatedBy,
	}
	return s.Repo.CreateExperiment(ctx, expt)
}

func (s *evaluationImpl) UpdateExperiment(ctx context.Context, id int64, name string, description string) (*entity.Experiment, error) {
	expt, err := s.Repo.GetExperimentByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if expt == nil {
		return nil, errorx.New(errno.ErrRecordNotFound)
	}
	if name != "" {
		expt.Name = name
	}
	if description != "" {
		expt.Description = description
	}
	if err := s.Repo.UpdateExperiment(ctx, expt); err != nil {
		return nil, err
	}
	return expt, nil
}

func (s *evaluationImpl) DeleteExperiment(ctx context.Context, id int64) error {
	expt, err := s.Repo.GetExperimentByID(ctx, id)
	if err != nil {
		return err
	}
	if expt == nil {
		return errorx.New(errno.ErrRecordNotFound)
	}
	return s.Repo.DeleteExperiment(ctx, id)
}

func (s *evaluationImpl) GetExperiment(ctx context.Context, id int64) (*entity.Experiment, error) {
	return s.Repo.GetExperimentByID(ctx, id)
}

func (s *evaluationImpl) ListExperiments(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.Experiment, int64, error) {
	return s.Repo.ListExperiments(ctx, createdBy, page, pageSize)
}

// ListTargets returns selectable evaluation targets for the given type.
// Only published resources are returned (the runners execute the RELEASE
// version of agents/workflows).
func (s *evaluationImpl) ListTargets(ctx context.Context, targetType int, createdBy string) ([]*entity.TargetInfo, error) {
	switch entity.TargetType(targetType) {
	case entity.TargetTypeAgent:
		return s.Repo.ListAgentTargets(ctx, createdBy)
	case entity.TargetTypeWorkflow:
		return s.Repo.ListWorkflowTargets(ctx, createdBy, 0, entity.TargetTypeWorkflow)
	case entity.TargetTypeChatflow:
		return s.Repo.ListWorkflowTargets(ctx, createdBy, 3, entity.TargetTypeChatflow)
	default:
		return []*entity.TargetInfo{}, nil
	}
}

func (s *evaluationImpl) StartExperiment(ctx context.Context, experimentID int64) error {
	expt, err := s.Repo.GetExperimentByID(ctx, experimentID)
	if err != nil {
		return err
	}
	if expt == nil {
		return errorx.New(errno.ErrRecordNotFound)
	}
	return s.Experiment.Run(ctx, expt)
}

func (s *evaluationImpl) GetExperimentDetail(ctx context.Context, experimentID int64, page, pageSize int) (*entity.Experiment, []*entity.ExperimentItemResult, int64, error) {
	expt, err := s.Repo.GetExperimentByID(ctx, experimentID)
	if err != nil {
		return nil, nil, 0, err
	}
	if expt == nil {
		return nil, nil, 0, errorx.New(errno.ErrRecordNotFound)
	}
	results, total, err := s.Repo.ListExperimentItemResults(ctx, experimentID, page, pageSize)
	if err != nil {
		return nil, nil, 0, err
	}
	return expt, results, total, nil
}
