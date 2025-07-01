package services

import (
	"context"
	"fmt"
	"path/filepath"
	"server/config"
	"server/internal/api/dto"
	"server/internal/models"
	"server/internal/repository"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azdatalake/filesystem"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

type LogServices struct {
	Repo   repository.LogRepo
	Config config.AppConfig
}

func (s *LogServices) GetLogs(filters *dto.LogFilter) ([]*models.Log, int64, error) {
	filters.Project = "logs-" + filters.Project
	return s.Repo.GetLogs(filters)
}

func (s *LogServices) CreateProjectIndex(projectID string) error {
	return s.Repo.CreateProjectIndex(projectID)
}

func (s *LogServices) GetLogsMinMaxDate(projectName string) ([]string, error) {
	projectName = "logs-" + projectName
	return s.Repo.GetLogsAvailabilities(projectName)
}

func (s *LogServices) CheckIfIndexExists(indexName string) (bool, error) {
	indexName = "logs-" + indexName
	return s.Repo.CheckIfIndexExists(indexName)
}

func (s *LogServices) ListAllLogsFromStorage(projectName string) ([]string, error) {

	storageAccName := s.Config.AzureStorageAccountName
	fileSystemName := s.Config.ColdStorageContainer
	accountKey := s.Config.AzureStorageKey

	// Create the filesystem URL
	filesystemURL := fmt.Sprintf("https://%s.dfs.core.windows.net/%s", storageAccName, fileSystemName)

	if accountKey == "" {
		return nil, fmt.Errorf("account key not found in environment variable 'KEY'")
	}

	// Create shared key credential
	credential, err := azdatalake.NewSharedKeyCredential(storageAccName, accountKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create credential: %w", err)
	}

	// Create a filesystem client with a shared key
	filesystemClient, err := filesystem.NewClientWithSharedKeyCredential(filesystemURL, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create filesystem client: %w", err)
	}

	ctx := context.Background()

	var allFiles []string

	var pathPrefix *string
	if projectName != "" {

		prefixValue := projectName
		if !strings.HasSuffix(prefixValue, "/") {
			prefixValue += "/"
		}
		pathPrefix = &prefixValue
	}

	pager := filesystemClient.NewListPathsPager(true, &filesystem.ListPathsOptions{
		Prefix: pathPrefix,
	})

	pageCount := 0

	// Continue fetching pages
	for pager.More() {
		pageCount++
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list paths on page %d: %w", pageCount, err)
		}

		// Process each path in the current page
		for _, path := range page.PathList.Paths {
			if path.Name != nil {
				if path.IsDirectory == nil || (path.IsDirectory != nil && !*path.IsDirectory) {
					_, file := filepath.Split(*path.Name)
					allFiles = append(allFiles, file)
				}
			}
		}
	}
	return allFiles, nil
}

func (s *LogServices) GetLogsFromArchive(p string, fileName string, filter *dto.LogFilter) ([]*models.Log, int64, error) {
	return s.Repo.GetLogsFromArchiveStorage(p, fileName, filter)
}

func (s *LogServices) GetArchiveMinMaxDate(P string, fileName string) ([]string, error) {
	return s.Repo.GetArchiveLogMinMaxDate(P, fileName)
}
