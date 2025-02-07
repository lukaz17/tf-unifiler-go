// Copyright (C) 2025 T-Force I/O
// This file is part of TF Unifiler
//
// TF Unifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TF Unifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TF Unifiler. If not, see <https://www.gnu.org/licenses/>.

package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var SchemaVersion = 1

type DbContext struct {
	db  *gorm.DB
	uri string
}

func Connect(uri string) (*DbContext, error) {
	db, err := gorm.Open(sqlite.Open(uri), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	c := &DbContext{db, uri}
	c.Migrate()
	return c, nil
}

func (c *DbContext) Disconnect() {
}

func (c *DbContext) Migrate() error {
	return c.db.AutoMigrate(&Hash{}, &Mapping{}, &Set{}, &SetHash{})
}

func (c *DbContext) Count(model interface{}, query, args interface{}) (int64, error) {
	var count int64
	if query == nil {
		result := c.db.Model(model).Count(&count)
		return count, result.Error
	}
	result := c.db.Model(model).Where(query, args).Count(&count)
	return count, result.Error
}

func (c *DbContext) Truncate(model interface{}) {
	c.db.Where("1 = 1").Delete(model)
}

func (c *DbContext) isEmptyResultError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return errStr == "record not found"
}
