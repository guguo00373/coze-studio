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

import { useCallback, useEffect, useRef, useState } from 'react';

import { useNavigate } from 'react-router-dom';
import {
  Button,
  Form,
  FormInput,
  FormSelect,
  InputNumber,
  Modal,
  Table,
  Tag,
  Toast,
  Typography,
  withField,
  type ColumnProps,
  type FormApi,
} from '@coze-arch/coze-design';
import { IconCozPlus, IconCozTrashCan } from '@coze-arch/coze-design/icons';

import { evaluationApi, getErrorMessage } from './api';
import {
  EXPERIMENT_STATUS_MAP,
  TARGET_TYPE_MAP,
  TARGET_TYPE_OPTIONS,
  formatDateTime,
  parseRunStats,
} from './constants';
import { useFetch, useMutation } from './hooks';
import { type Experiment, type EvalSet, type Evaluator } from './types';

const { Text, Title } = Typography;
const PAGE_SIZE = 20;

const FormInputNumber = withField(InputNumber);

interface ExperimentFormValues {
  name: string;
  description: string;
  eval_set_id: number;
  target_type: number;
  target_id: string;
  concurrency: number;
  evaluator_ids: number[];
}

const ExperimentFormModal = ({
  visible,
  evalSets,
  evaluators,
  saving,
  onOk,
  onCancel,
}: {
  visible: boolean;
  evalSets: EvalSet[];
  evaluators: Evaluator[];
  saving: boolean;
  onOk: (values: ExperimentFormValues) => void;
  onCancel: () => void;
}) => {
  const formApi = useRef<FormApi<ExperimentFormValues>>();
  const [valid, setValid] = useState(true);
  const [targetType, setTargetType] = useState<number>(1);

  const { data: targets = [] } = useFetch(
    () => evaluationApi.listTargets(targetType),
    [targetType, visible],
  );

  // Reset the target selection whenever the target type changes or the modal
  // reopens. Avoids calling setValue inside onValueChange (nested form
  // mutation can crash the Semi form when the target_id field swaps between
  // Select and Input).
  useEffect(() => {
    formApi.current?.setValue('target_id', undefined);
  }, [targetType, visible]);

  const evalSetOptions = evalSets.map(set => ({
    label: set.name,
    value: set.id,
  }));
  const evaluatorOptions = evaluators.map(item => ({
    label: `${item.name} (#${item.id})`,
    value: item.id,
  }));
  const targetOptions = targets.map(t => ({
    label: `${t.name} (${t.target_id})`,
    value: t.target_id,
  }));

  return (
    <Modal
      size="medium"
      title="创建实验"
      visible={visible}
      okText="确定"
      cancelText="取消"
      okButtonProps={{ disabled: !valid, loading: saving }}
      onOk={async () => {
        try {
          const values = await formApi.current?.validate();
          if (values) {
            onOk(values);
          }
        } catch {
          Toast.error('请完善必填项');
        }
      }}
      onCancel={onCancel}
    >
      <Form<ExperimentFormValues>
        initValues={{
          name: '',
          description: '',
          eval_set_id: undefined,
          target_type: 1,
          target_id: '',
          concurrency: 1,
          evaluator_ids: [],
        }}
        onValueChange={value => {
          const values = value as Partial<ExperimentFormValues>;
          const nextType =
            typeof values?.target_type === 'number' ? values.target_type : 1;
          if (nextType !== targetType) {
            setTargetType(nextType);
          }
          setValid(
            !!values?.name?.trim() &&
              !!values?.eval_set_id &&
              !!values?.target_id?.trim() &&
              !!nextType,
          );
        }}
        getFormApi={api => {
          formApi.current = api;
        }}
      >
        <FormInput
          field="name"
          label="名称"
          placeholder="请输入实验名称"
          rules={[{ required: true }]}
        />
        <FormInput
          field="description"
          label="描述"
          placeholder="请输入实验描述"
        />
        <FormSelect
          field="eval_set_id"
          label="评测集"
          placeholder="请选择评测集"
          optionList={evalSetOptions}
          rules={[{ required: true }]}
        />
        <FormSelect
          field="target_type"
          label="评测对象类型"
          placeholder="请选择评测对象类型"
          optionList={TARGET_TYPE_OPTIONS}
          rules={[{ required: true }]}
        />
        {targetOptions.length ? (
          <FormSelect
            field="target_id"
            label="评测对象"
            placeholder="请选择评测对象"
            optionList={targetOptions}
            rules={[{ required: true }]}
          />
        ) : (
          <FormInput
            field="target_id"
            label="评测对象 ID"
            placeholder="请输入 Agent / 工作流 / 对话流 ID"
            rules={[{ required: true }]}
          />
        )}
        <FormInputNumber
          field="concurrency"
          label="并发数"
          min={1}
          max={20}
          step={1}
          placeholder="1"
        />
        <FormSelect
          field="evaluator_ids"
          label="评估器"
          placeholder="请选择评估器（可多选）"
          optionList={evaluatorOptions}
          multiple
          rules={[{ required: true, type: 'array' }]}
        />
      </Form>
    </Modal>
  );
};

