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
  Input,
  Modal,
  Select,
  Space,
  Switch,
  Table,
  Tag,
  TextArea,
  Toast,
  Typography,
  type ColumnProps,
} from '@coze-arch/coze-design';
import { IconCozPlus, IconCozTrashCan, IconCozEdit } from '@coze-arch/coze-design/icons';

import { evaluationApi, getErrorMessage } from './api';
import { EVAL_SET_STATUS_MAP, formatDateTime } from './constants';
import { useFetch, useMutation } from './hooks';
import { type EvalSet } from './types';

const { Text, Title } = Typography;
const PAGE_SIZE = 20;

export interface ColumnDef {
  name: string;
  data_type: string;
  required: boolean;
  description: string;
}

interface EvalSetFormValues {
  name: string;
  description: string;
  columns: ColumnDef[];
}

const DATA_TYPE_OPTIONS = [
  { label: 'String', value: 'string' },
  { label: 'Number', value: 'number' },
  { label: 'Boolean', value: 'boolean' },
  { label: 'Object', value: 'object' },
];

const DEFAULT_COLUMNS: ColumnDef[] = [
  {
    name: 'input',
    data_type: 'string',
    required: false,
    description: '作为输入投递给评测对象',
  },
  {
    name: 'reference_output',
    data_type: 'string',
    required: false,
    description: '预期理想输出，可作为评估时的参考标准',
  },
];

export function parseSchemaToColumns(schemaJson?: string): ColumnDef[] {
  if (!schemaJson) {
    return [];
  }
  try {
    const obj = JSON.parse(schemaJson);
    if (obj && Array.isArray(obj.field_schemas)) {
      return obj.field_schemas.map((f: Record<string, unknown>) => ({
        name: typeof f.name === 'string' ? f.name : '',
        data_type: typeof f.data_type === 'string' ? f.data_type : 'string',
        required: !!f.required,
        description: typeof f.description === 'string' ? f.description : '',
      }));
    }
    if (obj && typeof obj === 'object') {
      return Object.entries(obj as Record<string, unknown>).map(([key, value]) => ({
        name: key,
        data_type: typeof value === 'string' ? value : 'string',
        required: false,
        description: '',
      }));
    }
  } catch {
    // ignore invalid schema
  }
  return [];
}

export function serializeSchemaToColumns(columns: ColumnDef[]): string {
  return JSON.stringify({
    field_schemas: columns.map(c => ({
      name: c.name,
      data_type: c.data_type,
      required: c.required,
      description: c.description,
    })),
  });
}

