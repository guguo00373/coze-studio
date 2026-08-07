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

import { useEffect, useState } from 'react';

import { useNavigate, useParams } from 'react-router-dom';
import {
  Button,
  Modal,
  Table,
  Tag,
  Toast,
  Typography,
  type ColumnProps,
} from '@coze-arch/coze-design';
import { IconCozArrowLeft } from '@coze-arch/coze-design/icons';

import { evaluationApi, getErrorMessage } from './api';
import {
  EXPERIMENT_STATUS_MAP,
  ITEM_RESULT_STATUS_MAP,
  TARGET_TYPE_MAP,
  formatDateTime,
  parseEvaluatorResults,
  parseRunStats,
} from './constants';
import { useFetch, useMutation } from './hooks';
import {
  type EvaluatorResult,
  type Experiment,
  type ExperimentDetail,
  type ExperimentItemResult,
} from './types';

const { Text, Title } = Typography;
const PAGE_SIZE = 20;

const InfoItem = ({ label, value }: { label: string; value: string }) => (
  <div className="flex items-center gap-[8px]">
    <Text type="secondary" fontSize="12px">
      {label}
    </Text>
    <Text>{value}</Text>
  </div>
);

const ScoreTag = ({ score }: { score: number }) => {
  const color = score >= 80 ? 'green' : score >= 60 ? 'orange' : 'red';
  return <Tag color={color}>{score.toFixed(1)}</Tag>;
};

const ResultDetailModal = ({
  record,
  onClose,
}: {
  record: ExperimentItemResult | null;
  onClose: () => void;
}) => {
  const results = record ? parseEvaluatorResults(record.evaluator_results_json) : [];

  return (
    <Modal
      size="large"
      title={`用例结果 #${record?.eval_set_item_id ?? ''}`}
      visible={!!record}
      okText="关闭"
      cancelText="取消"
      onCancel={onClose}
      onOk={onClose}
    >
      {record ? (
        <div className="flex flex-col gap-[12px]">
          <div className="flex items-center gap-[8px]">
            <Tag color={ITEM_RESULT_STATUS_MAP[record.status]?.color}>
              {ITEM_RESULT_STATUS_MAP[record.status]?.label ?? record.status}
            </Tag>
            <Text type="secondary" fontSize="12px">
              耗时 {record.latency_ms}ms
            </Text>
          </div>
          <div>
            <Text strong>输入</Text>
            <pre className="coz-bg-plus rounded-[8px] p-[12px] overflow-auto max-h-[200px]">
              {record.input_json}
            </pre>
          </div>
          {record.target_error_msg ? (
            <div>
              <Text strong>运行错误</Text>
              <pre className="coz-bg-plus rounded-[8px] p-[12px] overflow-auto max-h-[200px] coz-fg-danger">
                {record.target_error_msg}
              </pre>
            </div>
          ) : null}
          <div>
            <Text strong>实际输出</Text>
            <pre className="coz-bg-plus rounded-[8px] p-[12px] overflow-auto max-h-[240px]">
              {record.actual_output || record.output_json || '-'}
            </pre>
          </div>
          {results.length ? (
            <div>
              <Text strong>评估结果</Text>
              {results.map((item: EvaluatorResult) => (
                <div
                  key={item.evaluator_id}
                  className="coz-bg-plus rounded-[8px] p-[12px] mb-[8px]"
                >
                  <div className="flex items-center gap-[8px] mb-[4px]">
                    <Text strong>评估器 #{item.evaluator_id}</Text>
                    <ScoreTag score={item.score} />
                  </div>
                  <Text type="secondary" fontSize="12px">
                    {item.reasoning}
                  </Text>
                </div>
              ))}
            </div>
          ) : null}
        </div>
      ) : null}
    </Modal>
  );
};

