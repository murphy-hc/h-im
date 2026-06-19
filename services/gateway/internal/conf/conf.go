package conf

import "time"

// Server holds server configuration.
type Server struct {
	WS WSServer
}

// WSServer holds WebSocket server configuration.
type WSServer struct {
	Addr string
}

// JWT holds JWT configuration.
type JWT struct {
	Secret      string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
}

// Data holds data source configuration.
type Data struct {
	// Gateway doesn't need direct DB access
}
