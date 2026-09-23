package http

import "net/http"

type CompensationPeriodHandler interface {
	CreatePeriod(w http.ResponseWriter, r *http.Request)
	ListPeriods(w http.ResponseWriter, r *http.Request)
	GetPeriod(w http.ResponseWriter, r *http.Request)
	DraftPeriod(w http.ResponseWriter, r *http.Request)
	FinalizePeriod(w http.ResponseWriter, r *http.Request)
	ReopenPeriod(w http.ResponseWriter, r *http.Request)
	DeletePeriod(w http.ResponseWriter, r *http.Request)
	ListPeriodStaff(w http.ResponseWriter, r *http.Request)
	GetPeriodStaff(w http.ResponseWriter, r *http.Request)
}

type WorksheetHandler interface {
	CreateWorksheet(w http.ResponseWriter, r *http.Request)
	ListWorksheets(w http.ResponseWriter, r *http.Request)
	GetWorksheet(w http.ResponseWriter, r *http.Request)
	ListWorksheetCommissions(w http.ResponseWriter, r *http.Request)
	GenerateWorksheetCommissions(w http.ResponseWriter, r *http.Request)
	PatchWorksheet(w http.ResponseWriter, r *http.Request)
	DeleteWorksheet(w http.ResponseWriter, r *http.Request)
	FinalizeWorksheet(w http.ResponseWriter, r *http.Request)
}

type VisitCommissionHandler interface {
	GenerateVisitCommissions(w http.ResponseWriter, r *http.Request)
	ListVisitCommissions(w http.ResponseWriter, r *http.Request)
	PatchCommissionItem(w http.ResponseWriter, r *http.Request)
	ArchiveVisitCommission(w http.ResponseWriter, r *http.Request)
}
