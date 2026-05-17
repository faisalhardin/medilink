package institution

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/faisalhardin/medilink/internal/entity/model"
	"github.com/faisalhardin/medilink/internal/library/common/commonerr"
	"github.com/faisalhardin/medilink/internal/library/middlewares/auth"
	"github.com/pkg/errors"
)

const (
	productStatisticsMaxWindow = 31 * 24 * time.Hour
	productStatisticsTopN      = 5
)

var (
	WrapMsgGetProductStatistics = WrapErrMsgPrefix + "GetProductStatistics"
)

func (uc *InstitutionUC) GetProductStatistics(ctx context.Context, params model.ProductStatisticsParams) (result model.ProductStatisticsResponse, err error) {
	userDetail, found := auth.GetUserDetailFromCtx(ctx)
	if !found {
		err = commonerr.SetNewUnauthorizedAPICall()
		return
	}

	query, startLocal, endLocal, err := buildProductStatisticsQuery(params)
	if err != nil {
		return
	}
	query.IDMstInstitution = userDetail.InstitutionID

	rows, err := uc.InstitutionRepo.GetProductStatistics(ctx, query)
	if err != nil {
		err = errors.Wrap(err, WrapMsgGetProductStatistics)
		return
	}

	result = assembleProductStatisticsResponse(query, startLocal, endLocal, rows)
	return
}

func buildProductStatisticsQuery(params model.ProductStatisticsParams) (query model.ProductStatisticsQuery, startLocal, endLocal time.Time, err error) {
	if params.StartTime.IsZero() {
		err = commonerr.SetNewBadRequest("invalid start_time", "start_time must be RFC3339 with timezone offset")
		return
	}
	if params.EndTime.IsZero() {
		err = commonerr.SetNewBadRequest("invalid end_time", "end_time must be RFC3339 with timezone offset")
		return
	}

	startLocal = params.StartTime.Time
	endLocal = params.EndTime.Time

	if !endLocal.After(startLocal) {
		err = commonerr.SetNewBadRequest("invalid time range", "end_time must be after start_time")
		return
	}

	if endLocal.Sub(startLocal) > productStatisticsMaxWindow {
		err = commonerr.SetNewBadRequest("invalid time range", "time range must not exceed 31 days")
		return
	}

	granularity := params.Granularity
	if granularity == "" {
		granularity = model.ProductStatisticsGranularityDay
	}
	switch granularity {
	case model.ProductStatisticsGranularityHour,
		model.ProductStatisticsGranularityDay,
		model.ProductStatisticsGranularityWeek:
	default:
		err = commonerr.SetNewBadRequest("invalid granularity", "granularity must be hour, day, or week")
		return
	}

	query = model.ProductStatisticsQuery{
		StartTime:               startLocal,
		EndTime:                 endLocal,
		Granularity:             granularity,
		IDTrxInstitutionProduct: params.ProductID,
	}
	return
}

