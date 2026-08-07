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
	"bytes"
	"encoding/csv"
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/extrame/xls"
	"github.com/xuri/excelize/v2"
)

const (
	maxImportRows = 100000
	maxImportCols = 500
)

// ParseImportFile parses CSV / XLSX / XLS file content into raw rows. The
// first row is treated as the header row by the caller.
func ParseImportFile(filename string, data []byte) ([][]string, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".csv":
		return parseCSV(data)
	case ".xlsx":
		return parseXLSX(data)
	case ".xls":
		return parseXLS(data)
	default:
		return nil, fmt.Errorf("unsupported file type %q, only csv/xlsx/xls are supported", ext)
	}
}

func parseCSV(data []byte) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.FieldsPerRecord = -1
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("parse csv failed: %w", err)
	}
	return normalizeRows(rows), nil
}

func parseXLSX(data []byte) ([][]string, error) {
	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("parse xlsx failed: %w", err)
	}
	defer func() {
		_ = f.Close()
	}()
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, errors.New("xlsx file has no sheet")
	}
	rows, err := f.GetRows(sheets[0])
	if err != nil {
		return nil, fmt.Errorf("read xlsx sheet failed: %w", err)
	}
	return normalizeRows(rows), nil
}

func parseXLS(data []byte) ([][]string, error) {
	wb, err := xls.OpenReader(bytes.NewReader(data), "UTF-8")
	if err != nil {
		return nil, fmt.Errorf("parse xls failed: %w", err)
	}
	rows := wb.ReadAllCells(maxImportRows)
	return normalizeRows(rows), nil
}

func normalizeRows(rows [][]string) [][]string {
	out := make([][]string, 0, len(rows))
	for _, row := range rows {
		if len(row) > maxImportCols {
			row = row[:maxImportCols]
		}
		end := len(row)
		for end > 0 && strings.TrimSpace(row[end-1]) == "" {
			end--
		}
		row = row[:end]
		empty := true
		for _, cell := range row {
			if strings.TrimSpace(cell) != "" {
				empty = false
				break
			}
		}
		if !empty {
			out = append(out, row)
		}
	}
	return out
}
