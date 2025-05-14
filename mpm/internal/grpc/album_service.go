package grpc

import (
	pb "github.com/Tyatyushkin/mpm/proto/albums"
	"mpm/internal/storage"
)

type AlbumServer struct {
	pb.UnimplementedAlbumServiceServer
	albumStorage storage.AlbumStorage
}
