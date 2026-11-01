package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type toolSpec struct {
	Name        string
	Description string
	Properties  map[string]any
	Required    []string
	Mutates     bool
	Elevated    bool
	HRonly      bool
}

func field(description string) map[string]any {
	return map[string]any{"type": "string", "description": description}
}
func choice(description string, options ...string) map[string]any {
	return map[string]any{"type": "string", "description": description, "enum": options}
}

var toolSpecs = []toolSpec{
	{Name: "get_leave_balance", Description: "Retrieve the authenticated employee's or a permitted report's tracked leave balances; HR must provide employee_id.", Properties: map[string]any{"employee_id": field("Employee ID, required for HR/admin"), "year": map[string]any{"type": "integer", "description": "Year, default current year"}}},
	{Name: "get_leave_policies", Description: "Retrieve organization leave policies, entitlements, carry-over rules and approval requirements.", Properties: map[string]any{}},
	{Name: "get_company_holidays", Description: "Retrieve configured company holidays.", Properties: map[string]any{}},
	{Name: "list_leave_requests", Description: "List leave requests permitted for the caller. Managers see only self and direct reports.", Properties: map[string]any{}},
	{Name: "get_attendance", Description: "Retrieve schedule-aware attendance records for a date within the caller's scope.", Properties: map[string]any{"date": field("Date YYYY-MM-DD, defaults to today")}},
	{Name: "get_attendance_corrections", Description: "Retrieve pending and historical attendance correction requests visible to this caller.", Properties: map[string]any{}},
	{Name: "get_attendance_alerts", Description: "Analyze current authorized attendance data for recorded late/early/overtime minutes and previous-day missing check-outs. Does not change records.", Properties: map[string]any{"date": field("Work date YYYY-MM-DD, default today")}},
	{Name: "get_work_schedules", Description: "Get configured work schedules and timekeeping rules.", Properties: map[string]any{}},
	{Name: "get_attendance_summary", Description: "Retrieve a payroll-period attendance summary; HR/admin must specify employee_id. This is NOT payroll or salary calculation.", Properties: map[string]any{"from": field("Start date YYYY-MM-DD"), "to": field("End date YYYY-MM-DD"), "employee_id": field("Employee ID, required for HR/admin")}, Required: []string{"from", "to"}},
	{Name: "get_employee_directory", Description: "List employee directory records within caller's authorization scope.", Properties: map[string]any{}, Elevated: true},
	{Name: "get_workforce_dashboard", Description: "Return existing employee, leave and attendance metrics. Do not infer additional KPIs.", Properties: map[string]any{}, Elevated: true},
	{Name: "submit_leave_request", Description: "PREPARE a leave request for the caller or an HR-authorized employee. Requires separate human confirmation before API submission. Not an approval.", Properties: map[string]any{"employee_id": field("Required for HR/admin, omit for self-service"), "leave_type": choice("Type", "annual", "sick", "unpaid", "remote", "parental", "other"), "start_date": field("YYYY-MM-DD"), "end_date": field("YYYY-MM-DD"), "partial_day": choice("full or half-day", "full", "first_half", "second_half"), "reason": field("Reason; never invent one")}, Required: []string{"leave_type", "start_date", "end_date"}, Mutates: true},
	{Name: "submit_attendance_correction", Description: "PREPARE an attendance correction for later confirmation. NEVER fabricate timestamps or reasons.", Properties: map[string]any{"employee_id": field("Required for HR/admin"), "work_date": field("YYYY-MM-DD"), "requested_check_in": field("RFC3339 timestamp with timezone, if supplied"), "requested_check_out": field("RFC3339 timestamp with timezone, if supplied"), "work_mode": choice("Work mode", "office", "remote", "hybrid", "field"), "reason": field("Employee-provided reason")}, Required: []string{"work_date", "reason"}, Mutates: true},
	{Name: "review_leave_request", Description: "PREPARE a manager/HR decision on a pending leave request. Only a human can confirm a final decision.", Properties: map[string]any{"request_id": field("Existing leave request ID"), "decision": choice("Decision", "approved", "rejected"), "note": field("Manager's note")}, Required: []string{"request_id", "decision"}, Mutates: true, Elevated: true},
	{Name: "review_attendance_correction", Description: "PREPARE a manager/HR attendance correction decision, subject to human confirmation.", Properties: map[string]any{"request_id": field("Existing correction ID"), "decision": choice("Decision", "approved", "rejected"), "note": field("Manager's note")}, Required: []string{"request_id", "decision"}, Mutates: true, Elevated: true},
	{Name: "update_employee_department_or_location", Description: "PREPARE a narrowly scoped department/location change for HR approval. Never change pay, status, roles or permissions.", Properties: map[string]any{"employee_id": field("Employee ID"), "department": field("New department, if changing"), "location": field("New location, if changing")}, Required: []string{"employee_id"}, Mutates: true, HRonly: true},
	{Name: "create_employee", Description: "PREPARE creation of an employee record, subject to HR confirmation and backend validation. Does not send invitations.", Properties: map[string]any{"first_name": field("First name"), "last_name": field("Last name"), "work_email": field("Verified work email"), "department": field("Department"), "location": field("Location"), "employee_number": field("Employee number")}, Required: []string{"first_name", "last_name", "work_email"}, Mutates: true, HRonly: true},
}