func assembleProductStatisticsResponse(
	query model.ProductStatisticsQuery,
	startLocal, endLocal time.Time,
	rows []model.ProductStatisticsRow,
) model.ProductStatisticsResponse {
	loc := startLocal.Location()

	bucketIndex := make(map[time.Time]int)
	buckets := make([]model.ProductStatisticsBucket, 0)
	summaryByProduct := make(map[int64]*model.ProductStatisticsSummaryItem)

	var summaryRevenue float64
	var summaryQuantity int64

	for _, row := range rows {
		idx, ok := bucketIndex[row.PeriodStart]
		if !ok {
			periodEnd := bucketPeriodEnd(row.PeriodStart, query.Granularity)
			buckets = append(buckets, model.ProductStatisticsBucket{
				PeriodStart: formatTimeInLocation(row.PeriodStart, loc),
				PeriodEnd:   formatTimeInLocation(periodEnd, loc),
				Products:    []model.ProductStatisticsProductItem{},
			})
			idx = len(buckets) - 1
			bucketIndex[row.PeriodStart] = idx
		}

		item := model.ProductStatisticsProductItem{
			ProductID:     row.IDTrxInstitutionProduct,
			Name:          row.Name,
			UnitPrice:     row.AvgUnitPrice,
			TotalQuantity: row.TotalQuantity,
			TotalRevenue:  row.TotalRevenue,
		}
		buckets[idx].Products = append(buckets[idx].Products, item)
		buckets[idx].TotalRevenue += row.TotalRevenue
		buckets[idx].TotalQuantity += row.TotalQuantity

		summaryRevenue += row.TotalRevenue
		summaryQuantity += row.TotalQuantity

		agg, exists := summaryByProduct[row.IDTrxInstitutionProduct]
		if !exists {
			summaryByProduct[row.IDTrxInstitutionProduct] = &model.ProductStatisticsSummaryItem{
				ProductID:     row.IDTrxInstitutionProduct,
				Name:          row.Name,
				UnitPrice:     row.AvgUnitPrice,
				TotalQuantity: row.TotalQuantity,
				TotalRevenue:  row.TotalRevenue,
			}
			continue
		}
		agg.TotalQuantity += row.TotalQuantity
		agg.TotalRevenue += row.TotalRevenue
		if agg.TotalQuantity > 0 {
			agg.UnitPrice = agg.TotalRevenue / float64(agg.TotalQuantity)
		}
	}

	summaryItems := make([]model.ProductStatisticsSummaryItem, 0, len(summaryByProduct))
	for _, item := range summaryByProduct {
		summaryItems = append(summaryItems, *item)
	}

	return model.ProductStatisticsResponse{
		Period: model.ProductStatisticsPeriod{
			Start: startLocal.In(loc).Format(time.RFC3339),
			End:   endLocal.In(loc).Format(time.RFC3339),
		},
		Granularity: query.Granularity,
		UTCOffset:   formatOffsetSeconds(zoneOffsetSeconds(startLocal)),
		Buckets:     buckets,
		Summary: model.ProductStatisticsSummary{
			TotalRevenue:          summaryRevenue,
			TotalQuantity:         summaryQuantity,
			TopProductsByRevenue:  topProductSummaryItems(summaryItems, productStatisticsTopN, true),
			TopProductsByQuantity: topProductSummaryItems(summaryItems, productStatisticsTopN, false),
		},
	}
}

func topProductSummaryItems(items []model.ProductStatisticsSummaryItem, n int, byRevenue bool) []model.ProductStatisticsSummaryItem {
	if len(items) == 0 {
		return []model.ProductStatisticsSummaryItem{}
	}

	sort.Slice(items, func(i, j int) bool {
		if byRevenue {
			if items[i].TotalRevenue == items[j].TotalRevenue {
				return items[i].TotalQuantity > items[j].TotalQuantity
			}
			return items[i].TotalRevenue > items[j].TotalRevenue
		}
		if items[i].TotalQuantity == items[j].TotalQuantity {
			return items[i].TotalRevenue > items[j].TotalRevenue
		}
		return items[i].TotalQuantity > items[j].TotalQuantity
	})

	if len(items) < n {
		n = len(items)
	}
	return append([]model.ProductStatisticsSummaryItem(nil), items[:n]...)
}

func bucketPeriodEnd(start time.Time, granularity string) time.Time {
	switch granularity {
	case model.ProductStatisticsGranularityHour:
		return start.Add(time.Hour).Add(-time.Second)
	case model.ProductStatisticsGranularityWeek:
		return start.Add(7 * 24 * time.Hour).Add(-time.Second)
	default:
		return start.Add(24 * time.Hour).Add(-time.Second)
	}
}

func formatTimeInLocation(t time.Time, loc *time.Location) string {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc).Format(time.RFC3339)
}

func zoneOffsetSeconds(t time.Time) int {
	_, offsetSec := t.Zone()
	return offsetSec
}

func formatOffsetSeconds(offsetSec int) string {
	sign := "+"
	if offsetSec < 0 {
		sign = "-"
		offsetSec = -offsetSec
	}
	hours := offsetSec / 3600
	mins := (offsetSec % 3600) / 60
	return fmt.Sprintf("%s%02d:%02d", sign, hours, mins)
}
