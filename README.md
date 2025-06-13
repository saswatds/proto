# Proto CLI Tool

A command-line tool for managing and syncing Protocol Buffer files from Git repositories.

## Installation

You can install the tool in several ways:

### Latest Version
```bash
go install github.com/saswatds/proto/cmd/proto@latest
```

### Specific Version
```bash
go install github.com/saswatds/proto/cmd/proto@v0.1.1
```

### From Source
```bash
git clone https://github.com/saswatds/proto.git
cd proto
go install ./cmd/proto
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/). The current version is v0.1.1.

- v0.1.1: Added help command
  - Comprehensive help message with command descriptions
  - Better error messages with usage suggestions
  - Improved user experience for new users

- v0.1.0: Initial release
  - Basic proto file management
  - Git repository integration
  - Go and Python SDK generation

## Prerequisites

- Go 1.16 or later
- Git
- Protocol Buffers compiler (protoc)
- Go protobuf plugin (`go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`)
- Python protobuf plugin (`pip install protobuf`)

## Usage

### Initialize Configuration

```bash
proto init --url <github-repo-url> [--branch <branch-name>] [--remote-path <path>] [--proto-dir <proto-dir>] [--build-dir <build-dir>]
```

Example:
```bash
proto init \
  --url https://github.com/example/proto-files \
  --branch main \
  --remote-path api/proto \
  --proto-dir ./proto \
  --build-dir ./gen
```

### Sync Proto Files

```bash
proto sync
```

This command will:
1. Check if there are any new changes in the repository
2. Download and sync the proto files from the specified path to the proto directory if changes are detected
3. Update the last commit ID in the configuration

### Build SDKs

```bash
proto build [go|python]
```

Example:
```bash
proto build go    # Build Go SDK in the build directory
proto build python  # Build Python SDK in the build directory
```

## Configuration

The tool stores its configuration in `~/.protorc` with the following YAML structure:

```yaml
github_url: https://github.com/example/proto-files
branch: main
remote_path: api/proto
proto_dir: ./proto
build_dir: ./gen
last_commit_id: abc123...
```

## Directory Structure

The tool maintains separate directories for different purposes:
- `proto_dir`: Contains the synced .proto files from the repository
- `build_dir`: Contains all generated SDK files (both Go and Python)

The `remote_path` parameter allows you to specify a subdirectory within the repository where the proto files are located. For example, if your proto files are in the `api/proto` directory of your repository, you would set `remote_path: api/proto`.

## Error Handling

The tool provides clear error messages for common issues:
- Missing configuration
- Git repository access issues
- Protocol buffer compilation errors
- File system permission issues
- No proto files found in the specified path

## Testing

Run the tests using:

```bash
go test ./...
```

The test suite includes:
- Configuration file handling
- Command-line argument parsing
- Git repository operations
- Protocol buffer compilation