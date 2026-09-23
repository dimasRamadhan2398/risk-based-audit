package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ExecutiveSummary struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Quarter             int            `gorm:"type:int" json:"quarter"` // 1, 2, 3, 4
	PeriodeBulan        string         `gorm:"type:varchar(50)" json:"periodeBulan"`
	Tahun               int            `gorm:"type:int" json:"tahun"`
	NomorDokumen        string         `gorm:"type:varchar(100)" json:"nomorDokumen"`
	DokumenPath         string         `gorm:"type:varchar(255)" json:"dokumenPath"` // Uploaded file
	Status              string         `gorm:"type:varchar(50);default:'Draft'" json:"status"` // Draft, Approved, Rejected
	
	// Section I
	Narrative           string         `gorm:"type:text" json:"narrative"`

	// Section II
	JumlahLaporan       int            `gorm:"type:int" json:"jumlahLaporan"`
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
