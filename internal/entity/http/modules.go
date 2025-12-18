package http

type Handlers struct {
	InstitutionHandler InstitutionHandler
	PatientHandler     PatientHandler
	AuthHandler        AuthHandler
	ProductHandler     ProductHandler
	JourneyHandler     JourneyHandler
	OdontogramHandler  OdontogramHandler
}
