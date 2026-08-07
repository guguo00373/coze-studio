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

package repository

import (
	"context"
	"strconv"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/internal/dal"
	dalModel "github.com/coze-dev/coze-studio/backend/domain/evaluation/internal/dal/model"
)

type EvaluationRepository interface {
	EvalSetRepo
	EvalSetItemRepo
	EvaluatorRepo
	ExperimentRepo
	ExperimentItemResultRepo
	ExperimentAggrResultRepo
	ListAgentTargets(ctx context.Context, createdBy string) ([]*entity.TargetInfo, error)
	ListWorkflowTargets(ctx context.Context, createdBy string, mode int, targetType entity.TargetType) ([]*entity.TargetInfo, error)
	CreateCustomTemplate(ctx context.Context, tpl *entity.CustomEvaluatorTemplate) (*entity.CustomEvaluatorTemplate, error)
	DeleteCustomTemplate(ctx context.Context, id int64, createdBy string) error
	ListCustomTemplates(ctx context.Context, createdBy string) ([]*entity.CustomEvaluatorTemplate, error)
}

type EvalSetRepo interface {
	CreateEvalSet(ctx context.Context, es *entity.EvalSet) (*entity.EvalSet, error)
	UpdateEvalSet(ctx context.Context, es *entity.EvalSet) error
	DeleteEvalSet(ctx context.Context, id int64) error
	GetEvalSetByID(ctx context.Context, id int64) (*entity.EvalSet, error)
	ListEvalSets(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.EvalSet, int64, error)
	UpdateEvalSetItemCount(ctx context.Context, id int64, delta int64) error
}

type EvalSetItemRepo interface {
	BatchCreateEvalSetItems(ctx context.Context, items []*entity.EvalSetItem) error
	BatchDeleteEvalSetItems(ctx context.Context, ids []int64) error
	UpdateEvalSetItem(ctx context.Context, id int64, dataJSON datatypes.JSON) error
	ListEvalSetItems(ctx context.Context, evalSetID int64, page, pageSize int) ([]*entity.EvalSetItem, int64, error)
	CountEvalSetItems(ctx context.Context, evalSetID int64) (int64, error)
}

type EvaluatorRepo interface {
	CreateEvaluator(ctx context.Context, ev *entity.Evaluator) (*entity.Evaluator, error)
	UpdateEvaluator(ctx context.Context, ev *entity.Evaluator) error
	DeleteEvaluator(ctx context.Context, id int64) error
	GetEvaluatorByID(ctx context.Context, id int64) (*entity.Evaluator, error)
	ListEvaluators(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.Evaluator, int64, error)
}

type ExperimentRepo interface {
	CreateExperiment(ctx context.Context, expt *entity.Experiment) (*entity.Experiment, error)
	UpdateExperiment(ctx context.Context, expt *entity.Experiment) error
	UpdateExperimentStatus(ctx context.Context, id int64, status int) error
	DeleteExperiment(ctx context.Context, id int64) error
	GetExperimentByID(ctx context.Context, id int64) (*entity.Experiment, error)
	ListExperiments(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.Experiment, int64, error)
}

type ExperimentItemResultRepo interface {
	CreateExperimentItemResult(ctx context.Context, r *entity.ExperimentItemResult) (*entity.ExperimentItemResult, error)
	UpdateExperimentItemResult(ctx context.Context, r *entity.ExperimentItemResult) error
	ListExperimentItemResults(ctx context.Context, experimentID int64, page, pageSize int) ([]*entity.ExperimentItemResult, int64, error)
	CountExperimentItemResultsByStatus(ctx context.Context, experimentID int64, status int) (int64, error)
}

type ExperimentAggrResultRepo interface {
	CreateExperimentAggrResult(ctx context.Context, r *entity.ExperimentAggrResult) (*entity.ExperimentAggrResult, error)
	GetExperimentAggrResult(ctx context.Context, experimentID int64) (*entity.ExperimentAggrResult, error)
}

type evaluationRepository struct {
	db                *gorm.DB
	EvalSetDAO        *dal.EvalSetDAO
	EvalSetItemDAO    *dal.EvalSetItemDAO
	EvaluatorDAO      *dal.EvaluatorDAO
	ExperimentDAO     *dal.ExperimentDAO
	ItemResultDAO     *dal.ExperimentItemResultDAO
	AggrResultDAO     *dal.ExperimentAggrResultDAO
	CustomTplDAO      *dal.EvaluatorCustomTemplateDAO
}

func NewEvaluationRepository(db *gorm.DB) EvaluationRepository {
	return &evaluationRepository{
		db:            db,
		EvalSetDAO:     dal.NewEvalSetDAO(db),
		EvalSetItemDAO: dal.NewEvalSetItemDAO(db),
		EvaluatorDAO:   dal.NewEvaluatorDAO(db),
		ExperimentDAO:  dal.NewExperimentDAO(db),
		ItemResultDAO:  dal.NewExperimentItemResultDAO(db),
		AggrResultDAO:  dal.NewExperimentAggrResultDAO(db),
		CustomTplDAO:   dal.NewEvaluatorCustomTemplateDAO(db),
	}
}

