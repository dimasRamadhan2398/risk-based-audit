package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExecutiveSummary struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AssignmentLetterID  string         `gorm:"type:varchar(100)" json:"assignmentLetterId"`
	Quarter             int            `gorm:"type:int" json:"quarter"` // 1, 2, 3, 4
	PeriodeBulan        string         `gorm:"type:varchar(50)" json:"periodeBulan"`
	Tahun               int            `gorm:"type:int" json:"tahun"`
	NomorDokumen        string         `gorm:"type:varchar(100)" json:"nomorDokumen"`
	DokumenPath         string         `gorm:"type:varchar(255)" json:"dokumenPath"` // Uploaded file
	Status              string         `gorm:"type:varchar(50);default:'Draft'" json:"status"` // Draft, Approved, Rejected
	ExecutiveNote       string         `gorm:"type:text" json:"executiveNote"`                // Note from Executive to Auditor

	// Section I
	Narrative           string         `gorm:"type:text" json:"narrative"`

	// Section II
	JumlahLaporan       int            `gorm:"type:int" json:"jumlahLaporan"`
	// Findings by the five working paper risk levels
	RisikoHigh          int            `gorm:"type:int;not null;default:0" json:"risikoHigh"`
	RisikoModerateHigh  int            `gorm:"type:int;not null;default:0" json:"risikoModerateHigh"`
	RisikoModerate      int            `gorm:"type:int;not null;default:0" json:"risikoModerate"`
	RisikoLowModerate   int            `gorm:"type:int;not null;default:0" json:"risikoLowModerate"`
	RisikoLow           int            `gorm:"type:int;not null;default:0" json:"risikoLow"`
	// Legacy three-level counts, kept for backward compatibility. Writes keep
	// them in step with the five levels (see SyncRiskCounts).
	RisikoTinggi        int            `gorm:"type:int" json:"risikoTinggi"`
	RisikoSedang        int            `gorm:"type:int" json:"risikoSedang"`
	RisikoRendah        int            `gorm:"type:int" json:"risikoRendah"`
	JumlahRekomendasi   int            `gorm:"type:int" json:"jumlahRekomendasi"`

	// Section III & IV & Matriks Induk are stored as serialized JSON strings for simplicity & database compatibility
	FollowUpTable       string         `gorm:"type:text" json:"followUpTable"` // JSON array
	TopFindings         string         `gorm:"type:text" json:"topFindings"`   // JSON array
	MatriksKompilasi    string         `gorm:"type:text" json:"matriksKompilasi"` // JSON array

	// Section V & VII
	AkarMasalah         string         `gorm:"type:text" json:"akarMasalah"`
	Kesimpulan          string         `gorm:"type:text" json:"kesimpulan"`

	// Signature
	SignatureTempat     string         `gorm:"type:varchar(100)" json:"signatureTempat"`
	SignatureTanggal    string         `gorm:"type:varchar(100)" json:"signatureTanggal"`
	SignatureNamaKepala string         `gorm:"type:varchar(200)" json:"signatureNamaKepala"`
	SignatureNIK        string         `gorm:"type:varchar(50)" json:"signatureNIK"`

	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
	DeletedAt           gorm.DeletedAt `gorm:"index" json:"-"`
}

func (e *ExecutiveSummary) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	e.SyncRiskCounts()
	return
}

// Five-level and legacy risk count columns of executive_summaries
var (
	ExecutiveSummaryRiskColumns       = []string{"risiko_high", "risiko_moderate_high", "risiko_moderate", "risiko_low_moderate", "risiko_low"}
	ExecutiveSummaryLegacyRiskColumns = []string{"risiko_tinggi", "risiko_sedang", "risiko_rendah"}
)

// LegacyRiskCounts collapses the five levels onto the legacy three:
// Tinggi = high + moderate-high, Sedang = moderate, Rendah = low-moderate + low.
func LegacyRiskCounts(high, moderateHigh, moderate, lowModerate, low int) (tinggi, sedang, rendah int) {
	return high + moderateHigh, moderate, lowModerate + low
}

// SyncRiskCounts keeps the two sets of risk counts in step on create. When
// any five-level count is set, the legacy counts are derived from them;
// otherwise legacy counts from an old client are copied into the five levels
// (Tinggi -> High, Sedang -> Moderate, Rendah -> Low).
func (e *ExecutiveSummary) SyncRiskCounts() {
	if e.RisikoHigh != 0 || e.RisikoModerateHigh != 0 || e.RisikoModerate != 0 || e.RisikoLowModerate != 0 || e.RisikoLow != 0 {
		e.RisikoTinggi, e.RisikoSedang, e.RisikoRendah = LegacyRiskCounts(e.RisikoHigh, e.RisikoModerateHigh, e.RisikoModerate, e.RisikoLowModerate, e.RisikoLow)
		return
	}
	e.RisikoHigh, e.RisikoModerate, e.RisikoLow = e.RisikoTinggi, e.RisikoSedang, e.RisikoRendah
}

// NormalizeExecutiveSummaryRiskCounts copies the legacy counts of rows saved
// before the five levels existed into the new columns (Tinggi -> High,
// Sedang -> Moderate, Rendah -> Low), only where every five-level count is
// still 0 and a legacy count is not. Idempotent: a copied row no longer
// matches, and writes keep both sets in step (an all-zero five-level save
// also zeroes the legacy counts). Timestamps are left untouched.
func NormalizeExecutiveSummaryRiskCounts(db *gorm.DB) (int64, error) {
	res := db.Unscoped().Model(&ExecutiveSummary{}).
		Where("COALESCE(risiko_high, 0) = 0 AND COALESCE(risiko_moderate_high, 0) = 0 AND COALESCE(risiko_moderate, 0) = 0 AND COALESCE(risiko_low_moderate, 0) = 0 AND COALESCE(risiko_low, 0) = 0").
		Where("COALESCE(risiko_tinggi, 0) <> 0 OR COALESCE(risiko_sedang, 0) <> 0 OR COALESCE(risiko_rendah, 0) <> 0").
		UpdateColumns(map[string]interface{}{
			"risiko_high":     gorm.Expr("COALESCE(risiko_tinggi, 0)"),
			"risiko_moderate": gorm.Expr("COALESCE(risiko_sedang, 0)"),
			"risiko_low":      gorm.Expr("COALESCE(risiko_rendah, 0)"),
		})
	return res.RowsAffected, res.Error
}

