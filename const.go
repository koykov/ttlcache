package ttlcache

const (
	cacheStatusNil    = 0
	cacheStatusActive = 1
	cacheStatusClosed = 2

	defaultEvictWorkers     = 16
	defaultDumpWriteWorkers = 16
	defaultDumpReadWorkers  = 16
)
