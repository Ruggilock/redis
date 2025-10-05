package service

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"time"

	"github.com/valkey-io/valkey-go"
)

// RepositoryConfig contiene la configuración del repository
type RepositoryConfig struct {
	Host           string
	Port           int
	Password       string
	UseTLS         bool
	RequestTimeout int
	ClientName     string
}

// CacheRepository maneja las operaciones con Valkey
type CacheRepository struct {
	client valkey.Client
}

// NewCacheRepository crea un nuevo repositorio conectado a Valkey
func NewCacheRepository(config RepositoryConfig) (*CacheRepository, error) {
	if config.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	log.Printf("🔧 RepositoryConfig: %+v", config)

	// Construir opciones del cliente
	options := valkey.ClientOption{
		InitAddress: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Password:    config.Password,
		ClientName:  config.ClientName,
	}

	// Configurar TLS si está habilitado
	if config.UseTLS {
		options.TLSConfig = &tls.Config{
			MinVersion:         tls.VersionTLS12,
			InsecureSkipVerify: true, // Para ElastiCache con certificados autofirmados
		}
	}

	// Crear cliente
	client, err := valkey.NewClient(options)
	if err != nil {
		return nil, fmt.Errorf("error creando cliente: %w", err)
	}

	// Probar conexión con PING
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Do(ctx, client.B().Ping().Build()).Error(); err != nil {
		client.Close()
		return nil, fmt.Errorf("error en ping: %w", err)
	}

	log.Printf("✅ Repository conectado a Valkey Standalone en %s:%d (TLS: %v)",
		config.Host, config.Port, config.UseTLS)

	return &CacheRepository{
		client: client,
	}, nil
}

// Set guarda un valor con TTL opcional
func (r *CacheRepository) Set(key string, value string, ttlSeconds int64) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	ctx := context.Background()

	// Construir comando SET
	var cmd valkey.Completed
	if ttlSeconds > 0 {
		// SET key value EX seconds
		cmd = r.client.B().Set().Key(key).Value(value).ExSeconds(ttlSeconds).Build()
	} else {
		// SET key value
		cmd = r.client.B().Set().Key(key).Value(value).Build()
	}

	if err := r.client.Do(ctx, cmd).Error(); err != nil {
		return fmt.Errorf("error en Set: %w", err)
	}

	return nil
}

// Get obtiene un valor por key
func (r *CacheRepository) Get(key string) (string, bool, error) {
	if key == "" {
		return "", false, fmt.Errorf("key cannot be empty")
	}

	ctx := context.Background()

	// Ejecutar GET
	result := r.client.Do(ctx, r.client.B().Get().Key(key).Build())

	// Si es nil (key no existe), retornar found=false
	if valkey.IsValkeyNil(result.Error()) {
		return "", false, nil
	}

	// Si hay otro error
	if err := result.Error(); err != nil {
		return "", false, err
	}

	// Obtener el valor
	value, err := result.ToString()
	if err != nil {
		return "", false, err
	}

	return value, true, nil
}

// Delete elimina una key
func (r *CacheRepository) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	ctx := context.Background()

	if err := r.client.Do(ctx, r.client.B().Del().Key(key).Build()).Error(); err != nil {
		return fmt.Errorf("error en Delete: %w", err)
	}

	return nil
}

// Exists verifica si una key existe
func (r *CacheRepository) Exists(key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	ctx := context.Background()

	count, err := r.client.Do(ctx, r.client.B().Exists().Key(key).Build()).AsInt64()
	if err != nil {
		return false, fmt.Errorf("error en Exists: %w", err)
	}

	return count > 0, nil
}

// Close cierra la conexión
func (r *CacheRepository) Close() {
	if r.client != nil {
		r.client.Close()
	}
}
