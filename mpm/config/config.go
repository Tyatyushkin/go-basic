package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	// Server configuration
	ServerPort string
	GRPCPort   string
	JWT        JWTConfig

	// Storage configuration
	StorageType  string // "json", "mongodb", or "postgres"
	JSONDataPath string
	MongoDB      MongoDBConfig
	Postgres     PostgresConfig
}

type JWTConfig struct {
	Secret string
	// TODO: Add expiration time configuration
}

type MongoDBConfig struct {
	// Connection settings
	Host     string
	Port     string
	Username string
	Password string
	Database string

	// Connection URI (if provided, overrides individual settings)
	URI string

	// Collection names
	Collections CollectionNames

	// Connection pool settings
	MaxPoolSize    uint64
	MinPoolSize    uint64
	MaxIdleTime    time.Duration
	ConnectTimeout time.Duration

	// Features
	EnableTransactions  bool
	EnableChangeStreams bool
}

type CollectionNames struct {
	Users    string
	Albums   string
	Photos   string
	Tags     string
	Comments string
}

type PostgresConfig struct {
	// Connection settings
	Host     string
	Port     string
	Username string
	Password string
	Database string
	SSLMode  string

	// Connection pool settings
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration

	// Additional settings
	Schema         string
	MigrationsPath string
	EnableQueryLog bool
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	cfg := &Config{
		// Server defaults
		ServerPort: getEnvOrDefault("SERVER_PORT", "8484"),
		GRPCPort:   getEnvOrDefault("GRPC_PORT", "50051"),

		// JWT configuration
		JWT: JWTConfig{
			Secret: getEnvOrDefault("JWT_SECRET", ""),
		},

		// Storage configuration
		StorageType:  getEnvOrDefault("MPM_STORAGE_TYPE", "json"),
		JSONDataPath: getEnvOrDefault("MPM_DATA_PATH", "/opt/mpm/data"),

		// MongoDB configuration
		MongoDB: MongoDBConfig{
			Host:     getEnvOrDefault("MONGODB_HOST", "localhost"),
			Port:     getEnvOrDefault("MONGODB_PORT", "27017"),
			Username: getEnvOrDefault("MONGO_USERNAME", ""),
			Password: getEnvOrDefault("MONGO_PASSWORD", ""),
			Database: getEnvOrDefault("MONGO_DATABASE", "mpm_db"),
			URI:      getEnvOrDefault("MONGODB_URI", ""),

			// Collection names
			Collections: CollectionNames{
				Users:    getEnvOrDefault("MONGO_COLLECTION_USERS", "users"),
				Albums:   getEnvOrDefault("MONGO_COLLECTION_ALBUMS", "albums"),
				Photos:   getEnvOrDefault("MONGO_COLLECTION_PHOTOS", "photos"),
				Tags:     getEnvOrDefault("MONGO_COLLECTION_TAGS", "tags"),
				Comments: getEnvOrDefault("MONGO_COLLECTION_COMMENTS", "comments"),
			},

			// Connection pool settings
			MaxPoolSize:    getEnvUint64OrDefault("MONGO_MAX_POOL_SIZE", 100),
			MinPoolSize:    getEnvUint64OrDefault("MONGO_MIN_POOL_SIZE", 5),
			MaxIdleTime:    getEnvDurationOrDefault("MONGO_MAX_IDLE_TIME", 30*time.Second),
			ConnectTimeout: getEnvDurationOrDefault("MONGO_CONNECT_TIMEOUT", 10*time.Second),

			// Features
			EnableTransactions:  getEnvBoolOrDefault("MONGO_ENABLE_TRANSACTIONS", true),
			EnableChangeStreams: getEnvBoolOrDefault("MONGO_ENABLE_CHANGE_STREAMS", false),
		},

		// PostgreSQL configuration
		Postgres: PostgresConfig{
			Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
			Port:     getEnvOrDefault("POSTGRES_PORT", "5432"),
			Username: getEnvOrDefault("POSTGRES_USER", "mpm_user"),
			Password: getEnvOrDefault("POSTGRES_PASSWORD", "mpm_password"),
			Database: getEnvOrDefault("POSTGRES_DATABASE", "mpm_db"),
			SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),

			// Connection pool settings
			MaxOpenConns:    getEnvIntOrDefault("POSTGRES_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvIntOrDefault("POSTGRES_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDurationOrDefault("POSTGRES_CONN_MAX_LIFETIME", 5*time.Minute),
			ConnMaxIdleTime: getEnvDurationOrDefault("POSTGRES_CONN_MAX_IDLE_TIME", 90*time.Second),

			// Additional settings
			Schema:         getEnvOrDefault("POSTGRES_SCHEMA", "public"),
			MigrationsPath: getEnvOrDefault("POSTGRES_MIGRATIONS_PATH", "internal/storage/postgres/migrations"),
			EnableQueryLog: getEnvBoolOrDefault("POSTGRES_ENABLE_QUERY_LOG", false),
		},
	}

	// If MongoDB URI is not provided, construct it from individual settings
	if cfg.MongoDB.URI == "" && cfg.MongoDB.Username != "" && cfg.MongoDB.Password != "" {
		cfg.MongoDB.URI = "mongodb://" + cfg.MongoDB.Username + ":" + cfg.MongoDB.Password + "@" +
			cfg.MongoDB.Host + ":" + cfg.MongoDB.Port + "/" + cfg.MongoDB.Database + "?authSource=admin"
	}

	return cfg
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvUint64OrDefault(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseUint(value, 10, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvDurationOrDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
