package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	DatabaseURL string
	RabbitURL   string

	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretkey string
	MinIOUseSSL    bool
	BucketUploads  string
	BucketSites    string

	SMTPHost string
	SMTPPort int
	MailFrom string

	APIHTTPAddr       string
	APIGRPCAddr       string
	EdgeHTTPAddr      string
	ControlGRPCTarget string

	SiteDomain   string
	PublicAPIURL string

	JWTSecret []byte
	JWTTTL    time.Duration

	MaxUploadBytes    int64
	MaxUnpackedBytes  int64
	MaxFilesPerDeploy int

	LogLevel slog.Level
}

type Service string

const (
	ServiceAPI    Service = "api"
	ServiceWorker Service = "worker"
	ServiceEdge   Service = "edge"
)

var requiredBy = map[Service][]string{
	ServiceAPI: {
		"DATABASE_URL", "RABBITMQ_URL",
		"MINIO_ENDPOINT", "MINIO_ACCESS_KEY", "MINIO_SECRET_KEY",
		"MINIO_BUCKET_UPLOADS",
		"API_HTTP_ADDR", "API_GRPC_ADDR",
		"SITE_DOMAIN", "PUBLIC_API_URL", "JWT_SECRET",
	},
	ServiceWorker: {
		"DATABASE_URL", "RABBITMQ_URL",
		"MINIO_ENDPOINT", "MINIO_ACCESS_KEY", "MINIO_SECRET_KEY", "MINIO_BUCKET_UPLOADS", "MINIO_BUCKET_SITES",
		"SMTP_HOST", "MAIL_FROM",
		"SITE_DOMAIN", "PUBLIC_API_URL",
	},
	ServiceEdge: {
		"MINIO_ENDPOINT", "MINIO_ACCESS_KEY", "MINIO_SECRET_KEY",
		"MINIO_BUCKET_SITES",
		"EDGE_HTTP_ADDR", "CONTROL_GRPC_TARGET", "SITE_DOMAIN",
	},
}

func Load(svc Service) (*Config, error) {
	required := map[string]bool{}

	for _, k := range requiredBy[svc] {
		required[k] = true
	}

	var missing []string
	get := func(key string) string {
		v := os.Getenv(key)
		if v == "" && required[key] {
			missing = append(missing, key)
		}
		return v
	}
	getInt := func(key string, def int) int {
		v := os.Getenv(key)
		if v == "" {
			return def
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			missing = append(missing, key+" (must be an integer)")
			return def
		}
		return n
	}
	getInt64 := func(key string, def int64) int64 {
		v := os.Getenv(key)
		if v == "" {
			return def
		}
		n, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			missing = append(missing, key+" (must be an integer)")
			return def
		}
		return n
	}
	c := &Config{
		DatabaseURL:    get("DATABASE_URL"),
		RabbitURL:      get("RABBITMQ_URL"),
		MinIOEndpoint:  get("MINIO_ENDPOINT"),
		MinIOAccessKey: get("MINIO_ACCESS_KEY"),
		MinIOSecretkey: get("MINIO_SECRET_KEY"),
		MinIOUseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
		BucketUploads:  get("MINIO_BUCKET_UPLOADS"),

		BucketSites: get("MINIO_BUCKET_SITES"),

		SMTPHost:    get("SMTP_HOST"),
		SMTPPort:    getInt("SMTP_PORT", 1025),
		MailFrom:    get("MAIL_FROM"),
		APIHTTPAddr: get("API_HTTP_ADDR"),

		APIGRPCAddr: get("API_GRPC_ADDR"),

		EdgeHTTPAddr: get("EDGE_HTTP_ADDR"),

		ControlGRPCTarget: get("CONTROL_GRPC_TARGET"),
		SiteDomain:        get("SITE_DOMAIN"),

		PublicAPIURL: get("PUBLIC_API_URL"),
		JWTSecret:    []byte(get("JWT_SECRET")),
		JWTTTL:       time.Duration(getInt("JWT_TTL_MINUTES", 60)) * time.Minute,

		MaxUploadBytes: getInt64("MAX_UPLOAD_BYTES", 52428800),

		MaxUnpackedBytes:  getInt64("MAX_UNPACKED_BYTES", 209715200),
		MaxFilesPerDeploy: getInt("MAX_FILES_PER_DEPLOY", 2000),
		LogLevel:          parseLevel(os.Getenv("LOG_LEVEL")),
	}
	if required["JWT_SECRET"] && len(c.JWTSecret) < 32 {
		missing = append(missing, "JWT_SECRET (must be atleast 32 characters)")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%s: invalid configuration: %v", svc, missing)
	}
	return c, nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
