package routes

import (
	ctrlActivity "audit-service/controllers/audit_activity"
	ctrlAssignment "audit-service/controllers/audit_assignment"
	ctrlCharter "audit-service/controllers/audit_charter"
	ctrlMandate "audit-service/controllers/audit_mandate"
	ctrlMedia "audit-service/controllers/media"
)

// RouteRegistry holds all controller references
type RouteRegistry struct {
	AuditActivity   ctrlActivity.AuditActivityControllerInterface
	AuditCharter    ctrlCharter.AuditCharterControllerInterface
	AuditMandate    ctrlMandate.AuditMandateControllerInterface
	AuditAssignment ctrlAssignment.AuditAssignmentControllerInterface
	Media           *ctrlMedia.MediaController
}

// NewRouteRegistry creates a new route registry
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{}
}

// SetAuditCharterController sets the audit charter controller
func (r *RouteRegistry) SetAuditCharterController(ctrl ctrlCharter.AuditCharterControllerInterface) {
	r.AuditCharter = ctrl
}

// SetAuditMandateController sets the audit mandate controller
func (r *RouteRegistry) SetAuditMandateController(ctrl ctrlMandate.AuditMandateControllerInterface) {
	r.AuditMandate = ctrl
}

// SetAuditAssignmentController sets the audit assignment controller
func (r *RouteRegistry) SetAuditAssignmentController(ctrl ctrlAssignment.AuditAssignmentControllerInterface) {
	r.AuditAssignment = ctrl
}

// SetAuditActivityController sets the audit activity controller
func (r *RouteRegistry) SetAuditActivityController(ctrl ctrlActivity.AuditActivityControllerInterface) {
	r.AuditActivity = ctrl
}

// SetMediaController sets the media controller
func (r *RouteRegistry) SetMediaController(ctrl *ctrlMedia.MediaController) {
	r.Media = ctrl
}
