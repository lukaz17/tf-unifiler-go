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
	"github.com/google/uuid"
	"github.com/tforce-io/tf-golib/stdx/opx"
)

type SetHash struct {
	SetID  uuid.UUID `gorm:"column:set_id;primaryKey"`
	HashID uuid.UUID `gorm:"column:hash_id;primaryKey"`
}

func NewSetHash(setID, hashID uuid.UUID) *SetHash {
	return &SetHash{
		SetID:  setID,
		HashID: hashID,
	}
}

func (ctx *DbContext) GetSetHashesByHashIDs(hashes uuid.UUIDs) ([]*SetHash, error) {
	return ctx.findSetHashesByHashIDs(hashes)
}

func (ctx *DbContext) GetSetHashesBySetIDs(hashes uuid.UUIDs) ([]*SetHash, error) {
	return ctx.findSetHashesBySetIDs(hashes)
}

func (ctx *DbContext) SaveSetHashes(setHashes []*SetHash) error {
	hashes := make([]uuid.UUID, len(setHashes))
	for i, setHash := range setHashes {
		hashes[i] = setHash.HashID
	}
	changedSetHashes, err := ctx.findSetHashesBySetIDs(hashes)
	if err != nil {
		return err
	}
	newSetHashes := []*SetHash{}
	for _, setHash := range setHashes {
		existed := opx.ContainsFunc(changedSetHashes, setHash, areEqualSetHashes)
		if !existed {
			newSetHashes = append(newSetHashes, setHash)
		}
	}
	return ctx.writeSetHashes(newSetHashes, []*SetHash{})
}

func (ctx *DbContext) findSetHashesByHashIDs(hashes uuid.UUIDs) ([]*SetHash, error) {
	var docs []*SetHash
	result := ctx.db.Model(&SetHash{}).
		Where("hash_id IN ?", hashes).
		Find(&docs)
	return docs, result.Error
}

func (ctx *DbContext) findSetHashesBySetIDs(hashes uuid.UUIDs) ([]*SetHash, error) {
	var docs []*SetHash
	result := ctx.db.Model(&SetHash{}).
		Where("set_id IN ?", hashes).
		Find(&docs)
	return docs, result.Error
}

func (ctx *DbContext) writeSetHashes(newSetHashes []*SetHash, _ []*SetHash) error {
	tx := ctx.db.Begin()
	for _, setHash := range newSetHashes {
		result := tx.Create(setHash)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	tx.Commit()
	return nil
}

func areEqualSetHashes(x, y *SetHash) bool {
	if x == nil && y == nil {
		return true
	}
	if x == nil || y == nil {
		return false
	}
	return x.SetID == y.SetID && x.HashID == y.HashID
}
