package repositories

import (
	"encoding/json"
	"github.com/SHOWROOM-inc/recursive_mock_gen/models"
	"os"
)

type CacheRepository interface {
	ReadCache(cacheFilePath string) (models.Cache, error)
	WriteCache(cacheFilePath string, m models.Cache) error
}

func NewCacheRepository() CacheRepository {
	return &cacheRepository{}
}

type cacheRepository struct {
}

// ReadCache キャッシュファイルを読み込みます。もしキャッシュが無い場合は初期値を返します。
func (r *cacheRepository) ReadCache(cacheFilePath string) (models.Cache, error) {
	b, err := os.ReadFile(cacheFilePath)
	if err != nil {
		return models.Cache{}, nil
	}

	var m models.Cache
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}

func (r *cacheRepository) WriteCache(cacheFilePath string, m models.Cache) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(cacheFilePath, b, 0o644)
}
