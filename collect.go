package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/pkoukk/tiktoken-go"
)

var totalTokens int

const maxTotalTokens = 50000
const maxFileSize = 1 * 1024 * 1024

var encoder, err = tiktoken.EncodingForModel("gpt-4o")

func init() {
	if err != nil {
		fmt.Println("Error initializing tokenizer:", err)
		os.Exit(1)
	}
}

func isIgnored(path string, ignorePatterns []string) bool {
	for _, pattern := range ignorePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			continue
		}
		if matched || strings.Contains(path, pattern) {
			return true
		}
	}
	return false
}

func isIncluded(path string, includePatterns []string) bool {
	if len(includePatterns) == 0 {
		return true
	}
	for _, pattern := range includePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err != nil {
			continue
		}
		if matched || strings.HasSuffix(path, pattern) {
			return true
		}
	}
	return false
}

func isBinaryFile(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()

	buf := make([]byte, 8000)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		return false, err
	}

	for i := 0; i < n; i++ {
		if buf[i] == 0 {
			return true, nil
		}
	}
	return false, nil
}

func countTokens(text string) int {
	tokens := encoder.Encode(text, nil, nil)
	return len(tokens)
}

func copyToClipboard(text string) {
	var cmd *exec.Cmd
	if _, err := exec.LookPath("pbcopy"); err == nil {
		cmd = exec.Command("pbcopy")
	} else if _, err := exec.LookPath("xclip"); err == nil {
		cmd = exec.Command("xclip", "-selection", "clipboard")
	} else {
		fmt.Println("Clipboard copy not supported on this platform.")
		return
	}
	in, _ := cmd.StdinPipe()
	cmd.Start()
	in.Write([]byte(text))
	in.Close()
	cmd.Wait()
}

func parseGitignore(rootDir string) ([]string, error) {
	var gitignorePatterns []string
	gitignorePath := filepath.Join(rootDir, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		if os.IsNotExist(err) {
			return gitignorePatterns, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		gitignorePatterns = append(gitignorePatterns, line)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return gitignorePatterns, nil
}

func collectFilesContent(rootDir string, includePatterns, ignorePatterns []string) (string, string) {
	var collectedContent strings.Builder
	var files []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relativePath, _ := filepath.Rel(rootDir, path)

		if d.IsDir() {
			if isIgnored(relativePath, ignorePatterns) {
				return filepath.SkipDir
			}
			return nil
		}

		if isIgnored(relativePath, ignorePatterns) {
			return nil
		}

		if !isIncluded(relativePath, includePatterns) {
			return nil
		}

		files = append(files, path)
		return nil
	})

	if err != nil {
		fmt.Println("Error:", err)
	}

	var wg sync.WaitGroup
	maxGoroutines := 10
	sem := make(chan struct{}, maxGoroutines)

	mu := &sync.Mutex{}

	for _, path := range files {
		mu.Lock()
		if totalTokens >= maxTotalTokens {
			mu.Unlock()
			fmt.Println("Reached maximum token limit.")
			break
		}
		mu.Unlock()

		wg.Add(1)
		sem <- struct{}{}
		go func(path string) {
			defer wg.Done()
			defer func() { <-sem }()

			content, tokenCount, err := processFile(path, rootDir)
			if err != nil {
				fmt.Printf("Error processing file %s: %s\n", path, err)
				return
			}

			mu.Lock()
			if totalTokens+tokenCount <= maxTotalTokens {
				collectedContent.WriteString(content)
				totalTokens += tokenCount
			} else {
				fmt.Printf("Skipping file %s to stay within token limit.\n", path)
			}
			mu.Unlock()
		}(path)
	}

	wg.Wait()

	fileTree := buildFileTree(files, rootDir)

	return fileTree, collectedContent.String()
}

func processFile(path, rootDir string) (string, int, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", 0, fmt.Errorf("Error stating file %s: %s", path, err)
	}
	relativePath, _ := filepath.Rel(rootDir, path)
	if info.Size() > maxFileSize {
		fmt.Printf("Skipping large file (>1MB): %s\n", relativePath)
		return "", 0, nil
	}

	isBinary, err := isBinaryFile(path)
	if err != nil {
		return "", 0, fmt.Errorf("Error checking if file is binary: %s", err)
	}
	if isBinary {
		fmt.Printf("Skipping binary file: %s\n", relativePath)
		return "", 0, nil
	}

	file, err := os.Open(path)
	if err != nil {
		return "", 0, fmt.Errorf("Error opening file %s: %s", relativePath, err)
	}
	defer file.Close()

	var fileContent strings.Builder
	fileContent.WriteString(fmt.Sprintf("File: %s\n", relativePath))

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fileContent.WriteString(scanner.Text() + "\n")
	}
	if err := scanner.Err(); err != nil {
		return "", 0, fmt.Errorf("Error reading file %s: %s", relativePath, err)
	}
	fileContent.WriteString("\n")

	tokenCount := countTokens(fileContent.String())

	return fileContent.String(), tokenCount, nil
}

func buildFileTree(files []string, rootDir string) string {
	var builder strings.Builder
	for _, path := range files {
		relativePath, _ := filepath.Rel(rootDir, path)
		builder.WriteString(relativePath + "\n")
	}
	return builder.String()
}

func main() {
	includePtr := flag.String("include", "", "Comma-separated list of file extensions or patterns to include (e.g., .go,.txt).")
	ignorePtr := flag.String("ignore", "", "Comma-separated list of patterns to ignore.")
	parseGitignorePtr := flag.Bool("gitignore", true, "Parse .gitignore files to exclude patterns.")
	flag.Parse()

	includePatterns := strings.Split(*includePtr, ",")
	if *includePtr == "" {
		includePatterns = []string{}
	}

	userIgnorePatterns := strings.Split(*ignorePtr, ",")
	if *ignorePtr == "" {
		userIgnorePatterns = []string{}
	}

	rootDir := "."

	defaultIgnorePatterns := []string{
		".git", ".svn", ".hg",
		"node_modules", "venv", "env", "__pycache__", "target", "bin", "obj",
		"build", "dist", "out",
		".idea", ".vscode", ".settings",
		"*.log", "*.tmp", "*.swp",
		"*.exe", "*.dll", "*.so", "*.bin", "*.class", "*.jar", "*.war",
		"*.jpg", "*.jpeg", "*.png", "*.gif", "*.mp3", "*.mp4",
		"*.zip", "*.tar", "*.gz", "*.7z", "*.rar",
		"_build", "site",
	}

	ignorePatterns := append(defaultIgnorePatterns, userIgnorePatterns...)

	if *parseGitignorePtr {
		gitignorePatterns, err := parseGitignore(rootDir)
		if err != nil {
			fmt.Printf("Error parsing .gitignore: %s\n", err)
		} else {
			ignorePatterns = append(ignorePatterns, gitignorePatterns...)
		}
	}

	fileTree, collectedContent := collectFilesContent(rootDir, includePatterns, ignorePatterns)
	totalContent := fmt.Sprintf("File Tree:\n%s\n\nContents:\n%s", fileTree, collectedContent)

	copyToClipboard(totalContent)
	fmt.Printf("Total tokens used: %d\n", totalTokens)

}


