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

package rdb

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/infra/idgen"
	"github.com/coze-dev/coze-studio/backend/infra/rdb"
	"github.com/coze-dev/coze-studio/backend/infra/rdb/entity"
	"github.com/coze-dev/coze-studio/backend/infra/sqlparser"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

type mysqlService struct {
	db        *gorm.DB
	generator idgen.IDGenerator
}

func NewService(db *gorm.DB, generator idgen.IDGenerator) rdb.RDB {
	return &mysqlService{db: db, generator: generator}
}

// CreateTable create table
func (m *mysqlService) CreateTable(ctx context.Context, req *rdb.CreateTableRequest) (*rdb.CreateTableResponse, error) {
	if req == nil || req.Table == nil {
		return nil, fmt.Errorf("invalid request")
	}

	// build column definitions
	columnDefs := make([]string, 0, len(req.Table.Columns))
	for _, col := range req.Table.Columns {
		if !isValidIdentifier(col.Name) {
			return nil, fmt.Errorf("invalid column name format: %s", col.Name)
		}
		colDef := fmt.Sprintf("%s %s", m.quoteIdent(col.Name), m.dataType(col))

		if col.Length != nil && !(m.isPostgres() && col.AutoIncrement) {
			colDef += fmt.Sprintf("(%d)", *col.Length)
		} else if col.Length == nil && col.DataType == entity.TypeVarchar {
			colDef += fmt.Sprintf("(%d)", 255)
		}

		if col.NotNull && !(m.isPostgres() && col.AutoIncrement) {
			colDef += " NOT NULL"
		}
		if col.DefaultValue != nil && !(m.isPostgres() && col.AutoIncrement) {
			if col.DataType == entity.TypeTimestamp {
				colDef += fmt.Sprintf(" DEFAULT %s", *col.DefaultValue)
			} else if col.DataType == entity.TypeText {
				// do nothing
			} else {
				colDef += fmt.Sprintf(" DEFAULT '%s'", escapeString(*col.DefaultValue))
			}
		}
		if col.AutoIncrement && !m.isPostgres() {
			colDef += " AUTO_INCREMENT"
		}
		if col.Comment != nil && !m.isPostgres() {
			colDef += fmt.Sprintf(" COMMENT '%s'", escapeString(*col.Comment))
		}

		columnDefs = append(columnDefs, colDef)
	}

	// build index definitions
	for _, idx := range req.Table.Indexes {
		if idx.Name != "" && !isValidIdentifier(idx.Name) {
			return nil, fmt.Errorf("invalid index name format: %s", idx.Name)
		}
		for _, col := range idx.Columns {
			if !isValidIdentifier(col) {
				return nil, fmt.Errorf("invalid column name in index: %s", col)
			}
		}

		var idxDef string
		switch idx.Type {
		case entity.PrimaryKey:
			idxDef = fmt.Sprintf("PRIMARY KEY (%s)", m.quoteIdents(idx.Columns))
		case entity.UniqueKey:
			if m.isPostgres() {
				idxDef = fmt.Sprintf("CONSTRAINT %s UNIQUE (%s)", m.quoteIdent(idx.Name), m.quoteIdents(idx.Columns))
			} else {
				idxDef = fmt.Sprintf("UNIQUE KEY `%s` (`%s`)", idx.Name, strings.Join(idx.Columns, "`,`"))
			}
		default:
			if m.isPostgres() {
				continue
			}
			idxDef = fmt.Sprintf("KEY `%s` (`%s`)", idx.Name, strings.Join(idx.Columns, "`,`"))
		}
		columnDefs = append(columnDefs, idxDef)
	}

	tableOptions := make([]string, 0)
	if req.Table.Options != nil {
		if req.Table.Options.Collate != nil && !m.isPostgres() {
			tableOptions = append(tableOptions, fmt.Sprintf("COLLATE=%s", *req.Table.Options.Collate))
		}
		if req.Table.Options.AutoIncrement != nil && !m.isPostgres() {
			tableOptions = append(tableOptions, fmt.Sprintf("AUTO_INCREMENT=%d", *req.Table.Options.AutoIncrement))
		}
		if req.Table.Options.Comment != nil && !m.isPostgres() {
			tableOptions = append(tableOptions, fmt.Sprintf("COMMENT='%s'", escapeString(*req.Table.Options.Comment)))
		}
	}

	tableName := req.Table.Name
	if req.Table.Name == "" {
		genName, err := m.genTableName(ctx)
		if err != nil {
			return nil, err
		}

		tableName = genName
	}

	if !isValidIdentifier(tableName) {
		return nil, fmt.Errorf("invalid table name format: %s", tableName)
	}

	createSQL := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n) %s",
		m.quoteIdent(tableName),
		strings.Join(columnDefs, ",\n  "),
		strings.Join(tableOptions, " "),
	)

	logs.CtxInfof(ctx, "[CreateTable] execute sql is %s, req is %v", createSQL, req)

	err := m.db.WithContext(ctx).Exec(createSQL).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %v", err)
	}

	if m.isPostgres() {
		for _, idx := range req.Table.Indexes {
			if idx.Type != entity.NormalKey {
				continue
			}
			idxSQL := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)",
				m.quoteIdent(idx.Name), m.quoteIdent(tableName), m.quoteIdents(idx.Columns))
			if err := m.db.WithContext(ctx).Exec(idxSQL).Error; err != nil {
				return nil, fmt.Errorf("failed to create index: %v", err)
			}
		}
	}

	resTable := req.Table
	resTable.Name = tableName
	return &rdb.CreateTableResponse{Table: resTable}, nil
}

