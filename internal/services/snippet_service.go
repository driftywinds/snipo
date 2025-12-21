package services

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/validation"
)

// Common errors
var (
	ErrSnippetNotFound = errors.New("snippet not found")
	ErrValidation      = errors.New("validation error")
)

// SnippetService handles snippet business logic
type SnippetService struct {
	repo               *repository.SnippetRepository
	tagRepo            *repository.TagRepository
	folderRepo         *repository.FolderRepository
	fileRepo           *repository.SnippetFileRepository
	logger             *slog.Logger
	maxFilesPerSnippet int
}

// NewSnippetService creates a new snippet service
func NewSnippetService(repo *repository.SnippetRepository, logger *slog.Logger) *SnippetService {
	return &SnippetService{
		repo:               repo,
		logger:             logger,
		maxFilesPerSnippet: 10, // Default
	}
}

// WithTagRepo adds tag repository to the service
func (s *SnippetService) WithTagRepo(tagRepo *repository.TagRepository) *SnippetService {
	s.tagRepo = tagRepo
	return s
}

// WithFolderRepo adds folder repository to the service
func (s *SnippetService) WithFolderRepo(folderRepo *repository.FolderRepository) *SnippetService {
	s.folderRepo = folderRepo
	return s
}

// WithFileRepo adds file repository to the service
func (s *SnippetService) WithFileRepo(fileRepo *repository.SnippetFileRepository) *SnippetService {
	s.fileRepo = fileRepo
	return s
}

// WithMaxFiles sets the maximum files per snippet
func (s *SnippetService) WithMaxFiles(max int) *SnippetService {
	s.maxFilesPerSnippet = max
	return s
}

// Create creates a new snippet
func (s *SnippetService) Create(ctx context.Context, input *models.SnippetInput) (*models.Snippet, error) {
	// Validate input
	if errs := validation.ValidateSnippetInput(input); errs.HasErrors() {
		return nil, errs
	}

	snippet, err := s.repo.Create(ctx, input)
	if err != nil {
		s.logger.Error("failed to create snippet", "error", err)
		return nil, err
	}

	// Set tags if provided
	if s.tagRepo != nil && len(input.Tags) > 0 {
		if err := s.tagRepo.SetSnippetTags(ctx, snippet.ID, input.Tags); err != nil {
			s.logger.Warn("failed to set snippet tags", "id", snippet.ID, "error", err)
		} else {
			// Fetch tags to include in response
			tags, _ := s.tagRepo.GetSnippetTags(ctx, snippet.ID)
			snippet.Tags = tags
		}
	}

	// Set folder if provided
	if s.folderRepo != nil && input.FolderID != nil {
		if err := s.folderRepo.SetSnippetFolder(ctx, snippet.ID, input.FolderID); err != nil {
			s.logger.Warn("failed to set snippet folder", "id", snippet.ID, "error", err)
		} else {
			// Fetch folders to include in response
			folders, _ := s.folderRepo.GetSnippetFolders(ctx, snippet.ID)
			snippet.Folders = folders
		}
	}

	// Create files if provided
	if s.fileRepo != nil && len(input.Files) > 0 {
		// Limit files
		files := input.Files
		if len(files) > s.maxFilesPerSnippet {
			files = files[:s.maxFilesPerSnippet]
		}
		createdFiles, err := s.fileRepo.SyncFiles(ctx, snippet.ID, files)
		if err != nil {
			s.logger.Warn("failed to create snippet files", "id", snippet.ID, "error", err)
		} else {
			snippet.Files = createdFiles
		}
	}

	s.logger.Info("snippet created", "id", snippet.ID, "title", snippet.Title)
	return snippet, nil
}

// GetByID retrieves a snippet by ID
func (s *SnippetService) GetByID(ctx context.Context, id string) (*models.Snippet, error) {
	snippet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("failed to get snippet", "id", id, "error", err)
		return nil, err
	}

	if snippet == nil {
		return nil, ErrSnippetNotFound
	}

	// Fetch tags
	if s.tagRepo != nil {
		tags, _ := s.tagRepo.GetSnippetTags(ctx, id)
		snippet.Tags = tags
	}

	// Fetch folders
	if s.folderRepo != nil {
		folders, _ := s.folderRepo.GetSnippetFolders(ctx, id)
		snippet.Folders = folders
	}

	// Fetch files
	if s.fileRepo != nil {
		files, _ := s.fileRepo.GetBySnippetID(ctx, id)
		snippet.Files = files
	}

	return snippet, nil
}

