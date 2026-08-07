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

package model

import "time"

const TableNameEvaluatorCustomTemplate = "evaluator_custom_template"

// EvaluatorCustomTemplate maps to the evaluator_custom_template table.
type EvaluatorCustomTemplate struct {
	ID          int64     `gorm:"column:id;primaryKey;autoIncrement:true;comment:id" json:"id"`
	Name        string    `gorm:"column:name;not null;comment:template name" json:"name"`
	Description string    `gorm:"column:description;not null;default:'';comment:description" json:"description"`
	Prompt      string    `gorm:"column:prompt;not null;comment:scoring prompt template" json:"prompt"`
	CreatedBy   string    `gorm:"column:created_by;not null;default:'';comment:creator id" json:"created_by"`
	CreatedAt   time.Time `gorm:"column:created_at;not null;default:now;comment:create time" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at;not null;default:now;comment:update time" json:"updated_at"`
}

// TableName sets the table name.
func (*EvaluatorCustomTemplate) TableName() string {
	return TableNameEvaluatorCustomTemplate
}
