package database

import (
	"github.com/liteseed/transit/internal/database/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	DB *gorm.DB
}

func New(url string) (*Database, error) {
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  url,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{
		CreateBatchSize: 200,
		Logger:          logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	c := &Database{DB: db}

	return c, nil
}

func FromDialector(d gorm.Dialector) (*Database, error) {
	db, err := gorm.Open(
		d,
		&gorm.Config{
			CreateBatchSize: 200,
			Logger:          logger.Default.LogMode(logger.Warn),
		},
	)
	if err != nil {
		return nil, err
	}
	c := &Database{DB: db}

	return c, nil
}

func (c *Database) Migrate() error {
	err := c.DB.AutoMigrate(&schema.Order{})
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

func (c *Database) GetOrder(id string) (*schema.Order, error) {
	order := &schema.Order{}
	err := c.DB.First(&order, "id = ?", id).Error
	return order, err
}

func (c *Database) UpdateOrder(id string, o *schema.Order) error {
	return c.DB.Model(&schema.Order{}).Where("id = ?", id).Updates(&o).Error
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

type Scope = func(*gorm.DB) *gorm.DB
