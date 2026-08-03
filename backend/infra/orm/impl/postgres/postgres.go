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

package postgres

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/coze-dev/coze-studio/backend/pkg/envkey"
	"github.com/coze-dev/coze-studio/backend/pkg/logs"
)

func New() (*gorm.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Build DSN from individual environment variables
		host := os.Getenv("DATABASE_HOST")
		port := envkey.GetStringD("DATABASE_PORT", "5432")
		user := os.Getenv("DATABASE_USER")
		password := os.Getenv("DATABASE_PASSWORD")
		dbname := os.Getenv("DATABASE_NAME")
		sslmode := envkey.GetStringD("DATABASE_SSLMODE", "disable")

		if host == "" || user == "" || dbname == "" {
			return nil, fmt.Errorf("missing required PostgreSQL connection parameters: DATABASE_HOST, DATABASE_USER, DATABASE_NAME")
		}

		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("postgres open, dsn: %s, err: %w", dsn, err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		logs.Errorf("InitDB. db.DB() fail. err:%v", err)
		return nil, err
	}

	// Connection pool configuration
	sqlDB.SetMaxIdleConns(envkey.GetIntD("DATABASE_MAX_IDLE_CONNS", 10))
	sqlDB.SetMaxOpenConns(envkey.GetIntD("DATABASE_MAX_OPEN_CONNS", 100))
	sqlDB.SetConnMaxLifetime(time.Duration(envkey.GetIntD("DATABASE_CONN_MAX_LIFETIME", 3600)) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(envkey.GetIntD("DATABASE_CONN_MAX_IDLE_TIME", 600)) * time.Second)

	return db, nil
}
