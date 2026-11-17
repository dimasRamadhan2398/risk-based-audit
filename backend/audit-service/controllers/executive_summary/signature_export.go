package executivesummary

import (
	apperrors "audit-service/pkg/errors"
)

// ValidateSignatureImage checks an approval signature data URL the same way
// the executive summary approval does: a PNG or JPEG data URL of at most
// MaxSignatureImageBytes whose bytes really are that type. Other approvals
// (the COSO internal control assessment) reuse it so every stored signature
// image passes one rule. An empty value is an error here; callers with an
// optional signature check for "" first.
func ValidateSignatureImage(raw string) (string, *apperrors.AppError) {
	return validateSignatureImage(raw)
}

// MaxApproveBodyBytes bounds an approval request body that carries a
// signature image
const MaxApproveBodyBytes = maxApproveBodyBytes