func (r *evaluationRepository) CreateEvalSet(ctx context.Context, es *entity.EvalSet) (*entity.EvalSet, error) {
	return r.EvalSetDAO.Create(ctx, es)
}
func (r *evaluationRepository) UpdateEvalSet(ctx context.Context, es *entity.EvalSet) error {
	return r.EvalSetDAO.Update(ctx, es)
}
func (r *evaluationRepository) DeleteEvalSet(ctx context.Context, id int64) error {
	return r.EvalSetDAO.Delete(ctx, id)
}
func (r *evaluationRepository) GetEvalSetByID(ctx context.Context, id int64) (*entity.EvalSet, error) {
	return r.EvalSetDAO.GetByID(ctx, id)
}
func (r *evaluationRepository) ListEvalSets(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.EvalSet, int64, error) {
	return r.EvalSetDAO.List(ctx, createdBy, page, pageSize)
}
func (r *evaluationRepository) UpdateEvalSetItemCount(ctx context.Context, id int64, delta int64) error {
	return r.EvalSetDAO.UpdateItemCount(ctx, id, delta)
}

func (r *evaluationRepository) BatchCreateEvalSetItems(ctx context.Context, items []*entity.EvalSetItem) error {
	return r.EvalSetItemDAO.BatchCreate(ctx, items)
}
func (r *evaluationRepository) BatchDeleteEvalSetItems(ctx context.Context, ids []int64) error {
	return r.EvalSetItemDAO.BatchDelete(ctx, ids)
}
func (r *evaluationRepository) UpdateEvalSetItem(ctx context.Context, id int64, dataJSON datatypes.JSON) error {
	return r.EvalSetItemDAO.Update(ctx, id, dataJSON)
}
func (r *evaluationRepository) ListEvalSetItems(ctx context.Context, evalSetID int64, page, pageSize int) ([]*entity.EvalSetItem, int64, error) {
	return r.EvalSetItemDAO.List(ctx, evalSetID, page, pageSize)
}
func (r *evaluationRepository) CountEvalSetItems(ctx context.Context, evalSetID int64) (int64, error) {
	return r.EvalSetItemDAO.CountBySetID(ctx, evalSetID)
}

func (r *evaluationRepository) CreateEvaluator(ctx context.Context, ev *entity.Evaluator) (*entity.Evaluator, error) {
	return r.EvaluatorDAO.Create(ctx, ev)
}
func (r *evaluationRepository) UpdateEvaluator(ctx context.Context, ev *entity.Evaluator) error {
	return r.EvaluatorDAO.Update(ctx, ev)
}
func (r *evaluationRepository) DeleteEvaluator(ctx context.Context, id int64) error {
	return r.EvaluatorDAO.Delete(ctx, id)
}
func (r *evaluationRepository) GetEvaluatorByID(ctx context.Context, id int64) (*entity.Evaluator, error) {
	return r.EvaluatorDAO.GetByID(ctx, id)
}
func (r *evaluationRepository) ListEvaluators(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.Evaluator, int64, error) {
	return r.EvaluatorDAO.List(ctx, createdBy, page, pageSize)
}

func (r *evaluationRepository) CreateExperiment(ctx context.Context, expt *entity.Experiment) (*entity.Experiment, error) {
	return r.ExperimentDAO.Create(ctx, expt)
}
func (r *evaluationRepository) UpdateExperiment(ctx context.Context, expt *entity.Experiment) error {
	return r.ExperimentDAO.Update(ctx, expt)
}
func (r *evaluationRepository) UpdateExperimentStatus(ctx context.Context, id int64, status int) error {
	return r.ExperimentDAO.UpdateStatus(ctx, id, status)
}
func (r *evaluationRepository) DeleteExperiment(ctx context.Context, id int64) error {
	if err := r.ItemResultDAO.DeleteByExperimentID(ctx, id); err != nil {
		return err
	}
	if err := r.AggrResultDAO.DeleteByExperimentID(ctx, id); err != nil {
		return err
	}
	return r.ExperimentDAO.Delete(ctx, id)
}
func (r *evaluationRepository) GetExperimentByID(ctx context.Context, id int64) (*entity.Experiment, error) {
	return r.ExperimentDAO.GetByID(ctx, id)
}
func (r *evaluationRepository) ListExperiments(ctx context.Context, createdBy string, page, pageSize int) ([]*entity.Experiment, int64, error) {
	return r.ExperimentDAO.List(ctx, createdBy, page, pageSize)
}

