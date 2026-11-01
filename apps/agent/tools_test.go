package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestEmployeeCannotSeeHRTools(t *testing.T) {
	tools := llmTools(principal{Role: "employee"})
	for _, item := range tools {
		fn := item["function"].(map[string]any)
		name := fn["name"].(string)
		if name == "create_employee" || name == "review_leave_request" || name == "get_employee_directory" || name == "update_employee_department_or_location" {
			t.Fatalf("employee exposed privileged tool %s", name)
		}
	}
}

func TestLeaveSubmissionIsOnlyStaged(t *testing.T) {
	plan, err := planTool(principal{Role: "employee"}, "submit_leave_request", map[string]any{
		"leave_type": "annual", "start_date": "2026-11-16", "end_date": "2026-11-20", "reason": "Holiday",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !plan.Mutation || !strings.Contains(plan.Preview, "Remains subject") {
		t.Fatalf("unexpected mutation plan: %#v", plan)
	}
	if plan.Method != "POST" || plan.Path != "/api/v1/leave/requests" {
		t.Fatalf("unexpected endpoint: %#v", plan)
	}
	var payload map[string]any
	if err := json.Unmarshal(plan.Body, &payload); err != nil {
		t.Fatal(err)
	}
	if payload["leave_type"] != "annual" || payload["partial_day"] != "full" {
		t.Fatalf("incorrect API request: %#v", payload)
	}
}

func TestSensitiveToolsStillRoleChecked(t *testing.T) {
	_, err := planTool(principal{Role: "employee"}, "review_leave_request", map[string]any{"request_id": "xyz", "decision": "approved"})
	if err == nil {
		t.Fatal("employee should not be allowed to stage leave approval")
	}
	_, err = planTool(principal{Role: "manager"}, "create_employee", map[string]any{"first_name": "A", "last_name": "B", "work_email": "a@example.com"})
	if err == nil {
		t.Fatal("manager should not create employees")
	}
}

func TestInvalidWorkflowInputsRejected(t *testing.T) {
	cases := []struct {
		name string
		args map[string]any
	}{
		{"submit_leave_request", map[string]any{"leave_type": "annual", "start_date": "2026-11-25", "end_date": "2026-11-12"}},
		{"submit_attendance_correction", map[string]any{"work_date": "2026-11-03", "reason": "Missing"}},
		{"get_attendance", map[string]any{"date": "not-a-date"}},
		{"get_attendance_summary", map[string]any{"from": "2026-01-01", "to": "2028-02-01"}},
		{"submit_leave_request", map[string]any{"leave_type": "annual", "start_date": "2026-11-12", "end_date": "2026-11-13", "organization_id": "other-company"}},
	}
	for _, c := range cases {
		if _, err := planTool(principal{Role: "employee"}, c.name, c.args); err == nil {
			t.Errorf("expected rejection for %s: %#v", c.name, c.args)
		}
	}
}

func TestHRMustChooseEmployeeForBalance(t *testing.T) {
	plan, err := planTool(principal{Role: "hr"}, "get_leave_balance", map[string]any{"employee_id": "emp_001", "year": json.Number("2026")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(plan.Path, "employee_id=emp_001") || !strings.Contains(plan.Path, "year=2026") {
		t.Fatal(plan.Path)
	}
}

func TestToolSchemasContainRequiredArrays(t *testing.T) {
	for _, tool := range llmTools(principal{Role: "hr"}) {
		fn := tool["function"].(map[string]any)
		schema := fn["parameters"].(map[string]any)
		if _, ok := schema["required"].([]string); !ok {
			t.Fatalf("tool %v has null or invalid required array", fn["name"])
		}
	}
}

func TestAttendanceAlertsAreBasedOnRecordedMetrics(t *testing.T) {
	input := []byte(`{"date":"2026-10-29","data":[{"id":"a1","employee":"Ava","work_date":"2026-10-29","late_minutes":22,"overtime_minutes":15,"early_leave_minutes":0,"check_in":"2026-10-29T09:22:00Z","check_out":null}]}`)
	alerts := attendanceAlerts(input)
	if !strings.Contains(alerts, "late: 22 minutes") || !strings.Contains(alerts, "overtime: 15 minutes") || !strings.Contains(alerts, "missing checkout") {
		t.Fatalf("expected derived flags: %s", alerts)
	}
}