const ExperimentDetailPage = () => {
  const { id } = useParams<{ id: string }>();
  const experimentId = Number(id);
  const navigate = useNavigate();

  const [page, setPage] = useState(1);
  const [detailRecord, setDetailRecord] = useState<ExperimentItemResult | null>(
    null,
  );

  const { data, loading, refresh } = useFetch(
    () => evaluationApi.getExperimentDetail(experimentId, page, PAGE_SIZE),
    [experimentId, page],
  );

  const { run: startExperiment } = useMutation(
    async () => {
      try {
        await evaluationApi.startExperiment(experimentId);
        Toast.success('实验已开始运行');
        refresh();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const experiment: Experiment | undefined = (data as ExperimentDetail | undefined)
    ?.experiment;

  useEffect(() => {
    if (!experiment) {
      return;
    }
    if (experiment.status !== 1) {
      return;
    }
    const timer = setInterval(() => {
      refresh();
    }, 3000);
    return () => clearInterval(timer);
  }, [experiment?.status, refresh]);

  const statusConf = experiment
    ? EXPERIMENT_STATUS_MAP[experiment.status]
    : undefined;
  const stats = parseRunStats(experiment?.run_stats_json);

  const columns: ColumnProps<ExperimentItemResult>[] = [
    {
      title: '用例 ID',
      dataIndex: 'eval_set_item_id',
      width: 100,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: value => (
        <Tag color={ITEM_RESULT_STATUS_MAP[value]?.color}>
          {ITEM_RESULT_STATUS_MAP[value]?.label ?? value}
        </Tag>
      ),
    },
    {
      title: '输入',
      dataIndex: 'input_json',
      ellipsis: { showTooltip: true },
    },
    {
      title: '实际输出',
      dataIndex: 'actual_output',
      ellipsis: { showTooltip: true },
      render: (value, record) => value || record.output_json || '-',
    },
    {
      title: '评分',
      dataIndex: 'evaluator_results_json',
      width: 150,
      render: (value: string) => {
        const results = parseEvaluatorResults(value);
        if (!results.length) {
          return '-';
        }
        return (
          <span className="flex items-center gap-[6px] flex-wrap">
            {results.map(item => (
              <span
                key={item.evaluator_id}
                className="flex items-center gap-[4px]"
              >
                <Text type="secondary" fontSize="12px">
                  #{item.evaluator_id}
                </Text>
                <ScoreTag score={item.score} />
              </span>
            ))}
          </span>
        );
      },
    },
    {
      title: '耗时(ms)',
      dataIndex: 'latency_ms',
      width: 90,
    },
    {
      title: '操作',
      dataIndex: 'action',
      width: 80,
      render: (_value, record) => (
        <Button
          theme="borderless"
          type="tertiary"
          onClick={() => setDetailRecord(record)}
        >
          详情
        </Button>
      ),
    },
  ];

  const running = experiment?.status === 1;

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center gap-[8px] mb-[16px]">
        <Button
          theme="borderless"
          type="tertiary"
          icon={<IconCozArrowLeft />}
          onClick={() => navigate('/evaluation/experiments')}
        >
          返回
        </Button>
      </div>
      <div className="flex items-center justify-between mb-[16px]">
        <div className="flex items-center gap-[12px]">
          <Title heading={5} style={{ margin: 0 }}>
            {experiment?.name ?? '实验详情'}
          </Title>
          {statusConf ? (
            <Tag color={statusConf.color}>{statusConf.label}</Tag>
          ) : null}
        </div>
        {experiment ? (
          <Button
            theme="solid"
            type="primary"
            disabled={running}
            loading={running}
            onClick={() => startExperiment()}
          >
            {running ? '运行中...' : '重新运行'}
          </Button>
        ) : null}
      </div>
      {experiment ? (
        <>
          <div className="grid grid-cols-2 gap-[12px] mb-[16px] coz-bg-plus rounded-[12px] p-[16px]">
            <InfoItem label="评测对象" value={`${TARGET_TYPE_MAP[experiment.target_type] ?? '未知'} / ${experiment.target_id}`} />
            <InfoItem label="评测集" value={`#${experiment.eval_set_id}`} />
            <InfoItem
              label="评估器"
              value={
                experiment.evaluator_ids
                  ? experiment.evaluator_ids
                      .split(',')
                      .filter(Boolean)
                      .map(id => `#${id}`)
                      .join(', ')
                  : '-'
              }
            />
            <InfoItem label="并发数" value={String(experiment.concurrency)} />
            <InfoItem label="创建时间" value={formatDateTime(experiment.created_at)} />
          </div>
          {stats ? (
            <div className="flex items-center gap-[16px] mb-[16px]">
              <Text>
                成功 <Tag color="green">{stats.success_count ?? 0}</Tag>
              </Text>
              <Text>
                失败 <Tag color="red">{stats.failed_count ?? 0}</Tag>
              </Text>
              {stats.scores
                ? Object.entries(stats.scores).map(
                    ([evaluatorId, score]) => (
                      <Text key={evaluatorId}>
                        评估器 #{evaluatorId} 均分
                        <ScoreTag score={score.average} />
                      </Text>
                    ),
                  )
                : null}
            </div>
          ) : null}
          {experiment.error_msg ? (
            <pre className="coz-bg-plus rounded-[8px] p-[12px] overflow-auto max-h-[120px] coz-fg-danger mb-[16px]">
              {experiment.error_msg}
            </pre>
          ) : null}
          <Table<ExperimentItemResult>
            tableProps={{
              columns,
              dataSource: data?.results ?? [],
              rowKey: 'id',
              loading,
              pagination: {
                total: data?.total ?? 0,
                currentPage: page,
                pageSize: PAGE_SIZE,
                onChange: setPage,
              },
            }}
            empty={
              <div className="py-[40px] text-center coz-fg-secondary">
                暂无结果
              </div>
            }
          />
        </>
      ) : null}
      <ResultDetailModal record={detailRecord} onClose={() => setDetailRecord(null)} />
    </div>
  );
};

export default ExperimentDetailPage;
