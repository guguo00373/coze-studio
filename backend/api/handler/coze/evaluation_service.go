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

package coze

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/coze-dev/coze-studio/backend/api/internal/httputil"
	"github.com/coze-dev/coze-studio/backend/application/evaluation"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
)

type EvalSetCreateReq struct {
	SpaceID     int64  `json:"space_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SchemaJSON  string `json:"schema_json"`
}

type EvalSetUpdateReq struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	SchemaJSON  string `json:"schema_json"`
}

type EvalSetItemAddReq struct {
	EvalSetID int64  `json:"eval_set_id"`
	DataJSON  string `json:"data_json"`
}

type EvaluatorCreateReq struct {
	SpaceID     int64   `json:"space_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Type        int     `json:"type"`
	ModelID     string  `json:"model_id"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
}

type EvaluatorUpdateReq struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	ModelID     string  `json:"model_id"`
	Prompt      string  `json:"prompt"`
	Temperature float64 `json:"temperature"`
}

type ExperimentCreateReq struct {
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

type ExperimentUpdateReq struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type PageReq struct {
	Page     int `json:"page" query:"page"`
	PageSize int `json:"page_size" query:"page_size"`
}

// CreateEvalSet .
// @router /api/evaluation_api/create_eval_set [POST]
func CreateEvalSet(ctx context.Context, c *app.RequestContext) {
	var req EvalSetCreateReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.CreateEvalSet(ctx, &evaluation.EvalSetCreateRequest{
		SpaceID:     req.SpaceID,
		Name:        req.Name,
		Description: req.Description,
		SchemaJSON:  req.SchemaJSON,
	})
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// UpdateEvalSet .
// @router /api/evaluation_api/update_eval_set [POST]
func UpdateEvalSet(ctx context.Context, c *app.RequestContext) {
	var req EvalSetUpdateReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.UpdateEvalSet(ctx, &evaluation.EvalSetUpdateRequest{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		SchemaJSON:  req.SchemaJSON,
	})
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// DeleteEvalSet .
// @router /api/evaluation_api/delete_eval_set [POST]
func DeleteEvalSet(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if err := evaluation.EvalSVC.DeleteEvalSet(ctx, req.ID); err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]bool{"success": true})
}

// GetEvalSet .
// @router /api/evaluation_api/get_eval_set [GET]
func GetEvalSet(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id" query:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.GetEvalSet(ctx, req.ID)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// ListEvalSets .
// @router /api/evaluation_api/list_eval_sets [GET]
func ListEvalSets(ctx context.Context, c *app.RequestContext) {
	var req PageReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	list, total, err := evaluation.EvalSVC.ListEvalSets(ctx, req.Page, req.PageSize)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]any{"list": list, "total": total})
}

// AddEvalSetItems .
// @router /api/evaluation_api/add_eval_set_items [POST]
func AddEvalSetItems(ctx context.Context, c *app.RequestContext) {
	var req struct {
		EvalSetID int64              `json:"eval_set_id"`
		Items     []*EvalSetItemAddReq `json:"items"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	items := make([]*evaluation.EvalSetItemCreatePayload, 0, len(req.Items))
	for _, it := range req.Items {
		if it == nil {
			continue
		}
		items = append(items, &evaluation.EvalSetItemCreatePayload{DataJSON: it.DataJSON})
	}
	resp, err := evaluation.EvalSVC.AddEvalSetItems(ctx, &evaluation.EvalSetItemAddRequest{
		EvalSetID: req.EvalSetID,
		Items:     items,
	})
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// DeleteEvalSetItems .
// @router /api/evaluation_api/delete_eval_set_items [POST]
func DeleteEvalSetItems(ctx context.Context, c *app.RequestContext) {
	var req struct {
		EvalSetID int64   `json:"eval_set_id"`
		ItemIDs   []int64 `json:"item_ids"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if err := evaluation.EvalSVC.DeleteEvalSetItems(ctx, req.EvalSetID, req.ItemIDs); err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]bool{"success": true})
}

// UpdateEvalSetItem .
// @router /api/evaluation_api/update_eval_set_item [POST]
func UpdateEvalSetItem(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID       int64  `json:"id"`
		DataJSON string `json:"data_json"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if err := evaluation.EvalSVC.UpdateEvalSetItem(ctx, req.ID, req.DataJSON); err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]bool{"success": true})
}

