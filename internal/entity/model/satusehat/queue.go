package satusehat

import (
	"encoding/json"
	"time"
)

const (
	TRX_SATUSEHAT_QUEUE_TABLE = "mdl_trx_satusehat_queue"

	EventTypeDiagnosisSave = "diagnosis_save"
	EventTypeAnamnesaSave  = "anamnesa_save"

	QueueStatusPending    = "pending"
	QueueStatusProcessing = "processing"
	QueueStatusDone       = "done"
	QueueStatusFailed     = "failed"
)

// SatuSehatQueueEntry is a row in the transactional outbox table.
// It is written inside the same DB transaction as the primary data (diagnosis/anamnesa)
// and processed asynchronously by the background worker.
type SatuSehatQueueEntry struct {
	ID           string          `xorm:"'id' pk" json:"id"`
	VisitID      int64           `xorm:"'visit_id'" json:"visit_id"`
	EventType    string          `xorm:"'event_type'" json:"event_type"`
	Payload      json.RawMessage `xorm:"'payload'" json:"payload"`
	Status       string          `xorm:"'status'" json:"status"`
	Attempts     int16           `xorm:"'attempts'" json:"attempts"`
	LastError    string          `xorm:"'last_error'" json:"last_error,omitempty"`
	ProcessAfter time.Time       `xorm:"'process_after'" json:"process_after"`
	CreatedAt    time.Time       `xorm:"'created_at' created" json:"created_at"`
	UpdatedAt    time.Time       `xorm:"'updated_at' updated" json:"updated_at"`
}
