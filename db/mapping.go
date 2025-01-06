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

import "github.com/google/uuid"

type Mapping struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	HashID    uuid.UUID `gorm:"column:hash_id"`
	Name      string    `gorm:"column:name"`
	Extension string    `gorm:"column:extension"`
}

func (e *Mapping) FullName() string {
	if e.Extension == "" {
		return e.Name
	}
	return e.Name + e.Extension
}