// ListEvalSetItems .
// @router /api/evaluation_api/list_eval_set_items [GET]
func ListEvalSetItems(ctx context.Context, c *app.RequestContext) {
	var req struct {
		EvalSetID int64 `json:"eval_set_id" query:"eval_set_id"`
		PageReq
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	list, total, err := evaluation.EvalSVC.ListEvalSetItems(ctx, req.EvalSetID, req.Page, req.PageSize)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]any{"list": list, "total": total})
}

// CreateEvaluator .
// @router /api/evaluation_api/create_evaluator [POST]
func CreateEvaluator(ctx context.Context, c *app.RequestContext) {
	var req EvaluatorCreateReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.CreateEvaluator(ctx, &evaluation.EvaluatorCreateRequest{
		SpaceID:     req.SpaceID,
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		ModelID:     req.ModelID,
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
	})
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// UpdateEvaluator .
// @router /api/evaluation_api/update_evaluator [POST]
func UpdateEvaluator(ctx context.Context, c *app.RequestContext) {
	var req EvaluatorUpdateReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.UpdateEvaluator(ctx, &evaluation.EvaluatorUpdateRequest{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
		ModelID:     req.ModelID,
		Prompt:      req.Prompt,
		Temperature: req.Temperature,
	})
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// DeleteEvaluator .
// @router /api/evaluation_api/delete_evaluator [POST]
func DeleteEvaluator(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if err := evaluation.EvalSVC.DeleteEvaluator(ctx, req.ID); err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]bool{"success": true})
}

// GetEvaluator .
// @router /api/evaluation_api/get_evaluator [GET]
func GetEvaluator(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id" query:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.GetEvaluator(ctx, req.ID)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// ListEvaluators .
// @router /api/evaluation_api/list_evaluators [GET]
func ListEvaluators(ctx context.Context, c *app.RequestContext) {
	var req PageReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	list, total, err := evaluation.EvalSVC.ListEvaluators(ctx, req.Page, req.PageSize)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]any{"list": list, "total": total})
}

// CreateExperiment .
// @router /api/evaluation_api/create_experiment [POST]
func CreateExperiment(ctx context.Context, c *app.RequestContext) {
	var req ExperimentCreateReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.CreateExperiment(ctx, &evaluation.ExperimentCreateRequest{
		SpaceID:      req.SpaceID,
		Name:         req.Name,
		Description:  req.Description,
		EvalSetID:    req.EvalSetID,
		TargetType:   req.TargetType,
		TargetID:     req.TargetID,
		TargetConfig: req.TargetConfig,
		EvaluatorIDs: req.EvaluatorIDs,
		Concurrency:  req.Concurrency,
	})
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// UpdateExperiment .
// @router /api/evaluation_api/update_experiment [POST]
func UpdateExperiment(ctx context.Context, c *app.RequestContext) {
	var req ExperimentUpdateReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.UpdateExperiment(ctx, &evaluation.ExperimentUpdateRequest{
		ID:          req.ID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// DeleteExperiment .
// @router /api/evaluation_api/delete_experiment [POST]
func DeleteExperiment(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if err := evaluation.EvalSVC.DeleteExperiment(ctx, req.ID); err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]bool{"success": true})
}

// GetExperiment .
// @router /api/evaluation_api/get_experiment [GET]
func GetExperiment(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id" query:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	resp, err := evaluation.EvalSVC.GetExperiment(ctx, req.ID)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// ListExperiments .
// @router /api/evaluation_api/list_experiments [GET]
func ListExperiments(ctx context.Context, c *app.RequestContext) {
	var req PageReq
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	list, total, err := evaluation.EvalSVC.ListExperiments(ctx, req.Page, req.PageSize)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]any{"list": list, "total": total})
}

// StartExperiment .
// @router /api/evaluation_api/start_experiment [POST]
func StartExperiment(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if err := evaluation.EvalSVC.StartExperiment(ctx, req.ID); err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]bool{"success": true})
}

