package database

import (
	"github.com/liteseed/transit/internal/database/schema"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func New(url string) (*Database, error) {
	db, err := gorm.Open(sqlite.Open(url), &gorm.Config{
		CreateBatchSize: 200,
		Logger:          logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	c := &Database{DB: db}
	err = c.Migrate()
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Database) Migrate() error {
	err := c.DB.AutoMigrate(&schema.Order{}, &schema.Bundler{})
	return err
}

func (c *Database) CreateOrder(o *schema.Order) error {
	return c.DB.Create(&o).Error
}

func (c *Database) GetOrders(o *schema.Order, scopes ...Scope) (*[]schema.Order, error) {
	orders := &[]schema.Order{}
	err := c.DB.Scopes(scopes...).Where(o).Limit(25).Find(&orders).Error
	return orders, err
}

func (c *Database) UpdateOrder(o *schema.Order) error {
	return c.DB.Updates(o).Error
}

func (c *Database) UpdateOrders(orders *[]schema.Order, scopes ...Scope) error {
	return c.DB.Updates(orders).Error
}

func (c *Database) DeleteOrder(id string) error {
	return c.DB.Delete(&schema.Order{ID: id}).Error
}

func (c *Database) Shutdown() error {
	db, err := c.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}



func (c *Database) AssignBundler(b *schema.Bundler) error {
	return c.DB.Create(&b).Error
}

type Scope = func(*gorm.DB) *gorm.DB

// Scope for filtering records where confirmations is greater than 25
func ConfirmationsGreaterThanEqualTo25(db *gorm.DB) *gorm.DB {
	return db.Where("confirmations >= ?", 25)
}

// Scope for filtering records where confirmations is greater than 25
func ConfirmationsLessThan25(db *gorm.DB) *gorm.DB {
	return db.Where("confirmations < ?", 25)
}
