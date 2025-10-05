package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/Ruggilock/redis/internal/service"
	pb "github.com/Ruggilock/redis/proto"
	"google.golang.org/grpc"
)

func main() {
	// Leer configuración desde variables de entorno
	valkeyHost := getEnv("VALKEY_HOST", "localhost")
	valkeyPort := getEnvAsInt("VALKEY_PORT", 6379)
	valkeyPassword := os.Getenv("VALKEY_PASSWORD") // Obligatorio

	// Validar password
	if valkeyPassword == "" {
		log.Fatal("❌ VALKEY_PASSWORD es requerido")
	}

	// Crear configuración del repository
	repoConfig := service.RepositoryConfig{
		Host:           valkeyHost,
		Port:           valkeyPort,
		Password:       valkeyPassword,
		RequestTimeout: 5000,
		ClientName:     "grpc-cache-service",
	}

	// Crear repository
	repo, err := service.NewCacheRepository(repoConfig)
	if err != nil {
		log.Fatalf("❌ Error creando repository: %v", err)
	}
	defer repo.Close()

	// Crear listener gRPC
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("❌ Error: %v", err)
	}

	log.Println("🚀 Servidor gRPC en puerto 50051")

	// Crear servidor gRPC
	grpcServer := grpc.NewServer()

	// Crear y registrar servicio
	cacheService := service.NewCacheServer(repo)
	pb.RegisterCacheServiceServer(grpcServer, cacheService)

	// Graceful shutdown
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		log.Println("\n🛑 Deteniendo servidor...")
		grpcServer.GracefulStop()
	}()

	// Iniciar servidor
	log.Println("✅ Servidor listo")
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("❌ Error: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
