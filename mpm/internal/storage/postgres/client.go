package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"mpm/config"
)

// Client представляет клиент для работы с PostgreSQL
type Client struct {
	db     *sql.DB
	config *config.PostgresConfig
}

// NewClient создает новое подключение к PostgreSQL
func NewClient(cfg *config.PostgresConfig) (*Client, error) {
	// Формирование строки подключения
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Username, cfg.Password, cfg.Database, cfg.SSLMode)

	// Добавление схемы, если указана
	if cfg.Schema != "" && cfg.Schema != "public" {
		dsn += fmt.Sprintf(" search_path=%s", cfg.Schema)
	}

	// Открытие соединения
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	// Проверка подключения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Printf("Успешное подключение к PostgreSQL: %s:%s/%s", cfg.Host, cfg.Port, cfg.Database)

	client := &Client{
		db:     db,
		config: cfg,
	}

	return client, nil
}

// GetDB возвращает экземпляр базы данных
func (c *Client) GetDB() *sql.DB {
	return c.db
}

// Close закрывает соединение с базой данных
func (c *Client) Close() error {
	if c.db != nil {
		if err := c.db.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
		log.Println("Подключение к PostgreSQL закрыто")
	}
	return nil
}

// HealthCheck проверяет состояние подключения
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.db.PingContext(ctx)
}

// BeginTx начинает транзакцию
func (c *Client) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return c.db.BeginTx(ctx, nil)
}

// ExecContext выполняет запрос без возврата результата
func (c *Client) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if c.config.EnableQueryLog {
		log.Printf("PostgreSQL Exec: %s", query)
	}
	return c.db.ExecContext(ctx, query, args...)
}

// QueryContext выполняет запрос с возвратом результата
func (c *Client) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if c.config.EnableQueryLog {
		log.Printf("PostgreSQL Query: %s", query)
	}
	return c.db.QueryContext(ctx, query, args...)
}

// QueryRowContext выполняет запрос с возвратом одной строки
func (c *Client) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if c.config.EnableQueryLog {
		log.Printf("PostgreSQL QueryRow: %s", query)
	}
	return c.db.QueryRowContext(ctx, query, args...)
}

// Migrate выполняет миграции базы данных
func (c *Client) Migrate() error {
	driver, err := postgres.WithInstance(c.db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", c.config.MigrationsPath),
		c.config.Database,
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// Выполнение миграций
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Миграции PostgreSQL успешно выполнены")
	return nil
}
