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
go install github.com/saswatds/proto/cmd/proto@v0.9.3
```

### From Source
```bash
git clone https://github.com/saswatds/proto.git
cd proto
go install ./cmd/proto
```

## Versioning

This project follows [Semantic Versioning](https://semver.org/). The current version is v0.9.3.

For a detailed list of changes, see the [CHANGELOG.md](CHANGELOG.md).

## Prerequisites

- Go 1.16 or later
- Git
- Protocol Buffers compiler (protoc)
- Go protobuf plugins:
  ```bash
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
- Python protobuf and gRPC plugins:
  ```bash
  pip install protobuf grpcio grpcio-tools
  ```

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
3. Update the git head in the configuration

### Generate SDKs

```bash
proto gen [go|python]
```

Example:
```bash
proto gen go    # Generate Go SDK in the build directory
proto gen python  # Generate Python SDK in the build directory
```

## Configuration

The tool stores its configuration in `.protorc` in the current working directory with the following YAML structure:

```yaml
github_url: https://github.com/example/proto-files
branch: main
remote_path: api/proto  # Path within the repository containing proto files (quotes optional)
proto_dir: ./proto
build_dir: ./gen
gitHead: abc123...  # Latest commit ID from the repository
```

## Directory Structure

The tool maintains separate directories for different purposes:
- `proto_dir`: Contains the synced .proto files from the repository
- `build_dir`: Contains all generated SDK files (both Go and Python)

The `remote_path` parameter allows you to specify a subdirectory within the repository where the proto files are located. For example, if your proto files are in the `api/proto` directory of your repository, you would set `remote_path: api/proto`. The path can be specified with or without quotes, and both forward slashes and backslashes are supported.

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