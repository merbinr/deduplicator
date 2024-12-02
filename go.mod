module github.com/merbinr/deduplicator

go 1.22.5

require github.com/redis/go-redis/v9 v9.7.0

require github.com/merbinr/log_models v0.0.0-20241201181219-2e7df76c9734

require (
	github.com/buger/jsonparser v1.1.1
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/rabbitmq/amqp091-go v1.10.0
	gopkg.in/yaml.v3 v3.0.1
)
