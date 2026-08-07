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

type EvaluatorDAO struct {
	db *gorm.DB
}

func (dao *EvaluatorDAO) Create(ctx context.Context, ev *model.Evaluator) (*model.Evaluator, error) {
	if err := dao.db.WithContext(ctx).Create(ev).Error; err != nil {
		return nil, err
	}
	return ev, nil
}

func (dao *EvaluatorDAO) Update(ctx context.Context, ev *model.Evaluator) error {
	return dao.db.WithContext(ctx).Model(&model.Evaluator{}).Where("id = ?", ev.ID).Updates(ev).Error
}

func (dao *EvaluatorDAO) Delete(ctx context.Context, id int64) error {
	return dao.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Evaluator{}).Error
}

func (dao *EvaluatorDAO) GetByID(ctx context.Context, id int64) (*model.Evaluator, error) {
	var ev model.Evaluator
	err := dao.db.WithContext(ctx).Where("id = ?", id).First(&ev).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ev, nil
}

func (dao *EvaluatorDAO) List(ctx context.Context, createdBy string, page int, pageSize int) ([]*model.Evaluator, int64, error) {
	var list []*model.Evaluator
	var total int64
	query := dao.db.WithContext(ctx).Model(&model.Evaluator{})
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
