package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"mpm/config"
)

// Client обертка для MongoDB клиента с конфигурацией
type Client struct {
	client *mongo.Client
	db     *mongo.Database
	config *config.MongoDBConfig
}

// NewClient создает новое подключение к MongoDB
func NewClient(cfg *config.MongoDBConfig) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ConnectTimeout)
	defer cancel()

	// Настройка опций клиента
	clientOptions := options.Client()

	// Используем URI если он задан, иначе формируем из отдельных параметров
	if cfg.URI != "" {
		clientOptions.ApplyURI(cfg.URI)
	} else {
		connectionString := fmt.Sprintf("mongodb://%s:%s", cfg.Host, cfg.Port)
		clientOptions.ApplyURI(connectionString)
	}

	// Настройка пула соединений
	clientOptions.SetMaxPoolSize(cfg.MaxPoolSize)
	clientOptions.SetMinPoolSize(cfg.MinPoolSize)
	clientOptions.SetMaxConnIdleTime(cfg.MaxIdleTime)
	clientOptions.SetConnectTimeout(cfg.ConnectTimeout)

	// Подключение к MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Проверка подключения
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Получение базы данных
	db := client.Database(cfg.Database)

	log.Printf("Успешное подключение к MongoDB: %s/%s", cfg.Host, cfg.Database)

	mongoClient := &Client{
		client: client,
		db:     db,
		config: cfg,
	}

	// Инициализация индексов
	if err := mongoClient.ensureIndexes(); err != nil {
		log.Printf("Предупреждение: не удалось создать индексы: %v", err)
	}

	return mongoClient, nil
}

// GetDatabase возвращает экземпляр базы данных
func (c *Client) GetDatabase() *mongo.Database {
	return c.db
}

// GetCollection возвращает коллекцию по имени
func (c *Client) GetCollection(name string) *mongo.Collection {
	return c.db.Collection(name)
}

// GetAlbumsCollection возвращает коллекцию альбомов
func (c *Client) GetAlbumsCollection() *mongo.Collection {
	return c.GetCollection(c.config.Collections.Albums)
}

// GetUsersCollection возвращает коллекцию пользователей
func (c *Client) GetUsersCollection() *mongo.Collection {
	return c.GetCollection(c.config.Collections.Users)
}

// GetPhotosCollection возвращает коллекцию фотографий
func (c *Client) GetPhotosCollection() *mongo.Collection {
	return c.GetCollection(c.config.Collections.Photos)
}

// GetTagsCollection возвращает коллекцию тегов
func (c *Client) GetTagsCollection() *mongo.Collection {
	return c.GetCollection(c.config.Collections.Tags)
}

// GetCommentsCollection возвращает коллекцию комментариев
func (c *Client) GetCommentsCollection() *mongo.Collection {
	return c.GetCollection(c.config.Collections.Comments)
}

// ensureIndexes создает индексы для всех коллекций
func (c *Client) ensureIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Индексы для коллекции albums
	albumsCol := c.GetAlbumsCollection()
	albumIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "tags", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "name", Value: 1}, {Key: "user_id", Value: 1}},
		},
	}

	if _, err := albumsCol.Indexes().CreateMany(ctx, albumIndexes); err != nil {
		return fmt.Errorf("failed to create albums indexes: %w", err)
	}

	// Индексы для коллекции users (подготовка на будущее)
	usersCol := c.GetUsersCollection()
	userIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "email", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := usersCol.Indexes().CreateMany(ctx, userIndexes); err != nil {
		return fmt.Errorf("failed to create users indexes: %w", err)
	}

	// Индексы для коллекции tags
	tagsCol := c.GetTagsCollection()
	tagIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}

	if _, err := tagsCol.Indexes().CreateMany(ctx, tagIndexes); err != nil {
		return fmt.Errorf("failed to create tags indexes: %w", err)
	}

	log.Println("MongoDB индексы успешно созданы")
	return nil
}

// Close закрывает подключение к MongoDB
func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	log.Println("Подключение к MongoDB закрыто")
	return nil
}

// HealthCheck проверяет состояние подключения
func (c *Client) HealthCheck() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return c.client.Ping(ctx, nil)
}
