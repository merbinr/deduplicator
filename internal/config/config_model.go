package config

type ConfigModel struct {
	RedisCache     redis_cache_model `yaml:"redis_cache"`
	IncommingQueue queue_model       `yaml:"incomming_queue"`
	QutgoingQueue  queue_model       `yaml:"outgoing_queue"`
}

type queue_model struct {
	User string `yaml:"user"`
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
	Name string `yaml:"queue_name"`
}

type redis_cache_model struct {
	Host string `yaml:"host"`
	Port uint16 `yaml:"port"`
	DB   int    `yaml:"db"`
}
