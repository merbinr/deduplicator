package config

type ConfigModel struct {
	StageName    string        `yaml:"stage_name"`
	OutputMethod string        `yaml:"output_method"`
	LogSources   logSources    `yaml:"log_sources"`
	Services     servicesModel `yaml:"services"`
}

type logSources struct {
	AwsVpcLogs awsVpcLogsModel `yaml:"aws_vpc_logs"`
}

type awsVpcLogsModel struct {
	UniqueStringFields string `yaml:"unique_string_fields"`
}

type servicesModel struct {
	IncommingQueue queueModel      `yaml:"incomming_queue"`
	OutgoingQueue  queueModel      `yaml:"outgoing_queue"`
	RedisCache     redisCacheModel `yaml:"redis_cache"`
	Opensearch     opensearchModel `yaml:"opensearch"`
}

type queueModel struct {
	User      string `yaml:"user"`
	Port      int    `yaml:"port"`
	QueueName string `yaml:"queue_name"`
}

type redisCacheModel struct {
	Port   int `yaml:"port"`
	Db     int `yaml:"db"`
	Expiry int `yaml:"expiry"`
}

type opensearchModel struct {
	Port               int    `yaml:"port"`
	Username           string `yaml:"username"`
	PreferredBatchSize int    `yaml:"preferred_batch_size"`
	Retries            int    `yaml:"retries"`
	RetryDelay         int    `yaml:"retry_delay"`
}
