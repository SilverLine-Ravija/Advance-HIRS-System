package httpapi

import (
	"net/http"

	"github.com/savisaluwadana/Advance-HIRS-System/apps/api/internal/auth"
	"github.com/savisaluwadana/Advance-HIRS-System/apps/api/internal/store"
)

// NewHandlerV5 preserves the access, leave and audit surfaces from V4 and
// adds schedule-aware attendance, correction approvals and period summaries.
func NewHandlerV5(dataStore *store.Store, authManager *auth.Manager, allowedOrigin string) http.Handler {
	server := &Server{store: dataStore, auth: authManager, allowedOrigin: allowedOrigin}
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("POST /api/v1/auth/login", server.login)
	mux.HandleFunc("GET /api/v1/auth/me", server.authorize(server.me, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/auth/logout", server.authorize(server.atomic((*Server).logout), "admin", "hr", "manager", "employee"))

	mux.HandleFunc("POST /api/v1/access/invitations/inspect", server.inspectAccessInvitation)
	mux.HandleFunc("POST /api/v1/access/invitations/accept", server.atomic((*Server).acceptAccessInvitation))
	mux.HandleFunc("GET /api/v1/access/invitations", server.authorize(server.listAccessInvitations, "admin", "hr"))
	mux.HandleFunc("POST /api/v1/access/invitations", server.authorize(server.atomic((*Server).createAccessInvitation), "admin", "hr"))
	mux.HandleFunc("POST /api/v1/access/invitations/{id}/revoke", server.authorize(server.atomic((*Server).revokeAccessInvitation), "admin", "hr"))

	mux.HandleFunc("GET /api/v1/dashboard", server.authorize(server.dashboard, "admin", "hr", "manager"))
	mux.HandleFunc("GET /api/v1/employees", server.authorize(server.listEmployees, "admin", "hr", "manager"))
	mux.HandleFunc("POST /api/v1/employees", server.authorize(server.atomic((*Server).createEmployee), "admin", "hr"))
	mux.HandleFunc("PATCH /api/v1/employees/{id}", server.authorize(server.atomic((*Server).patchEmployee), "admin", "hr"))
	mux.HandleFunc("DELETE /api/v1/employees/{id}", server.authorize(server.atomic((*Server).archiveEmployee), "admin", "hr"))

	mux.HandleFunc("GET /api/v1/leave/policies", server.authorize(server.leavePoliciesV2, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("PUT /api/v1/leave/policies", server.authorize(server.atomic((*Server).upsertLeavePolicyV2), "admin", "hr"))
	mux.HandleFunc("POST /api/v1/leave/policies/{id}/assignments", server.authorize(server.atomic((*Server).assignLeavePolicyV2), "admin", "hr"))
	mux.HandleFunc("GET /api/v1/leave/balances", server.authorize(server.leaveBalancesV2, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("GET /api/v1/leave/balances/ledger", server.authorize(server.leaveBalanceLedgerV2, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/leave/balances/adjustments", server.authorize(server.atomic((*Server).adjustLeaveBalanceV2), "admin", "hr"))
	mux.HandleFunc("GET /api/v1/leave/holidays", server.authorize(server.holidaysV2, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/leave/holidays", server.authorize(server.atomic((*Server).createHolidayV2), "admin", "hr"))
	mux.HandleFunc("DELETE /api/v1/leave/holidays/{id}", server.authorize(server.atomic((*Server).deleteHolidayV2), "admin", "hr"))
	mux.HandleFunc("GET /api/v1/leave/requests", server.authorize(server.workflowLeaveRequests, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/leave/requests/preview", server.authorize(server.previewLeaveDays, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/leave/requests", server.authorize(server.atomic((*Server).createLeaveRequestV2), "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/leave/requests/{id}/decision", server.authorize(server.atomic((*Server).decideLeaveRequestV2), "admin", "hr", "manager"))
	mux.HandleFunc("POST /api/v1/leave/requests/{id}/cancel", server.authorize(server.atomic((*Server).cancelLeaveRequestV2), "admin", "hr", "manager", "employee"))

	mux.HandleFunc("GET /api/v1/attendance/schedules", server.authorize(server.workSchedulesV2, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("PUT /api/v1/attendance/schedules", server.authorize(server.atomic((*Server).upsertWorkScheduleV2), "admin", "hr"))
	mux.HandleFunc("POST /api/v1/attendance/schedules/{id}/assignments", server.authorize(server.atomic((*Server).assignWorkScheduleSafeV2), "admin", "hr"))
	mux.HandleFunc("GET /api/v1/attendance", server.authorize(server.attendanceV2, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/attendance/check-in", server.authorize(server.atomic((*Server).checkInV2), "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/attendance/check-out", server.authorize(server.atomic((*Server).checkOutTodayV2), "admin", "hr", "manager", "employee"))
	mux.HandleFunc("GET /api/v1/attendance/corrections", server.authorize(server.attendanceCorrectionsV2, "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/attendance/corrections", server.authorize(server.atomic((*Server).createAttendanceCorrectionV2), "admin", "hr", "manager", "employee"))
	mux.HandleFunc("POST /api/v1/attendance/corrections/{id}/decision", server.authorize(server.atomic((*Server).decideAttendanceCorrectionV2), "admin", "hr", "manager"))
	mux.HandleFunc("GET /api/v1/attendance/timesheet", server.authorize(server.attendanceTimesheetV2, "admin", "hr", "manager", "employee"))

	mux.HandleFunc("GET /api/v1/audit", server.authorize(server.auditEvents, "admin", "hr"))
	return server.withMiddleware(mux)
}
