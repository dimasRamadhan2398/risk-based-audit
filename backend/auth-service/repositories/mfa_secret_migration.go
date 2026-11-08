package repositories

import (
	"errors"
	"fmt"

	"auth-service/models"
	"auth-service/pkg/secretcrypt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MFASecretMigrationResult counts what EncryptMFASecrets found and did.
type MFASecretMigrationResult struct {
	Total      int // rows in mfa_setups
	Empty      int // empty secret (email MFA): nothing to encrypt
	Current    int // already encrypted with the primary key
	Legacy     int // plaintext rows encrypted (or that would be, on dry run)
	Rotated    int // old-key rows re-encrypted with the primary key
	Unreadable int // encrypted rows no configured key opens (left untouched)
}

// Changed is the number of rows rewritten (or to be rewritten on dry run).
func (r MFASecretMigrationResult) Changed() int { return r.Legacy + r.Rotated }

// EncryptMFASecrets encrypts every legacy plaintext mfa_setups.secret_key with
// the primary key of c and re-encrypts values only an old key opens. It runs
// in one transaction with the rows locked, and is idempotent: a second run
// finds only Current rows and changes nothing.
//
// If any row is Unreadable the transaction is rolled back and an error is
// returned, so a wrong key never leaves the table half migrated. dryRun counts
// without writing.
func EncryptMFASecrets(db *gorm.DB, c *secretcrypt.Cipher, dryRun bool) (MFASecretMigrationResult, error) {
	var res MFASecretMigrationResult
	if !c.Enabled() {
		return res, secretcrypt.ErrNoKey
	}
	err := db.Transaction(func(tx *gorm.DB) error {
		var rows []models.MFASetup
		q := tx.Model(&models.MFASetup{}).Select("id", "user_id", "secret_key").Order("id")
		if tx.Dialector.Name() == "postgres" {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.Find(&rows).Error; err != nil {
			return fmt.Errorf("read mfa_setups: %w", err)
		}
		res.Total = len(rows)
		for _, row := range rows {
			if row.SecretKey == "" {
				res.Empty++
				continue
			}
			aad := mfaSecretAAD(row.UserID)
			plain, rewrite, err := c.Open(row.SecretKey, aad)
			if err != nil {
				res.Unreadable++
				continue
			}
			if !rewrite {
				res.Current++
				continue
			}
			if secretcrypt.IsEncrypted(row.SecretKey) {
				res.Rotated++
			} else {
				res.Legacy++
			}
			if dryRun {
				continue
			}
			enc, err := c.Encrypt(plain, aad)
			if err != nil {
				return err
			}
			if len(enc) > 255 {
				return fmt.Errorf("mfa_setups %s: encrypted secret is %d chars, column is varchar(255)", row.ID, len(enc))
			}
			upd := tx.Model(&models.MFASetup{}).
				Where("id = ? AND secret_key = ?", row.ID, row.SecretKey).
				UpdateColumn("secret_key", enc)
			if upd.Error != nil {
				return fmt.Errorf("update mfa_setups %s: %w", row.ID, upd.Error)
			}
			if upd.RowsAffected != 1 {
				return fmt.Errorf("update mfa_setups %s: row changed during migration", row.ID)
			}
		}
		if res.Unreadable > 0 {
			return errUnreadableRows
		}
		if dryRun {
			return errDryRun
		}
		return nil
	})
	switch {
	case errors.Is(err, errDryRun):
		return res, nil
	case errors.Is(err, errUnreadableRows):
		return res, fmt.Errorf("%d mfa_setups row(s) cannot be decrypted with MFA_ENCRYPTION_KEY or MFA_ENCRYPTION_KEY_OLD; nothing was changed", res.Unreadable)
	}
	return res, err
}

var (
	errDryRun         = errors.New("dry run")
	errUnreadableRows = errors.New("unreadable rows")
)
