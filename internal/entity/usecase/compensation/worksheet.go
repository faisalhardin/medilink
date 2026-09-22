package compensation

import (
	"context"

	"github.com/faisalhardin/medilink/internal/entity/model"
)

// WorksheetUC is the worksheet lifecycle contract.
type WorksheetUC interface {
	CreateWorksheet(ctx context.Context, req model.CreateWorksheetRequest) (model.WorksheetResponse, error)
	ListWorksheets(ctx context.Context, req model.ListWorksheetsRequest) (model.ListWorksheetsResponse, error)
	GetWorksheet(ctx context.Context, worksheetUUID string) (model.WorksheetResponse, error)
	PatchWorksheet(ctx context.Context, req model.PatchWorksheetRequest) (model.WorksheetResponse, error)
	DeleteWorksheet(ctx context.Context, worksheetUUID string) (model.DeleteWorksheetResponse, error)
	FinalizeWorksheet(ctx context.Context, worksheetUUID string) (model.FinalizeWorksheetResponse, error)
}
