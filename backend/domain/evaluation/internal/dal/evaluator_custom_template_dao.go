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

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/internal/dal/model"
)

type EvaluatorCustomTemplateDAO struct {
	db *gorm.DB
}

func NewEvaluatorCustomTemplateDAO(db *gorm.DB) *EvaluatorCustomTemplateDAO {
	return &EvaluatorCustomTemplateDAO{db: db}
}

func (dao *EvaluatorCustomTemplateDAO) Create(ctx context.Context, tpl *model.EvaluatorCustomTemplate) (*model.EvaluatorCustomTemplate, error) {
	if err := dao.db.WithContext(ctx).Create(tpl).Error; err != nil {
		return nil, err
	}
	return tpl, nil
}

func (dao *EvaluatorCustomTemplateDAO) Delete(ctx context.Context, id int64, createdBy string) error {
	return dao.db.WithContext(ctx).
		Where("id = ? AND created_by = ?", id, createdBy).
		Delete(&model.EvaluatorCustomTemplate{}).Error
}

func (dao *EvaluatorCustomTemplateDAO) ListByCreatedBy(ctx context.Context, createdBy string) ([]*model.EvaluatorCustomTemplate, error) {
	var list []*model.EvaluatorCustomTemplate
	err := dao.db.WithContext(ctx).
		Where("created_by = ?", createdBy).
		Order("id ASC").
		Find(&list).Error
	return list, err
}
