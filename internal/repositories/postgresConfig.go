package repositories

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
}
