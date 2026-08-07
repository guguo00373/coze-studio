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
	"gorm.io/gorm"
)

// NewEvalSetDAO creates eval set data access object.
func NewEvalSetDAO(db *gorm.DB) *EvalSetDAO {
	return &EvalSetDAO{db: db}
}

// NewEvalSetItemDAO creates eval set item data access object.
func NewEvalSetItemDAO(db *gorm.DB) *EvalSetItemDAO {
	return &EvalSetItemDAO{db: db}
}

// NewEvaluatorDAO creates evaluator data access object.
func NewEvaluatorDAO(db *gorm.DB) *EvaluatorDAO {
	return &EvaluatorDAO{db: db}
}

// NewExperimentDAO creates experiment data access object.
func NewExperimentDAO(db *gorm.DB) *ExperimentDAO {
	return &ExperimentDAO{db: db}
}

// NewExperimentItemResultDAO creates experiment item result data access object.
func NewExperimentItemResultDAO(db *gorm.DB) *ExperimentItemResultDAO {
	return &ExperimentItemResultDAO{db: db}
}

// NewExperimentAggrResultDAO creates experiment aggr result data access object.
func NewExperimentAggrResultDAO(db *gorm.DB) *ExperimentAggrResultDAO {
	return &ExperimentAggrResultDAO{db: db}
}
