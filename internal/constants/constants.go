package constants

const (
	// User validation constants
	MinUsernameLength = 3
	MaxUsernameLength = 32

	// User naming constants
	UsernameSeparator = "-"

	// Traffic constants
	BytesInGB = 1024 * 1024 * 1024

	// Duration constants
	MillisecondsInDay = 24 * 60 * 60 * 1000

	// Network constants
	DefaultTimeout          = 30
	DefaultRetryCount       = 3
	DefaultRetryWaitTime    = 5
	DefaultRetryMaxWaitTime = 20

	// Cache constants
	CacheExpiration      = 30 // minutes
	CacheCleanupInterval = 10 // minutes

	// Formatting constants
	MaxEmailDisplayLength = 17
	MaxEmailSuffixLength  = 14
	TimestampFormat       = "2006-01-02 15:04:05"
	DateFormat            = "2006-01-02"
)
