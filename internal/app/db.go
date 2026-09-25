package app

import (
	"basket/internal/config"
	"basket/internal/model"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func OpenDatabase(c config.Config) (*gorm.DB, error) {
	var d gorm.Dialector
	switch c.Database.Driver {
	case "mysql":
		d = mysql.Open(c.Database.DSN)
	case "postgres", "postgresql":
		d = postgres.Open(c.Database.DSN)
	case "mssql", "sqlserver":
		d = sqlserver.Open(c.Database.DSN)
	case "sqlite":
		d = sqlite.Open(c.Database.DSN)
	default:
		return nil, fmt.Errorf("unsupported database driver %q", c.Database.Driver)
	}
	level := logger.Silent
	if c.Logging.SQL {
		level = logger.Info
	}
	db, e := gorm.Open(d, &gorm.Config{
		Logger:         logger.Default.LogMode(level),
		TranslateError: true,
		PrepareStmt:    true,
	})
	if e != nil {
		return nil, e
	}
	sql, e := db.DB()
	if e != nil {
		return nil, e
	}
	sql.SetMaxOpenConns(c.Database.MaxOpen)
	sql.SetMaxIdleConns(c.Database.MaxIdle)
	sql.SetConnMaxLifetime(c.Database.ConnMaxLifetime)
	if c.Database.AutoMigrate {
		if e = Migrate(db); e != nil {
			sql.Close()
			return nil, e
		}
	}
	return db, nil
}
func Migrate(db *gorm.DB) error {
	models := model.Entities()
	// Extended models include the JDBC-only columns and replace their JPA subset.
	for i, m := range models {
		switch m.(type) {
		case *model.ThematicExecDetails:
			models[i] = &model.ExecutionDetail{}
		case *model.ResearchcallOrderEntity:
			models[i] = &model.ResearchMaster{}
		}
	}
	models = append(models, &model.ThematicMaster{}, &model.ThematicScrip{}, &model.RebalanceScrip{}, &model.ThematicExecution{})
	return db.AutoMigrate(models...)
}
