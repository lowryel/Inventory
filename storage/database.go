package database

import (
	"xorm.io/xorm"
	_ "github.com/lib/pq"
	// "gorm.io/driver/postgres"
	"fmt"
	// "gorm.io/gorm"
)

// type Config struct{
// 	Host 		string
// 	Port 		string
// 	User 		string
// 	DBName 		string
// 	Password 	string
// 	SSLMode 	string
// }

type Inventory struct{
	ID 			int64	`json:"id"`
	Name 		*string	`json:"name"`
	Price 		*float32	`json:"price"`
	Instock 	*bool	`json:"instock"`
	Quantity 	*int	`json:"quantity"`
}

func NewConnection() (*xorm.Engine, error){
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s  dbname=%s sslmode=disable", "localhost", 5432, "lotus_api", "lotus@api", "shops")

	engine, err := xorm.NewEngine("postgres", dsn)
	if err != nil{
		return nil, err
	}
	fmt.Println("Hello database launching...")
	if err := engine.Ping(); err != nil{
		return nil, err
	}
	if err := engine.Sync(new(Inventory)); err != nil{
		return nil, err
	}
	return engine, err
}