func allowed(p principal, t toolSpec) bool {
	if t.HRonly {
		return p.Role == "hr" || p.Role == "admin"
	}
	if t.Elevated {
		return p.Role == "hr" || p.Role == "admin" || p.Role == "manager"
	}
	return true
}

func llmTools(p principal) []map[string]any {
	out := make([]map[string]any, 0, len(toolSpecs))
	for _, t := range toolSpecs {
		if !allowed(p, t) {
			continue
		}
		required := t.Required
		if required == nil {
			required = []string{}
		}
		out = append(out, map[string]any{"type": "function", "function": map[string]any{
			"name": t.Name, "description": t.Description, "parameters": map[string]any{"type": "object", "properties": t.Properties, "required": required, "additionalProperties": false},
		}})
	}
	return out
}

type plannedTool struct {
	Name     string
	Method   string
	Path     string
	Body     []byte
	Preview  string
	Mutation bool
}

func getString(args map[string]any, key string) (string, error) {
	v, ok := args[key]
	if !ok || v == nil {
		return "", nil
	}
	value, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%s must be text", key)
	}
	value = strings.TrimSpace(value)
	if len(value) > 512 {
		return "", fmt.Errorf("%s is too long", key)
	}
	return value, nil
}
func requiredString(args map[string]any, key string) (string, error) {
	v, err := getString(args, key)
	if err != nil {
		return "", err
	}
	if v == "" {
		return "", fmt.Errorf("%s is required", key)
	}
	return v, nil
}
func parseDate(s string) error {
	if _, err := time.Parse("2006-01-02", s); err != nil {
		return errors.New("date must be YYYY-MM-DD")
	}
	return nil
}
func validChoice(s string, options ...string) bool {
	for _, o := range options {
		if s == o {
			return true
		}
	}
	return false
}

var idPattern = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

func validateID(s string) error {
	if !idPattern.MatchString(s) {
		return errors.New("invalid record identifier")
	}
	return nil
}
func marshal(v any) []byte { data, _ := json.Marshal(v); return data }

