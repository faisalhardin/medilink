package constant

// Event types for odontogram operations
const (
	// Whole tooth codes
	EventTypeToothCodeInsert = "tooth_code_insert"
	EventTypeToothCodeRemove = "tooth_code_remove"

	// General notes
	EventTypeToothGeneralNoteUpdate = "tooth_general_note_update" // Empty string clears note

	// Surface codes
	EventTypeToothSurfaceCodeSet    = "tooth_surface_code_set"
	EventTypeToothSurfaceCodeRemove = "tooth_surface_code_remove"

	// Surface notes
	EventTypeToothSurfaceNoteUpdate = "tooth_surface_note_update" // Empty string clears note

	// Utility
	EventTypeToothReset = "tooth_reset" // Clear all tooth data
)

// Valid tooth numbers (FDI two-digit notation)
var ValidToothNumbers = map[string]bool{
	// Upper right (1st quadrant)
	"11": true, "12": true, "13": true, "14": true, "15": true, "16": true, "17": true, "18": true,
	// Upper left (2nd quadrant)
	"21": true, "22": true, "23": true, "24": true, "25": true, "26": true, "27": true, "28": true,
	// Lower left (3rd quadrant)
	"31": true, "32": true, "33": true, "34": true, "35": true, "36": true, "37": true, "38": true,
	// Lower right (4th quadrant)
	"41": true, "42": true, "43": true, "44": true, "45": true, "46": true, "47": true, "48": true,
}

// Valid surface codes
var ValidSurfaceCodes = map[string]bool{
	"M": true, // Mesial
	"D": true, // Distal
	"L": true, // Lingual
	"O": true, // Occlusal
	"V": true, // Vestibular
}

// Valid whole tooth codes
var ValidWholeToothCodes = map[string]bool{
	"amf": true, // Amalgam filling
	"rct": true, // Root canal treatment
	"poc": true, // Post and core
	"mis": true, // Missing
	"imp": true, // Implant
	"cfr": true, // Crown/Fixed restoration
	"non": true, // Non-erupted
	"une": true, // Unerupted
	"sou": true, // Sound (healthy)
}

// Valid surface codes (for surface-specific treatments)
var ValidSurfaceTreatmentCodes = map[string]bool{
	"car": true, // Caries
	"amf": true, // Amalgam filling
	"gif": true, // Glass ionomer filling
	"rcf": true, // Resin composite filling
	"cof": true, // Composite filling
	"tmf": true, // Temporary filling
	"sea": true, // Sealant
	"abr": true, // Abrasion
	"att": true, // Attrition
	"ero": true, // Erosion
	"fra": true, // Fracture
}
