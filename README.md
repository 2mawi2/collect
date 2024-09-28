# Collect Files Content Utility

This utility recursively scans the current directory, collects the content of files, and copies the combined content along with a file tree structure to your clipboard. It reports the total number of tokens used, utilizing the [`tiktoken-go`](https://github.com/pkoukk/tiktoken-go) library for token counting. You can specify which files to include or ignore through command-line arguments.

## Features

- **Recursive Scanning**: Walks through directories starting from the current location.
- **File Inclusion/Exclusion**: Supports specifying patterns to include or ignore.
- **Token Counting**: Ensures the collected content doesn't exceed a token limit.
- **Clipboard Copying**: Automatically copies the collected content to your clipboard.
- **File Tree Generation**: Generates a file tree structure of the collected files.
- **Concurrency**: Efficient processing using goroutines.

## Installation

### Prerequisites

- **Go Language**: Version 1.16 or higher
- **Git**: To clone the repository

### macOS

1. **Install Go** (if not already installed):

   ```bash
   brew install go
   ```

2. **Clone the Repository**:

   ```bash
   git clone https://github.com/yourusername/collect-files-content.git
   cd collect-files-content
   ```

3. **Build the Executable**:

   ```bash
   go build -o collect
   ```

4. **Move Executable to PATH**:

   ```bash
   sudo mv collect /usr/local/bin/
   ```

5. **Make Executable Globally Accessible**:

   If you're using **Fish Shell** or any other shell, ensure `/usr/local/bin` is in your `PATH`. For Fish Shell:

   ```fish
   set -Ua fish_user_paths /usr/local/bin
   ```

### Windows

1. **Install Go**:

   Download and install Go from the [official website](https://golang.org/dl/).

2. **Clone the Repository**:

   ```cmd
   git clone https://github.com/yourusername/collect-files-content.git
   cd collect-files-content
   ```

3. **Build the Executable**:

   ```cmd
   go build -o collect.exe
   ```

4. **Add Executable to PATH**:

   Move `collect.exe` to a directory that's in your `PATH`, or add the directory containing `collect.exe` to your `PATH` environment variable.

## Usage

Run the `collect` command in the directory you want to scan.

```bash
collect [options]
```

### Options

- `-include`: **(Optional)** Comma-separated list of file extensions or patterns to include.

  Example:

  ```bash
  collect -include=".go,.txt"
  ```

- `-ignore`: **(Optional)** Comma-separated list of patterns to ignore.

  Example:

  ```bash
  collect -ignore="testdata,*.md"
  ```

- `-gitignore`: **(Optional)** Parse `.gitignore` files to exclude patterns. Defaults to `true`. Set to `false` to ignore `.gitignore`.

  ```bash
  collect -gitignore=false
  ```

### Example Commands

- **Collect all files**:

  ```bash
  collect
  ```

- **Include only specific file types**:

  ```bash
  collect -include=".go,.md,.txt"
  ```

- **Ignore specific directories or files**:

  ```bash
  collect -ignore="vendor,node_modules,*.test.go"
  ```

- **Do not parse `.gitignore`**:

  ```bash
  collect -gitignore=false
  ```

## How It Works

1. **Scanning**: The script walks through the current directory recursively, respecting the include and ignore patterns provided.

2. **File Processing**:

   - Skips directories and files matching ignore patterns.
   - Includes files matching the include patterns.
   - Skips binary files and files larger than 1 MB.
   - Reads file content and accumulates tokens using `tiktoken-go`.

3. **Token Counting**:

   - Uses `tiktoken-go` to tokenize file content.
   - Ensures the total tokens do not exceed `50,000`.

4. **Content Collection**:

   - Builds a string containing the file tree and contents.
   - Formats each file with its relative path and content.

5. **Copy to Clipboard**:

   - Copies the collected content to the system clipboard.
   - Supports both macOS (`pbcopy`) and Linux (`xclip`).

6. **Output**:

   - Prints the total number of tokens used.
   - Alerts if the token limit is reached or files are skipped.

## Notes

- **No External Dependencies**: Aside from Go and `tiktoken-go`, no additional installations are required.
- **Clipboard Support**:

  - On macOS, `pbcopy` is used (which is available by default).
  - On Windows, clipboard copying is not implemented in this script. Integration can be added if needed.
  - On Linux systems, `xclip` is required. Install it via your package manager.

## Troubleshooting

- **Clipboard Not Working**:

  - Ensure `pbcopy` (macOS) or `xclip` (Linux) is installed and accessible.
  - For Linux, install `xclip`:

    ```bash
    sudo apt-get install xclip
    ```

- **Token Limit Reached**:

  - Adjust the `maxTotalTokens` constant in the script if you need a higher limit.
  - Include fewer files or more specific patterns.

- **Binary Files Detected as Text**:

  - Ensure that binary files have appropriate extensions or are properly detected.
  - Modify the `isBinaryFile` function if necessary.

## Customization

- **Change Token Limit**:

  Modify the `maxTotalTokens` constant at the top of the script.

  ```go
  const maxTotalTokens = 50000 // Adjust as needed
  ```

- **Adjust Max File Size**:

  Modify the `maxFileSize` constant.

  ```go
  const maxFileSize = 1 * 1024 * 1024 // 1 MB
  ```

- **Default Ignore Patterns**:

  Update the `defaultIgnorePatterns` slice with any additional patterns you wish to ignore by default.

## Contributing

Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License

This project is licensed under the MIT License.

---

**Disclaimer**: Ensure you comply with your organization's policies and any relevant laws when collecting and copying file content.