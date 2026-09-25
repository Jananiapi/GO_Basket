package model

import (
	"strconv"
	"testing"
	"time"
)

func TestTimestampNowAndTime(t *testing.T) {
	before := time.Now().Add(-time.Second)
	now := Now()
	after := time.Now().Add(time.Second)

	if now == nil {
		t.Fatal("expected non-nil Timestamp from Now()")
	}
	tt := now.Time()
	if tt.Before(before) || tt.After(after) {
		t.Fatalf("Now() returned time %v outside expected window [%v, %v]", tt, before, after)
	}
}

func TestTimestampMarshalJSON(t *testing.T) {
	fixed := time.Date(2026, 9, 21, 10, 30, 0, 0, time.UTC)
	ts := Timestamp(fixed)

	b, err := ts.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	expected := strconv.FormatInt(fixed.UnixMilli(), 10)
	if string(b) != expected {
		t.Fatalf("expected %s, got %s", expected, string(b))
	}
}

func TestTimestampUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantYear  int
		wantZero  bool
		shouldErr bool
	}{
		{
			name:     "null literal",
			input:    "null",
			wantZero: true,
		},
		{
			name:     "empty string literal",
			input:    `""`,
			wantZero: true,
		},
		{
			name:     "epoch milliseconds integer",
			input:    "1790073000000",
			wantYear: 2026,
		},
		{
			name:     "RFC3339Nano string",
			input:    `"2026-09-21T10:30:00.123456789Z"`,
			wantYear: 2026,
		},
		{
			name:     "datetime space string",
			input:    `"2026-09-21 10:30:00"`,
			wantYear: 2026,
		},
		{
			name:     "date only string",
			input:    `"2026-09-21"`,
			wantYear: 2026,
		},
		{
			name:      "invalid date string format",
			input:     `"21-09-2026"`,
			shouldErr: true,
		},
		{
			name:      "malformed json",
			input:     `{not a date}`,
			shouldErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var ts Timestamp
			err := ts.UnmarshalJSON([]byte(tc.input))
			assertTimestampUnmarshal(t, tc.input, err, tc.shouldErr, tc.wantZero, tc.wantYear, ts)
		})
	}
}

func assertTimestampUnmarshal(t *testing.T, input string, err error, shouldErr, wantZero bool, wantYear int, ts Timestamp) {
	t.Helper()
	if shouldErr {
		if err == nil {
			t.Fatalf("expected error for input %q, got nil", input)
		}
		return
	}
	if err != nil {
		t.Fatalf("unexpected error for input %q: %v", input, err)
	}
	if wantZero && !ts.Time().IsZero() {
		t.Fatalf("expected zero time for input %q, got %v", input, ts.Time())
	}
	if !wantZero && ts.Time().Year() != wantYear {
		t.Fatalf("expected year %d for input %q, got %d (%v)", wantYear, input, ts.Time().Year(), ts.Time())
	}
}

func TestTimestampValueAndScan(t *testing.T) {
	fixed := time.Date(2026, 9, 21, 12, 0, 0, 0, time.UTC)
	ts := Timestamp(fixed)

	// Value()
	val, err := ts.Value()
	if err != nil {
		t.Fatalf("Value() returned error: %v", err)
	}
	tv, ok := val.(time.Time)
	if !ok || !tv.Equal(fixed) {
		t.Fatalf("Value() expected time.Time %v, got %v", fixed, val)
	}

	// Scan nil
	var s1 Timestamp
	if err := s1.Scan(nil); err != nil {
		t.Fatalf("Scan(nil) failed: %v", err)
	}
	if !s1.Time().IsZero() {
		t.Fatalf("expected zero time after Scan(nil), got %v", s1.Time())
	}

	// Scan time.Time
	var s2 Timestamp
	if err := s2.Scan(fixed); err != nil {
		t.Fatalf("Scan(time.Time) failed: %v", err)
	}
	if !s2.Time().Equal(fixed) {
		t.Fatalf("expected %v, got %v", fixed, s2.Time())
	}

	// Scan string RFC3339
	var s3 Timestamp
	if err := s3.Scan("2026-09-21T12:00:00Z"); err != nil {
		t.Fatalf("Scan(string) failed: %v", err)
	}
	if s3.Time().Year() != 2026 {
		t.Fatalf("expected year 2026, got %v", s3.Time())
	}

	// Scan []byte
	var s4 Timestamp
	if err := s4.Scan([]byte("2026-09-21T12:00:00Z")); err != nil {
		t.Fatalf("Scan([]byte) failed: %v", err)
	}
	if s4.Time().Year() != 2026 {
		t.Fatalf("expected year 2026, got %v", s4.Time())
	}

	// Scan unsupported type
	var s5 Timestamp
	if err := s5.Scan(12345); err == nil {
		t.Fatal("expected error when scanning unsupported int type, got nil")
	}

	// GormDataType()
	if ts.GormDataType() != "time" {
		t.Fatalf("expected GormDataType 'time', got %q", ts.GormDataType())
	}
}

func TestTableNamesAndEntities(t *testing.T) {
	tableTests := []struct {
		table    interface{ TableName() string }
		expected string
	}{
		{BasketNameEntity{}, "tbl_basket_order"},
		{BasketScripEntity{}, "tbl_basket_order_scrip"},
		{DeviceMappingEntity{}, "tbl_device_mapping"},
		{OrderStatusFeedEntity{}, "tbl_order_status_feed"},
		{ReasearchCallUsers{}, "tbl_researchcall_usermapping"},
		{ResearchCallStatusEntity{}, "tbl_researchcall_status"},
		{ResearchcallOrderEntity{}, "tbl_researchcall_master"},
		{ResearchcallScripEntity{}, "tbl_research_scrip"},
		{SectorReportsEntity{}, "tbl_sector_reports"},
		{ThematicExeMasterEntity{}, "tbl_user_thematic_master"},
		{ThematicExecDetails{}, "tbl_user_thematic_exec_details"},
		{UserNotification{}, "tbl_user_notification"},
		{VendorAppEntity{}, "tbl_vendor_app"},
		{ThematicMaster{}, "tbl_thematic_basket_master"},
		{ThematicScrip{}, "tbl_thematic_basket_scrips"},
		{RebalanceScrip{}, "tbl_thematic_basket_rebalance_scrips"},
		{ThematicExecution{}, "tbl_user_thematic_exec"},
		{ExecutionDetail{}, "tbl_user_thematic_exec_details"},
		{ResearchMaster{}, "tbl_researchcall_master"},
	}

	for _, tt := range tableTests {
		if got := tt.table.TableName(); got != tt.expected {
			t.Errorf("expected TableName %q, got %q", tt.expected, got)
		}
	}

	entities := Entities()
	if len(entities) == 0 {
		t.Fatal("expected non-empty Entities() slice")
	}
}
