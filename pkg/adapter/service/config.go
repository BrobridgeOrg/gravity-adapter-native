package adapter

type SourceConfig struct {
	Sources map[string]SourceEntry `json:"sources"`
}

type SourceEntry struct {
	SubscriberID        string    `json:"subscriber_id"`
	SubscriberName      string    `json:"subscriber_name"`
	Host                string    `json:"host"`
	Port                int       `json:"port"`
	WorkerCount         *int      `json:"worker_count",omitempty`
	PingInterval        *int64    `json:"ping_interval",omitempty`
	MaxPingsOutstanding *int      `json:"max_pings_outstanding",omitempty`
	MaxReconnects       *int      `json:"max_reconnects",omitempty`
	Collections         *[]string `json:"collections"`
}

type SourceInfo struct {
	SubscriberID        string
	SubscriberName      string
	Host                string
	Port                int
	WorkerCount         int
	PingInterval        int64
	MaxPingsOutstanding int
	MaxReconnects       int
	Collections         []string
}

func NewSourceInfo(entry *SourceEntry) *SourceInfo {

	info := &SourceInfo{
		SubscriberID:        entry.SubscriberID,
		SubscriberName:      entry.SubscriberName,
		Host:                entry.Host,
		Port:                entry.Port,
		WorkerCount:         16,
		PingInterval:        10,
		MaxPingsOutstanding: 3,
		MaxReconnects:       -1,
		Collections:         make([]string, 0),
	}

	// default settings
	if entry.PingInterval != nil {
		info.PingInterval = *entry.PingInterval
	}

	if entry.MaxPingsOutstanding != nil {
		info.MaxPingsOutstanding = *entry.MaxPingsOutstanding
	}

	if entry.MaxReconnects != nil {
		info.MaxReconnects = *entry.MaxReconnects
	}

	if entry.WorkerCount != nil {
		info.WorkerCount = *entry.WorkerCount
	}

	if entry.Collections != nil {
		for _, collection := range *entry.Collections {
			info.Collections = append(info.Collections, collection)
		}
	}

	return info
}
