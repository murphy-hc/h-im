package conf

// Server holds server configuration.
type Server struct {
	GRPC GRPCServer
}

// GRPCServer holds gRPC server configuration.
type GRPCServer struct {
	Addr string
}

// Data holds data source configuration.
type Data struct {
	Database Database
	Redis    Redis
}

// Database holds database configuration.
type Database struct {
	Driver string
	Source string
}

// Redis holds Redis configuration.
type Redis struct {
	Addr     string
	Password string
	DB       int
}
