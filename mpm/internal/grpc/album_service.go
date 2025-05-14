package grpc

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"mpm/internal/repository"
	pb "mpm/proto/albums"
	"time"
)

type AlbumServer struct {
	pb.UnimplementedAlbumServiceServer
	repository *repository.Repository
}

func NewAlbumServer(repo *repository.Repository) *AlbumServer {
	return &AlbumServer{
		repository: repo,
	}
}

func (s *AlbumServer) GetAlbums(ctx context.Context, req *pb.GetAlbumsRequest) (*pb.GetAlbumsResponse, error) {
	albums, err := s.repository.GetAllAlbums(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка получения списка альбомов: %v", err)
	}

	result := &pb.GetAlbumsResponse{
		Albums: make([]*pb.Album, 0, len(albums)),
	}

	for _, album := range albums {
		result.Albums = append(result.Albums, &pb.Album{
			Id:          int32(album.ID),
			Name:        album.Name,
			Description: album.Description,
			CreatedAt:   album.CreatedAt.Format(time.RFC3339),
		})
	}
	return result, nil
}
