package db

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

type Database interface {
	ListOrders() []Order
	AddOrder(author User, order Order) error
	UpdateOrder(order Order) error

	ListUsers() []User
	AddUser(user User) error

	Close() error
}

var _ Database = &DB{}

type User struct {
	ID      int64
	Name    string `gorm:"not null;unique"`
	Skills  []Skill
	Created time.Time
	Balance float64
}

type Order struct {
	ID       int64
	UserID   int64  `gorm:"index"`
	Assigned int64  `gorm:"index"`
	Title    string `gorm:"not null"`
	Duration time.Duration
	Fee      float64
	Created  time.Time
	Done     time.Time
}

type Skill struct {
	ID     int64
	UserID int64
	Name   string
}

type DB struct {
	db *gorm.DB
}

func NewDB() Database {
	db, err := gorm.Open("postgres", "host=localhost port=5432 dbname=freelance sslmode=disable")
	if err != nil {
		panic(err)
	}

	db.LogMode(false)
	db.AutoMigrate(&User{}, &Order{}, &Skill{})

	return &DB{db: db}
}

func (db *DB) ListOrders() (list []Order) {
	db.db.Find(&list)
	return list
}

func (db *DB) UpdateOrder(order Order) error {
	tx := db.db.Begin()

	var state Order

	tx.Where(&Order{ID: order.ID}).First(&state)
	if !state.Done.IsZero() {
		tx.Rollback()
		return fmt.Errorf("Order already done")
	}

	if !order.Done.IsZero() {
		tx.Model(&User{ID: order.Assigned}).UpdateColumn("balance", gorm.Expr("balance + ?", state.Fee))
	}

	tx.Model(&order).UpdateColumns(Order{Assigned: order.Assigned, Done: order.Done})

	rdb := tx.Commit()

	return rdb.Error
}

func (db *DB) AddOrder(author User, order Order) error {
	order.UserID = author.ID
	order.Created = time.Now()

	tx := db.db.Begin()

	tx.Model(&author).UpdateColumn("balance", gorm.Expr("balance - ?", order.Fee))

	tx.Where(&User{ID: author.ID}).First(&author)
	if author.Balance < 0 {
		tx.Rollback()
		return fmt.Errorf("Not enough money for placing the order")
	}

	tx.Create(&order)
	rdb := tx.Commit()

	return rdb.Error
}

func (db *DB) ListUsers() (list []User) {
	db.db.Find(&list)
	return list
}

func (db *DB) AddUser(user User) error {
	user.Created = time.Now()
	rdb := db.db.Create(&user)

	return rdb.Error
}

func (db *DB) Close() error {
	return db.db.Close()
}
