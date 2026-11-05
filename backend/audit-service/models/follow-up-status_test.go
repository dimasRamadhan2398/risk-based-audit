package models

import (
	"testing"
	"time"
)

func fuDate(y int, m time.Month, d int) *time.Time {
	t := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return &t
}

func TestClassifyFollowUp(t *testing.T) {
	// 2026-10-14 09:00 UTC = 16:00 WIB, Jakarta date 14 Oct
	asOf := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	approvedLongAgo := time.Date(2026, 9, 1, 3, 0, 0, 0, time.UTC)
	approvedRecently := time.Date(2026, 10, 10, 3, 0, 0, 0, time.UTC)

	cases := []struct {
		name       string
		in         FollowUpInput
		excluded   bool
		bucket     string
		reason     string
		overdueDay int
	}{
		{name: "missing ATR row", in: FollowUpInput{HasATR: false}, bucket: FollowUpBucketNoAction, reason: NoActionReasonNoPIC},
		{name: "missing ATR row ignores other fields", in: FollowUpInput{HasATR: false, Status: ATRStatusCompleted, PICUserID: "u1"}, bucket: FollowUpBucketNoAction, reason: NoActionReasonNoPIC},
		{name: "cancelled excluded", in: FollowUpInput{HasATR: true, Status: ATRStatusCancelled, DueDate: fuDate(2026, 1, 1)}, excluded: true},
		{name: "canceled spelling excluded", in: FollowUpInput{HasATR: true, Status: "canceled"}, excluded: true},
		{name: "completed past due is closed", in: FollowUpInput{HasATR: true, Status: "completed", DueDate: fuDate(2026, 1, 1)}, bucket: FollowUpBucketClosed},
		{name: "planned overdue wins over no action", in: FollowUpInput{HasATR: true, Status: ATRStatusPlanned, DueDate: fuDate(2026, 10, 1)}, bucket: FollowUpBucketOverdue, overdueDay: 13},
		{name: "in progress overdue", in: FollowUpInput{HasATR: true, Status: ATRStatusInProgress, DueDate: fuDate(2026, 10, 13), PICUserID: "u"}, bucket: FollowUpBucketOverdue, overdueDay: 1},
		{name: "pending review overdue", in: FollowUpInput{HasATR: true, Status: ATRStatusPendingReview, DueDate: fuDate(2026, 10, 12)}, bucket: FollowUpBucketOverdue, overdueDay: 2},
		{name: "due today not overdue", in: FollowUpInput{HasATR: true, Status: ATRStatusInProgress, DueDate: fuDate(2026, 10, 14)}, bucket: FollowUpBucketInProgress},
		{name: "pending review", in: FollowUpInput{HasATR: true, Status: ATRStatusPendingReview}, bucket: FollowUpBucketInProgress},
		{name: "unknown legacy status is in progress", in: FollowUpInput{HasATR: true, Status: "Waiting"}, bucket: FollowUpBucketInProgress},
		{name: "unknown legacy status overdue", in: FollowUpInput{HasATR: true, Status: "Waiting", DueDate: fuDate(2026, 9, 30)}, bucket: FollowUpBucketOverdue, overdueDay: 14},
		{name: "planned no PIC", in: FollowUpInput{HasATR: true, Status: ATRStatusPlanned, LHAApprovedAt: &approvedLongAgo}, bucket: FollowUpBucketNoAction, reason: NoActionReasonNoPIC},
		{name: "blank status is planned", in: FollowUpInput{HasATR: true, Status: " "}, bucket: FollowUpBucketNoAction, reason: NoActionReasonNoPIC},
		{name: "planned PIC name only counts as assigned", in: FollowUpInput{HasATR: true, Status: ATRStatusPlanned, PICName: "Budi"}, bucket: FollowUpBucketNoAction, reason: NoActionReasonNoPlan},
		{name: "planned PIC recent approval", in: FollowUpInput{HasATR: true, Status: ATRStatusPlanned, PICUserID: "u", LHAApprovedAt: &approvedRecently, DueDate: fuDate(2026, 12, 1)}, bucket: FollowUpBucketNoAction, reason: NoActionReasonNoPlan},
		{name: "planned PIC old approval", in: FollowUpInput{HasATR: true, Status: ATRStatusPlanned, PICUserID: "u", LHAApprovedAt: &approvedLongAgo}, bucket: FollowUpBucketNoAction, reason: NoActionReasonNoPlanEscalated},
		{name: "planned PIC unknown approval", in: FollowUpInput{HasATR: true, Status: ATRStatusPlanned, PICUserID: "u"}, bucket: FollowUpBucketNoAction, reason: NoActionReasonNoPlan},
	}
	for _, c := range cases {
		got := ClassifyFollowUp(c.in, asOf)
		if got.Excluded != c.excluded || got.Bucket != c.bucket || got.NoActionReason != c.reason || got.OverdueDays != c.overdueDay {
			t.Errorf("%s: got %+v, want excluded=%v bucket=%q reason=%q days=%d", c.name, got, c.excluded, c.bucket, c.reason, c.overdueDay)
		}
	}
}

// The as-of date is the Asia/Jakarta calendar date: 16:59:59Z on the due date
// is still the due date in Jakarta, 17:00Z is the next day.
func TestClassifyFollowUpJakartaBoundary(t *testing.T) {
	in := FollowUpInput{HasATR: true, Status: ATRStatusInProgress, DueDate: fuDate(2026, 10, 14)}

	before := time.Date(2026, 10, 14, 16, 59, 59, 0, time.UTC) // 23:59:59 WIB 14 Oct
	if got := ClassifyFollowUp(in, before); got.Bucket != FollowUpBucketInProgress {
		t.Errorf("23:59:59 WIB on the due date: got %+v, want IN_PROGRESS", got)
	}
	after := time.Date(2026, 10, 14, 17, 0, 0, 0, time.UTC) // 00:00 WIB 15 Oct
	if got := ClassifyFollowUp(in, after); got.Bucket != FollowUpBucketOverdue || got.OverdueDays != 1 {
		t.Errorf("00:00 WIB the day after: got %+v, want OVERDUE 1 day", got)
	}
}

// NO_PLAN_14D starts on the 14th Jakarta calendar day after the approval
func TestClassifyFollowUpNoPlan14DBoundary(t *testing.T) {
	// Approved 2026-10-01 23:30 WIB (16:30Z): Jakarta date 1 Oct
	approved := time.Date(2026, 10, 1, 16, 30, 0, 0, time.UTC)
	in := FollowUpInput{HasATR: true, Status: ATRStatusPlanned, PICUserID: "u", LHAApprovedAt: &approved}

	cases := []struct {
		asOf time.Time
		want string
	}{
		{time.Date(2026, 10, 14, 16, 59, 59, 0, time.UTC), NoActionReasonNoPlan},        // 14 Oct 23:59 WIB: 13 days
		{time.Date(2026, 10, 14, 17, 0, 0, 0, time.UTC), NoActionReasonNoPlanEscalated}, // 15 Oct 00:00 WIB: 14 days
		{time.Date(2026, 10, 20, 0, 0, 0, 0, time.UTC), NoActionReasonNoPlanEscalated},
	}
	for _, c := range cases {
		got := ClassifyFollowUp(in, c.asOf)
		if got.Bucket != FollowUpBucketNoAction || got.NoActionReason != c.want {
			t.Errorf("asOf %s: got %+v, want NO_ACTION/%s", c.asOf, got, c.want)
		}
	}
}