// GetExperimentDetail .
// @router /api/evaluation_api/get_experiment_detail [GET]
func GetExperimentDetail(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id" query:"id"`
		PageReq
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	resp, err := evaluation.EvalSVC.GetExperimentDetail(ctx, req.ID, req.Page, req.PageSize)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, resp)
}

// ListEvaluatorTemplates .
// @router /api/evaluation_api/list_evaluator_templates [GET]
func ListEvaluatorTemplates(ctx context.Context, c *app.RequestContext) {
	templates := make([]map[string]any, 0, len(entity.BuiltinEvaluatorTemplates))
	for _, t := range entity.BuiltinEvaluatorTemplates {
		templates = append(templates, map[string]any{
			"key":        t.Key,
			"name_zh":    t.NameZh,
			"name_en":    t.NameEn,
			"prompt_zh":  t.PromptZh,
			"prompt_en":  t.PromptEn,
			"is_builtin": true,
		})
	}
	customs, err := evaluation.EvalSVC.ListCustomTemplates(ctx)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	for _, t := range customs {
		templates = append(templates, map[string]any{
			"key":        fmt.Sprintf("custom:%d", t.ID),
			"name_zh":    t.Name,
			"name_en":    t.Name,
			"prompt_zh":  t.Prompt,
			"prompt_en":  t.Prompt,
			"is_builtin": false,
		})
	}
	c.JSON(consts.StatusOK, map[string]any{"list": templates})
}

// CreateEvaluatorTemplate .
// @router /api/evaluation_api/create_evaluator_template [POST]
func CreateEvaluatorTemplate(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Prompt      string `json:"prompt"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	tpl, err := evaluation.EvalSVC.CreateCustomTemplate(ctx, req.Name, req.Description, req.Prompt)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, tpl)
}

// DeleteEvaluatorTemplate .
// @router /api/evaluation_api/delete_evaluator_template [POST]
func DeleteEvaluatorTemplate(ctx context.Context, c *app.RequestContext) {
	var req struct {
		ID int64 `json:"id"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if err := evaluation.EvalSVC.DeleteCustomTemplate(ctx, req.ID); err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	c.JSON(consts.StatusOK, map[string]bool{"success": true})
}

// ParseEvalSetFile .
// @router /api/evaluation_api/parse_eval_set_file [POST]
func ParseEvalSetFile(ctx context.Context, c *app.RequestContext) {
	file, err := c.FormFile("file")
	if err != nil {
		httputil.BadRequest(c, "missing file")
		return
	}
	src, err := file.Open()
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	defer func() {
		_ = src.Close()
	}()
	content, err := io.ReadAll(src)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	rows, err := evaluation.EvalSVC.ParseImportFile(file.Filename, content)
	if err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if len(rows) == 0 {
		httputil.BadRequest(c, "file is empty")
		return
	}
	headers := rows[0]
	dataRows := rows[1:]
	c.JSON(consts.StatusOK, map[string]any{
		"headers": headers,
		"rows":    dataRows,
		"total":   len(dataRows),
	})
}

// ListTargets .
// @router /api/evaluation_api/list_targets [GET]
func ListTargets(ctx context.Context, c *app.RequestContext) {
	var req struct {
		Type int `json:"type" query:"type"`
	}
	if err := c.BindAndValidate(&req); err != nil {
		httputil.BadRequest(c, err.Error())
		return
	}
	if req.Type == 0 {
		req.Type = 1
	}
	targets, err := evaluation.EvalSVC.ListTargets(ctx, req.Type)
	if err != nil {
		httputil.InternalError(ctx, c, err)
		return
	}
	if targets == nil {
		targets = []*entity.TargetInfo{}
	}
	c.JSON(consts.StatusOK, map[string]any{"list": targets})
}
