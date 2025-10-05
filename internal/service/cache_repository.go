package service

import (
	"fmt"
	"log"

	"github.com/valkey-io/valkey-glide/go/api"
)

// ValkeyClient es nuestra interfaz común que funciona para ambos modos
type ValkeyClient interface {
	Set(key string, value string) (string, error)
	Get(key string) (api.Result[string], error)
	Del(keys []string) (int64, error)
	Exists(keys []string) (int64, error)
	Expire(key string, seconds int64) (bool, error)
	Ping() (string, error)
	Close()
}

// RepositoryConfig contiene la configuración del repository
type RepositoryConfig struct {
	Host           string
	Port           int
	Password       string
	IsCluster      bool
	RequestTimeout int
	ClientName     string
}

// CacheRepository maneja las operaciones con Valkey
type CacheRepository struct {
	client ValkeyClient // ← Usamos nuestra interfaz
}

// NewCacheRepository crea un nuevo repositorio conectado a Valkey
func NewCacheRepository(config RepositoryConfig) (*CacheRepository, error) {
	if config.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	var client ValkeyClient
	var err error

	client, err = createClusterClient(config)

	if err != nil {
		return nil, err
	}

	// Probar conexión
	result, err := client.Ping()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("error en ping a Valkey: %w", err)
	}

	mode := "Cluster"

	log.Printf("✅ Repository conectado a Valkey %s en %s:%d (PING: %s, TLS: enabled)",
		mode, config.Host, config.Port, result)

	return &CacheRepository{
		client: client,
	}, nil
}

// createClusterClient crea un cliente para modo Cluster con TLS
func createClusterClient(config RepositoryConfig) (ValkeyClient, error) {
	clientConfig := api.NewGlideClusterClientConfiguration().
		WithAddress(&api.NodeAddress{Host: config.Host, Port: config.Port}).
		WithCredentials(api.NewServerCredentialsWithDefaultUsername(config.Password)).
		WithUseTLS(true)

	if config.RequestTimeout > 0 {
		clientConfig = clientConfig.WithRequestTimeout(config.RequestTimeout)
	}

	if config.ClientName != "" {
		clientConfig = clientConfig.WithClientName(config.ClientName)
	}

	return api.NewGlideClusterClient(clientConfig)
}

// Set guarda un valor con TTL opcional
func (r *CacheRepository) Set(key string, value string, ttlSeconds int64) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	_, err := r.client.Set(key, value)
	if err != nil {
		return fmt.Errorf("error en Set: %w", err)
	}

	if ttlSeconds > 0 {
		_, err = r.client.Expire(key, ttlSeconds)
		if err != nil {
			return fmt.Errorf("error configurando TTL: %w", err)
		}
	}

	return nil
}

// Get obtiene un valor por key
func (r *CacheRepository) Get(key string) (string, bool, error) {
	if key == "" {
		return "", false, fmt.Errorf("key cannot be empty")
	}

	result, err := r.client.Get(key)
	if err != nil {
		return "", false, nil
	}

	if result.IsNil() {
		return "", false, nil
	}

	value := result.Value()

	if value == "" {
		return "", false, nil
	}

	return value, true, nil
}

// Delete elimina una key
func (r *CacheRepository) Delete(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	_, err := r.client.Del([]string{key})
	if err != nil {
		return fmt.Errorf("error en Delete: %w", err)
	}

	return nil
}

// Exists verifica si una key existe
func (r *CacheRepository) Exists(key string) (bool, error) {
	if key == "" {
		return false, fmt.Errorf("key cannot be empty")
	}

	count, err := r.client.Exists([]string{key})
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
