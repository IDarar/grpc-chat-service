package config

import (
	"strings"

	"github.com/spf13/viper"
)

type (
	Config struct {
		Environment string
		MySQL       MySQLConfig
		GRPC        GRPCConfig
	}
	MySQLConfig struct {
		Port     string `mapstructure:"MYSQL_PORT"`
		DBname   string `mapstructure:"MYSQL_DBNAME"`
		User     string `mapstructure:"MYSQL_USER"`
		Password string `mapstructure:"MYSQL_PASSWORDGO"`
	}
	GRPCConfig struct {
		Port             string `mapstructure:"port"`
		ServerCertFile   string `mapstructure:"servercertfile"`
		ServerKeyFile    string `mapstructure:"serverkeyfile"`
		ClientCACertFile string `mapstructure:"clientcacertfile"`
		ClientKeyFile    string `mapstructure:"clinetkeyfile"`
		ClientCertFile   string `mapstructure:"clinetcertfile"`
	}
)

func Init(path string) (*Config, error) {

	if err := parseEnv(); err != nil {
		return nil, err
	}

	if err := parseConfigFile(path); err != nil {
		return nil, err
	}

	var cfg Config
	if err := unmarshal(&cfg); err != nil {
		return nil, err
	}

	setFromEnv(&cfg)
	return &cfg, nil
}

func unmarshal(cfg *Config) error {
	/*if err := viper.UnmarshalKey("mysql", &cfg.MySQL); err != nil {
		return err
	}*/
	if err := viper.UnmarshalKey("grpc", &cfg.GRPC); err != nil {
		return err
	}

	return nil
}
func setFromEnv(cfg *Config) {
	cfg.MySQL.Port = viper.GetString("port")
	cfg.MySQL.DBname = viper.GetString("dbname")
	cfg.MySQL.User = viper.GetString("user")
	cfg.MySQL.Password = viper.GetString("passwordgo")
}

func parseConfigFile(filepath string) error {
	path := strings.Split(filepath, "/")

	viper.AddConfigPath(path[0]) // folder
	viper.SetConfigName(path[1]) // config file name

	return viper.ReadInConfig()
}

func parseEnv() error {
	return parseMySQLEnvVariables()
}

func parseMySQLEnvVariables() error {

	viper.SetEnvPrefix("mysql")
	if err := viper.BindEnv("user"); err != nil {
		return err
	}

	if err := viper.BindEnv("database"); err != nil {
		return err
	}
	if err := viper.BindEnv("passwordgo"); err != nil {
		return err
	}
	if err := viper.BindEnv("port"); err != nil {
		return err
	}

	viper.SetEnvPrefix("database")
	if err := viper.BindEnv("url"); err != nil {
		return err
	}
	return nil

}
