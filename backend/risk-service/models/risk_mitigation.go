package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MonitoringCheck struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	Checked   bool      `json:"checked"`
	Notes     string    `json:"notes"`
	StartDate time.Time `json:"startDate"`
	EndDate   time.Time `json:"endDate"`
}

type RiskMitigation struct {
	ID             uuid.UUID         `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RiskID         uuid.UUID         `gorm:"type:uuid;not null;index" json:"riskId"`
	RiskEvent      string            `gorm:"type:text;not null" json:"riskEvent"`
	MitigationPlan string            `gorm:"type:text;not null" json:"mitigationPlan"`
	Supervisor     string            `gorm:"type:varchar(200)" json:"supervisor"`
	PIC            string            `gorm:"type:varchar(200)" json:"pic"`
	UnitInCharge   string            `gorm:"type:varchar(200)" json:"unitInCharge"`
	StartDate      time.Time         `json:"start_date"`
	EndDate        time.Time         `json:"end_date"`
	Notes          string            `gorm:"type:text" json:"notes"`
	MonitoringData string            `gorm:"type:text" json:"-"`
	CreatedAt      time.Time         `json:"created_at"`
	UpdatedAt      time.Time         `json:"updated_at"`
	DeletedAt      gorm.DeletedAt    `gorm:"index" json:"-"`
	Monitoring     []MonitoringCheck `gorm:"-" json:"monitoring"`
}

func (RiskMitigation) TableName() string {
	return "risk_mitigation"
}

func GenerateMonitoringChecks(start, end time.Time) []MonitoringCheck {
	checks := make([]MonitoringCheck, 0)
	if start.IsZero() || end.IsZero() {
		return checks
	}

	durationDays := int(end.Sub(start).Hours() / 24)
	// Under 2 months (approx 60 days) -> Weekly
	if durationDays < 60 {
		numWeeks := (durationDays + 6) / 7
		if numWeeks <= 0 {
			numWeeks = 1
		}
		for i := 1; i <= numWeeks; i++ {
			wStart := start.AddDate(0, 0, (i-1)*7)
			wEnd := start.AddDate(0, 0, i*7).Add(-time.Second)
			if wEnd.After(end) {
				wEnd = end
			}
			checks = append(checks, MonitoringCheck{
				ID:        fmt.Sprintf("W%d", i),
				Label:     fmt.Sprintf("Week %d (%s - %s)", i, wStart.Format("02 Jan"), wEnd.Format("02 Jan")),
				Checked:   false,
				Notes:     "",
				StartDate: wStart,
				EndDate:   wEnd,
			})
		}
	} else {
		// Monthly
		current := start
		monthNames := []string{"Januari", "Februari", "Maret", "April", "Mei", "Juni", "Juli", "Agustus", "September", "Oktober", "November", "Desember"}

		i := 1
		for !current.After(end) {
			mStart := time.Date(current.Year(), current.Month(), 1, 0, 0, 0, 0, current.Location())
			if mStart.Before(start) {
				mStart = start
			}

			// Next month
			nextMonth := current.AddDate(0, 1, 0)
			mEnd := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, 0, current.Location()).Add(-time.Second)
			if mEnd.After(end) {
				mEnd = end
			}

			checks = append(checks, MonitoringCheck{
				ID:        fmt.Sprintf("M%d", i),
				Label:     fmt.Sprintf("%s %d", monthNames[current.Month()-1], current.Year()),
				Checked:   false,
				Notes:     "",
				StartDate: mStart,
				EndDate:   mEnd,
			})

			current = time.Date(current.Year(), current.Month()+1, 1, 0, 0, 0, 0, current.Location())
			i++
		}
	}
	return checks
}
