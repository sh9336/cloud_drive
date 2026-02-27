// internal/validator/validator.go
package validator

import (
	"fmt"
	"mime"
	"path/filepath"
)

func ValidateFileUpload(filename string, fileSize int64, mimeType string, maxSize int64, allowedTypes []string) error {
	// Check file size
	if fileSize > maxSize {
		return fmt.Errorf("file size %d exceeds maximum allowed size %d", fileSize, maxSize)
	}

	if fileSize <= 0 {
		return fmt.Errorf("invalid file size")
	}

	// Check MIME type
	if !isAllowedMimeType(mimeType, allowedTypes) {
		return fmt.Errorf("file type %s is not allowed", mimeType)
	}

	// Check file extension matches MIME type
	ext := filepath.Ext(filename)
	if ext != "" {
		expectedMime := mime.TypeByExtension(ext)
		if expectedMime != "" && expectedMime != mimeType {
			return fmt.Errorf("file extension does not match MIME type")
		}
	}

	return nil
}

func isAllowedMimeType(mimeType string, allowedTypes []string) bool {
	for _, allowed := range allowedTypes {
		if mimeType == allowed {
			return true
		}
	}
	return false
}
