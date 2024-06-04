package schema

import "database/sql/driver"

type Status string

const (
	Created   = "created"   // Order Created
	Queued    = "queued"    // Order Transaction ID added
	Confirmed = "confirmed" // Order Transaction has > 10 confirmation
	Paid      = "paid"      // Ready to Send
	Invalid   = "invalid"   // Not enough AR
	Posted    = "posted"    // Sent to Arweave
	Release   = "release"   // Request Liteseed Reward
	Permanent = "permanent" //
	Error     = "error"
)

func (s *Status) Scan(value any) error {
	*s = Status(value.(string))
	return nil
}

func (s Status) Value() (driver.Value, error) {
	return string(s), nil
}

type Order struct {
	ID            string `json:"id"`
	TransactionID string `json:"transaction_id"`
	URL           string `json:"url"`
	Address       string `json:"address"`
	Status        Status `gorm:"index:idx_status;default:created" sql:"type:status" json:"status"`
	Size          uint   `json:"size"`
}
