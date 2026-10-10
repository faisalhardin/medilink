package model

import (
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestWageCadence_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		value WageCadence
		want  bool
	}{
		{"monthly", WageCadenceMonthly, true},
		{"weekly", WageCadenceWeekly, true},
		{"literal monthly", WageCadence("monthly"), true},
		{"literal weekly", WageCadence("weekly"), true},
		{"daily", WageCadenceDaily, true},
		{"literal daily", WageCadence("daily"), true},
		{"empty", WageCadence(""), false},
		{"hourly", WageCadence("hourly"), false},
		{"MONTHLY", WageCadence("MONTHLY"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Errorf("WageCadence(%q).IsValid() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestCompensationPeriodStatus_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		value CompensationPeriodStatus
		want  bool
	}{
		{"open", CompensationPeriodStatusOpen, true},
		{"draft", CompensationPeriodStatusDraft, true},
		{"finalized", CompensationPeriodStatusFinalized, true},
		{"empty", CompensationPeriodStatus(""), false},
		{"closed", CompensationPeriodStatus("closed"), false},
		{"OPEN", CompensationPeriodStatus("OPEN"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Errorf("CompensationPeriodStatus(%q).IsValid() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestCompensationPeriodStatus_CanTransition(t *testing.T) {
	tests := []struct {
		name string
		from CompensationPeriodStatus
		to   CompensationPeriodStatus
		want bool
	}{
		{"open to draft", CompensationPeriodStatusOpen, CompensationPeriodStatusDraft, true},
		{"open to finalized", CompensationPeriodStatusOpen, CompensationPeriodStatusFinalized, false},
		{"open to open", CompensationPeriodStatusOpen, CompensationPeriodStatusOpen, false},
		{"draft to draft", CompensationPeriodStatusDraft, CompensationPeriodStatusDraft, true},
		{"draft to finalized", CompensationPeriodStatusDraft, CompensationPeriodStatusFinalized, true},
		{"draft to open", CompensationPeriodStatusDraft, CompensationPeriodStatusOpen, false},
		{"finalized to draft", CompensationPeriodStatusFinalized, CompensationPeriodStatusDraft, true},
		{"finalized to finalized", CompensationPeriodStatusFinalized, CompensationPeriodStatusFinalized, false},
		{"finalized to open", CompensationPeriodStatusFinalized, CompensationPeriodStatusOpen, false},
		{"invalid from", CompensationPeriodStatus("closed"), CompensationPeriodStatusDraft, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.from.CanTransition(tt.to); got != tt.want {
				t.Errorf("CompensationPeriodStatus(%q).CanTransition(%q) = %v, want %v", tt.from, tt.to, got, tt.want)
			}
		})
	}
}

func TestCommissionType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		value CommissionType
		want  bool
	}{
		{"percent", CommissionTypePercent, true},
		{"flat", CommissionTypeFlat, true},
		{"empty", CommissionType(""), false},
		{"fixed", CommissionType("fixed"), false},
		{"PERCENT", CommissionType("PERCENT"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Errorf("CommissionType(%q).IsValid() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestContributionSourceType_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		value ContributionSourceType
		want  bool
	}{
		{"procedure", ContributionSourceTypeProcedure, true},
		{"diagnosis", ContributionSourceTypeDiagnosis, true},
		{"anamnesa", ContributionSourceTypeAnamnesa, true},
		{"journey", ContributionSourceTypeJourney, true},
		{"manual", ContributionSourceTypeManual, true},
		{"empty", ContributionSourceType(""), false},
		{"other", ContributionSourceType("other"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.value.IsValid(); got != tt.want {
				t.Errorf("ContributionSourceType(%q).IsValid() = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestTrxCompensationPeriod_ToResponse(t *testing.T) {
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	drafted := time.Date(2026, 9, 1, 10, 0, 0, 0, time.UTC)
	row := TrxCompensationPeriod{
		UUID:            "period-1",
		Label:           "Aug 2026",
		PeriodStart:     start,
		PeriodEnd:       end,
		Status:          CompensationPeriodStatusDraft,
		TotalCommission: sql.NullInt64{Int64: 850000, Valid: true},
		TotalPayout:     sql.NullInt64{Int64: 850000, Valid: true},
		StaffCount:      sql.NullInt64{Int64: 3, Valid: true},
		VisitCount:      sql.NullInt64{Int64: 12, Valid: true},
		DraftedAt:       sql.NullTime{Time: drafted, Valid: true},
	}
	got := row.ToResponse(1)
	if got.UUID != "period-1" || got.PeriodStart != "2026-08-01" || got.PeriodEnd != "2026-08-31" {
		t.Fatalf("unexpected dates/uuid: %+v", got)
	}
	if got.TotalWage != 0 || got.TotalCommission != 850000 || got.NoContributorCount != 1 {
		t.Fatalf("unexpected totals: %+v", got)
	}
	if !got.DraftedAt.Valid || got.FinalizedAt.Valid {
		t.Fatalf("unexpected audit times: drafted=%v finalized=%v", got.DraftedAt, got.FinalizedAt)
	}
}

func TestMstStaffWage_ToResponse(t *testing.T) {
	from := time.Date(2026, 3, 1, 15, 0, 0, 0, time.FixedZone("WIB", 7*3600))
	row := MstStaffWage{
		ID:            9,
		StaffID:       "staff-1",
		WageAmount:    500000,
		WageCadence:   WageCadenceDaily,
		IsActive:      true,
		EffectiveFrom: from,
	}
	got := row.ToResponse()
	if got.ID != 9 || got.StaffID != "staff-1" || got.WageAmount != 500000 || got.WageCadence != WageCadenceDaily || !got.IsActive {
		t.Fatalf("unexpected wage response: %+v", got)
	}
	if got.EffectiveFrom != "2026-03-01" || got.EffectiveTo.Valid {
		t.Fatalf("unexpected dates: from=%s to=%+v", got.EffectiveFrom, got.EffectiveTo)
	}

	row.EffectiveTo = sql.NullTime{Time: time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC), Valid: true}
	got = row.ToResponse()
	if !got.EffectiveTo.Valid || got.EffectiveTo.String != "2026-03-31" {
		t.Fatalf("effective_to = %+v", got.EffectiveTo)
	}
}

func TestCompensationTableNames(t *testing.T) {
	if got := (MstStaffWage{}).TableName(); got != MstStaffWageTableName {
		t.Errorf("MstStaffWage.TableName() = %q, want %q", got, MstStaffWageTableName)
	}
	if got := (TrxWagePeriodSnapshot{}).TableName(); got != TrxWagePeriodSnapshotTableName {
		t.Errorf("TrxWagePeriodSnapshot.TableName() = %q, want %q", got, TrxWagePeriodSnapshotTableName)
	}
	if got := (TrxCompensationPeriod{}).TableName(); got != TrxCompensationPeriodTableName {
		t.Errorf("TrxCompensationPeriod.TableName() = %q, want %q", got, TrxCompensationPeriodTableName)
	}
	if got := (TrxVisitCommission{}).TableName(); got != TrxVisitCommissionTableName {
		t.Errorf("TrxVisitCommission.TableName() = %q, want %q", got, TrxVisitCommissionTableName)
	}
	if got := (MapVisitContributor{}).TableName(); got != MapVisitContributorTableName {
		t.Errorf("MapVisitContributor.TableName() = %q, want %q", got, MapVisitContributorTableName)
	}
	if got := (TrxWorksheet{}).TableName(); got != TrxWorksheetTableName {
		t.Errorf("TrxWorksheet.TableName() = %q, want %q", got, TrxWorksheetTableName)
	}
}

func TestApplySnapshotWage(t *testing.T) {
	days := func(n int64) sql.NullInt64 { return sql.NullInt64{Int64: n, Valid: true} }
	empty := sql.NullInt64{}

	t.Run("prorate", func(t *testing.T) {
		got, err := ApplySnapshotWage(1000, days(22), days(20), days(500))
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if !got.Valid || got.Int64 != 909 {
			t.Fatalf("total = %+v, want 909", got)
		}
	})

	t.Run("final when days missing", func(t *testing.T) {
		got, err := ApplySnapshotWage(1000, empty, empty, days(750))
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if !got.Valid || got.Int64 != 750 {
			t.Fatalf("total = %+v, want 750", got)
		}
	})

	t.Run("one day uses final", func(t *testing.T) {
		got, err := ApplySnapshotWage(1000, days(22), empty, days(400))
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if !got.Valid || got.Int64 != 400 {
			t.Fatalf("total = %+v, want 400", got)
		}
	})

	t.Run("neither stays empty", func(t *testing.T) {
		got, err := ApplySnapshotWage(1000, empty, empty, empty)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if got.Valid {
			t.Fatalf("total = %+v, want null", got)
		}
	})

	t.Run("zero staff days", func(t *testing.T) {
		got, err := ApplySnapshotWage(1000, days(22), days(0), empty)
		if err != nil {
			t.Fatalf("apply: %v", err)
		}
		if !got.Valid || got.Int64 != 0 {
			t.Fatalf("total = %+v, want 0", got)
		}
	})

	t.Run("mandatory zero", func(t *testing.T) {
		_, err := ApplySnapshotWage(1000, days(0), days(20), empty)
		var input *SnapshotWageInputError
		if err == nil || !errors.As(err, &input) || input.Code != SnapshotDaysInvalidCode {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("negative day", func(t *testing.T) {
		_, err := ApplySnapshotWage(1000, days(22), days(-1), empty)
		var input *SnapshotWageInputError
		if err == nil || !errors.As(err, &input) || input.Code != SnapshotDaysInvalidCode {
			t.Fatalf("error = %v", err)
		}
	})

	t.Run("negative final", func(t *testing.T) {
		_, err := ApplySnapshotWage(1000, empty, empty, days(-1))
		var input *SnapshotWageInputError
		if err == nil || !errors.As(err, &input) || input.Code != SnapshotFinalInvalidCode {
			t.Fatalf("error = %v", err)
		}
	})
}
