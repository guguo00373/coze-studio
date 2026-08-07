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

import { useEffect, useRef, useState } from 'react';

import { useNavigate, useParams } from 'react-router-dom';
import {
  Button,
  Form,
  FormTextArea,
  Input,
  Modal,
  Select,
  Table,
  Tag,
  TextArea,
  Toast,
  Typography,
  type ColumnProps,
} from '@coze-arch/coze-design';
import {
  IconCozArrowLeft,
  IconCozEdit,
  IconCozPlus,
  IconCozTrashCan,
  IconCozUpload,
} from '@coze-arch/coze-design/icons';

import { evaluationApi, getErrorMessage } from './api';
import { EVAL_SET_STATUS_MAP, formatDateTime } from './constants';
import { useFetch, useMutation } from './hooks';
import { parseSchemaToColumns, type ColumnDef } from './EvalSetListPage';
import { type EvalSet, type EvalSetItem, type ParsedImportFile } from './types';

const { Text, Title } = Typography;
const PAGE_SIZE = 20;

interface AddItemsValues {
  raw: string;
}

function parseItemsInput(raw: string): { data_json: string }[] {
  const trimmed = raw.trim();
  if (!trimmed) {
    return [];
  }
  const parsed = JSON.parse(trimmed);
  if (Array.isArray(parsed)) {
    return parsed.map(item => ({ data_json: JSON.stringify(item) }));
  }
  return [{ data_json: JSON.stringify(parsed) }];
}

const AddItemsModal = ({
  visible,
  adding,
  onOk,
  onCancel,
}: {
  visible: boolean;
  adding: boolean;
  onOk: (raw: string) => void;
  onCancel: () => void;
}) => {
  const [raw, setRaw] = useState('');

  return (
    <Modal
      size="medium"
      title="批量添加用例"
      visible={visible}
      okText="添加"
      cancelText="取消"
      okButtonProps={{ disabled: !raw.trim(), loading: adding }}
      onOk={() => onOk(raw)}
      onCancel={onCancel}
    >
      <Form<AddItemsValues>
        initValues={{ raw: '' }}
        onValueChange={value =>
          setRaw((value as Partial<AddItemsValues>)?.raw ?? '')
        }
      >
        <FormTextArea
          field="raw"
          label="JSON 数据"
          placeholder={
            '粘贴单个 JSON 对象，或以 JSON 数组粘贴多个用例，例如：\n{"input":"你好","reference_output":"你好"}\n或\n[{"input":"你好"},{"input":"再见"}]'
          }
          autosize={{ minRows: 8 }}
        />
      </Form>
    </Modal>
  );
};

