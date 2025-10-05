package service

import (
	"context"
	"log"

	pb "github.com/Ruggilock/redis/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CacheServer implementa el servicio gRPC
type CacheServer struct {
	pb.UnimplementedCacheServiceServer
	repo *CacheRepository // ← Usa el repository
}

// NewCacheServer crea el servidor gRPC con su repository
func NewCacheServer(repo *CacheRepository) *CacheServer {
	return &CacheServer{
		repo: repo,
	}
}

// Set - Maneja la petición gRPC y delega al repository
func (s *CacheServer) Set(ctx context.Context, req *pb.SetRequest) (*pb.SetResponse, error) {
	log.Printf("📝 gRPC Set: key=%s", req.Key)

	// Llamar al repository
	err := s.repo.Set(req.Key, req.Value, req.TtlSeconds)
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to set: %v", err)
	}

	return &pb.SetResponse{
		Success: true,
		Message: "OK",
	}, nil
}

// Get - Maneja la petición gRPC y delega al repository
func (s *CacheServer) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	log.Printf("🔍 gRPC Get: key=%s", req.Key)

	value, found, err := s.repo.Get(req.Key)
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to get: %v", err)
	}

	return &pb.GetResponse{
		Found: found,
		Value: value,
	}, nil
}

// Delete - Maneja la petición gRPC y delega al repository
func (s *CacheServer) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	log.Printf("🗑️  gRPC Delete: key=%s", req.Key)

	err := s.repo.Delete(req.Key)
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to delete: %v", err)
	}

	return &pb.DeleteResponse{
		Success: true,
	}, nil
}

// Exists - Maneja la petición gRPC y delega al repository
func (s *CacheServer) Exists(ctx context.Context, req *pb.ExistsRequest) (*pb.ExistsResponse, error) {
	log.Printf("❓ gRPC Exists: key=%s", req.Key)

	exists, err := s.repo.Exists(req.Key)
	if err != nil {
		log.Printf("❌ Error: %v", err)
		return nil, status.Errorf(codes.Internal, "failed to check existence: %v", err)
	}

	return &pb.ExistsResponse{
		Exists: exists,
	}, nil
}