func planTool(p principal, name string, args map[string]any) (plannedTool, error) {
	plan := plannedTool{Name: name, Method: "GET"}
	var spec *toolSpec
	for i := range toolSpecs {
		if toolSpecs[i].Name == name {
			spec = &toolSpecs[i]
			break
		}
	}
	if spec == nil || !allowed(p, *spec) {
		return plan, errors.New("tool unavailable for this role")
	}
	for key := range args {
		if _, ok := spec.Properties[key]; !ok {
			return plan, fmt.Errorf("unsupported argument: %s", key)
		}
	}
	str := func(key string) string { v, _ := getString(args, key); return v }
	for _, key := range spec.Required {
		if _, err := requiredString(args, key); err != nil {
			return plan, err
		}
	}
	for key := range args {
		if _, err := getString(args, key); err != nil && key != "year" {
			return plan, err
		}
	}
	query := url.Values{}
	plan.Mutation = spec.Mutates
	switch name {
	case "get_leave_balance":
		if employee := str("employee_id"); employee != "" {
			if err := validateID(employee); err != nil {
				return plan, err
			}
			query.Set("employee_id", employee)
		}
		if year, ok := args["year"]; ok {
			f, ok := year.(json.Number)
			if !ok {
				return plan, errors.New("year must be integer")
			}
			n, err := strconv.Atoi(f.String())
			if err != nil || n < 2000 || n > 2200 {
				return plan, errors.New("year out of range")
			}
			query.Set("year", strconv.Itoa(n))
		}
		plan.Path = "/api/v1/leave/balances"
	case "get_leave_policies":
		plan.Path = "/api/v1/leave/policies"
	case "get_company_holidays":
		plan.Path = "/api/v1/leave/holidays"
	case "list_leave_requests":
		plan.Path = "/api/v1/leave/requests"
	case "get_attendance":
		if date := str("date"); date != "" {
			if err := parseDate(date); err != nil {
				return plan, err
			}
			query.Set("date", date)
		}
		plan.Path = "/api/v1/attendance"
	case "get_attendance_corrections":
		plan.Path = "/api/v1/attendance/corrections"
	case "get_attendance_alerts":
		if date := str("date"); date != "" {
			if err := parseDate(date); err != nil {
				return plan, err
			}
			query.Set("date", date)
		}
		plan.Path = "/api/v1/attendance"
	case "get_work_schedules":
		plan.Path = "/api/v1/attendance/schedules"
	case "get_attendance_summary":
		from, to := str("from"), str("to")
		if err := parseDate(from); err != nil {
			return plan, err
		}
		if err := parseDate(to); err != nil {
			return plan, err
		}
		a, _ := time.Parse("2006-01-02", from)
		b, _ := time.Parse("2006-01-02", to)
		if b.Before(a) || b.Sub(a) > 366*24*time.Hour {
			return plan, errors.New("date range must be at most 366 days")
		}
		query.Set("from", from)
		query.Set("to", to)
		if emp := str("employee_id"); emp != "" {
			if err := validateID(emp); err != nil {
				return plan, err
			}
			query.Set("employee_id", emp)
		}
		plan.Path = "/api/v1/attendance/timesheet"
	case "get_employee_directory":
		plan.Path = "/api/v1/employees"
	case "get_workforce_dashboard":
		plan.Path = "/api/v1/dashboard"
	case "submit_leave_request":
		typeName := str("leave_type")
		if !validChoice(typeName, "annual", "sick", "unpaid", "remote", "parental", "other") {
			return plan, errors.New("invalid leave type")
		}
		start, end := str("start_date"), str("end_date")
		if err := parseDate(start); err != nil {
			return plan, err
		}
		if err := parseDate(end); err != nil {
			return plan, err
		}
		a, _ := time.Parse("2006-01-02", start)
		b, _ := time.Parse("2006-01-02", end)
		if b.Before(a) || b.Sub(a) > 366*24*time.Hour {
			return plan, errors.New("invalid leave date range")
		}
		partial := str("partial_day")
		if partial == "" {
			partial = "full"
		}
		if !validChoice(partial, "full", "first_half", "second_half") {
			return plan, errors.New("invalid partial day")
		}
		if partial != "full" && !a.Equal(b) {
			return plan, errors.New("half day must be a single date")
		}
		employee := str("employee_id")
		if employee != "" {
			if err := validateID(employee); err != nil {
				return plan, err
			}
		}
		if (p.Role == "hr" || p.Role == "admin") && employee == "" {
			return plan, errors.New("HR must specify employee_id")
		}
		plan.Method = "POST"
		plan.Path = "/api/v1/leave/requests"
		plan.Body = marshal(map[string]any{"employee_id": employee, "leave_type": typeName, "start_date": start, "end_date": end, "partial_day": partial, "reason": str("reason")})
		plan.Preview = fmt.Sprintf("Submit %s leave from %s to %s (%s); employee: %s. Remains subject to the existing approval workflow.", typeName, start, end, partial, employeeLabel(employee))
	case "submit_attendance_correction":
		date := str("work_date")
		if err := parseDate(date); err != nil {
			return plan, err
		}
		in, out := str("requested_check_in"), str("requested_check_out")
		if in == "" && out == "" {
			return plan, errors.New("at least one corrected timestamp is required")
		}
		if in != "" {
			if _, err := time.Parse(time.RFC3339, in); err != nil {
				return plan, errors.New("check-in must include RFC3339 timezone")
			}
		}
		if out != "" {
			if _, err := time.Parse(time.RFC3339, out); err != nil {
				return plan, errors.New("check-out must include RFC3339 timezone")
			}
		}
		mode := str("work_mode")
		if mode == "" {
			mode = "office"
		}
		if !validChoice(mode, "office", "remote", "hybrid", "field") {
			return plan, errors.New("invalid work mode")
		}
		reason := str("reason")
		if reason == "" {
			return plan, errors.New("reason required")
		}
		employee := str("employee_id")
		if employee != "" {
			if err := validateID(employee); err != nil {
				return plan, err
			}
		}
		if (p.Role == "hr" || p.Role == "admin") && employee == "" {
			return plan, errors.New("HR must specify employee_id")
		}
		plan.Method = "POST"
		plan.Path = "/api/v1/attendance/corrections"
		plan.Body = marshal(map[string]any{"employee_id": employee, "work_date": date, "requested_check_in": in, "requested_check_out": out, "work_mode": mode, "reason": reason})
		plan.Preview = fmt.Sprintf("Request attendance correction for %s on %s: check-in %s, check-out %s. Approval remains with authorized reviewers.", employeeLabel(employee), date, label(in), label(out))
	case "review_leave_request", "review_attendance_correction":
		id := str("request_id")
		if err := validateID(id); err != nil {
			return plan, err
		}
		decision := str("decision")
		if !validChoice(decision, "approved", "rejected") {
			return plan, errors.New("invalid decision")
		}
		plan.Method = "POST"
		plan.Body = marshal(map[string]any{"decision": decision, "note": str("note")})
		if name == "review_leave_request" {
			plan.Path = "/api/v1/leave/requests/" + id + "/decision"
		} else {
			plan.Path = "/api/v1/attendance/corrections/" + id + "/decision"
		}
		plan.Preview = fmt.Sprintf("%s %s request %s. This is a final decision requiring your confirmation.", strings.ToUpper(decision[:1])+decision[1:], name, id)
	case "update_employee_department_or_location":
		id := str("employee_id")
		if err := validateID(id); err != nil {
			return plan, err
		}
		patch := map[string]string{}
		if d := str("department"); d != "" {
			patch["department"] = d
		}
		if loc := str("location"); loc != "" {
			patch["location"] = loc
		}
		if len(patch) == 0 {
			return plan, errors.New("department or location required")
		}
		plan.Method = "PATCH"
		plan.Path = "/api/v1/employees/" + id
		plan.Body = marshal(patch)
		plan.Preview = fmt.Sprintf("Update employee %s: %s (HR confirmation required).", id, string(plan.Body))
	case "create_employee":
		first, last, email := str("first_name"), str("last_name"), str("work_email")
		if !strings.Contains(email, "@") || len(email) > 254 {
			return plan, errors.New("valid work email required")
		}
		if len(first) > 100 || len(last) > 100 {
			return plan, errors.New("name too long")
		}
		payload := map[string]string{"first_name": first, "last_name": last, "work_email": email, "department": str("department"), "location": str("location"), "employee_number": str("employee_number")}
		plan.Method = "POST"
		plan.Path = "/api/v1/employees"
		plan.Body = marshal(payload)
		plan.Preview = fmt.Sprintf("Create employee %s %s (%s). This does not invite or grant application access.", first, last, email)
	default:
		return plan, errors.New("unknown tool")
	}
	if len(query) > 0 {
		plan.Path += "?" + query.Encode()
	}
	return plan, nil
}

