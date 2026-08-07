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
	"errors"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/internal/dal/model"
)

type EvalSetDAO struct {
	db *gorm.DB
}

func (dao *EvalSetDAO) Create(ctx context.Context, es *model.EvalSet) (*model.EvalSet, error) {
	if err := dao.db.WithContext(ctx).Create(es).Error; err != nil {
		return nil, err
	}
	return es, nil
}

func (dao *EvalSetDAO) Update(ctx context.Context, es *model.EvalSet) error {
	return dao.db.WithContext(ctx).Model(&model.EvalSet{}).Where("id = ?", es.ID).Updates(es).Error
}

func (dao *EvalSetDAO) Delete(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&model.EvalSet{}).Error
}

func (dao *EvalSetDAO) GetByID(ctx context.Context, id int64) (*model.EvalSet, error) {
	var es model.EvalSet
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&es).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &es, nil
}

func (dao *EvalSetDAO) List(ctx context.Context, createdBy string, page int, pageSize int) ([]*model.EvalSet, int64, error) {
	var list []*model.EvalSet
	var total int64
	query := dao.db.WithContext(ctx).Model(&model.EvalSet{})
	if createdBy != "" {
		query = query.Where("created_by = ?", createdBy)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (dao *EvalSetDAO) UpdateItemCount(ctx context.Context, id int64, delta int64) error {
	return dao.db.WithContext(ctx).Model(&model.EvalSet{}).
		Where("id = ?", id).
		UpdateColumn("item_count", gorm.Expr("item_count + ?", delta)).Error
}