// GetByIDPublic retrieves a public snippet by ID and increments view count
func (s *SnippetService) GetByIDPublic(ctx context.Context, id string) (*models.Snippet, error) {
	snippet, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if snippet == nil || !snippet.IsPublic {
		return nil, ErrSnippetNotFound
	}

	// Increment view count asynchronously
	go func() {
		if err := s.repo.IncrementViewCount(context.Background(), id); err != nil {
			s.logger.Warn("failed to increment view count", "id", id, "error", err)
		}
	}()

	// Fetch files for public view
	if s.fileRepo != nil {
		files, _ := s.fileRepo.GetBySnippetID(ctx, id)
		snippet.Files = files
	}

	return snippet, nil
}

// Update updates an existing snippet
func (s *SnippetService) Update(ctx context.Context, id string, input *models.SnippetInput) (*models.Snippet, error) {
	// Validate input
	if errs := validation.ValidateSnippetInput(input); errs.HasErrors() {
		return nil, errs
	}

	// Check if snippet exists
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrSnippetNotFound
	}

	snippet, err := s.repo.Update(ctx, id, input)
	if err != nil {
		s.logger.Error("failed to update snippet", "id", id, "error", err)
		return nil, err
	}

	// Update tags if provided
	if s.tagRepo != nil && input.Tags != nil {
		if err := s.tagRepo.SetSnippetTags(ctx, id, input.Tags); err != nil {
			s.logger.Warn("failed to update snippet tags", "id", id, "error", err)
		}
		tags, _ := s.tagRepo.GetSnippetTags(ctx, id)
		snippet.Tags = tags
	}

	// Update folder if provided
	if s.folderRepo != nil {
		if err := s.folderRepo.SetSnippetFolder(ctx, id, input.FolderID); err != nil {
			s.logger.Warn("failed to update snippet folder", "id", id, "error", err)
		}
		folders, _ := s.folderRepo.GetSnippetFolders(ctx, id)
		snippet.Folders = folders
	}

	// Update files if provided
	if s.fileRepo != nil && input.Files != nil {
		// Limit files
		files := input.Files
		if len(files) > s.maxFilesPerSnippet {
			files = files[:s.maxFilesPerSnippet]
		}
		syncedFiles, err := s.fileRepo.SyncFiles(ctx, id, files)
		if err != nil {
			s.logger.Warn("failed to update snippet files", "id", id, "error", err)
		} else {
			snippet.Files = syncedFiles
		}
	}

	s.logger.Info("snippet updated", "id", id)
	return snippet, nil
}

// Delete removes a snippet
func (s *SnippetService) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrSnippetNotFound
		}
		s.logger.Error("failed to delete snippet", "id", id, "error", err)
		return err
	}

	s.logger.Info("snippet deleted", "id", id)
	return nil
}

// List retrieves snippets with filtering and pagination
func (s *SnippetService) List(ctx context.Context, filter models.SnippetFilter) (*models.SnippetListResponse, error) {
	// Apply defaults
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}
	if filter.SortBy == "" {
		filter.SortBy = "updated_at"
	}
	if filter.SortOrder == "" {
		filter.SortOrder = "desc"
	}

	return s.repo.List(ctx, filter)
}

// ToggleFavorite toggles the favorite status of a snippet
func (s *SnippetService) ToggleFavorite(ctx context.Context, id string) (*models.Snippet, error) {
	snippet, err := s.repo.ToggleFavorite(ctx, id)
	if err != nil {
		s.logger.Error("failed to toggle favorite", "id", id, "error", err)
		return nil, err
	}

	if snippet == nil {
		return nil, ErrSnippetNotFound
	}

	s.logger.Info("snippet favorite toggled", "id", id, "is_favorite", snippet.IsFavorite)
	return snippet, nil
}

// ToggleArchive toggles the archive status of a snippet
func (s *SnippetService) ToggleArchive(ctx context.Context, id string) (*models.Snippet, error) {
	snippet, err := s.repo.ToggleArchive(ctx, id)
	if err != nil {
		s.logger.Error("failed to toggle archive", "id", id, "error", err)
		return nil, err
	}

	if snippet == nil {
		return nil, ErrSnippetNotFound
	}

	s.logger.Info("snippet archive toggled", "id", id, "is_archived", snippet.IsArchived)
	return snippet, nil
}

// Search performs full-text search on snippets
func (s *SnippetService) Search(ctx context.Context, query string, limit int) ([]models.Snippet, error) {
	if query == "" {
		return []models.Snippet{}, nil
	}

	snippets, err := s.repo.Search(ctx, query, limit)
	if err != nil {
		s.logger.Error("failed to search snippets", "query", query, "error", err)
		return nil, err
	}

	return snippets, nil
}

// Duplicate creates a copy of an existing snippet
func (s *SnippetService) Duplicate(ctx context.Context, id string) (*models.Snippet, error) {
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrSnippetNotFound
	}

	input := &models.SnippetInput{
		Title:       existing.Title + " (copy)",
		Description: existing.Description,
		Content:     existing.Content,
		Language:    existing.Language,
		IsPublic:    false, // Copies are private by default
	}

	return s.repo.Create(ctx, input)
}
