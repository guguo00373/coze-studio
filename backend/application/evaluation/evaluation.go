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
	"context"

	"github.com/coze-dev/coze-studio/backend/application/base/ctxutil"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/service"
	"github.com/coze-dev/coze-studio/backend/pkg/lang/conv"
)

type EvaluationApplicationService struct {
	EvalDomainSVC service.Evaluation
}

func (s *EvaluationApplicationService) CreateEvalSet(ctx context.Context, req *EvalSetCreateRequest) (*EvalSetVO, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	es, err := s.EvalDomainSVC.CreateEvalSet(ctx, &entity.EvalSetCreateMeta{
		SpaceID:     req.SpaceID,
		Name:        req.Name,
		Description: req.Description,
		SchemaJSON:  req.SchemaJSON,
		CreatedBy:   conv.Int64ToStr(uid),
	})
	if err != nil {
		return nil, err
	}
	return convertEvalSet(es), nil
}

func (s *EvaluationApplicationService) UpdateEvalSet(ctx context.Context, req *EvalSetUpdateRequest) (*EvalSetVO, error) {
	es, err := s.EvalDomainSVC.UpdateEvalSet(ctx, req.ID, req.Name, req.Description, req.SchemaJSON)
	if err != nil {
		return nil, err
	}
	return convertEvalSet(es), nil
}

func (s *EvaluationApplicationService) DeleteEvalSet(ctx context.Context, id int64) error {
	return s.EvalDomainSVC.DeleteEvalSet(ctx, id)
}

func (s *EvaluationApplicationService) GetEvalSet(ctx context.Context, id int64) (*EvalSetVO, error) {
	es, err := s.EvalDomainSVC.GetEvalSet(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertEvalSet(es), nil
}

func (s *EvaluationApplicationService) ListEvalSets(ctx context.Context, page, pageSize int) ([]*EvalSetVO, int64, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	list, total, err := s.EvalDomainSVC.ListEvalSets(ctx, conv.Int64ToStr(uid), page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*EvalSetVO, 0, len(list))
	for _, es := range list {
		res = append(res, convertEvalSet(es))
	}
	return res, total, nil
}

func (s *EvaluationApplicationService) AddEvalSetItems(ctx context.Context, req *EvalSetItemAddRequest) ([]*EvalSetItemVO, error) {
	items := make([]*entity.EvalSetItemCreateMeta, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, &entity.EvalSetItemCreateMeta{
			EvalSetID: req.EvalSetID,
			DataJSON:  item.DataJSON,
		})
	}
	created, err := s.EvalDomainSVC.AddEvalSetItems(ctx, req.EvalSetID, items)
	if err != nil {
		return nil, err
	}
	res := make([]*EvalSetItemVO, 0, len(created))
	for _, item := range created {
		res = append(res, convertEvalSetItem(item))
	}
	return res, nil
}

func (s *EvaluationApplicationService) DeleteEvalSetItems(ctx context.Context, evalSetID int64, itemIDs []int64) error {
	return s.EvalDomainSVC.DeleteEvalSetItems(ctx, evalSetID, itemIDs)
}

func (s *EvaluationApplicationService) UpdateEvalSetItem(ctx context.Context, id int64, dataJSON string) error {
	return s.EvalDomainSVC.UpdateEvalSetItem(ctx, id, dataJSON)
}

func (s *EvaluationApplicationService) ListEvalSetItems(ctx context.Context, evalSetID int64, page, pageSize int) ([]*EvalSetItemVO, int64, error) {
	list, total, err := s.EvalDomainSVC.ListEvalSetItems(ctx, evalSetID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*EvalSetItemVO, 0, len(list))
	for _, item := range list {
		res = append(res, convertEvalSetItem(item))
	}
	return res, total, nil
}

func (s *EvaluationApplicationService) CreateEvaluator(ctx context.Context, req *EvaluatorCreateRequest) (*EvaluatorVO, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	ev, err := s.EvalDomainSVC.CreateEvaluator(ctx, &entity.EvaluatorCreateMeta{
		SpaceID:     req.SpaceID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		ModelID:     req.ModelID,
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
		CreatedBy:   conv.Int64ToStr(uid),
	})
	if err != nil {
		return nil, err
	}
	return convertEvaluator(ev), nil
}

func (s *EvaluationApplicationService) UpdateEvaluator(ctx context.Context, req *EvaluatorUpdateRequest) (*EvaluatorVO, error) {
	ev, err := s.EvalDomainSVC.UpdateEvaluator(ctx, req.ID, req.Name, req.Description, req.ModelID, req.Prompt, req.Temperature)
	if err != nil {
		return nil, err
	}
	return convertEvaluator(ev), nil
}

func (s *EvaluationApplicationService) DeleteEvaluator(ctx context.Context, id int64) error {
	return s.EvalDomainSVC.DeleteEvaluator(ctx, id)
}

