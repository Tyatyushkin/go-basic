package grpc

import (
	"mpm/internal/repository"
	pb "mpm/proto/albums"
)

type AlbumServer struct {
	pb.UnimplementedAlbumServiceServer
	repository *repository.Repository
}
