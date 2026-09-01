package model

import "testing"

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
		{"empty", WageCadence(""), false},
		{"daily", WageCadence("daily"), false},
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

func TestCompensationTableNames(t *testing.T) {
	if got := (MstStaffWage{}).TableName(); got != MstStaffWageTableName {
		t.Errorf("MstStaffWage.TableName() = %q, want %q", got, MstStaffWageTableName)
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
}