const ExperimentListPage = () => {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [modalVisible, setModalVisible] = useState(false);

  const { data, loading, refresh: reload } = useFetch(
    () => evaluationApi.listExperiments(page, PAGE_SIZE),
    [page],
  );

  const { data: evalSetData } = useFetch(
    () => evaluationApi.listEvalSets(1, 200),
    [modalVisible],
  );
  const { data: evaluatorData } = useFetch(
    () => evaluationApi.listEvaluators(1, 200),
    [modalVisible],
  );
  const { data: agentTargets = [] } = useFetch(
    () => evaluationApi.listTargets(1),
    [modalVisible],
  );
  const { data: workflowTargets = [] } = useFetch(
    () => evaluationApi.listTargets(2),
    [modalVisible],
  );
  const { data: chatflowTargets = [] } = useFetch(
    () => evaluationApi.listTargets(3),
    [modalVisible],
  );

  const targetNameMap = Object.fromEntries(
    [...agentTargets, ...workflowTargets, ...chatflowTargets].map(t => [
      t.target_id,
      t.name,
    ]),
  );
  const evalSetNameMap = Object.fromEntries(
    (evalSetData?.list ?? []).map(set => [set.id, set.name]),
  );
  const evaluatorNameMap = Object.fromEntries(
    (evaluatorData?.list ?? []).map(item => [item.id, item.name]),
  );

  const { run: createExperiment, loading: saving } = useMutation(
    async (values: ExperimentFormValues) => {
      try {
        await evaluationApi.createExperiment({
          space_id: 0,
          ...values,
          target_config: '',
        });
        setModalVisible(false);
        Toast.success('创建成功');
        reload();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const { run: removeExperiment } = useMutation(
    async (id: number) => {
      try {
        await evaluationApi.deleteExperiment(id);
        Toast.success('删除成功');
        reload();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const { run: startExperiment } = useMutation(
    async (id: number) => {
      try {
        await evaluationApi.startExperiment(id);
        Toast.success('实验已开始运行');
        reload();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const handleDelete = useCallback(
    (record: Experiment) => {
      Modal.confirm({
        title: '删除实验',
        content: `确定删除实验「${record.name}」吗？该操作不可恢复。`,
        okText: '删除',
        cancelText: '取消',
        onOk: () => removeExperiment(record.id),
      });
    },
    [removeExperiment],
  );

  const handleStart = useCallback(
    (record: Experiment) => {
      Modal.confirm({
        title: '运行实验',
        content: `确定开始运行实验「${record.name}」吗？将使用评测集 #${record.eval_set_id} 的全部用例执行。`,
        okText: '运行',
        cancelText: '取消',
        onOk: () => startExperiment(record.id),
      });
    },
    [startExperiment],
  );

  const renderRunStats = (record: Experiment) => {
    const stats = parseRunStats(record.run_stats_json);
    if (!stats) {
      return '-';
    }
    const parts: string[] = [];
    if (stats.success_count !== undefined) {
      parts.push(`成功 ${stats.success_count}`);
    }
    if (stats.failed_count !== undefined) {
      parts.push(`失败 ${stats.failed_count}`);
    }
    if (stats.scores) {
      Object.entries(stats.scores).forEach(([evaluatorId, score]) => {
        parts.push(`评估器#${evaluatorId} 均分 ${score.average?.toFixed(1)}`);
      });
    }
    return parts.length ? parts.join(' / ') : '-';
  };

  const columns: ColumnProps<Experiment>[] = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (value, record) => (
        <Text
          strong
          onClick={() => navigate(`/evaluation/experiments/${record.id}`)}
          className="cursor-pointer"
        >
          {value}
        </Text>
      ),
    },
    {
      title: '评测对象',
      dataIndex: 'target_type',
      width: 180,
      render: (_value, record) => {
        const typeLabel = TARGET_TYPE_MAP[record.target_type] ?? '未知';
        const name = targetNameMap[record.target_id];
        return `${typeLabel} / ${name || record.target_id}`;
      },
    },
    {
      title: '评测集',
      dataIndex: 'eval_set_id',
      width: 120,
      render: value => evalSetNameMap[value] || `#${value}`,
    },
    {
      title: '评估器',
      dataIndex: 'evaluator_ids',
      width: 180,
      render: (value: string) => {
        if (!value) {
          return '-';
        }
        const ids: number[] = [];
        try {
          const parsed = JSON.parse(value);
          if (Array.isArray(parsed)) {
            parsed.forEach(v => {
              const n = Number(v);
              if (!isNaN(n)) ids.push(n);
            });
          }
        } catch {
          // not JSON array, fall through to comma-split
        }
        if (!ids.length) {
          value
            .split(',')
            .forEach(s => {
              const n = Number(s.trim());
              if (!isNaN(n)) ids.push(n);
            });
        }
        return ids
          .map(id => evaluatorNameMap[id] || `#${id}`)
          .join(', ');
      },
    },
    {
      title: '并发',
      dataIndex: 'concurrency',
      width: 70,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: value => {
        const conf = EXPERIMENT_STATUS_MAP[value];
        return <Tag color={conf?.color ?? 'grey'}>{conf?.label ?? value}</Tag>;
      },
    },
    {
      title: '运行结果',
      dataIndex: 'run_stats_json',
      render: (_value, record) => renderRunStats(record),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 170,
      render: value => formatDateTime(value as string),
    },
    {
      title: '操作',
      dataIndex: 'action',
      width: 150,
      render: (_value, record) => (
        <span className="flex items-center gap-[8px]">
          <Button
            theme="borderless"
            type="primary"
            disabled={record.status === 1}
            onClick={() => handleStart(record)}
          >
            运行
          </Button>
          <Button
            theme="borderless"
            type="danger"
            icon={<IconCozTrashCan />}
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
        </span>
      ),
    },
  ];

  return (
    <div className="flex flex-col h-full">
      <div className="flex items-center justify-between mb-[16px]">
        <Title heading={5} style={{ margin: 0 }}>
          实验
        </Title>
        <Button
          theme="solid"
          type="primary"
          icon={<IconCozPlus />}
          onClick={() => setModalVisible(true)}
        >
          新增实验
        </Button>
      </div>
      <Table<Experiment>
        tableProps={{
          columns,
          dataSource: data?.list ?? [],
          rowKey: 'id',
          loading,
          pagination: {
            total: data?.total ?? 0,
            currentPage: page,
            pageSize: PAGE_SIZE,
            onChange: setPage,
          },
        }}
        empty={<div className="py-[40px] text-center coz-fg-secondary">暂无实验</div>}
      />
      <ExperimentFormModal
        visible={modalVisible}
        evalSets={evalSetData?.list ?? []}
        evaluators={evaluatorData?.list ?? []}
        saving={saving}
        onOk={values => createExperiment(values)}
        onCancel={() => setModalVisible(false)}
      />
    </div>
  );
};

export default ExperimentListPage;
