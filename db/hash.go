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
	"github.com/tforceaio/tf-unifiler-go/core"
)

type Hash struct {
	ID     uuid.UUID `gorm:"column:id;primaryKey"`
	Md5    string    `gorm:"column:md5"`
	Sha1   string    `gorm:"column:sha1"`
	Sha256 string    `gorm:"column:sha256;uniqueIndex"`
	Sha512 string    `gorm:"column:sha512"`

	Size        uint32 `gorm:"column:size"`
	Description string `gorm:"column:description"`
	IsIgnored   bool   `gorm:"column:is_ignored"`
}

func NewHash(fileHashes *core.FileMultiHash, isIgnored bool) *Hash {
	return &Hash{
		Md5:         fileHashes.Md5.HexStr(),
		Sha1:        fileHashes.Sha1.HexStr(),
		Sha256:      fileHashes.Sha256.HexStr(),
		Sha512:      fileHashes.Sha512.HexStr(),
		Size:        fileHashes.Size,
		Description: fileHashes.FileName,
		IsIgnored:   isIgnored,
	}
}

func (ctx *DbContext) GetHash(id uuid.UUID) (*Hash, error) {
	return ctx.findHash(id)
}

func (ctx *DbContext) GetHashBySha256(hash string) (*Hash, error) {
	return ctx.findHashBySha256(hash)
}

func (ctx *DbContext) GetHashesBySetIDs(setIDs uuid.UUIDs) ([]*Hash, error) {
	return ctx.findHashesBySetIDs(setIDs)
}

func (ctx *DbContext) GetHashesBySha256s(hashes []string) ([]*Hash, error) {
	return ctx.findHashesBySha256s(hashes)
}

func (ctx *DbContext) SaveHash(hash *Hash) error {
	changedHash, err := ctx.findHashBySha256(hash.Sha256)
	if err != nil {
		return err
	}
	newHashes := []*Hash{}
	changedHashes := []*Hash{}
	if changedHash == nil {
		newHashes = append(newHashes, hash)
	} else {
		changedHashes = append(changedHashes, hash)
	}
	return ctx.writeHashes(newHashes, changedHashes)
}

func (ctx *DbContext) SaveHashes(hashes []*Hash) error {
	sha256s := make([]string, len(hashes))
	for i, hash := range hashes {
		sha256s[i] = hash.Sha256
	}
	changedHashes, err := ctx.findHashesBySha256s(sha256s)
	if err != nil {
		return err
	}
	changedHashesMap := map[string]uuid.UUID{}
	for _, hash := range changedHashes {
		changedHashesMap[hash.Sha256] = hash.ID
	}
	newHashes := []*Hash{}
	for _, hash := range hashes {
		if _, ok := changedHashesMap[hash.Sha256]; ok {
			continue
		}
		newHashes = append(newHashes, hash)
	}
	return ctx.writeHashes(newHashes, []*Hash{})
}

func (ctx *DbContext) findHash(id uuid.UUID) (*Hash, error) {
	var doc *Hash
	result := ctx.db.Model(&Hash{}).
		Where("id = ?", id).
		First(&doc)
	if ctx.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

func (ctx *DbContext) findHashBySha256(hash string) (*Hash, error) {
	var doc *Hash
	result := ctx.db.Model(&Hash{}).
		Where("sha256 = ?", hash).
		First(&doc)
	if ctx.isEmptyResultError(result.Error) {
		return nil, nil
	}
	return doc, result.Error
}

func (ctx *DbContext) findHashesBySetIDs(setIDs uuid.UUIDs) ([]*Hash, error) {
	var docs []*Hash
	result := ctx.db.Model(&Hash{}).
		InnerJoins("hashes ON hashes.id = set_hashes.hash_id AND set_hashes.set_id IN ?", setIDs).
		Find(&docs)
	return docs, result.Error
}

func (ctx *DbContext) findHashesBySha256s(hashes []string) ([]*Hash, error) {
	var docs []*Hash
	result := ctx.db.Model(&Hash{}).
		Where("sha256 IN ?", hashes).
		Find(&docs)
	return docs, result.Error
}

func (ctx *DbContext) writeHashes(newHashes []*Hash, changedHashes []*Hash) error {
	tx := ctx.db.Begin()
	for _, hash := range newHashes {
		if hash.ID == uuid.Nil {
			var err error
			hash.ID, err = uuid.NewV7()
			if err != nil {
				return err
			}
		}
		result := tx.Create(hash)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	for _, hash := range changedHashes {
		result := tx.Model(&Hash{}).
			Where("id = ?", hash.ID).
			Updates(map[string]interface{}{
				"md5":         hash.Md5,
				"sha1":        hash.Sha1,
				"sha256":      hash.Sha256,
				"sha512":      hash.Sha512,
				"size":        hash.Size,
				"description": hash.Description,
				"is_ignored":  hash.IsIgnored,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	tx.Commit()
	return nil
}

func (ctx *DbContext) setIgnoredBySha256s(hashesToIgnore, hashesToApprove []string) error {
	tx := ctx.db.Begin()
	if len(hashesToIgnore) > 0 {
		result := tx.Model(&Hash{}).
			Where("sha256 IN ?", hashesToIgnore).
			Updates(map[string]interface{}{
				"is_ignored": true,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	if len(hashesToApprove) > 0 {
		result := tx.Model(&Hash{}).
			Where("sha256 IN ?", hashesToApprove).
			Updates(map[string]interface{}{
				"is_ignored": false,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	return nil
}
