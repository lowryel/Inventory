package database

import (
	"xorm.io/xorm"
	_ "github.com/lib/pq"
	// "gorm.io/driver/postgres"
	"fmt"
)

type Inventory struct{
	ID 			int64	`json:"id"`
	Name 		*string	`json:"name"`
	Price 		*float32	`json:"price"`
	Instock 	*bool	`json:"instock"`
	Quantity 	*int	`json:"quantity"`
	UserID		*StoreUsers	`json:"user_id"`
}

type StoreUsers struct{
	ID 			int64	`json:"id"`
	Name 		*string	`json:"name"`
	Phone 		*string	`json:"phone"`
	Email 		*string	`json:"email"`
	// Inventory	*[]Inventory `json:"inventory"`
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
	if err := engine.Sync(new(Inventory), new(StoreUsers)); err != nil{
		return nil, err
	}
	return engine, err
}
