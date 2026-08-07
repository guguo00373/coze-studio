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

import { useCallback, useRef, useState } from 'react';

import {
  Button,
  Form,
  FormInput,
  FormSelect,
  FormTextArea,
  Input,
  InputNumber,
  Modal,
  Select,
  Table,
  Tag,
  Toast,
  Typography,
  withField,
  type ColumnProps,
  type FormApi,
} from '@coze-arch/coze-design';
import { IconCozEdit, IconCozPlus, IconCozTrashCan } from '@coze-arch/coze-design/icons';

import { evaluationApi, getErrorMessage } from './api';
import { EVALUATOR_STATUS_MAP, formatDateTime } from './constants';
import { useFetch, useMutation } from './hooks';
import { type Evaluator, type EvaluatorTemplate, type ModelOption } from './types';

const { Text, Title } = Typography;
const PAGE_SIZE = 20;

const FormInputNumber = withField(InputNumber);

interface EvaluatorFormValues {
  name: string;
  description: string;
  model_id: string;
  temperature: number;
  prompt: string;
}

const EvaluatorFormModal = ({
  visible,
  initValues,
  modelOptions,
  templates,
  saving,
  onTemplatesChanged,
  onOk,
  onCancel,
}: {
  visible: boolean;
  initValues?: Evaluator;
  modelOptions: ModelOption[];
  templates: EvaluatorTemplate[];
  saving: boolean;
  onTemplatesChanged: () => void;
  onOk: (values: EvaluatorFormValues) => void;
  onCancel: () => void;
}) => {
  const formApi = useRef<FormApi<EvaluatorFormValues>>();
  const [valid, setValid] = useState(true);
  const [selectedTemplate, setSelectedTemplate] = useState<string>();
  const [promptText, setPromptText] = useState('');
  const [saveTplVisible, setSaveTplVisible] = useState(false);
  const [tplName, setTplName] = useState('');

  const applyTemplate = (key: string) => {
    const tpl = templates.find(t => t.key === key);
    if (tpl) {
      setSelectedTemplate(key);
      setPromptText(tpl.prompt_zh);
      formApi.current?.setValue('prompt', tpl.prompt_zh);
    }
  };

  const selectedTpl = templates.find(t => t.key === selectedTemplate);

  const { run: saveTemplate, loading: savingTemplate } = useMutation(
    async () => {
      if (!tplName.trim()) {
        Toast.error('请输入模板名称');
        return;
      }
      if (!promptText.trim()) {
        Toast.error('请先编写评估 Prompt');
        return;
      }
      try {
        await evaluationApi.createEvaluatorTemplate({
          name: tplName.trim(),
          description: '',
          prompt: promptText,
        });
        setSaveTplVisible(false);
        setTplName('');
        Toast.success('模板已保存');
        onTemplatesChanged();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const { run: deleteTemplate } = useMutation(async () => {
    if (!selectedTpl || selectedTpl.is_builtin) {
      return;
    }
    const id = Number(selectedTpl.key.replace('custom:', ''));
    if (!id) {
      return;
    }
    try {
      await evaluationApi.deleteEvaluatorTemplate(id);
      setSelectedTemplate(undefined);
      Toast.success('模板已删除');
      onTemplatesChanged();
    } catch (e) {
      Toast.error(getErrorMessage(e));
    }
  });

  const templateOptions = templates.map(t => ({
    label: t.is_builtin ? t.name_zh : `${t.name_zh}（自定义）`,
    value: t.key,
  }));

  return (
    <Modal
      size="medium"
      title={initValues ? '编辑评估器' : '创建评估器'}
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
      <Form<EvaluatorFormValues>
        initValues={{
          name: initValues?.name ?? '',
          description: initValues?.description ?? '',
          model_id: initValues?.model_id ?? '',
          temperature: initValues?.temperature ?? 0,
          prompt: initValues?.prompt ?? '',
        }}
        onValueChange={value => {
          const values = value as Partial<EvaluatorFormValues>;
          setValid(!!values?.name?.trim() && !!values?.model_id);
          if (typeof values?.prompt === 'string') {
            setPromptText(values.prompt);
          }
        }}
        getFormApi={api => {
          formApi.current = api;
        }}
      >
        <FormInput
          field="name"
          label="名称"
          placeholder="请输入评估器名称"
          rules={[{ required: true }]}
        />
        <FormInput
          field="description"
          label="描述"
          placeholder="请输入评估器描述"
        />
        {modelOptions.length ? (
          <FormSelect
            field="model_id"
            label="模型"
            placeholder="请选择评估模型"
            optionList={modelOptions}
            rules={[{ required: true }]}
          />
        ) : (
          <FormInput
            field="model_id"
            label="模型 ID"
            placeholder="请输入模型 ID，如 100001"
            rules={[{ required: true }]}
          />
        )}
        <FormInputNumber
          field="temperature"
          label="温度"
          min={0}
          max={2}
          step={0.1}
          placeholder="0.7"
        />
        {templates.length ? (
          <div className="mb-[16px]">
            <div className="flex items-center justify-between mb-[4px]">
              <Text>模板</Text>
              <Button
                theme="borderless"
                type="tertiary"
                onClick={() => setSaveTplVisible(true)}
              >
                保存当前 Prompt 为模板
              </Button>
            </div>
            <div className="flex items-center gap-[8px]">
              <Select
                placeholder="选择模板自动填充评估 Prompt（含自定义）"
                optionList={templateOptions}
                value={selectedTemplate}
                onChange={(key: string) => applyTemplate(key)}
                style={{ flex: 1 }}
              />
              {selectedTpl && !selectedTpl.is_builtin ? (
                <Button
                  theme="borderless"
                  type="danger"
                  icon={<IconCozTrashCan />}
                  onClick={() => deleteTemplate()}
                >
                  删除模板
                </Button>
              ) : null}
            </div>
          </div>
        ) : null}
        <FormTextArea
          field="prompt"
          label="评估 Prompt"
          placeholder={
            '可选，使用 {{input}}、{{expected}}、{{output}} 占位符。留空使用默认评估模板。'
          }
          autosize={{ minRows: 5 }}
        />
      </Form>
      <Modal
        size="small"
        title="保存为模板"
        visible={saveTplVisible}
        okText="保存"
        cancelText="取消"
        okButtonProps={{ loading: savingTemplate }}
        onOk={() => saveTemplate()}
        onCancel={() => {
          setSaveTplVisible(false);
          setTplName('');
        }}
      >
        <Text className="block mb-[4px]">模板名称</Text>
        <Input
          value={tplName}
          onChange={(v: string) => setTplName(v ?? '')}
          placeholder="请输入模板名称"
          maxLength={100}
        />
      </Modal>
    </Modal>
  );
};

const EvaluatorListPage = () => {
  const [page, setPage] = useState(1);
  const [modalVisible, setModalVisible] = useState(false);
  const [editing, setEditing] = useState<Evaluator | undefined>(undefined);

  const { data, loading, refresh: reload } = useFetch(
    () => evaluationApi.listEvaluators(page, PAGE_SIZE),
    [page],
  );

  const { data: modelOptions = [] } = useFetch(
    () => evaluationApi.listModels(),
    [modalVisible],
  );

  const {
    data: templates = [],
    refresh: reloadTemplates,
  } = useFetch(() => evaluationApi.listEvaluatorTemplates(), [modalVisible]);

  const { run: saveEvaluator, loading: saving } = useMutation(
    async (values: EvaluatorFormValues) => {
      try {
        if (editing) {
          await evaluationApi.updateEvaluator({
            id: editing.id,
            ...values,
          });
        } else {
          await evaluationApi.createEvaluator({
            space_id: 0,
            type: 1,
            ...values,
          });
        }
        setModalVisible(false);
        Toast.success('保存成功');
        reload();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const { run: removeEvaluator } = useMutation(
    async (id: number) => {
      try {
        await evaluationApi.deleteEvaluator(id);
        Toast.success('删除成功');
        reload();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const handleDelete = useCallback(
    (record: Evaluator) => {
      Modal.confirm({
        title: '删除评估器',
        content: `确定删除评估器「${record.name}」吗？`,
        okText: '删除',
        cancelText: '取消',
        onOk: () => removeEvaluator(record.id),
      });
    },
    [removeEvaluator],
  );

  const columns: ColumnProps<Evaluator>[] = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (value, record) => (
        <Text strong>
          {value}
          <Text type="secondary" fontSize="12px">
            {' '}
            (#{record.id})
          </Text>
        </Text>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: { showTooltip: true },
    },
    {
      title: '模型',
      dataIndex: 'model_id',
      width: 160,
      render: value => {
        const option = modelOptions.find(o => o.value === value);
        return (option?.name ?? value) as string;
      },
    },
    {
      title: '温度',
      dataIndex: 'temperature',
      width: 90,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: value => {
        const conf = EVALUATOR_STATUS_MAP[value];
        return <Tag color={conf?.color ?? 'grey'}>{conf?.label ?? value}</Tag>;
      },
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
      width: 140,
      render: (_value, record) => (
        <span className="flex items-center gap-[8px]">
          <Button
            theme="borderless"
            type="tertiary"
            icon={<IconCozEdit />}
            onClick={() => {
              setEditing(record);
              setModalVisible(true);
            }}
          >
            编辑
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
          评估器
        </Title>
        <Button
          theme="solid"
          type="primary"
          icon={<IconCozPlus />}
          onClick={() => {
            setEditing(undefined);
            setModalVisible(true);
          }}
        >
          新增评估器
        </Button>
      </div>
      <Table<Evaluator>
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
        empty={<div className="py-[40px] text-center coz-fg-secondary">暂无评估器</div>}
      />
      <EvaluatorFormModal
        visible={modalVisible}
        initValues={editing}
        modelOptions={modelOptions}
        templates={templates}
        saving={saving}
        onTemplatesChanged={reloadTemplates}
        onOk={values => saveEvaluator(values)}
        onCancel={() => setModalVisible(false)}
      />
    </div>
  );
};

export default EvaluatorListPage;