const ImportItemsModal = ({
  visible,
  columns,
  importing,
  onOk,
  onCancel,
}: {
  visible: boolean;
  columns: ColumnDef[];
  importing: boolean;
  onOk: (items: { data_json: string }[]) => void;
  onCancel: () => void;
}) => {
  const fileRef = useRef<HTMLInputElement>(null);
  const [parsed, setParsed] = useState<ParsedImportFile | null>(null);
  const [mapping, setMapping] = useState<Record<string, string>>({});
  const [parsing, setParsing] = useState(false);
  const [fileName, setFileName] = useState('');
  const [error, setError] = useState('');

  const reset = () => {
    setParsed(null);
    setMapping({});
    setFileName('');
    setError('');
    if (fileRef.current) {
      fileRef.current.value = '';
    }
  };

  useEffect(() => {
    if (visible) {
      reset();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [visible]);

  const handleFile = async (file: File) => {
    setParsing(true);
    setError('');
    try {
      const data = await evaluationApi.parseEvalSetFile(file);
      setParsed(data);
      setFileName(file.name);
      const nextMapping: Record<string, string> = {};
      data.headers.forEach(header => {
        const match = columns.find(c => c.name === header);
        nextMapping[header] = match ? match.name : '';
      });
      setMapping(nextMapping);
    } catch (e) {
      setError(getErrorMessage(e));
    } finally {
      setParsing(false);
    }
  };

  const columnOptions = [
    { label: '不导入', value: '' },
    ...columns.map(c => ({ label: c.name, value: c.name })),
  ];

  const handleOk = () => {
    if (!parsed) {
      return;
    }
    const headerIdx: Record<string, number> = {};
    parsed.headers.forEach((header, index) => {
      const target = mapping[header];
      if (target) {
        headerIdx[target] = index;
      }
    });
    const mappedCols = Object.keys(headerIdx);
    if (!mappedCols.length) {
      Toast.error('请至少配置一列映射');
      return;
    }
    const items = parsed.rows.map(row => {
      const obj: Record<string, string> = {};
      mappedCols.forEach(col => {
        const index = headerIdx[col];
        const value = row[index];
        obj[col] = value !== undefined && value !== null ? String(value) : '';
      });
      return { data_json: JSON.stringify(obj) };
    });
    onOk(items);
  };

  const previewColumns: ColumnProps<Record<string, unknown>>[] = (parsed?.headers ?? []).map(
    header => ({ title: header, dataIndex: header, ellipsis: true }),
  );
  const previewRows = (parsed?.rows ?? [])
    .slice(0, 5)
    .map((row, index) => {
      const obj: Record<string, unknown> = { __key__: index };
      parsed!.headers.forEach((header, colIndex) => {
        obj[header] = row[colIndex];
      });
      return obj;
    });

  return (
    <Modal
      size="large"
      title="导入用例"
      visible={visible}
      okText="导入"
      cancelText="取消"
      okButtonProps={{
        disabled: !parsed || !Object.values(mapping).some(Boolean),
        loading: importing,
      }}
      onOk={handleOk}
      onCancel={onCancel}
    >
      <div className="flex items-center gap-[12px] mb-[12px]">
        <Button
          theme="solid"
          type="primary"
          icon={<IconCozUpload />}
          loading={parsing}
          onClick={() => fileRef.current?.click()}
        >
          选择文件
        </Button>
        <input
          ref={fileRef}
          type="file"
          accept=".csv,.xlsx,.xls"
          className="hidden"
          onChange={e => {
            const file = e.target.files?.[0];
            if (file) {
              void handleFile(file);
            }
          }}
        />
        <Text type="secondary">支持 CSV / XLSX / XLS，首行为表头</Text>
      </div>
      {fileName ? (
        <Text className="block mb-[12px]">已选择：{fileName}</Text>
      ) : null}
      {error ? (
        <Text type="danger" className="block mb-[12px]">
          {error}
        </Text>
      ) : null}
      {parsed ? (
        <>
          <div className="mb-[8px]">
            <Text strong>列映射</Text>
            <Text type="secondary" className="ml-[8px]">
              未配置映射的列不会被导入
            </Text>
          </div>
          <div className="mb-[16px]">
            {parsed.headers.map((header, index) => (
              <div
                key={`${index}-${header}`}
                className="flex items-center gap-[12px] mb-[8px]"
              >
                <Text className="w-[200px] truncate" title={header}>
                  {header || `列${index + 1}`}
                </Text>
                <Select
                  value={mapping[header] ?? ''}
                  optionList={columnOptions}
                  onChange={(value: string) =>
                    setMapping(m => ({ ...m, [header]: value }))
                  }
                  placeholder="选择目标列"
                  style={{ width: 220 }}
                />
              </div>
            ))}
          </div>
          <Text strong className="block mb-[8px]">
            预览（前 {Math.min(5, parsed.rows.length)} 行 / 共 {parsed.total} 行）
          </Text>
          <Table<Record<string, unknown>>
            tableProps={{
              columns: previewColumns,
              dataSource: previewRows,
              rowKey: '__key__',
              pagination: false,
              scroll: { x: 'max-content' },
            }}
            empty={null}
          />
        </>
      ) : null}
    </Modal>
  );
};

const EditItemModal = ({
  visible,
  item,
  columns,
  saving,
  onOk,
  onCancel,
}: {
  visible: boolean;
  item?: EvalSetItem;
  columns: ColumnDef[];
  saving: boolean;
  onOk: (dataJson: string) => void;
  onCancel: () => void;
}) => {
  const [values, setValues] = useState<Record<string, string>>({});
  const [raw, setRaw] = useState('');

  useEffect(() => {
    if (visible && item) {
      let obj: Record<string, unknown> = {};
      try {
        obj = JSON.parse(item.data_json) as Record<string, unknown>;
      } catch {
        obj = {};
      }
      const next: Record<string, string> = {};
      columns.forEach(c => {
        const val = obj[c.name];
        next[c.name] = val !== undefined && val !== null ? String(val) : '';
      });
      setValues(next);
      setRaw(item.data_json);
    }
  }, [visible, item, columns]);

  const handleOk = () => {
    if (!item) {
      return;
    }
    if (columns.length === 0) {
      try {
        JSON.parse(raw);
      } catch {
        Toast.error('数据不是合法 JSON');
        return;
      }
      onOk(raw);
      return;
    }
    let obj: Record<string, unknown> = {};
    try {
      obj = JSON.parse(item.data_json) as Record<string, unknown>;
    } catch {
      obj = {};
    }
    columns.forEach(c => {
      obj[c.name] = values[c.name] ?? '';
    });
    onOk(JSON.stringify(obj));
  };

  return (
    <Modal
      size="medium"
      title="编辑用例"
      visible={visible}
      okText="保存"
      cancelText="取消"
      okButtonProps={{ loading: saving }}
      onOk={handleOk}
      onCancel={onCancel}
    >
      {columns.length ? (
        columns.map(c => (
          <div key={c.name} className="mb-[12px]">
            <Text className="block mb-[4px]">{c.name}</Text>
            <Input
              value={values[c.name] ?? ''}
              onChange={(v: string) =>
                setValues(m => ({ ...m, [c.name]: v ?? '' }))
              }
              placeholder={c.description || `请输入 ${c.name}`}
            />
          </div>
        ))
      ) : (
        <div>
          <Text className="block mb-[4px]">数据(JSON)</Text>
          <TextArea
            value={raw}
            onChange={(v: string) => setRaw(v ?? '')}
            autosize={{ minRows: 6 }}
            placeholder='{"input": "问题", "reference_output": "期望输出"}'
          />
        </div>
      )}
    </Modal>
  );
};

const EvalSetDetailPage = () => {
  const { id } = useParams<{ id: string }>();
  const evalSetId = Number(id);
  const navigate = useNavigate();

  const [page, setPage] = useState(1);
  const [selectedRowKeys, setSelectedRowKeys] = useState<number[]>([]);
  const [addModalVisible, setAddModalVisible] = useState(false);
  const [importModalVisible, setImportModalVisible] = useState(false);
  const [editingItem, setEditingItem] = useState<EvalSetItem | undefined>(
    undefined,
  );
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [viewingItem, setViewingItem] = useState<EvalSetItem | undefined>(
    undefined,
  );
  const [viewModalVisible, setViewModalVisible] = useState(false);

  const {
    data: evalSetData,
    loading: evalSetLoading,
    refresh: refreshEvalSet,
  } = useFetch(() => evaluationApi.getEvalSet(evalSetId), [evalSetId]);

  const {
    data: itemsData,
    loading: itemsLoading,
    refresh: reloadItems,
  } = useFetch(() => evaluationApi.listEvalSetItems(evalSetId, page, PAGE_SIZE), [
    evalSetId,
    page,
  ]);

  const submitItems = async (payload: { data_json: string }[]) => {
    try {
      await evaluationApi.addEvalSetItems({
        eval_set_id: evalSetId,
        items: payload,
      });
      Toast.success(`已添加 ${payload.length} 条用例`);
      setAddModalVisible(false);
      setImportModalVisible(false);
      reloadItems();
      refreshEvalSet();
    } catch (e) {
      Toast.error(getErrorMessage(e));
    }
  };

  const { run: addItems, loading: adding } = useMutation(async (raw: string) => {
    let payload: { data_json: string }[];
    try {
      payload = parseItemsInput(raw);
    } catch {
      Toast.error('JSON 格式不正确，请检查后重试');
      return;
    }
    if (!payload.length) {
      Toast.error('未解析到有效数据');
      return;
    }
    await submitItems(payload);
  });

  const { run: importItems, loading: importing } = useMutation(
    async (payload: { data_json: string }[]) => {
      if (!payload.length) {
        Toast.error('文件无有效数据行');
        return;
      }
      await submitItems(payload);
    },
  );

  const { run: removeItems } = useMutation(
    async (itemIds: number[]) => {
      if (!itemIds.length) {
        return;
      }
      try {
        await evaluationApi.deleteEvalSetItems({
          eval_set_id: evalSetId,
          item_ids: itemIds,
        });
        Toast.success('删除成功');
        setSelectedRowKeys([]);
        reloadItems();
        refreshEvalSet();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const { run: updateItem, loading: updating } = useMutation(
    async (dataJson: string) => {
      if (!editingItem) {
        return;
      }
      try {
        await evaluationApi.updateEvalSetItem({
          id: editingItem.id,
          data_json: dataJson,
        });
        Toast.success('保存成功');
        setEditModalVisible(false);
        setEditingItem(undefined);
        reloadItems();
      } catch (e) {
        Toast.error(getErrorMessage(e));
      }
    },
  );

  const handleEdit = (record: EvalSetItem) => {
    setEditingItem(record);
    setEditModalVisible(true);
  };

  const handleDelete = (record: EvalSetItem) => {
    Modal.confirm({
      title: '删除用例',
      content: `确定删除用例 #${record.id} 吗？`,
      okText: '删除',
      cancelText: '取消',
      onOk: () => removeItems([record.id]),
    });
  };

  const evalSet: EvalSet | undefined = evalSetData;
  const statusConf = evalSet ? EVAL_SET_STATUS_MAP[evalSet.status] : undefined;
  const columns = evalSet ? parseSchemaToColumns(evalSet.schema_json) : [];

  const parseItemJson = (dataJson: string): Record<string, unknown> => {
    try {
      return JSON.parse(dataJson) as Record<string, unknown>;
    } catch {
      return {};
    }
  };

  const tableColumns: ColumnProps<EvalSetItem>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 90,
      render: value => `#${String(value).padStart(5, '0')}`,
    },
    ...columns.map(
      (col): ColumnProps<EvalSetItem> => ({
        title: col.name,
        dataIndex: col.name,
        ellipsis: { showTooltip: true },
        render: (_value, record) => {
          const v = parseItemJson(record.data_json)[col.name];
          return v !== undefined && v !== null ? String(v) : '';
        },
      }),
    ),
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      width: 170,
      render: value => formatDateTime(value as string),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 170,
      render: value => formatDateTime(value as string),
    },
    {
      title: '操作',
      dataIndex: 'action',
      width: 160,
      render: (_value, record) => (
        <span className="flex items-center gap-[8px]">
          <Button
            theme="borderless"
            type="tertiary"
            onClick={() => {
              setViewingItem(record);
              setViewModalVisible(true);
            }}
          >
            查看
          </Button>
          <Button
            theme="borderless"
            type="tertiary"
            icon={<IconCozEdit />}
            onClick={() => handleEdit(record)}
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
      <div className="flex items-center gap-[8px] mb-[16px]">
        <Button
          theme="borderless"
          type="tertiary"
          icon={<IconCozArrowLeft />}
          onClick={() => navigate('/evaluation/sets')}
        >
          返回
        </Button>
      </div>
      <div className="flex items-center justify-between mb-[16px]">
        <div className="flex items-center gap-[12px]">
          <Title heading={5} style={{ margin: 0 }}>
            {evalSet?.name ?? '评测集详情'}
          </Title>
          {statusConf ? (
            <Tag color={statusConf.color}>{statusConf.label}</Tag>
          ) : null}
          {!evalSetLoading && evalSet ? (
            <Text type="secondary" fontSize="12px">
              共 {evalSet.item_count} 条用例
            </Text>
          ) : null}
        </div>
        <div className="flex items-center gap-[8px]">
          <Button
            theme="solid"
            type="primary"
            icon={<IconCozUpload />}
            onClick={() => setImportModalVisible(true)}
          >
            导入用例
          </Button>
          <Button
            theme="solid"
            type="primary"
            icon={<IconCozPlus />}
            onClick={() => setAddModalVisible(true)}
          >
            添加用例
          </Button>
          <Button
            theme="solid"
            type="danger"
            icon={<IconCozTrashCan />}
            disabled={!selectedRowKeys.length}
            onClick={() => {
              Modal.confirm({
                title: '批量删除用例',
                content: `确定删除选中的 ${selectedRowKeys.length} 条用例吗？`,
                okText: '删除',
                cancelText: '取消',
                onOk: () => removeItems(selectedRowKeys),
              });
            }}
          >
            批量删除
          </Button>
        </div>
      </div>
      {evalSet?.description ? (
        <Text type="secondary" className="mb-[16px]">
          {evalSet.description}
        </Text>
      ) : null}
      <Table<EvalSetItem>
        tableProps={{
          columns: tableColumns,
          dataSource: itemsData?.list ?? [],
          rowKey: 'id',
          loading: itemsLoading,
          rowSelection: {
            fixed: true,
            selectedRowKeys,
            onChange: keys =>
              setSelectedRowKeys((keys as number[] | undefined) ?? []),
          },
          pagination: {
            total: itemsData?.total ?? 0,
            currentPage: page,
            pageSize: PAGE_SIZE,
            onChange: setPage,
          },
        }}
        empty={<div className="py-[40px] text-center coz-fg-secondary">暂无用例</div>}
      />
      <AddItemsModal
        visible={addModalVisible}
        adding={adding}
        onOk={raw => addItems(raw)}
        onCancel={() => setAddModalVisible(false)}
      />
      <ImportItemsModal
        visible={importModalVisible}
        columns={columns}
        importing={importing}
        onOk={items => importItems(items)}
        onCancel={() => setImportModalVisible(false)}
      />
      <EditItemModal
        visible={editModalVisible}
        item={editingItem}
        columns={columns}
        saving={updating}
        onOk={dataJson => updateItem(dataJson)}
        onCancel={() => {
          setEditModalVisible(false);
          setEditingItem(undefined);
        }}
      />
      <Modal
        size="medium"
        title="用例详情"
        visible={viewModalVisible}
        okText="关闭"
        cancelText={null}
        onCancel={() => {
          setViewModalVisible(false);
          setViewingItem(undefined);
        }}
      >
        {viewingItem ? (
          <div>
            {columns.map(col => (
              <div key={col.name} className="mb-[12px]">
                <Text strong className="block mb-[4px]">
                  {col.name}
                </Text>
                <div className="coz-bg-fill-1 rounded-[6px] px-[12px] py-[8px]">
                  {String(
                    parseItemJson(viewingItem.data_json)[col.name] ?? '',
                  ) || '-'}
                </div>
              </div>
            ))}
            <div className="flex gap-[24px]">
              <Text type="secondary" fontSize="12px">
                创建时间：{formatDateTime(viewingItem.created_at as string)}
              </Text>
              <Text type="secondary" fontSize="12px">
                更新时间：{formatDateTime(viewingItem.updated_at as string)}
              </Text>
            </div>
          </div>
        ) : null}
      </Modal>
    </div>
  );
};

export default EvalSetDetailPage;
