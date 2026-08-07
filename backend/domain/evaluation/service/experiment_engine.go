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

package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"

	"gorm.io/datatypes"

	"github.com/coze-dev/coze-studio/backend/domain/evaluation/entity"
	"github.com/coze-dev/coze-studio/backend/domain/evaluation/repository"
	"github.com/coze-dev/coze-studio/backend/pkg/taskgroup"
)

type experimentEngine struct {
	repo      repository.EvaluationRepository
	runner    *targetRunnerFactory
	evaluator *evaluatorExecutor
	nowMillis func() int64
}

func NewExperimentEngine(repo repository.EvaluationRepository, runner *targetRunnerFactory, evaluator *evaluatorExecutor) *experimentEngine {
	return &experimentEngine{
		repo:      repo,
		runner:    runner,
		evaluator: evaluator,
		nowMillis: nowMillis,
	}
}

func (e *experimentEngine) Run(ctx context.Context, expt *entity.Experiment) error {
	items, total, err := e.repo.ListEvalSetItems(ctx, expt.EvalSetID, 1, 1)
	if err != nil {
		return err
	}
	_ = items
	if total == 0 {
		return fmt.Errorf("evaluation set %d has no items", expt.EvalSetID)
	}

	runner := e.runner.GetRunner(expt.TargetType)
	if runner == nil {
		return fmt.Errorf("unsupported target type %d", expt.TargetType)
	}

	evaluatorIDs := parseInt64List(expt.EvaluatorIDs)
	evaluators := make([]*entity.Evaluator, 0, len(evaluatorIDs))
	for _, evID := range evaluatorIDs {
		ev, err := e.repo.GetEvaluatorByID(ctx, evID)
		if err != nil {
			return err
		}
		if ev == nil {
			return fmt.Errorf("evaluator %d not found", evID)
		}
		evaluators = append(evaluators, ev)
	}

	if err := e.repo.UpdateExperimentStatus(ctx, expt.ID, int(entity.ExperimentStatusRunning)); err != nil {
		return err
	}

	go e.execute(ctx, expt, runner, evaluators)
	return nil
}

func (e *experimentEngine) execute(ctx context.Context, expt *entity.Experiment, runner TargetRunner, evaluators []*entity.Evaluator) {
	concurrency := expt.Concurrency
	if concurrency <= 0 {
		concurrency = 1
	}

	var (
		mu          sync.Mutex
		successCnt  int64
		failedCnt   int64
		scoreSums   = make(map[int64]float64)
		scoreCounts = make(map[int64]int64)
	)

	page := 1
	pageSize := 200
	for {
		items, total, err := e.repo.ListEvalSetItems(ctx, expt.EvalSetID, page, pageSize)
		if err != nil {
			e.finishFailed(ctx, expt, err)
			return
		}
		if len(items) == 0 {
			break
		}

		g := taskgroup.NewTaskGroup(ctx, concurrency)
		for _, item := range items {
			item := item
			g.Go(func() error {
				// Only deliver the case input to the target. The reference /
				// expected output is an answer for the evaluator and must never
				// be sent to the target, otherwise the target can "cheat" by
				// echoing it back.
				targetInput := buildTargetInput(string(item.DataJSON), entity.TargetType(expt.TargetType))
				runResult, runErr := runner.Run(ctx, expt.TargetID, targetInput)
				if runResult != nil {
					runResult.Input, runResult.Expected = parseCaseData(string(item.DataJSON))
				}
				evaluatorResults, evalErr := e.score(ctx, evaluators, runResult, runErr)

				itemResult := &entity.ExperimentItemResult{
					ExperimentID:  expt.ID,
					EvalSetItemID: item.ID,
					InputJSON:     item.DataJSON,
				}
				mu.Lock()
				defer mu.Unlock()

				if runErr != nil {
					itemResult.Status = int(entity.ItemResultStatusFailed)
					itemResult.TargetErrorMsg = runErr.Error()
				} else {
					itemResult.Status = int(entity.ItemResultStatusSuccess)
					itemResult.ActualOutput = runResult.ActualOutput
					if runResult != nil {
						tokens, _ := json.Marshal(map[string]any{
							"input_tokens":  runResult.InputTokens,
							"output_tokens": runResult.OutputTokens,
							"latency_ms":    runResult.LatencyMS,
						})
						itemResult.TokensUsed = datatypes.JSON(tokens)
						itemResult.LatencyMS = runResult.LatencyMS
					}
				}

				if evalErr != nil {
					if itemResult.TargetErrorMsg == "" {
						itemResult.TargetErrorMsg = evalErr.Error()
					}
					if itemResult.Status == int(entity.ItemResultStatusSuccess) {
						itemResult.Status = int(entity.ItemResultStatusFailed)
					}
				} else if len(evaluatorResults) > 0 {
					b, _ := json.Marshal(evaluatorResults)
					itemResult.EvaluatorResults = datatypes.JSON(b)
					for _, er := range evaluatorResults {
						scoreSums[er.EvaluatorID] += er.Score
						scoreCounts[er.EvaluatorID]++
					}
				}

				// Count by the final status so stats stay consistent with the
				// stored item results (an evaluator failure demotes a success).
				if itemResult.Status == int(entity.ItemResultStatusSuccess) {
					successCnt++
				} else {
					failedCnt++
				}

				_, err := e.repo.CreateExperimentItemResult(ctx, itemResult)
				return err
			})
		}
		if err := g.Wait(); err != nil {
			e.finishFailed(ctx, expt, err)
			return
		}

		if int64((page-1)*pageSize)+int64(len(items)) >= total {
			break
		}
		page++
	}

	e.finishSuccess(ctx, expt, successCnt, failedCnt, scoreSums, scoreCounts)
}

