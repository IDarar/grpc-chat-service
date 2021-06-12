package mysql

import (
	"database/sql"
	"fmt"

	"github.com/IDarar/grpc-chat-service/internal/config"
	"github.com/IDarar/hub/pkg/logger"
	_ "github.com/go-sql-driver/mysql"
)

func NewMySQLDB(cfg *config.Config) (*sql.DB, error) {

	dbSource := fmt.Sprintf("root:%v@tcp(127.0.0.1:%v)/%v", cfg.MySQL.Password, cfg.MySQL.Port, cfg.MySQL.DBname)

	db, err := sql.Open("mysql", dbSource)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	return db, nil
}
