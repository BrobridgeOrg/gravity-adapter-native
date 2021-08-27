package adapter

type SourceConfig struct {
	Sources map[string]SourceEntry `json:"sources"`
}

type SourceEntry struct {
	SubscriberID        string       `json:"subscriber_id"`
	SubscriberName      string       `json:"subscriber_name"`
	Host                string       `json:"host"`
	Port                int          `json:"port"`
	WorkerCount         *int         `json:"worker_count",omitempty`
	ChunkSize           *int         `json:"chunk_size",omitempty`
	PingInterval        *int64       `json:"ping_interval",omitempty`
	MaxPingsOutstanding *int         `json:"max_pings_outstanding",omitempty`
	MaxReconnects       *int         `json:"max_reconnects",omitempty`
	Collections         *[]string    `json:"collections"`
	Verbose             *bool        `json:"verbose",omitempty`
	InitialLoad         *bool        `json:"initial_load", omitempty`
	OmittedCount        *uint64      `json:"omitted_count", omitempty`
	Events              SourceEvents `json:"events"`
}

type SourceEvents struct {
	Snapshot string `json:"snapshot"`
	Create   string `json:"create"`
	Update   string `json:"update"`
	Delete   string `json:"delete"`
}

type SourceInfo struct {
	SubscriberID        string
	SubscriberName      string
	Host                string
	Port                int
	WorkerCount         int
	ChunkSize           int
	PingInterval        int64
	MaxPingsOutstanding int
	MaxReconnects       int
	Collections         []string
	Verbose             bool
	InitialLoad         bool
	OmittedCount        uint64
	Events              SourceEvents
}

func NewSourceInfo(entry *SourceEntry) *SourceInfo {

	info := &SourceInfo{
		SubscriberID:        entry.SubscriberID,
		SubscriberName:      entry.SubscriberName,
		Host:                entry.Host,
		Port:                entry.Port,
		WorkerCount:         4,
		ChunkSize:           2048,
		PingInterval:        10,
		MaxPingsOutstanding: 3,
		MaxReconnects:       -1,
		Verbose:             true,
		InitialLoad:         true,
		OmittedCount:        100000,
		Collections:         make([]string, 0),
		Events:              entry.Events,
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

	if entry.ChunkSize != nil {
		info.ChunkSize = *entry.ChunkSize
	}

	if entry.Verbose != nil {
		info.Verbose = *entry.Verbose
	}

	if entry.InitialLoad != nil {
		info.InitialLoad = *entry.InitialLoad
	}

	if entry.OmittedCount != nil {
		info.OmittedCount = *entry.OmittedCount
	}

	if entry.Collections != nil {
		for _, collection := range *entry.Collections {
			info.Collections = append(info.Collections, collection)
		}
	}

	return info
}
