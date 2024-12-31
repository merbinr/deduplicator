package config

type ConfigModel struct {
	RedisCache     redis_cache_model `yaml:"redis_cache"`
	IncommingQueue queue_model       `yaml:"incomming_queue"`
	QutgoingQueue  queue_model       `yaml:"outgoing_queue"`
	LogSource      log_source_model  `yaml:"log_sources"`
	OutputMethod   string            `yaml:"output_method"`
	Opensearch     opensearch_model  `yaml:"opensearch"`
}

type opensearch_model struct {
	Host              string `yaml:"host"`
	Port              int    `yaml:"port"`
	Username          string `yaml:"username"`
	PreferedBatchSize int    `yaml:"prefered_patch_size"`
	Retries           int    `yaml:"retries"`
	RetryDelay        int    `yaml:"retry_delay"`
}

type queue_model struct {
	User string `yaml:"user"`
	Port uint16 `yaml:"port"`
	Name string `yaml:"queue_name"`
}

type redis_cache_model struct {
	Port   uint16 `yaml:"port"`
	DB     int    `yaml:"db"`
	Expiry int    `yaml:"expiry"`
}

type log_source_model struct {
	AwsVpcLogsModel aws_vpc_logs_model `yaml:"aws_vpc_logs"`
}

type aws_vpc_logs_model struct {
	UniqueStringFields string `yaml:"unique_string_fields"`
}