func (s *EvaluationApplicationService) GetEvaluator(ctx context.Context, id int64) (*EvaluatorVO, error) {
	ev, err := s.EvalDomainSVC.GetEvaluator(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertEvaluator(ev), nil
}

func (s *EvaluationApplicationService) ListEvaluators(ctx context.Context, page, pageSize int) ([]*EvaluatorVO, int64, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	list, total, err := s.EvalDomainSVC.ListEvaluators(ctx, conv.Int64ToStr(uid), page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*EvaluatorVO, 0, len(list))
	for _, ev := range list {
		res = append(res, convertEvaluator(ev))
	}
	return res, total, nil
}

func (s *EvaluationApplicationService) CreateCustomTemplate(ctx context.Context, name, description, prompt string) (*entity.CustomEvaluatorTemplate, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	return s.EvalDomainSVC.CreateCustomTemplate(ctx, &entity.CustomEvaluatorTemplateCreateMeta{
		Name:        name,
		Description: description,
		Prompt:      prompt,
		CreatedBy:   conv.Int64ToStr(uid),
	})
}

func (s *EvaluationApplicationService) DeleteCustomTemplate(ctx context.Context, id int64) error {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	return s.EvalDomainSVC.DeleteCustomTemplate(ctx, id, conv.Int64ToStr(uid))
}

func (s *EvaluationApplicationService) ListCustomTemplates(ctx context.Context) ([]*entity.CustomEvaluatorTemplate, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	return s.EvalDomainSVC.ListCustomTemplates(ctx, conv.Int64ToStr(uid))
}

func (s *EvaluationApplicationService) CreateExperiment(ctx context.Context, req *ExperimentCreateRequest) (*ExperimentVO, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	expt, err := s.EvalDomainSVC.CreateExperiment(ctx, &entity.ExperimentCreateMeta{
		SpaceID:      req.SpaceID,
		Name:         req.Name,
		Description:  req.Description,
		EvalSetID:    req.EvalSetID,
		TargetType:   req.TargetType,
		TargetID:     req.TargetID,
		TargetConfig: req.TargetConfig,
		EvaluatorIDs: req.EvaluatorIDs,
		Concurrency:  req.Concurrency,
		CreatedBy:    conv.Int64ToStr(uid),
	})
	if err != nil {
		return nil, err
	}
	return convertExperiment(expt), nil
}

func (s *EvaluationApplicationService) UpdateExperiment(ctx context.Context, req *ExperimentUpdateRequest) (*ExperimentVO, error) {
	expt, err := s.EvalDomainSVC.UpdateExperiment(ctx, req.ID, req.Name, req.Description)
	if err != nil {
		return nil, err
	}
	return convertExperiment(expt), nil
}

func (s *EvaluationApplicationService) DeleteExperiment(ctx context.Context, id int64) error {
	return s.EvalDomainSVC.DeleteExperiment(ctx, id)
}

func (s *EvaluationApplicationService) GetExperiment(ctx context.Context, id int64) (*ExperimentVO, error) {
	expt, err := s.EvalDomainSVC.GetExperiment(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertExperiment(expt), nil
}

func (s *EvaluationApplicationService) ListExperiments(ctx context.Context, page, pageSize int) ([]*ExperimentVO, int64, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	list, total, err := s.EvalDomainSVC.ListExperiments(ctx, conv.Int64ToStr(uid), page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	res := make([]*ExperimentVO, 0, len(list))
	for _, expt := range list {
		res = append(res, convertExperiment(expt))
	}
	return res, total, nil
}

// ListTargets lists selectable evaluation targets.
func (s *EvaluationApplicationService) ListTargets(ctx context.Context, targetType int) ([]*entity.TargetInfo, error) {
	uid := ctxutil.MustGetUIDFromCtx(ctx)
	return s.EvalDomainSVC.ListTargets(ctx, targetType, conv.Int64ToStr(uid))
}

func (s *EvaluationApplicationService) StartExperiment(ctx context.Context, id int64) error {
	return s.EvalDomainSVC.StartExperiment(ctx, id)
}

func (s *EvaluationApplicationService) GetExperimentDetail(ctx context.Context, id int64, page, pageSize int) (*ExperimentDetailVO, error) {
	expt, results, total, err := s.EvalDomainSVC.GetExperimentDetail(ctx, id, page, pageSize)
	if err != nil {
		return nil, err
	}
	vo := &ExperimentDetailVO{
		Experiment: convertExperiment(expt),
		Total:      total,
	}
	for _, r := range results {
		vo.Results = append(vo.Results, convertItemResult(r))
	}
	return vo, nil
}

// ParseImportFile parses an uploaded CSV/XLSX/XLS file into raw rows.
func (s *EvaluationApplicationService) ParseImportFile(filename string, data []byte) ([][]string, error) {
	return service.ParseImportFile(filename, data)
}
