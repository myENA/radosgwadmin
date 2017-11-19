package radosgwadmin

import (
	"context"
)

// UsageRequest - desribes a usage request
type UsageRequest struct {
	UID         string    `url:"uid,omitempty"`
	Start       RadosTime `url:"start,omitempty"`
	End         RadosTime `url:"end,omitempty"`
	ShowEntries bool      `url:"show-entries,omitempty"`
	ShowSummary bool      `url:"show-summary,omitempty"`
}

// TrimUsageRequest - describes a trim usage request
type TrimUsageRequest struct {
	UID       string    `url:"uid,omitempty"`
	Start     RadosTime `url:"start,omitempty"`
	End       RadosTime `url:"end,omitempty"`
	RemoveAll bool      `url:"remove-all,omitempty"`
}

// UsageTrim - trim usage data
func (aa *AdminAPI) UsageTrim(ctx context.Context, treq *TrimUsageRequest) error {
	err := aa.Delete(ctx, "/usage", treq, nil)
	return err
}

// Usage - Get usage data.
func (aa *AdminAPI) Usage(ctx context.Context, ureq *UsageRequest) (*UsageResponse, error) {
	uresp := new(UsageResponse)

	err := aa.Get(ctx, "/usage", ureq, uresp)
	return uresp, err
}

// UsageResponse - response from a GetUsage()
type UsageResponse struct {
	Entries []UsageEntry   `json:"entries"`
	Summary []UsageSummary `json:"summary"`
}

// UsageEntry - usage entry.
type UsageEntry struct {
	Buckets []UsageBucket `json:"buckets"`
	User    string        `json:"user"`
}

// UsageSummary - usage summary info.
type UsageSummary struct {
	Categories []UsageCategory `json:"categories"`
	Total      *UsageTotal     `json:"total"`
}

// UsageBucket - bucket usage metadata
type UsageBucket struct {
	Bucket     string          `json:"bucket"`
	Owner      string          `json:"owner"`
	Categories []UsageCategory `json:"categories"`
	Epoch      int             `json:"epoch"`
	Time       RadosTime       `json:"time"`
}

// UsageCategory - usage by category.
type UsageCategory struct {
	BytesSent     int    `json:"bytes_sent"`
	BytesReceived int    `json:"bytes_received"`
	Ops           int    `json:"ops"`
	SuccessfulOps int    `json:"successful_ops"`
	Category      string `json:"category"`
}

// UsageTotal - overall usage totals
type UsageTotal struct {
	BytesSent     int `json:"bytes_sent"`
	BytesReceived int `json:"bytes_received"`
	Ops           int `json:"ops"`
	SuccessfulOps int `json:"successful_ops"`
}
