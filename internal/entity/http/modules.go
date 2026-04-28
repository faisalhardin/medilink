package http

type Handlers struct {
	InstitutionHandler  InstitutionHandler
	PatientHandler      PatientHandler
	AuthHandler         AuthHandler
	ProductHandler      ProductHandler
	JourneyHandler      JourneyHandler
	OdontogramHandler   OdontogramHandler
	RecallHandler       RecallHandler
	ICD10Handler        ICD10Handler
	PractitionerHandler PractitionerHandler
	DiagnosisHandler    DiagnosisHandler
	AnamnesaHandler     AnamnesaHandler
}
