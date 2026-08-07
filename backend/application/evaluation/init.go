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
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/repository"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/service"
	"github.com/coze-dev/coze-studio/backend/infra/idgen"
)

var EvalSVC *EvaluationApplicationService

func InitService(db *gorm.DB, idGenSVC idgen.IDGenerator) *EvaluationApplicationService {
	repo := repository.NewEvaluationRepository(db)
	runner := service.NewTargetRunnerFactory()
	evaluator := service.NewEvaluatorExecutor()

	components := &service.Components{
		Repo:       repo,
		Runner:     runner,
		Evaluator:  evaluator,
		Experiment: service.NewExperimentEngine(repo, runner, evaluator),
	}

	EvalSVC = &EvaluationApplicationService{
		EvalDomainSVC: service.NewEvaluationService(components),
	}
	return EvalSVC
}
