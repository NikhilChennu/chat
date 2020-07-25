package db

import (
	"fmt"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

//docker run --name postgres-docker -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres

type DBObject struct {
	conn *gorm.DB
}

const (
	tableName = "accounts"
	mobileLen = 10
)

//GetDBObject - Retruns new instance of DB
func GetDBObject(host string, port string, username string, password string, dbName string) (*DBObject, error) {
	dbUri := fmt.Sprintf("host=%s port=%s user=%s dbname=%s sslmode=disable password=%s", host, port, username, dbName, password) //Build connection string
	print(dbUri)
	conn, err := gorm.Open("postgres", dbUri)
	if err != nil {
		return nil, err
	}
	conn.Debug().AutoMigrate(&Account{})

	return &DBObject{conn}, nil

}

//CheckAccount - checks existance of an account
func (db *DBObject) CheckAccount(mobile string) (bool, error) {

	if mobile == "" || len(mobile) != 10 {
		return false, fmt.Errorf("invalid mobile number")
	}

	temp := &Account{}

	err := db.conn.Table(tableName).Where("mobile = ?", mobile).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, fmt.Errorf("Connection error. Please retry")
	}
	if temp.Mobile != "" {
		return true, nil
	}
	return false, nil

}

//GetAccount - retrieves account details
func (db *DBObject) GetAccount(mobile string) (*Account, error) {
	if mobile == "" || len(mobile) != 10 {
		return nil, fmt.Errorf("invalid mobile number")
	}

	temp := &Account{}

	err := db.conn.Table(tableName).Where("mobile = ?", mobile).First(temp).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("Connection error. Please retry")
	}
	if temp.Mobile != "" {
		return temp, nil
	}
	return nil, fmt.Errorf("Account doesn't exists")

}

//DeleteAccount - deletes account
func (db *DBObject) DeleteAccount(mobile string) (bool, error) {

	deleteAccnt := Account{Mobile: mobile}
	db.conn.Delete(&deleteAccnt)

	if deleteAccnt.DeletedAt != nil {
		return false, fmt.Errorf("Connection error. Please retry")
	}

	return true, nil
}

//AddAccount - adds account
func (db *DBObject) AddAccount(account *Account) error {

	if len(account.UserID) == 0 {
		return fmt.Errorf("Invalid UserId")
	}

	if account.Mobile == "" || len(account.Mobile) != 10 {
		return fmt.Errorf("invalid mobile number")
	}

	db.conn.Create(account)
	if account.ID <= 0 {
		return fmt.Errorf("Failed to create account, connection error.")
	}
	return nil
}

func (db *DBObject) Close() {
	if db.conn != nil {
		db.conn.Close()
	}
}
