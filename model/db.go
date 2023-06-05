package model

import (
	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type Config struct {
	Type        string `yaml:"Type"`
	User        string `yaml:"User"`
	Password    string `yaml:"Password"`
	Host        string `yaml:"Host"`
	Name        string `yaml:"Name"`
	TablePrefix string `yaml:"TablePrefix"`
	Debug       bool   `yaml:"debug"`
}

//func Open(database Config) (*gorm.DB, error) {
//	db, err := gorm.Open(database.Type, fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
//		database.User,
//		database.Password,
//		database.Host,
//		database.Name))
//	if err != nil {
//		//log.Fatalf("models.Setup err: %v", err)
//		return nil, err
//	}
//
//	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
//		return database.TablePrefix + defaultTableName
//	}
//
//	db.SingularTable(true)
//	//db.LogMode(true)
//	db.DB().SetMaxIdleConns(10)
//	db.DB().SetMaxOpenConns(100)
//
//	if database.Debug {
//		db = db.Debug()
//	}
//	return db, nil
//}

func IsDuplicatedEntryError(err error) bool {
	mysqlErr, ok := err.(*mysql.MySQLError)
	if ok {
		if mysqlErr.Number == 1062 {
			return true
		}
	}
	return false
}

func IsNotFoundError(err error) bool {
	return err == gorm.ErrRecordNotFound
}