// ListAgentTargets lists the user's published agents (with their names) as
// selectable evaluation targets.
func (r *evaluationRepository) ListAgentTargets(ctx context.Context, createdBy string) ([]*entity.TargetInfo, error) {
	type agentTargetRow struct {
		AgentID int64  `gorm:"column:agent_id"`
		Name    string `gorm:"column:name"`
	}
	var rows []agentTargetRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.agent_id, v.name
		FROM single_agent_publish p
		LEFT JOIN single_agent_version v ON v.agent_id = p.agent_id
		WHERE p.creator_id = ?
		GROUP BY p.agent_id, v.name
	`, createdBy).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	seen := make(map[int64]bool, len(rows))
	res := make([]*entity.TargetInfo, 0, len(rows))
	for _, row := range rows {
		if seen[row.AgentID] {
			continue
		}
		seen[row.AgentID] = true
		res = append(res, &entity.TargetInfo{
			TargetID: strconv.FormatInt(row.AgentID, 10),
			Name:     row.Name,
			Type:     entity.TargetTypeAgent,
		})
	}
	return res, nil
}

// ListWorkflowTargets lists the user's published workflows of the given mode
// (0 = workflow, 3 = chatflow) as selectable evaluation targets.
func (r *evaluationRepository) ListWorkflowTargets(ctx context.Context, createdBy string, mode int, targetType entity.TargetType) ([]*entity.TargetInfo, error) {
	type workflowTargetRow struct {
		ID   int64  `gorm:"column:id"`
		Name string `gorm:"column:name"`
	}
	var rows []workflowTargetRow
	err := r.db.WithContext(ctx).Raw(`
		SELECT id, name
		FROM workflow_meta
		WHERE creator_id = ? AND mode = ? AND latest_version IS NOT NULL AND latest_version_ts > 0
		ORDER BY updated_at DESC
	`, createdBy, mode).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	res := make([]*entity.TargetInfo, 0, len(rows))
	for _, row := range rows {
		res = append(res, &entity.TargetInfo{
			TargetID: strconv.FormatInt(row.ID, 10),
			Name:     row.Name,
			Type:     targetType,
		})
	}
	return res, nil
}

func (r *evaluationRepository) CreateCustomTemplate(ctx context.Context, tpl *entity.CustomEvaluatorTemplate) (*entity.CustomEvaluatorTemplate, error) {
	created, err := r.CustomTplDAO.Create(ctx, &dalModel.EvaluatorCustomTemplate{
		Name:        tpl.Name,
		Description: tpl.Description,
		Prompt:      tpl.Prompt,
		CreatedBy:   tpl.CreatedBy,
	})
	if err != nil {
		return nil, err
	}
	return &entity.CustomEvaluatorTemplate{
		ID:          created.ID,
		Name:        created.Name,
		Description: created.Description,
		Prompt:      created.Prompt,
		CreatedBy:   created.CreatedBy,
	}, nil
}

func (r *evaluationRepository) DeleteCustomTemplate(ctx context.Context, id int64, createdBy string) error {
	return r.CustomTplDAO.Delete(ctx, id, createdBy)
}

func (r *evaluationRepository) ListCustomTemplates(ctx context.Context, createdBy string) ([]*entity.CustomEvaluatorTemplate, error) {
	list, err := r.CustomTplDAO.ListByCreatedBy(ctx, createdBy)
	if err != nil {
		return nil, err
	}
	res := make([]*entity.CustomEvaluatorTemplate, 0, len(list))
	for _, tpl := range list {
		res = append(res, &entity.CustomEvaluatorTemplate{
			ID:          tpl.ID,
			Name:        tpl.Name,
			Description: tpl.Description,
			Prompt:      tpl.Prompt,
			CreatedBy:   tpl.CreatedBy,
		})
	}
	return res, nil
}

func (r *evaluationRepository) CreateExperimentItemResult(ctx context.Context, res *entity.ExperimentItemResult) (*entity.ExperimentItemResult, error) {
	return r.ItemResultDAO.Create(ctx, res)
}
func (r *evaluationRepository) UpdateExperimentItemResult(ctx context.Context, res *entity.ExperimentItemResult) error {
	return r.ItemResultDAO.Update(ctx, res)
}
func (r *evaluationRepository) ListExperimentItemResults(ctx context.Context, experimentID int64, page, pageSize int) ([]*entity.ExperimentItemResult, int64, error) {
	return r.ItemResultDAO.List(ctx, experimentID, page, pageSize)
}
func (r *evaluationRepository) CountExperimentItemResultsByStatus(ctx context.Context, experimentID int64, status int) (int64, error) {
	return r.ItemResultDAO.CountByStatus(ctx, experimentID, status)
}

func (r *evaluationRepository) CreateExperimentAggrResult(ctx context.Context, res *entity.ExperimentAggrResult) (*entity.ExperimentAggrResult, error) {
	return r.AggrResultDAO.Create(ctx, res)
}
func (r *evaluationRepository) GetExperimentAggrResult(ctx context.Context, experimentID int64) (*entity.ExperimentAggrResult, error) {
	return r.AggrResultDAO.GetByExperimentID(ctx, experimentID)
}
