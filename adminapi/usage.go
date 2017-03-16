package adminapi

type UsageRequest struct {
	Uid         string    `url:uid,omitempty`
	Start       RadosTime `url:start,omitempty`
	End         RadosTime `url:end,omitempty`
	ShowEntries bool      `url:show-entries,omitempty`
	ShowSummary bool      `url:show-summary,omitempty`
}

func (aa *AdminApi) GetUsage(ureq *UsageRequest) (*UsageResponse, error) {
	uresp := new(UsageResponse)

	err := aa.Get("/usage", ureq, uresp)
	if err != nil {
		return uresp, err
	}
	return uresp, nil
}

type UsageResponse struct {
	Entries []UsageEntry   `json:entries`
	Summary []UsageSummary `json:summary`
}

type UsageEntry struct {
	Buckets []UsageBucket `json:bucket`
	User    string        `json:user`
}

type UsageSummary struct {
	Categories []UsageCategory `json:categories`
	Total      *UsageTotal     `json:total`
}

type UsageBucket struct {
	Bucket     string          `json:bucket`
	Owner      string          `json:owner`
	Categories []UsageCategory `json:categories`
	Epoch      int             `json:epoch`
	Time       RadosTime       `json:time`
}

type UsageCategory struct {
	BytesSent     int    `json:bytes_sent`
	BytesReceived int    `json:bytes_received`
	Ops           int    `json:ops`
	SuccessfulOps int    `json:successful_ops`
	Category      string `json:category`
}

type UsageTotal struct {
	BytesSent     int `json:bytes_sent`
	BytesReceived int `json:bytes_received`
	Ops           int `json:ops`
	SuccessfulOps int `json:successful_ops`
}
