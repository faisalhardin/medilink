package permission

// Patient permissions
const (
	PatientRead   = "patient.read"
	PatientCreate = "patient.create"
	PatientUpdate = "patient.update"
)

// Visit permissions
const (
	VisitRead   = "visit.read"
	VisitCreate = "visit.create"
	VisitUpdate = "visit.update"
)

// Diagnosis permissions
const (
	DiagnosisRead   = "diagnosis.read"
	DiagnosisCreate = "diagnosis.create"
	DiagnosisUpdate = "diagnosis.update"
	DiagnosisDelete = "diagnosis.delete"
)

// Anamnesa permissions
const (
	AnamnesaRead   = "anamnesa.read"
	AnamnesaCreate = "anamnesa.create"
	AnamnesaUpdate = "anamnesa.update"
	AnamnesaDelete = "anamnesa.delete"
)

// Product permissions
const (
	ProductRead       = "product.read"
	ProductCreate     = "product.create"
	ProductUpdate     = "product.update"
	ProductDelete     = "product.delete"
	ProductStatistics = "product.statistics"
)

// Journey permissions
const (
	JourneyRead   = "journey.read"
	JourneyCreate = "journey.create"
	JourneyUpdate = "journey.update"
	JourneyDelete = "journey.delete"
)

// Recall permissions
const (
	RecallRead   = "recall.read"
	RecallCreate = "recall.create"
	RecallUpdate = "recall.update"
	RecallDelete = "recall.delete"
)

// Odontogram permissions
const (
	OdontogramRead   = "odontogram.read"
	OdontogramCreate = "odontogram.create"
	OdontogramUpdate = "odontogram.update"
	OdontogramDelete = "odontogram.delete"
)

// Reference permissions
const (
	ReferenceSearch = "reference.search"
)

// Staff permissions
const (
	StaffRead       = "staff.read"
	StaffCreate     = "staff.create"
	StaffUpdate     = "staff.update"
	StaffDelete     = "staff.delete"
	StaffRoleAssign = "staff.role.assign"
)

// Compensation permissions
const (
	CompensationRead     = "compensation.read"
	CompensationAssign   = "compensation.assign"
	CompensationFinalize = "compensation.finalize"
	CompensationManage   = "compensation.manage"
)