func isValidIdentifier(name string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z_][a-zA-Z0-9_]*$`, name)
	return match
}

func escapeString(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}

func (m *mysqlService) isPostgres() bool {
	return m.db != nil && m.db.Dialector != nil && m.db.Dialector.Name() == "postgres"
}

func (m *mysqlService) quoteIdent(name string) string {
	if m.isPostgres() {
		return fmt.Sprintf("\"%s\"", name)
	}
	return fmt.Sprintf("`%s`", name)
}

func (m *mysqlService) quoteIdents(names []string) string {
	quoted := make([]string, 0, len(names))
	for _, name := range names {
		quoted = append(quoted, m.quoteIdent(name))
	}
	return strings.Join(quoted, ",")
}

func (m *mysqlService) dataType(col *entity.Column) string {
	if !m.isPostgres() {
		return string(col.DataType)
	}
	if col.AutoIncrement {
		if col.DataType == entity.TypeBigInt {
			return "BIGSERIAL"
		}
		return "SERIAL"
	}
	switch col.DataType {
	case entity.TypeJson:
		return "JSONB"
	case entity.TypeDouble:
		return "DOUBLE PRECISION"
	default:
		return string(col.DataType)
	}
}

func postgresDataType(dataType string) entity.DataType {
	switch strings.ToLower(dataType) {
	case "integer":
		return entity.TypeInt
	case "bigint":
		return entity.TypeBigInt
	case "character varying":
		return entity.TypeVarchar
	case "text":
		return entity.TypeText
	case "boolean":
		return entity.TypeBoolean
	case "json", "jsonb":
		return entity.TypeJson
	case "timestamp without time zone", "timestamp with time zone":
		return entity.TypeTimestamp
	case "real":
		return entity.TypeFloat
	case "double precision":
		return entity.TypeDouble
	default:
		return entity.DataType(strings.ToUpper(dataType))
	}
}

// AlterTable alter table
func (m *mysqlService) AlterTable(ctx context.Context, req *rdb.AlterTableRequest) (*rdb.AlterTableResponse, error) {
	if req == nil || len(req.Operations) == 0 {
		return nil, fmt.Errorf("invalid request")
	}

	alterSQL := fmt.Sprintf("ALTER TABLE %s", m.quoteIdent(req.TableName))
	operations := make([]string, 0, len(req.Operations))
	postAlterSQLs := make([]string, 0)

	for _, op := range req.Operations {
		switch op.Action {
		case entity.AddColumn:
			if op.Column == nil {
				return nil, fmt.Errorf("column is required for ADD COLUMN operation")
			}
			colDef := fmt.Sprintf("ADD COLUMN %s %s", m.quoteIdent(op.Column.Name), m.dataType(op.Column))
			if op.Column.Length != nil {
				colDef += fmt.Sprintf("(%d)", *op.Column.Length)
			} else if op.Column.Length == nil && op.Column.DataType == entity.TypeVarchar {
				colDef += fmt.Sprintf("(%d)", 255)
			}

			if op.Column.NotNull {
				colDef += " NOT NULL"
			}

			if op.Column.DefaultValue != nil {
				if op.Column.DataType == entity.TypeTimestamp {
					colDef += fmt.Sprintf(" DEFAULT %s", *op.Column.DefaultValue)
				} else {
					colDef += fmt.Sprintf(" DEFAULT '%s'", *op.Column.DefaultValue)
				}
			}

			operations = append(operations, colDef)

		case entity.DropColumn:
			if op.Column == nil {
				return nil, fmt.Errorf("column is required for DROP COLUMN operation")
			}
			operations = append(operations, fmt.Sprintf("DROP COLUMN %s", m.quoteIdent(op.Column.Name)))

		case entity.ModifyColumn:
			if op.Column == nil {
				return nil, fmt.Errorf("column is required for MODIFY COLUMN operation")
			}
			var colDef string
			if m.isPostgres() {
				colDef = fmt.Sprintf("ALTER COLUMN %s TYPE %s", m.quoteIdent(op.Column.Name), m.dataType(op.Column))
			} else {
				colDef = fmt.Sprintf("MODIFY COLUMN `%s` %s", op.Column.Name, op.Column.DataType)
			}
			if op.Column.Length != nil {
				colDef += fmt.Sprintf("(%d)", *op.Column.Length)
			} else if op.Column.Length == nil && op.Column.DataType == entity.TypeVarchar {
				colDef += fmt.Sprintf("(%d)", 255)
			}
			operations = append(operations, colDef)

		case entity.RenameColumn:
			if op.Column == nil || op.OldName == nil {
				return nil, fmt.Errorf("column and old name are required for RENAME COLUMN operation")
			}
			operations = append(operations, fmt.Sprintf("RENAME COLUMN %s TO %s", m.quoteIdent(*op.OldName), m.quoteIdent(op.Column.Name)))

		case entity.AddIndex:
			if op.Index == nil {
				return nil, fmt.Errorf("index is required for ADD INDEX operation")
			}
			var idxDef string
			switch op.Index.Type {
			case entity.PrimaryKey:
				idxDef = fmt.Sprintf("ADD PRIMARY KEY (%s)", m.quoteIdents(op.Index.Columns))
			case entity.UniqueKey:
				if m.isPostgres() {
					idxDef = fmt.Sprintf("ADD CONSTRAINT %s UNIQUE (%s)", m.quoteIdent(op.Index.Name), m.quoteIdents(op.Index.Columns))
				} else {
					idxDef = fmt.Sprintf("ADD UNIQUE INDEX `%s` (`%s`)", op.Index.Name, strings.Join(op.Index.Columns, "`,`"))
				}
			default:
				if m.isPostgres() {
					postAlterSQLs = append(postAlterSQLs, fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (%s)", m.quoteIdent(op.Index.Name), m.quoteIdent(req.TableName), m.quoteIdents(op.Index.Columns)))
					continue
				} else {
					idxDef = fmt.Sprintf("ADD INDEX `%s` (`%s`)", op.Index.Name, strings.Join(op.Index.Columns, "`,`"))
				}
			}
			operations = append(operations, idxDef)
		}
	}

	if len(operations) > 0 {
		alterSQL += " " + strings.Join(operations, ", ")

		logs.CtxInfof(ctx, "[AlterTable] execute sql is %s, req is %v", alterSQL, req)

		err := m.db.WithContext(ctx).Exec(alterSQL).Error
		if err != nil {
			return nil, fmt.Errorf("failed to alter table: %v", err)
		}
	}

	for _, sql := range postAlterSQLs {
		logs.CtxInfof(ctx, "[AlterTable] execute sql is %s, req is %v", sql, req)
		if err := m.db.WithContext(ctx).Exec(sql).Error; err != nil {
			return nil, fmt.Errorf("failed to alter table: %v", err)
		}
	}

	table, err := m.getTableInfo(ctx, req.TableName)
	if err != nil {
		return nil, fmt.Errorf("failed to get table info: %v", err)
	}

	return &rdb.AlterTableResponse{Table: table}, nil
}

// DropTable drop table
func (m *mysqlService) DropTable(ctx context.Context, req *rdb.DropTableRequest) (*rdb.DropTableResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	dropSQL := "DROP TABLE"
	if req.IfExists {
		dropSQL += " IF EXISTS"
	}
	dropSQL += fmt.Sprintf(" %s", m.quoteIdent(req.TableName))

	logs.CtxInfof(ctx, "[DropTable] execute sql is %s, req is %v", dropSQL, req)

	err := m.db.WithContext(ctx).Exec(dropSQL).Error
	if err != nil {
		return nil, fmt.Errorf("failed to drop table: %v", err)
	}

	return &rdb.DropTableResponse{Success: true}, nil
}

// GetTable get table schema info
func (m *mysqlService) GetTable(ctx context.Context, req *rdb.GetTableRequest) (*rdb.GetTableResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	table, err := m.getTableInfo(ctx, req.TableName)
	if err != nil {
		return nil, err
	}

	return &rdb.GetTableResponse{Table: table}, nil
}

func (m *mysqlService) InsertData(ctx context.Context, req *rdb.InsertDataRequest) (*rdb.InsertDataResponse, error) {
	if req == nil || len(req.Data) == 0 {
		return nil, fmt.Errorf("invalid request")
	}

	fields := make([]string, 0)
	for field := range req.Data[0] {
		fields = append(fields, field)
	}

	const batchSize = 1000
	var totalAffected int64

	for i := 0; i < len(req.Data); i += batchSize {
		end := i + batchSize
		if end > len(req.Data) {
			end = len(req.Data)
		}

		currentBatch := req.Data[i:end]

		placeholderGroups := make([]string, 0, len(currentBatch))
		values := make([]interface{}, 0, len(currentBatch)*len(fields))

		for _, row := range currentBatch {
			placeholders := make([]string, len(fields))
			for j := range placeholders {
				placeholders[j] = "?"
			}
			placeholderGroups = append(placeholderGroups, "("+strings.Join(placeholders, ",")+")")

			for _, field := range fields {
				values = append(values, row[field])
			}
		}

		insertSQL := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
			m.quoteIdent(req.TableName),
			m.quoteIdents(fields),
			strings.Join(placeholderGroups, ","),
		)

		logs.CtxInfof(ctx, "[InsertData] execute sql is %s, value is %v in batch %d", insertSQL, values, i)

		result := m.db.WithContext(ctx).Exec(insertSQL, values...)
		if result.Error != nil {
			return nil, result.Error
		}

		affected := result.RowsAffected
		totalAffected += affected
	}

	return &rdb.InsertDataResponse{AffectedRows: totalAffected}, nil
}

// UpdateData Update data
func (m *mysqlService) UpdateData(ctx context.Context, req *rdb.UpdateDataRequest) (*rdb.UpdateDataResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	setClauses := make([]string, 0)
	values := make([]interface{}, 0)
	for field, value := range req.Data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", m.quoteIdent(field)))
		values = append(values, value)
	}

	whereClause, whereValues, err := m.buildWhereClause(req.Where)
	if err != nil {
		return nil, fmt.Errorf("failed to build where clause: %v", err)
	}
	values = append(values, whereValues...)

	limitClause := ""
	if req.Limit != nil {
		limitClause = fmt.Sprintf(" LIMIT %d", *req.Limit)
	}

	updateSQL := fmt.Sprintf("UPDATE %s SET %s%s%s",
		m.quoteIdent(req.TableName),
		strings.Join(setClauses, ", "),
		whereClause,
		limitClause,
	)

	logs.CtxInfof(ctx, "[UpdateData] execute sql is %s, value is %v, req is %v", updateSQL, values, req)

	result := m.db.WithContext(ctx).Exec(updateSQL, values...)
	if result.Error != nil {
		return nil, result.Error
	}

	affectedRows := result.RowsAffected

	return &rdb.UpdateDataResponse{AffectedRows: affectedRows}, nil
}

// DeleteData delete data
func (m *mysqlService) DeleteData(ctx context.Context, req *rdb.DeleteDataRequest) (*rdb.DeleteDataResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	whereClause, whereValues, err := m.buildWhereClause(req.Where)
	if err != nil {
		return nil, fmt.Errorf("failed to build where clause: %v", err)
	}

	limitClause := ""
	if req.Limit != nil {
		limitClause = fmt.Sprintf(" LIMIT %d", *req.Limit)
	}

	deleteSQL := fmt.Sprintf("DELETE FROM %s%s%s",
		m.quoteIdent(req.TableName),
		whereClause,
		limitClause,
	)

	logs.CtxInfof(ctx, "[DeleteData] execute sql is %s, value is %v, req is %v", deleteSQL, whereValues, req)

	result := m.db.WithContext(ctx).Exec(deleteSQL, whereValues...)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to delete data: %v", result.Error)
	}

	affectedRows := result.RowsAffected

	return &rdb.DeleteDataResponse{AffectedRows: affectedRows}, nil
}

// SelectData select data
func (m *mysqlService) SelectData(ctx context.Context, req *rdb.SelectDataRequest) (*rdb.SelectDataResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	fields := "*"
	if len(req.Fields) > 0 {
		fields = m.quoteIdents(req.Fields)
	}

	whereClause := ""
	whereValues := make([]interface{}, 0)
	if req.Where != nil {
		clause, values, err := m.buildWhereClause(req.Where)
		if err != nil {
			return nil, fmt.Errorf("failed to build where clause: %v", err)
		}
		whereClause = clause
		whereValues = values
	}

	orderByClause := ""
	if len(req.OrderBy) > 0 {
		orders := make([]string, len(req.OrderBy))
		for i, order := range req.OrderBy {
			orders[i] = fmt.Sprintf("%s %s", m.quoteIdent(order.Field), order.Direction)
		}
		orderByClause = " ORDER BY " + strings.Join(orders, ", ")
	}

	limitClause := ""
	if req.Limit != nil {
		limitClause = fmt.Sprintf(" LIMIT %d", *req.Limit)
		if req.Offset != nil {
			limitClause += fmt.Sprintf(" OFFSET %d", *req.Offset)
		}
	}

	selectSQL := fmt.Sprintf("SELECT %s FROM %s%s%s%s",
		fields,
		m.quoteIdent(req.TableName),
		whereClause,
		orderByClause,
		limitClause,
	)

	logs.CtxInfof(ctx, "[SelectData] execute sql is %s, value is %v, req is %v", selectSQL, whereValues, req)

	rows, err := m.db.WithContext(ctx).Raw(selectSQL, whereValues...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute select: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	resultSet := &entity.ResultSet{
		Columns: columns,
		Rows:    make([]map[string]interface{}, 0),
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			rowData[col] = values[i]
		}
		resultSet.Rows = append(resultSet.Rows, rowData)
	}

	// get total count
	var total int64
	if whereClause != "" {
		countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s%s", m.quoteIdent(req.TableName), whereClause)
		err = m.db.WithContext(ctx).Raw(countSQL, whereValues...).Scan(&total).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get total count: %v", err)
		}
	} else {
		total = int64(len(resultSet.Rows))
	}

	return &rdb.SelectDataResponse{
		ResultSet: resultSet,
		Total:     total,
	}, nil
}

// UpsertData upsert data
func (m *mysqlService) UpsertData(ctx context.Context, req *rdb.UpsertDataRequest) (*rdb.UpsertDataResponse, error) {
	if req == nil || len(req.Data) == 0 {
		return nil, fmt.Errorf("invalid request: empty data")
	}

	keys := req.Keys
	if len(keys) == 0 {
		primaryKeys, err := m.getTablePrimaryKeys(ctx, req.TableName)
		if err != nil {
			return nil, fmt.Errorf("failed to get primary keys: %v", err)
		}

		if len(primaryKeys) == 0 {
			return nil, fmt.Errorf("table %s has no primary key, keys are required for upsert operation", req.TableName)
		}

		keys = primaryKeys
	}

	fields := make([]string, 0)
	for field := range req.Data[0] {
		fields = append(fields, field)
	}

	const batchSize = 1000
	var totalAffected, totalInserted, totalUpdated int64

	for i := 0; i < len(req.Data); i += batchSize {
		end := i + batchSize
		if end > len(req.Data) {
			end = len(req.Data)
		}

		currentBatch := req.Data[i:end]

		placeholderGroups := make([]string, 0, len(currentBatch))
		values := make([]interface{}, 0, len(currentBatch)*len(fields))

		for _, row := range currentBatch {
			placeholders := make([]string, len(fields))
			for j := range placeholders {
				placeholders[j] = "?"
			}
			placeholderGroups = append(placeholderGroups, "("+strings.Join(placeholders, ",")+")")

			for _, field := range fields {
				values = append(values, row[field])
			}
		}

		updateClauses := make([]string, 0, len(fields))
		for _, field := range fields {
			isKey := false
			for _, key := range keys {
				if field == key {
					isKey = true
					break
				}
			}
			if !isKey {
				if m.isPostgres() {
					updateClauses = append(updateClauses, fmt.Sprintf("%s=EXCLUDED.%s", m.quoteIdent(field), m.quoteIdent(field)))
				} else {
					updateClauses = append(updateClauses, fmt.Sprintf("`%s`=VALUES(`%s`)", field, field))
				}
			}
		}

		var upsertSQL string
		if m.isPostgres() {
			conflictAction := "DO NOTHING"
			if len(updateClauses) > 0 {
				conflictAction = "DO UPDATE SET " + strings.Join(updateClauses, ",")
			}
			upsertSQL = fmt.Sprintf(
				"INSERT INTO %s (%s) VALUES %s ON CONFLICT (%s) %s",
				m.quoteIdent(req.TableName),
				m.quoteIdents(fields),
				strings.Join(placeholderGroups, ","),
				m.quoteIdents(keys),
				conflictAction,
			)
		} else {
			upsertSQL = fmt.Sprintf(
				"INSERT INTO `%s` (`%s`) VALUES %s ON DUPLICATE KEY UPDATE %s",
				req.TableName,
				strings.Join(fields, "`,`"),
				strings.Join(placeholderGroups, ","),
				strings.Join(updateClauses, ","),
			)
		}

		logs.CtxInfof(ctx, "[UpsertData] execute sql is %s, value is %v, batch is %d", upsertSQL, values, i)

		result := m.db.WithContext(ctx).Exec(upsertSQL, values...)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to upsert data: %v", result.Error)
		}

		total, inserted, updated := calculateInsertedUpdated(result.RowsAffected, len(currentBatch))
		totalInserted += inserted
		totalUpdated += updated
		totalAffected += total
	}

	return &rdb.UpsertDataResponse{
		AffectedRows:  totalAffected,
		InsertedRows:  totalInserted,
		UpdatedRows:   totalUpdated,
		UnchangedRows: int64(len(req.Data)) - totalAffected,
	}, nil
}

func (m *mysqlService) getTablePrimaryKeys(ctx context.Context, tableName string) ([]string, error) {
	if m.isPostgres() {
		query := `
        SELECT a.attname
        FROM pg_index i
        JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
        WHERE i.indrelid = ?::regclass
          AND i.indisprimary
        ORDER BY array_position(i.indkey, a.attnum)
    `

		var primaryKeys []string
		rows, err := m.db.WithContext(ctx).Raw(query, tableName).Rows()
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var columnName string
			if err := rows.Scan(&columnName); err != nil {
				return nil, err
			}
			primaryKeys = append(primaryKeys, columnName)
		}

		return primaryKeys, nil
	}

	query := `
        SELECT COLUMN_NAME
        FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
        WHERE TABLE_SCHEMA = DATABASE()
          AND TABLE_NAME = ?
          AND CONSTRAINT_NAME = 'PRIMARY'
        ORDER BY ORDINAL_POSITION
    `

	var primaryKeys []string
	rows, err := m.db.WithContext(ctx).Raw(query, tableName).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		primaryKeys = append(primaryKeys, columnName)
	}

	return primaryKeys, nil
}

// calculateInsertedUpdated function remains unchanged
func calculateInsertedUpdated(affectedRows int64, batchSize int) (int64, int64, int64) {
	updated := int64(0)
	inserted := affectedRows
	if affectedRows > int64(batchSize) {
		updated = affectedRows - int64(batchSize)
		inserted = int64(batchSize) - updated
	}

	return inserted + updated, inserted, updated
}

// ExecuteSQL Execute SQL
func (m *mysqlService) ExecuteSQL(ctx context.Context, req *rdb.ExecuteSQLRequest) (*rdb.ExecuteSQLResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("invalid request")
	}

	logs.CtxInfof(ctx, "[ExecuteSQL] req is %v", req)

	var processedSQL string
	var processedParams []interface{}
	var err error

	// Handle SQLType: if raw, do not process params
	if req.SQLType == entity.SQLType_Raw {
		processedSQL = req.SQL
		processedParams = nil
	} else {
		processedSQL, processedParams, err = m.processSliceParams(req.SQL, req.Params)
		if err != nil {
			return nil, fmt.Errorf("failed to process parameters: %v", err)
		}
	}

	operation, err := sqlparser.New().GetSQLOperation(processedSQL)
	if err != nil {
		return nil, err
	}

	if operation != sqlparser.OperationTypeSelect {
		result := m.db.WithContext(ctx).Exec(processedSQL, processedParams...)
		if result.Error != nil {
			return nil, fmt.Errorf("failed to execute SQL: %v", result.Error)
		}

		resultSet := &entity.ResultSet{
			Columns:      []string{},
			Rows:         []map[string]interface{}{},
			AffectedRows: result.RowsAffected,
		}

		return &rdb.ExecuteSQLResponse{
			ResultSet: resultSet,
		}, nil
	}

	rows, err := m.db.WithContext(ctx).Raw(processedSQL, processedParams...).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to execute SQL: %v", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %v", err)
	}

	resultSet := &entity.ResultSet{
		Columns: columns,
		Rows:    make([]map[string]interface{}, 0),
	}

	for rows.Next() {
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))
		for i := range values {
			valuePtrs[i] = &values[i]
		}

		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		rowData := make(map[string]interface{})
		for i, col := range columns {
			rowData[col] = values[i]
		}
		resultSet.Rows = append(resultSet.Rows, rowData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error while reading rows: %v", err)
	}

	return &rdb.ExecuteSQLResponse{
		ResultSet: resultSet,
	}, nil
}

func (m *mysqlService) processSliceParams(sql string, params []interface{}) (string, []interface{}, error) {
	if len(params) == 0 {
		return sql, params, nil
	}

	processedParams := make([]interface{}, 0)
	paramIndex := 0
	resultSQL := ""
	lastPos := 0

	// get all ? positions
	for i := 0; i < len(sql); i++ {
		if sql[i] == '?' && paramIndex < len(params) {
			resultSQL += sql[lastPos:i]
			lastPos = i + 1

			param := params[paramIndex]
			paramIndex++

			if m.isSlice(param) {
				sliceValues, err := m.getSliceValues(param)
				if err != nil {
					return "", nil, err
				}

				if len(sliceValues) == 0 {
					resultSQL += "(NULL)"
				} else {
					// (?, ?, ...)
					placeholders := make([]string, len(sliceValues))
					for j := range placeholders {
						placeholders[j] = "?"
					}
					resultSQL += "(" + strings.Join(placeholders, ", ") + ")"

					processedParams = append(processedParams, sliceValues...)
				}
			} else {
				resultSQL += "?"
				processedParams = append(processedParams, param)
			}
		}
	}

	resultSQL += sql[lastPos:]

	return resultSQL, processedParams, nil
}

func (m *mysqlService) isSlice(param interface{}) bool {
	if param == nil {
		return false
	}

	rv := reflect.ValueOf(param)
	return rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() != reflect.Uint8 // exclude []byte
}

func (m *mysqlService) getSliceValues(param interface{}) ([]interface{}, error) {
	rv := reflect.ValueOf(param)
	if rv.Kind() != reflect.Slice {
		return nil, fmt.Errorf("parameter is not a slice")
	}

	length := rv.Len()
	values := make([]interface{}, length)

	for i := 0; i < length; i++ {
		values[i] = rv.Index(i).Interface()
	}

	return values, nil
}

func (m *mysqlService) genTableName(ctx context.Context) (string, error) {
	id, err := m.generator.GenID(ctx)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("table_%d", id), nil
}

func (m *mysqlService) getTableInfo(ctx context.Context, tableName string) (*entity.Table, error) {
	if m.isPostgres() {
		columnsSQL := `
        SELECT
            column_name,
            data_type,
            character_maximum_length,
            is_nullable,
            column_default
        FROM information_schema.columns
        WHERE table_schema = current_schema()
          AND table_name = ?
        ORDER BY ordinal_position
    `

		type columnInfo struct {
			ColumnName   string  `gorm:"column:column_name"`
			DataType     string  `gorm:"column:data_type"`
			CharLength   *int    `gorm:"column:character_maximum_length"`
			IsNullable   string  `gorm:"column:is_nullable"`
			DefaultValue *string `gorm:"column:column_default"`
		}

		var columnsData []columnInfo
		if err := m.db.WithContext(ctx).Raw(columnsSQL, tableName).Scan(&columnsData).Error; err != nil {
			return nil, err
		}
		if len(columnsData) == 0 {
			return nil, fmt.Errorf("table not found: %s", tableName)
		}

		columns := make([]*entity.Column, len(columnsData))
		for i, colData := range columnsData {
			columns[i] = &entity.Column{
				Name:          colData.ColumnName,
				DataType:      postgresDataType(colData.DataType),
				Length:        colData.CharLength,
				NotNull:       colData.IsNullable == "NO",
				DefaultValue:  colData.DefaultValue,
				AutoIncrement: colData.DefaultValue != nil && strings.Contains(*colData.DefaultValue, "nextval"),
			}
		}

		indexesSQL := `
        SELECT
            i.relname AS index_name,
            ix.indisunique AS is_unique,
            ix.indisprimary AS is_primary,
            array_to_string(array_agg(a.attname ORDER BY array_position(ix.indkey, a.attnum)), ',') AS columns
        FROM pg_class t
        JOIN pg_index ix ON t.oid = ix.indrelid
        JOIN pg_class i ON i.oid = ix.indexrelid
        JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
        JOIN pg_namespace n ON n.oid = t.relnamespace
        WHERE n.nspname = current_schema()
          AND t.relname = ?
        GROUP BY i.relname, ix.indisunique, ix.indisprimary
    `

		type indexInfo struct {
			IndexName string `gorm:"column:index_name"`
			IsUnique  bool   `gorm:"column:is_unique"`
			IsPrimary bool   `gorm:"column:is_primary"`
			Columns   string `gorm:"column:columns"`
		}

		var indexesData []indexInfo
		if err := m.db.WithContext(ctx).Raw(indexesSQL, tableName).Scan(&indexesData).Error; err != nil {
			return nil, err
		}

		indexes := make([]*entity.Index, 0, len(indexesData))
		for _, idxData := range indexesData {
			idxType := entity.NormalKey
			if idxData.IsPrimary {
				idxType = entity.PrimaryKey
			} else if idxData.IsUnique {
				idxType = entity.UniqueKey
			}
			indexes = append(indexes, &entity.Index{
				Name:    idxData.IndexName,
				Type:    idxType,
				Columns: strings.Split(idxData.Columns, ","),
			})
		}

		return &entity.Table{
			Name:    tableName,
			Columns: columns,
			Indexes: indexes,
		}, nil
	}

	tableInfoSQL := `
        SELECT 
            TABLE_NAME,
            TABLE_COLLATION,
            AUTO_INCREMENT,
            TABLE_COMMENT
        FROM information_schema.TABLES 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = ?
    `

	var (
		name          string
		collation     *string
		autoIncrement *int64
		comment       *string
	)

	err := m.db.WithContext(ctx).Raw(tableInfoSQL, tableName).Row().Scan(
		&name,
		&collation,
		&autoIncrement,
		&comment,
	)
	if err != nil {
		return nil, err
	}

	columnsSQL := `
        SELECT 
            COLUMN_NAME,
            DATA_TYPE,
            CHARACTER_MAXIMUM_LENGTH,
            IS_NULLABLE,
            COLUMN_DEFAULT,
            EXTRA,
            COLUMN_COMMENT
        FROM information_schema.COLUMNS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = ?
        ORDER BY ORDINAL_POSITION
    `

	type columnInfo struct {
		ColumnName    string  `gorm:"column:COLUMN_NAME"`
		DataType      string  `gorm:"column:DATA_TYPE"`
		CharLength    *int    `gorm:"column:CHARACTER_MAXIMUM_LENGTH"`
		IsNullable    string  `gorm:"column:IS_NULLABLE"`
		DefaultValue  *string `gorm:"column:COLUMN_DEFAULT"`
		Extra         string  `gorm:"column:EXTRA"`
		ColumnComment *string `gorm:"column:COLUMN_COMMENT"`
	}

	var columnsData []columnInfo
	err = m.db.WithContext(ctx).Raw(columnsSQL, tableName).Scan(&columnsData).Error
	if err != nil {
		return nil, err
	}

	columns := make([]*entity.Column, len(columnsData))
	for i, colData := range columnsData {
		column := &entity.Column{
			Name:          colData.ColumnName,
			DataType:      entity.DataType(colData.DataType),
			Length:        colData.CharLength,
			NotNull:       colData.IsNullable == "NO",
			DefaultValue:  colData.DefaultValue,
			AutoIncrement: strings.Contains(colData.Extra, "auto_increment"),
			Comment:       colData.ColumnComment,
		}
		columns[i] = column
	}

	indexesSQL := `
        SELECT 
            INDEX_NAME,
            NON_UNIQUE,
            INDEX_TYPE,
            GROUP_CONCAT(COLUMN_NAME ORDER BY SEQ_IN_INDEX)
        FROM information_schema.STATISTICS 
        WHERE TABLE_SCHEMA = DATABASE() 
        AND TABLE_NAME = ?
        GROUP BY INDEX_NAME, NON_UNIQUE, INDEX_TYPE
    `

	type indexInfo struct {
		IndexName string `gorm:"column:INDEX_NAME"`
		NonUnique int    `gorm:"column:NON_UNIQUE"`
		IndexType string `gorm:"column:INDEX_TYPE"`
		Columns   string `gorm:"column:GROUP_CONCAT"`
	}

	var indexesData []indexInfo
	err = m.db.WithContext(ctx).Raw(indexesSQL, tableName).Scan(&indexesData).Error
	if err != nil {
		return nil, err
	}

	indexes := make([]*entity.Index, 0, len(indexesData))
	for _, idxData := range indexesData {
		index := &entity.Index{
			Name:    idxData.IndexName,
			Type:    entity.IndexType(idxData.IndexType),
			Columns: strings.Split(idxData.Columns, ","),
		}
		indexes = append(indexes, index)
	}

	return &entity.Table{
		Name:    name,
		Columns: columns,
		Indexes: indexes,
		Options: &entity.TableOption{
			Collate:       collation,
			AutoIncrement: autoIncrement,
			Comment:       comment,
		},
	}, nil
}

func (m *mysqlService) buildWhereClause(condition *rdb.ComplexCondition) (string, []interface{}, error) {
	if condition == nil {
		return "", nil, nil
	}
	if condition.Operator == "" {
		condition.Operator = entity.AND
	}
	if len(condition.NestedConditions) > 0 {
		return m.buildNestedConditions(condition)
	} else if len(condition.Conditions) > 0 {
		whereClauseString, values, err := m.buildWhereCondition(condition)
		return " WHERE " + whereClauseString, values, err
	} else {
		return "", nil, fmt.Errorf("empty condition: no nested or direct conditions found")
	}

}

func (m *mysqlService) buildWhereCondition(condition *rdb.ComplexCondition) (string, []interface{}, error) {
	var whereClause strings.Builder
	values := make([]interface{}, 0)
	for i, cond := range condition.Conditions {
		if i > 0 {
			whereClause.WriteString(fmt.Sprintf(" %s ", condition.Operator))
		}

		if cond.Operator == entity.OperatorIn || cond.Operator == entity.OperatorNotIn {
			if m.isSlice(cond.Value) {
				sliceValues, err := m.getSliceValues(cond.Value)
				if err != nil {
					return "", nil, fmt.Errorf("failed to process slice values: %v", err)
				}

				if len(sliceValues) == 0 {
					whereClause.WriteString(fmt.Sprintf("%s %s (NULL)", m.quoteIdent(cond.Field), string(cond.Operator)))
				} else {
					placeholders := make([]string, len(sliceValues))
					for i := range placeholders {
						placeholders[i] = "?"
					}
					whereClause.WriteString(fmt.Sprintf("%s %s (%s)", m.quoteIdent(cond.Field), string(cond.Operator), strings.Join(placeholders, ",")))

					values = append(values, sliceValues...)
				}
			} else {
				return "", nil, fmt.Errorf("IN operator requires a slice of values")
			}
		} else if cond.Operator == entity.OperatorIsNull || cond.Operator == entity.OperatorIsNotNull {
			whereClause.WriteString(fmt.Sprintf("%s %s", m.quoteIdent(cond.Field), cond.Operator))
		} else {
			whereClause.WriteString(fmt.Sprintf("%s %s ?", m.quoteIdent(cond.Field), cond.Operator))
			values = append(values, cond.Value)
		}
	}
	if whereClause.Len() > 0 {
		return whereClause.String(), values, nil
	}
	return "", values, nil
}

func (m *mysqlService) buildNestedConditions(condition *rdb.ComplexCondition) (string, []interface{}, error) {
	var whereClause strings.Builder
	values := make([]interface{}, 0)

	whereClause.WriteString(" WHERE (")
	for i, nested := range condition.NestedConditions {
		if i > 0 {
			whereClause.WriteString(fmt.Sprintf(" %s ", nested.Operator))
		}
		nestedClause, nestedValues, err := m.buildWhereCondition(nested)
		if err != nil {
			return "", nil, err
		}
		whereClause.WriteString(nestedClause)
		if i < len(condition.NestedConditions)-1 {
			whereClause.WriteString(" " + string(condition.Operator))
		}
		values = append(values, nestedValues...)
	}
	whereClause.WriteString(")")

	if whereClause.Len() > 0 {
		return whereClause.String(), values, nil
	}
	return "", values, nil
}
