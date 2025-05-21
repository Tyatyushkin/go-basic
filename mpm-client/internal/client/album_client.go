package client

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "mpm-client/proto/albums"
	"time"
)

// AlbumClient представляет клиент для работы с альбомами через gRPC
type AlbumClient struct {
	client pb.AlbumServiceClient
	conn   *grpc.ClientConn
}

// NewAlbumClient создает новый клиент для работы с альбомами
func NewAlbumClient(serverAddr string) (*AlbumClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Хотя DialContext помечен как устаревший, он все еще работает
	// и это простейший способ использовать контекст для таймаута
	conn, err := grpc.DialContext(ctx,
		serverAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к серверу: %w", err)
	}

	client := pb.NewAlbumServiceClient(conn)

	return &AlbumClient{
		client: client,
		conn:   conn,
	}, nil
}

// Close закрывает соединение с gRPC сервером
func (c *AlbumClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// GetAlbums получает список всех альбомов
func (c *AlbumClient) GetAlbums(ctx context.Context) ([]*pb.Album, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	response, err := c.client.GetAlbums(ctx, &pb.GetAlbumsRequest{})
	if err != nil {
		return nil, fmt.Errorf("ошибка при получении альбомов: %w", err)
	}

	return response.Albums, nil
}

// CreateAlbum создает новый альбом
func (c *AlbumClient) CreateAlbum(ctx context.Context, name, description string) (*pb.Album, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	response, err := c.client.CreateAlbum(ctx, &pb.CreateAlbumRequest{
		Name:        name,
		Description: description,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка при создании альбома: %w", err)
	}

	return response, nil
}
