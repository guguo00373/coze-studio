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

package dal

import (
	"context"
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/internal/dal/model"
)

type EvalSetItemDAO struct {
	db *gorm.DB
}

func (dao *EvalSetItemDAO) BatchCreate(ctx context.Context, items []*model.EvalSetItem) error {
	if len(items) == 0 {
		return nil
	}
	return dao.db.WithContext(ctx).Create(&items).Error
}

func (dao *EvalSetItemDAO) Update(ctx context.Context, id int64, dataJSON datatypes.JSON) error {
	return dao.db.WithContext(ctx).
		Model(&model.EvalSetItem{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"data_json":  dataJSON,
			"updated_at": time.Now(),
		}).Error
}

func (dao *EvalSetItemDAO) BatchDelete(ctx context.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return dao.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model.EvalSetItem{}).Error
}

func (dao *EvalSetItemDAO) List(ctx context.Context, evalSetID int64, page int, pageSize int) ([]*model.EvalSetItem, int64, error) {
	var list []*model.EvalSetItem
	var total int64
	query := dao.db.WithContext(ctx).Model(&model.EvalSetItem{}).Where("eval_set_id = ?", evalSetID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("id ASC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (dao *EvalSetItemDAO) CountBySetID(ctx context.Context, evalSetID int64) (int64, error) {
	var total int64
	err := dao.db.WithContext(ctx).Model(&model.EvalSetItem{}).Where("eval_set_id = ?", evalSetID).Count(&total).Error
	return total, err
}
