package config

type MysqlConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"db"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type ServerConfig struct {
	Host      string      `mapstructure:"host"`
	Port      int32       `mapstructure:"port"`
	MysqlInfo MysqlConfig `mapstructure:"mysql"`
	UserSrv UserSrvConfig `mapstructure:"usersrv"`
}
type UserSrvConfig struct {
	Host      string      `mapstructure:"host"`
	Port      int32       `mapstructure:"port"`
}
