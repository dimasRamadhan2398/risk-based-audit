package actiontakenreport

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"audit-service/models"
	"audit-service/pkg/authclient"
	"audit-service/pkg/logger"
	"audit-service/pkg/response"
	atrsvc "audit-service/services/action_taken_report"

	"github.com/gin-gonic/gin"
)

// Error codes of the reminder endpoint (response "error" field)
const (
	ErrCodeATRCancelled         = "ATR_CANCELLED"
	ErrCodeEmailNotConfigured   = "EMAIL_NOT_CONFIGURED"
	ErrCodeRecipientLookup      = "RECIPIENT_LOOKUP_FAILED"
	ErrCodeNoReminderRecipients = "NO_REMINDER_RECIPIENTS"
	ErrCodeEmailSendFailed      = "EMAIL_SEND_FAILED"
)

// Reminder recipient roles
const (
	reminderRoleAuditee  = "auditee"
	reminderRoleDeptHead = "department_head"
)

// ReminderRecipient is one address a reminder was (or would be) sent to
type ReminderRecipient struct {
	Role  string `json:"role"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// ReminderSkip is a recipient that got no e-mail, with the reason
type ReminderSkip struct {
	Role   string `json:"role"`
	Reason string `json:"reason"`
}

// ReminderResponse is the data of POST /action-taken-reports/:id/reminder
type ReminderResponse struct {
	SentTo  []ReminderRecipient `json:"sent_to"`
	Skipped []ReminderSkip      `json:"skipped"`
	Failed  []ReminderRecipient `json:"failed"`
}

// reminderPlan is the recipient list of one reminder, before sending
type reminderPlan struct {
	recipients []ReminderRecipient
	skipped    []ReminderSkip
}

// resolveReminderRecipients works out who gets the reminder:
//
//   - the auditee: the ATR's PIC (pic_user_id), looked up in auth-service;
//   - the department head: every active DEPARTMENT_HEAD of the audited
//     department (the LHA's department; the PIC's department when the LHA is
//     gone), looked up in auth-service.
//
// A recipient without an e-mail address, or one no longer active, is skipped
// and listed with the reason. Addresses are de-duplicated (case-insensitive).
// An error means auth-service could not be asked.
func (ctl *Controller) resolveReminderRecipients(ctx context.Context, authorization string, r *models.ActionTakenReport, lha *models.AuditResultReport) (reminderPlan, error) {
	var plan reminderPlan
	var candidates []ReminderRecipient

	var pic *authclient.Contact
	if strings.TrimSpace(r.PICUserID) == "" {
		plan.skipped = append(plan.skipped, ReminderSkip{Role: reminderRoleAuditee, Reason: "no PIC is assigned to this action taken report"})
	} else {
		contacts, err := ctl.Contacts.Contacts(ctx, authorization, []string{r.PICUserID}, "")
		if err != nil {
			return plan, err
		}
		if len(contacts) == 0 {
			plan.skipped = append(plan.skipped, ReminderSkip{Role: reminderRoleAuditee, Reason: "the PIC is not an active user"})
		} else {
			pic = &contacts[0]
			candidates = append(candidates, ReminderRecipient{Role: reminderRoleAuditee, Name: pic.FullName, Email: pic.Email})
		}
	}

	dept := ""
	if lha != nil {
		dept = strings.TrimSpace(lha.Department)
	}
	if dept == "" && pic != nil {
		dept = strings.TrimSpace(pic.Department)
	}
	if dept == "" {
		plan.skipped = append(plan.skipped, ReminderSkip{Role: reminderRoleDeptHead, Reason: "no department is known for this action taken report"})
	} else {
		heads, err := ctl.Contacts.Contacts(ctx, authorization, nil, dept)
		if err != nil {
			return plan, err
		}
		if len(heads) == 0 {
			plan.skipped = append(plan.skipped, ReminderSkip{Role: reminderRoleDeptHead, Reason: "no active department head in " + dept})
		}
		for _, h := range heads {
			candidates = append(candidates, ReminderRecipient{Role: reminderRoleDeptHead, Name: h.FullName, Email: h.Email})
		}
	}

	seen := map[string]bool{}
	for _, cand := range candidates {
		email := strings.TrimSpace(cand.Email)
		if email == "" {
			plan.skipped = append(plan.skipped, ReminderSkip{Role: cand.Role, Reason: "no e-mail address on file for " + displayName(cand.Name)})
			continue
		}
		key := strings.ToLower(email)
		if seen[key] {
			continue
		}
		seen[key] = true
		cand.Email = email
		plan.recipients = append(plan.recipients, cand)
	}
	return plan, nil
}

func displayName(name string) string {
	if strings.TrimSpace(name) == "" {
		return "a user"
	}
	return strings.TrimSpace(name)
}

// reminderMessage builds the subject and plain-text body for one recipient
func reminderMessage(name string, r *models.ActionTakenReport, lha *models.AuditResultReport, now time.Time) (subject, text string) {
	title := strings.TrimSpace(r.FindingTitle)
	if title == "" {
		title = strings.TrimSpace(r.ReportNumber)
	}
	// One line only: a title with line breaks must not add mail headers
	subject = strings.Join(strings.Fields("Pengingat tindak lanjut temuan: "+title), " ")

	reportLine := strings.TrimSpace(r.ReportNumber + " " + r.ReportTitle)
	if lha != nil && strings.TrimSpace(lha.ReportNumber+" "+lha.ReportTitle) != "" {
		reportLine = strings.TrimSpace(lha.ReportNumber + " " + lha.ReportTitle)
	}
	due := "belum ditetapkan"
	if r.DueDate != nil && !r.DueDate.IsZero() {
		due = models.ATRDueDate(r.DueDate.UTC()).Format("2006-01-02")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Yth. %s,\n\n", displayName(name))
	b.WriteString("Ini adalah pengingat untuk tindak lanjut temuan audit berikut:\n\n")
	fmt.Fprintf(&b, "LHA       : %s\n", reportLine)
	fmt.Fprintf(&b, "Temuan    : %s\n", title)
	if rec := strings.TrimSpace(r.Recommendation); rec != "" {
		fmt.Fprintf(&b, "Rekomendasi: %s\n", rec)
	}
	fmt.Fprintf(&b, "Status    : %s\n", r.Status)
	fmt.Fprintf(&b, "Progres   : %d%%\n", r.Progress)
	fmt.Fprintf(&b, "Batas waktu: %s\n", due)
	if overdue, days := models.ComputeATROverdue(r.DueDate, r.Status, now); overdue {
		fmt.Fprintf(&b, "\nBatas waktu telah terlewat %d hari.\n", days)
	}
	b.WriteString("\nMohon segera menyampaikan rencana tindak lanjut atau progres melalui AuditSphere.\n\n")
	b.WriteString("Salam,\nAuditSphere\n")
	return subject, b.String()
}

// SendReminder e-mails the PIC (auditee) and the department head(s) of the
// audited department about an open ATR (POST /action-taken-reports/:id/reminder).
//
//   - Permission: assign (ADMIN, AUDIT_MANAGER, AUDITOR).
//   - A CANCELLED ATR is rejected with 409 ATR_CANCELLED and nothing is sent.
//   - Each recipient gets their own e-mail (no shared "To" list).
//   - 503 when e-mail (RESEND_API_KEY) or auth-service lookups are not
//     configured/reachable; 422 when no recipient has an address; 502 when
//     every send failed. A partial failure answers 200 with "failed" listed.
func (ctl *Controller) SendReminder(c *gin.Context) {
	a := actor(c)
	r, ok := ctl.load(c)
	if !ok {
		return
	}
	if !a.Perms[atrsvc.PermAssign] {
		forbidden(c, "You are not allowed to send reminders for action taken reports")
		return
	}
	if r.Status == models.ATRStatusCancelled {
		response.Error(c, http.StatusConflict, ErrCodeATRCancelled,
			"A cancelled action taken report cannot receive reminders", "")
		return
	}
	if ctl.Mailer == nil || ctl.Contacts == nil {
		response.Error(c, http.StatusServiceUnavailable, ErrCodeEmailNotConfigured,
			"Reminder e-mail is not configured on this server", "")
		return
	}

	lhas, err := ctl.lhasFor([]models.ActionTakenReport{*r})
	if err != nil {
		response.InternalServerError(c, "Failed to fetch action taken report")
		return
	}
	lha := lhaOf(lhas, r)

	authorization := c.GetHeader("Authorization")
	plan, err := ctl.resolveReminderRecipients(c.Request.Context(), authorization, r, lha)
	if err != nil {
		if logger.Log != nil {
			logger.Warn("ATR reminder: recipient lookup failed", logger.LogField("atr_id", r.ID.String()), logger.LogField("error", err.Error()))
		}
		response.Error(c, http.StatusServiceUnavailable, ErrCodeRecipientLookup,
			"Could not look up the reminder recipients; try again later", "")
		return
	}
	if len(plan.recipients) == 0 {
		response.Error(c, http.StatusUnprocessableEntity, ErrCodeNoReminderRecipients,
			"No reminder recipient has an e-mail address", "")
		return
	}

	out := ReminderResponse{SentTo: []ReminderRecipient{}, Skipped: plan.skipped, Failed: []ReminderRecipient{}}
	if out.Skipped == nil {
		out.Skipped = []ReminderSkip{}
	}
	now := ctl.now()
	for _, rcp := range plan.recipients {
		subject, text := reminderMessage(rcp.Name, r, lha, now)
		if err := ctl.Mailer.Send(c.Request.Context(), []string{rcp.Email}, subject, text); err != nil {
			if logger.Log != nil {
				logger.Warn("ATR reminder: send failed", logger.LogField("atr_id", r.ID.String()),
					logger.LogField("role", rcp.Role), logger.LogField("error", err.Error()))
			}
			out.Failed = append(out.Failed, rcp)
			continue
		}
		out.SentTo = append(out.SentTo, rcp)
	}

	if len(out.SentTo) == 0 {
		response.Error(c, http.StatusBadGateway, ErrCodeEmailSendFailed, "The reminder could not be sent to any recipient", "")
		return
	}
	if logger.Log != nil {
		logger.Info("ATR reminder sent", logger.LogField("atr_id", r.ID.String()),
			logger.LogField("sent", len(out.SentTo)), logger.LogField("failed", len(out.Failed)),
			logger.LogField("by", a.DisplayName()))
	}
	response.OK(c, "Reminder sent", out)
}
