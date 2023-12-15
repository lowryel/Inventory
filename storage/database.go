package database

import (
	_ "github.com/lib/pq"
	"xorm.io/xorm"

	// "gorm.io/driver/postgres"
	"fmt"
	"time"
)

type Inventory struct{
	ID 			int64			`json:"id"`
	Name 		*string			`json:"name" validate:"required"`
	Price 		*float32		`json:"price" validate:"required"`
	Instock 	*bool			`json:"instock" validate:"required"`
	Quantity 	*int			`json:"quantity" validate:"required"`
	UserID		*StoreUsers		`json:"user_id"`
}

type StoreUsers struct{
	ID 				int64		`json:"id"`
	First_name		*string		`json:"first_name" validate:"required"`
	Username 		*string		`json:"username" validate:"required, min=3, max=50"`
	Phone 			*string		`json:"phone" validate:"required, min=10, max=15"`
	Email 			*string		`json:"email" validate:"required, email"`
	Password		*string		`json:"password" validate:"required, min=8"`
	User_type		*string		`json:"user_type" validate:"required, eq=ADMIN|eq=USER"`
	Token			*string		`json:"token"`
	Created			time.Time	`json:"created"`
	Refresh_token	*string		`json:"refresh_token"`
}

type Login struct {
	Username string	`json:"username"`
	Password string	`json:"password"`
}

type LoginData struct{
	ID 				int64		`json:"id" validate:"autoIncrement"`
	Username 		string		`json:"username" validate:"required, min=3, max=50"`
	Phone 			string		`json:"phone" validate:"required, min=10, max=15"`
	Email 			string		`json:"email" validate:"required, email"`
	Password		string		`json:"password" validate:"required, min=8"`
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
	if err := engine.Sync(new(Inventory), new(StoreUsers), new(LoginData)); err != nil{
		return nil, err
	}
	return engine, err
}