const EvalSetFormModal = ({
  visible,
  initValues,
  saving,
  onOk,
  onCancel,
}: {
  visible: boolean;
  initValues?: EvalSet;
  saving: boolean;
  onOk: (values: EvalSetFormValues) => void;
  onCancel: () => void;
}) => {
  const [name, setName] = useState('');
  const [description, setDescription] = useState('');
  const [columns, setColumns] = useState<ColumnDef[]>([]);

  useEffect(() => {
    if (visible) {
      setName(initValues?.name ?? '');
      setDescription(initValues?.description ?? '');
      const cols = parseSchemaToColumns(initValues?.schema_json);
      setColumns(cols.length ? cols : DEFAULT_COLUMNS);
    }
  }, [visible, initValues]);

  const updateColumn = (index: number, patch: Partial<ColumnDef>) => {
    setColumns(list => list.map((c, i) => (i === index ? { ...c, ...patch } : c)));
  };

  const removeColumn = (index: number) => {
    setColumns(list => list.filter((_, i) => i !== index));
  };

  const handleOk = () => {
    if (!name.trim()) {
      Toast.error('请输入评测集名称');
      return;
    }
    onOk({
      name: name.trim(),
      description,
      columns: columns.filter(c => c.name.trim()),
    });
  };

  return (
    <Modal
      size="large"
      title={initValues ? '编辑评测集' : '新增评测集'}
      visible={visible}
      okText="确定"
      cancelText="取消"
      okButtonProps={{ disabled: !name.trim(), loading: saving }}
      onOk={handleOk}
      onCancel={onCancel}
    >
      <div className="coz-bg-fill-1 rounded-[8px] p-[20px] mb-[16px]">
        <Title heading={6} style={{ margin: '0 0 12px' }}>
          基本信息
        </Title>
        <div className="mb-[12px]">
          <div className="flex items-center justify-between mb-[4px]">
            <Text>名称</Text>
            <Text type="secondary" fontSize="12px">
              {name.length}/50
            </Text>
          </div>
          <Input
            value={name}
            onChange={(v: string) => setName(v ?? '')}
            maxLength={50}
            placeholder="请输入评测集名称"
          />
        </div>
        <div>
          <div className="flex items-center justify-between mb-[4px]">
            <Text>描述</Text>
            <Text type="secondary" fontSize="12px">
              {description.length}/200
            </Text>
          </div>
          <TextArea
            value={description}
            onChange={(v: string) => setDescription(v ?? '')}
            maxLength={200}
            placeholder="请输入评测集描述"
            autosize={{ minRows: 2 }}
          />
        </div>
      </div>

      <div className="coz-bg-fill-1 rounded-[8px] p-[20px]">
        <div className="flex items-center justify-between mb-[12px]">
          <Title heading={6} style={{ margin: 0 }}>
            配置列
          </Title>
          <Button
            theme="solid"
            type="primary"
            icon={<IconCozPlus />}
            onClick={() =>
              setColumns(list => [
                ...list,
                { name: '', data_type: 'string', required: false, description: '' },
              ])
            }
          >
            添加列
          </Button>
        </div>
        {columns.length === 0 ? (
          <div className="py-[24px] text-center coz-fg-secondary">
            暂无配置列，点击「添加列」新增
          </div>
        ) : (
          <Space vertical spacing="medium" style={{ width: '100%' }}>
            {columns.map((col, index) => (
              <div
                key={index}
                className="border coz-bd-fill-3 rounded-[8px] p-[16px] relative"
              >
                <Button
                  theme="borderless"
                  type="danger"
                  icon={<IconCozTrashCan />}
                  className="absolute top-[8px] right-[8px]"
                  onClick={() => removeColumn(index)}
                />
                <div className="flex items-start gap-[16px]">
                  <div style={{ width: '220px' }}>
                    <Text className="block mb-[4px]">名称</Text>
                    <Input
                      value={col.name}
                      onChange={(v: string) => updateColumn(index, { name: v ?? '' })}
                      maxLength={50}
                      placeholder="列名称"
                    />
                  </div>
                  <div style={{ width: '140px' }}>
                    <Text className="block mb-[4px]">数据类型</Text>
                    <Select
                      value={col.data_type}
                      optionList={DATA_TYPE_OPTIONS}
                      onChange={(v: string) => updateColumn(index, { data_type: v })}
                      style={{ width: '100%' }}
                    />
                  </div>
                  <div>
                    <Text className="block mb-[4px]">必填</Text>
                    <Switch
                      checked={col.required}
                      onChange={(checked: boolean) =>
                        updateColumn(index, { required: checked })
                      }
                    />
                  </div>
                </div>
                <div className="mt-[12px]">
                  <Text className="block mb-[4px]">描述</Text>
                  <Input
                    value={col.description}
                    onChange={(v: string) => updateColumn(index, { description: v ?? '' })}
                    maxLength={200}
                    placeholder="列描述"
                  />
                </div>
              </div>
            ))}
          </Space>
        )}
      </div>
    </Modal>
  );
};

const EvalSetListPage = () => {
  const navigate = useNavigate();
  const [page, setPage] = useState(1);
  const [modalVisible, setModalVisible] = useState(false);
  const [editing, setEditing] = useState<EvalSet | undefined>(undefined);

  const { data, loading, refresh: reload } = useFetch(
    () => evaluationApi.listEvalSets(page, PAGE_SIZE),
    [page],
  );

  const { run: saveEvalSet, loading: saving } = useMutation(
    async (values: EvalSetFormValues) => {
      const schemaJson = serializeSchemaToColumns(values.columns);
      try {
        if (editing) {
          await evaluationApi.updateEvalSet({
            id: editing.id,
            name: values.name,
            description: values.description,
            schema_json: schemaJson,
          });
        } else {
          await evaluationApi.createEvalSet({
            space_id: 0,
            name: values.name,
            description: values.description,
            schema_json: schemaJson,
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

  const { run: removeEvalSet } = useMutation(
    async (id: number) => {
      try {
        await evaluationApi.deleteEvalSet(id);
        Toast.success('删除成功');
        reload();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const handleDelete = useCallback(
    (record: EvalSet) => {
      Modal.confirm({
        title: '删除评测集',
        content: `确定删除评测集「${record.name}」吗？该操作不可恢复。`,
        okText: '删除',
        cancelText: '取消',
        onOk: () => removeEvalSet(record.id),
      });
    },
    [removeEvalSet],
  );

  const columns: ColumnProps<EvalSet>[] = [
    {
      title: '名称',
      dataIndex: 'name',
      render: (_value, record) => (
        <Text
          strong
          onClick={() => navigate(`/evaluation/sets/${record.id}`)}
          className="cursor-pointer"
        >
          {record.name}
        </Text>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: { showTooltip: true },
    },
    {
      title: '用例数',
      dataIndex: 'item_count',
      width: 90,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 90,
      render: value => {
        const conf = EVAL_SET_STATUS_MAP[value];
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
          评测集
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
          新增评测集
        </Button>
      </div>
      <Table<EvalSet>
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
        empty={<div className="py-[40px] text-center coz-fg-secondary">暂无评测集</div>}
      />
      <EvalSetFormModal
        visible={modalVisible}
        initValues={editing}
        saving={saving}
        onOk={values => saveEvalSet(values)}
        onCancel={() => setModalVisible(false)}
      />
    </div>
  );
};

export default EvalSetListPage;
