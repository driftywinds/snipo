package validation

import (
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/MohamedElashri/snipo/internal/models"
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return ""
	}
	var msgs []string
	for _, e := range ve {
		msgs = append(msgs, e.Field+": "+e.Message)
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are validation errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Allowed languages (subset of Prism.js supported languages)
var allowedLanguages = map[string]bool{
	"plaintext": true, "javascript": true, "typescript": true, "python": true,
	"go": true, "rust": true, "java": true, "c": true, "cpp": true, "csharp": true,
	"php": true, "ruby": true, "swift": true, "kotlin": true, "scala": true,
	"html": true, "css": true, "scss": true, "json": true, "yaml": true, "xml": true,
	"markdown": true, "sql": true, "bash": true, "shell": true, "powershell": true,
	"dockerfile": true, "nginx": true, "toml": true, "ini": true, "makefile": true,
	"lua": true, "perl": true, "r": true, "haskell": true, "elixir": true,
	"clojure": true, "graphql": true, "protobuf": true, "terraform": true,
}

// tagRegex validates tag names
var tagRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// ValidateSnippetInput validates snippet input
func ValidateSnippetInput(input *models.SnippetInput) ValidationErrors {
	var errs ValidationErrors

	// Title validation
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		errs = append(errs, ValidationError{Field: "title", Message: "Title is required"})
	} else if utf8.RuneCountInString(input.Title) > 200 {
		errs = append(errs, ValidationError{Field: "title", Message: "Title must be less than 200 characters"})
	}

	// Content validation
	if strings.TrimSpace(input.Content) == "" {
		errs = append(errs, ValidationError{Field: "content", Message: "Content is required"})
	} else if len(input.Content) > 1024*1024 { // 1MB limit
		errs = append(errs, ValidationError{Field: "content", Message: "Content must be less than 1MB"})
	}

	// Language validation
	input.Language = strings.ToLower(strings.TrimSpace(input.Language))
	if input.Language == "" {
		input.Language = "plaintext"
	} else if !allowedLanguages[input.Language] {
		errs = append(errs, ValidationError{Field: "language", Message: "Invalid language"})
	}

	// Description length
	input.Description = strings.TrimSpace(input.Description)
	if utf8.RuneCountInString(input.Description) > 500 {
		errs = append(errs, ValidationError{Field: "description", Message: "Description must be less than 500 characters"})
	}

	// Tag validation
	for i, tag := range input.Tags {
		tag = strings.TrimSpace(tag)
		input.Tags[i] = tag
		if tag == "" {
			continue
		}
		if len(tag) > 50 {
			errs = append(errs, ValidationError{Field: "tags", Message: "Tag name must be less than 50 characters"})
		} else if !tagRegex.MatchString(tag) {
			errs = append(errs, ValidationError{Field: "tags", Message: "Tag can only contain letters, numbers, underscores, and hyphens"})
		}
	}

	return errs
}

// GetAllowedLanguages returns a list of allowed language identifiers
func GetAllowedLanguages() []string {
	languages := make([]string, 0, len(allowedLanguages))
	for lang := range allowedLanguages {
		languages = append(languages, lang)
	}
	return languages
}
