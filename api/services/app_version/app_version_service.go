package app_version_service

import (
	"encoding/json"
	"errors"
	"os"
	"sync"
	"time"
)

const versionCacheTTL = time.Hour

type AppVersionService struct {
	metadataPath string

	mu             sync.Mutex
	cachedVersion  *AppVersionModel
	cacheExpiresAt time.Time
}

func NewAppVersionService(metadataPath string) *AppVersionService {
	return &AppVersionService{
		metadataPath: metadataPath,
	}
}

// GetCurrentVersion возвращает версию приложения из метаданных APK-сборки.
// Результат кешируется в памяти на один час.
func (srv *AppVersionService) GetCurrentVersion() (AppVersionModel, error) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if srv.cachedVersion != nil && time.Now().Before(srv.cacheExpiresAt) {
		return *srv.cachedVersion, nil
	}

	version, err := srv.readCurrentVersion()
	if err != nil {
		return AppVersionModel{}, err
	}

	srv.cachedVersion = &version
	srv.cacheExpiresAt = time.Now().Add(versionCacheTTL)

	return version, nil
}

func (srv *AppVersionService) readCurrentVersion() (AppVersionModel, error) {
	body, err := os.ReadFile(srv.metadataPath)
	if err != nil {
		return AppVersionModel{}, err
	}

	var metadata apkOutputMetadata
	if err := json.Unmarshal(body, &metadata); err != nil {
		return AppVersionModel{}, err
	}

	element := metadata.universalElement()
	if element == nil {
		return AppVersionModel{}, errors.New("apk metadata does not contain elements")
	}

	return AppVersionModel{
		VersionCode: element.VersionCode,
		VersionName: element.VersionName,
	}, nil
}

func (metadata *apkOutputMetadata) universalElement() *apkOutputMetadataElement {
	for i := range metadata.Elements {
		if metadata.Elements[i].Type == "UNIVERSAL" {
			return &metadata.Elements[i]
		}
	}

	if len(metadata.Elements) == 0 {
		return nil
	}

	return &metadata.Elements[0]
}