func (e *experimentEngine) score(ctx context.Context, evaluators []*entity.Evaluator, runResult *entity.TargetRunResult, runErr error) ([]*entity.EvaluatorResult, error) {
	if runErr != nil {
		return nil, runErr
	}
	if runResult == nil {
		return nil, fmt.Errorf("target run returned nil result")
	}
	if len(evaluators) == 0 {
		return nil, fmt.Errorf("experiment has no evaluators")
	}

	results := make([]*entity.EvaluatorResult, 0, len(evaluators))
	for _, ev := range evaluators {
		er, err := e.evaluator.Evaluate(ctx, ev, runResult)
		if err != nil {
			return nil, err
		}
		if er != nil {
			er.EvaluatorID = ev.ID
			results = append(results, er)
		}
	}
	return results, nil
}

func (e *experimentEngine) finishSuccess(ctx context.Context, expt *entity.Experiment, successCnt, failedCnt int64, scoreSums map[int64]float64, scoreCounts map[int64]int64) {
	status := entity.ExperimentStatusSuccess
	if failedCnt > 0 && successCnt > 0 {
		status = entity.ExperimentStatusPartial
	} else if failedCnt > 0 {
		status = entity.ExperimentStatusFailed
	}

	scores := make(map[string]any)
	for evID, sum := range scoreSums {
		scores[strconv.FormatInt(evID, 10)] = map[string]any{
			"average": sum / float64(scoreCounts[evID]),
			"count":   scoreCounts[evID],
		}
	}
	stats, _ := json.Marshal(map[string]any{
		"success_count": successCnt,
		"failed_count":  failedCnt,
		"scores":        scores,
	})

	if err := e.repo.UpdateExperimentStatus(ctx, expt.ID, int(status)); err != nil {
		return
	}
	_ = e.repo.UpdateExperiment(ctx, &entity.Experiment{
		ID:       expt.ID,
		RunStats: datatypes.JSON(stats),
	})
	_, _ = e.repo.CreateExperimentAggrResult(ctx, &entity.ExperimentAggrResult{
		ExperimentID:     expt.ID,
		EvaluatorScores:  datatypes.JSON(mustJSON(scores)),
		TotalItemCount:   successCnt + failedCnt,
		SuccessItemCount: successCnt,
	})
}

func (e *experimentEngine) finishFailed(ctx context.Context, expt *entity.Experiment, err error) {
	_ = e.repo.UpdateExperimentStatus(ctx, expt.ID, int(entity.ExperimentStatusFailed))
	_ = e.repo.UpdateExperiment(ctx, &entity.Experiment{
		ID:       expt.ID,
		ErrorMsg: err.Error(),
	})
}

func parseInt64List(raw datatypes.JSON) []int64 {
	var list []int64
	if len(raw) == 0 {
		return nil
	}
	_ = json.Unmarshal([]byte(raw), &list)
	return list
}

func mustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