func label(s string) string {
	if s == "" {
		return "unchanged"
	}
	return s
}
func employeeLabel(s string) string {
	if s == "" {
		return "myself"
	}
	return s
}

func (s *server) readTool(ctx context.Context, token string, plan plannedTool) string {
	body, status, err := s.apiCall(ctx, token, plan.Method, plan.Path, plan.Body)
	if err != nil {
		return "HRIS API temporarily unavailable: " + err.Error()
	}
	if status >= 400 {
		return fmt.Sprintf("HRIS API rejected this operation with HTTP %d: %s", status, limitText(string(body), 2000))
	}
	if plan.Name == "get_attendance_alerts" {
		return attendanceAlerts(body)
	}
	return limitText(string(body), 15000)
}
func limitText(v string, max int) string {
	if len(v) > max {
		return v[:max] + " [truncated]"
	}
	return v
}

// attendanceAlerts reports only facts already calculated and stored by the HRIS.
// It does not infer health, intent, performance or payroll entitlement.
func attendanceAlerts(raw []byte) string {
	var response struct {
		Date string           `json:"date"`
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "Could not interpret HRIS attendance records"
	}
	alerts := make([]map[string]any, 0)
	counts := map[string]int{"late": 0, "early_leave": 0, "overtime": 0, "missing_checkout": 0}
	today := time.Now().UTC().Format("2006-01-02")
	for _, entry := range response.Data {
		reasons := make([]string, 0, 4)
		for _, metric := range []struct{ field, label string }{{"late_minutes", "late"}, {"early_leave_minutes", "early_leave"}, {"overtime_minutes", "overtime"}} {
			if amount, ok := entry[metric.field].(float64); ok && amount > 0 {
				counts[metric.label]++
				reasons = append(reasons, fmt.Sprintf("%s: %.0f minutes", metric.label, amount))
			}
		}
		checkIn, _ := entry["check_in"].(string)
		checkOut, _ := entry["check_out"].(string)
		workDate, _ := entry["work_date"].(string)
		if checkIn != "" && checkOut == "" && workDate != "" && workDate < today {
			counts["missing_checkout"]++
			reasons = append(reasons, "missing checkout on a previous UTC day (confirm local timezone)")
		}
		if len(reasons) > 0 {
			alerts = append(alerts, map[string]any{"id": entry["id"], "employee": entry["employee"], "work_date": workDate, "reasons": reasons})
		}
	}
	payload := map[string]any{"source": "Authorized HRIS attendance; not an employment decision", "date": response.Date, "records_checked": len(response.Data), "counts": counts, "exceptions": alerts}
	return limitText(string(marshal(payload)), 15000)
}
