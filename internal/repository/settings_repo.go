package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/MohamedElashri/snipo/internal/models"
)

// SettingsRepository handles settings database operations
type SettingsRepository struct {
	db *sql.DB
}

// NewSettingsRepository creates a new settings repository
func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

// Get retrieves application settings
func (r *SettingsRepository) Get(ctx context.Context) (*models.Settings, error) {
	query := `
		SELECT id, app_name, custom_css, theme, default_language, 
		       s3_enabled, s3_endpoint, s3_bucket, s3_region, 
		       backup_encryption_enabled, archive_enabled, created_at, updated_at
		FROM settings
		WHERE id = 1
	`

	settings := &models.Settings{}
	err := r.db.QueryRowContext(ctx, query).Scan(
		&settings.ID,
		&settings.AppName,
		&settings.CustomCSS,
		&settings.Theme,
		&settings.DefaultLanguage,
		&settings.S3Enabled,
		&settings.S3Endpoint,
		&settings.S3Bucket,
		&settings.S3Region,
		&settings.BackupEncryptionEnabled,
		&settings.ArchiveEnabled,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}

	return settings, nil
}

// Update updates application settings
func (r *SettingsRepository) Update(ctx context.Context, input *models.SettingsInput) (*models.Settings, error) {
	// First get existing settings to preserve fields if needed
	// But here we assume input covers what we want to update
	// Note: S3 secrets are environment variables usually, but stored in DB here?
	// The schema has s3_endpoint but not secret key. Wait.
	// Looking at schema:
	// s3_endpoint, s3_bucket, s3_region.
	// Where are keys?
	// Config.go loads keys from ENV.
	// Schema doesn't store keys.
	// `internal/config/config.go` says:
	// cfg.S3.AccessKeyID = os.Getenv("SNIPO_S3_ACCESS_KEY")
	// But `settings` table has s3 related columns.
	// "s3_endpoint TEXT DEFAULT '', s3_bucket TEXT DEFAULT '', s3_region TEXT DEFAULT 'us-east-1'"
	// It seems the app might be moving to DB-based settings or supports both?
	// Creating the repo based on the table schema.

	query := `
		UPDATE settings
		SET app_name = ?, custom_css = ?, theme = ?, default_language = ?,
		    s3_enabled = ?, s3_endpoint = ?, s3_bucket = ?, s3_region = ?,
		    backup_encryption_enabled = ?, archive_enabled = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
		RETURNING id, app_name, custom_css, theme, default_language,
		          s3_enabled, s3_endpoint, s3_bucket, s3_region,
		          backup_encryption_enabled, archive_enabled, created_at, updated_at
	`

	settings := &models.Settings{}
	err := r.db.QueryRowContext(ctx, query,
		input.AppName,
		input.CustomCSS,
		input.Theme,
		input.DefaultLanguage,
		input.S3Enabled,
		input.S3Endpoint,
		input.S3Bucket,
		input.S3Region,
		input.BackupEncryptionEnabled,
		input.ArchiveEnabled,
	).Scan(
		&settings.ID,
		&settings.AppName,
		&settings.CustomCSS,
		&settings.Theme,
		&settings.DefaultLanguage,
		&settings.S3Enabled,
		&settings.S3Endpoint,
		&settings.S3Bucket,
		&settings.S3Region,
		&settings.BackupEncryptionEnabled,
		&settings.ArchiveEnabled,
		&settings.CreatedAt,
		&settings.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update settings: %w", err)
	}

	return settings, nil
}
