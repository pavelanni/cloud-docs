package main

import (
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/pavelanni/cloud-docs/internal/config"
	"github.com/pavelanni/cloud-docs/internal/storage"
)

type UploadStats struct {
	mu            sync.Mutex
	filesUploaded int
	filesSkipped  int
	totalBytes    int64
	errors        []error
}

func (s *UploadStats) AddUploaded(bytes int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesUploaded++
	s.totalBytes += bytes
}

func (s *UploadStats) AddSkipped() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.filesSkipped++
}

func (s *UploadStats) AddError(err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors = append(s.errors, err)
}

func (s *UploadStats) GetStats() (int, int, int64, []error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.filesUploaded, s.filesSkipped, s.totalBytes, append([]error{}, s.errors...)
}

type Config struct {
	SourceDir  string
	BucketName string
	Prefix     string
	Exclude    []string
	DryRun     bool
	Verbose    bool
}

func main() {
	var (
		sourceDir  = flag.String("source", ".", "Source directory to upload")
		bucketName = flag.String("bucket", "", "GCS bucket name (can also use BUCKET_NAME env var)")
		prefix     = flag.String("prefix", "", "Prefix to add to all uploaded files")
		exclude    = flag.String("exclude", "", "Comma-separated list of patterns to exclude")
		dryRun     = flag.Bool("dry-run", false, "Show what would be uploaded without actually uploading")
		verbose    = flag.Bool("verbose", false, "Verbose output")
	)
	flag.Parse()

	if *bucketName == "" {
		cfg := config.Load()
		*bucketName = cfg.BucketName
	}

	if *bucketName == "" {
		log.Fatal("Bucket name is required (use -bucket flag or BUCKET_NAME env var)")
	}

	uploadConfig := &Config{
		SourceDir:  *sourceDir,
		BucketName: *bucketName,
		Prefix:     *prefix,
		Exclude:    parseExcludePatterns(*exclude),
		DryRun:     *dryRun,
		Verbose:    *verbose,
	}

	if err := runUpload(uploadConfig); err != nil {
		log.Fatalf("Upload failed: %v", err)
	}
}

func parseExcludePatterns(exclude string) []string {
	defaultPatterns := []string{".git/*", ".DS_Store", "*.tmp", "*.log"}
	
	if exclude == "" {
		return defaultPatterns
	}
	
	customPatterns := strings.Split(exclude, ",")
	for i, pattern := range customPatterns {
		customPatterns[i] = strings.TrimSpace(pattern)
	}
	
	return append(defaultPatterns, customPatterns...)
}

func runUpload(cfg *Config) error {
	ctx := context.Background()

	storageClient, err := storage.NewClient(ctx, cfg.BucketName)
	if err != nil {
		return fmt.Errorf("failed to create storage client: %w", err)
	}
	defer storageClient.Close()

	fmt.Printf("Uploading from %s to gs://%s", cfg.SourceDir, cfg.BucketName)
	if cfg.Prefix != "" {
		fmt.Printf(" with prefix %s", cfg.Prefix)
	}
	fmt.Println()

	if cfg.DryRun {
		fmt.Println("DRY RUN - No files will be uploaded")
	}

	stats := &UploadStats{}
	startTime := time.Now()

	err = filepath.WalkDir(cfg.SourceDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(cfg.SourceDir, path)
		if err != nil {
			return err
		}

		if shouldExclude(relPath, cfg.Exclude) {
			if cfg.Verbose {
				fmt.Printf("Skipping: %s (excluded)\n", relPath)
			}
			stats.AddSkipped()
			return nil
		}

		objectPath := relPath
		if cfg.Prefix != "" {
			objectPath = filepath.Join(cfg.Prefix, relPath)
		}
		
		objectPath = filepath.ToSlash(objectPath)

		if cfg.Verbose || cfg.DryRun {
			fmt.Printf("Uploading: %s -> %s\n", relPath, objectPath)
		}

		if !cfg.DryRun {
			if err := uploadFile(ctx, storageClient, path, objectPath); err != nil {
				fmt.Printf("Error uploading %s: %v\n", relPath, err)
				stats.AddError(err)
				return nil
			}
		}

		fileInfo, err := os.Stat(path)
		if err != nil {
			return err
		}

		stats.AddUploaded(fileInfo.Size())
		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk directory: %w", err)
	}

	duration := time.Since(startTime)
	uploaded, skipped, totalBytes, errors := stats.GetStats()

	fmt.Printf("\nUpload complete:\n")
	fmt.Printf("  Files uploaded: %d\n", uploaded)
	fmt.Printf("  Files skipped: %d\n", skipped)
	fmt.Printf("  Total bytes: %d (%.2f MB)\n", totalBytes, float64(totalBytes)/1024/1024)
	fmt.Printf("  Duration: %v\n", duration.Round(time.Millisecond))
	if len(errors) > 0 {
		fmt.Printf("  Errors: %d\n", len(errors))
	}

	return nil
}

func shouldExclude(path string, patterns []string) bool {
	filename := filepath.Base(path)
	
	for _, pattern := range patterns {
		if matched, _ := filepath.Match(pattern, path); matched {
			return true
		}
		if matched, _ := filepath.Match(pattern, filename); matched {
			return true
		}
		if strings.HasSuffix(pattern, "/*") {
			prefix := strings.TrimSuffix(pattern, "/*")
			if strings.HasPrefix(path, prefix+"/") || path == prefix {
				return true
			}
		}
	}
	return false
}

func uploadFile(ctx context.Context, client *storage.Client, filePath, objectPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return client.UploadFile(ctx, objectPath, file, "")
}