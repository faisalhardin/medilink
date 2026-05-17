package institution

import (
	"testing"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	customtime "github.com/faisalhardin/medilink/pkg/type/time"
)

func mustParseRFC3339(t *testing.T, value string) customtime.Time {
	t.Helper()
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		t.Fatalf("parse time: %v", err)
	}
	return customtime.Time{Time: parsed}
}

func TestBuildProductStatisticsQuery(t *testing.T) {
	t.Parallel()

	query, _, _, err := buildProductStatisticsQuery(model.ProductStatisticsParams{
		StartTime:   mustParseRFC3339(t, "2026-04-01T00:00:00+07:00"),
		EndTime:     mustParseRFC3339(t, "2026-04-15T12:30:00+07:00"),
		Granularity: model.ProductStatisticsGranularityHour,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if query.Granularity != model.ProductStatisticsGranularityHour {
		t.Fatalf("expected hour granularity, got %s", query.Granularity)
	}
	if !query.StartTime.UTC().Equal(time.Date(2026, 3, 31, 17, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected start UTC: %s", query.StartTime.UTC())
	}
}

func TestBuildProductStatisticsQueryRejectsRangeOver31Days(t *testing.T) {
	t.Parallel()

	_, _, _, err := buildProductStatisticsQuery(model.ProductStatisticsParams{
		StartTime: mustParseRFC3339(t, "2026-04-01T00:00:00+07:00"),
		EndTime:   mustParseRFC3339(t, "2026-05-03T00:00:00+07:00"),
	})
	if err == nil {
		t.Fatal("expected error for range over 31 days")
	}
}

func TestAssembleProductStatisticsResponse(t *testing.T) {
	t.Parallel()

	periodStart := time.Date(2026, 4, 1, 0, 0, 0, 0, time.UTC)
	loc := time.FixedZone("", 7*3600)
	resp := assembleProductStatisticsResponse(
		model.ProductStatisticsQuery{
			Granularity: model.ProductStatisticsGranularityDay,
		},
		time.Date(2026, 4, 1, 0, 0, 0, 0, loc),
		time.Date(2026, 4, 2, 0, 0, 0, 0, loc),
		[]model.ProductStatisticsRow{
			{
				PeriodStart:             periodStart,
				IDTrxInstitutionProduct: 1,
				Name:                    "Paracetamol",
				TotalQuantity:           2,
				TotalRevenue:            10000,
				AvgUnitPrice:            5000,
			},
			{
				PeriodStart:             periodStart,
				IDTrxInstitutionProduct: 2,
				Name:                    "Vitamin C",
				TotalQuantity:           1,
				TotalRevenue:            3000,
				AvgUnitPrice:            3000,
			},
		},
	)

	if len(resp.Buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(resp.Buckets))
	}
	if resp.Buckets[0].TotalRevenue != 13000 {
		t.Fatalf("expected bucket revenue 13000, got %v", resp.Buckets[0].TotalRevenue)
	}
	if resp.Summary.TotalQuantity != 3 {
		t.Fatalf("expected summary quantity 3, got %d", resp.Summary.TotalQuantity)
	}
	if len(resp.Summary.TopProductsByRevenue) != 2 {
		t.Fatalf("expected 2 top products, got %d", len(resp.Summary.TopProductsByRevenue))
	}
}
