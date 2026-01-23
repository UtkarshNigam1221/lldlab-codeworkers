# Code Workers - Secure Code Execution Service

A lightweight, secure code execution service with Docker sandboxing, queue management, and test case evaluation.

## Features

- **Docker Sandboxing**: Isolated execution environment with resource limits
- **Multi-Language Support**: Python, C++, Java, JavaScript, TypeScript, Go
- **Queue System**: In-memory queue with concurrent workers
- **Test Case Evaluation**: Automatic test case validation
- **Resource Limits**: Configurable time and memory constraints
- **REST API**: Simple HTTP endpoints for submission and status checking
- **Dependency Injection**: Uber FX framework for clean architecture
- **Strategy Pattern**: Extensible language execution strategies

## Architecture

- **Queue**: Go channels for lightweight, in-memory job queue
- **Workers**: Concurrent goroutines processing submissions in parallel
- **Sandbox**: Alpine-based Docker container with security restrictions
- **Runner**: Executes code with timeout, memory limits, and network isolation
- **FX Lifecycle**: Managed startup/shutdown with dependency injection
- **Strategy Registry**: Pluggable language execution strategies

## Setup

### Prerequisites

- Docker
- Docker Compose
- Go 1.23+ (for local development)

### Quick Start

1. **Build the sandbox image:**
```bash
make build-sandbox
```

2. **Start the service:**
```bash
make run
```

The service will be available at `http://localhost:8080`

### Local Development

```bash
go mod download
make dev
```

## API Endpoints

### Submit Code

**POST** `/api/submit`

```json
{
  "language": "python",
  "code": "n = int(input())\nprint(n * 2)",
  "test_cases": [
    {"input": "5", "expected": "10"},
    {"input": "10", "expected": "20"}
  ],
  "time_limit": 5,
  "memory_limit": 256
}
```

**Response:**
```json
{
  "submission_id": "uuid-here",
  "status": "QUEUED"
}
```

### Submit Code (Synchronous)

**POST** `/api/submit/sync`

Same request format as `/api/submit`, but blocks until execution completes and returns results immediately.

**Response:**
```json
{
  "submission_id": "uuid-here",
  "status": "PASSED",
  "test_results": [
    {
      "passed": true,
      "output": "10\n",
      "error": "",
      "time": 0.123
    }
  ]
}
```

### Check Status

**GET** `/api/status/{submission_id}`

**Response:**
```json
{
  "submission_id": "uuid-here",
  "status": "PASSED",
  "test_results": [
    {
      "passed": true,
      "output": "10\n",
      "error": "",
      "time": 0.123
    }
  ]
}
```

### Health Check

**GET** `/health`

**Response:**
```
OK
```

## Supported Languages

- **Python** (python3)
- **C++** (g++)
- **Java** (OpenJDK 17)
- **JavaScript** (Node.js)
- **TypeScript** (ts-node)
- **Go** (go run)

## Security Features

- Network isolation (`--network none`)
- Memory limits (configurable, default 256MB)
- CPU limits (1 CPU core)
- Process limits (max 50 processes)
- Non-root user execution
- Timeout enforcement

## Configuration

Default configuration in `internal/providers/providers.go`:

- **Port**: `:8080` - API server port
- **Workers**: `5` - Number of concurrent workers
- **Queue Capacity**: `100` - Maximum queue capacity
- **Time Limit**: `5` seconds (default per submission)
- **Memory Limit**: `256` MB (default per submission)

Docker resource limits in `internal/executor/executor.go`:

- **CPU Limit**: `1.0` core
- **Process Limit**: `50` processes
- **Network**: Isolated (none)

## Status Codes

- `QUEUED`: Submission received and waiting
- `PROCESSING`: Currently being executed
- `PASSED`: All test cases passed
- `FAILED`: One or more test cases failed

## Example Usage

See `scripts/test_api.sh` for example API calls or run `make test`.

## Testing

### Run Tests

**Unit tests only:**
```bash
make test-unit
```

**Integration tests (requires Docker sandbox):**
```bash
make test-integration
```

**All tests:**
```bash
make test-all
```

**With coverage report:**
```bash
make coverage
```

### Code Quality

**Lint code:**
```bash
make lint
```

**Auto-fix linting issues:**
```bash
make lint-fix
```

**Format code:**
```bash
make fmt
```

**Static analysis:**
```bash
make vet
make staticcheck
```

**Security scan:**
```bash
make security
```

**Run all CI checks:**
```bash
make ci
```

### Test Coverage

The project includes:
- **Unit tests** for all core packages (queue, executor, strategy, runner, API handlers)
- **Integration tests** for end-to-end API workflows
- **Race detection** enabled in all test runs
- **Coverage reporting** with HTML visualization

View coverage report:
```bash
make coverage
open coverage.html
```

### Linting Configuration

The project uses `golangci-lint` with comprehensive checks:
- Error checking (errcheck)
- Code simplification (gosimple)
- Static analysis (staticcheck)
- Security scanning (gosec)
- Code formatting (gofmt, goimports)
- Cyclomatic complexity checks (gocyclo)
- Duplicate code detection (dupl)
- And 15+ other linters

Configuration: `.golangci.yml`

## Development

### Project Structure

```
codeworkers/
├── cmd/
│   └── worker/
│       └── main.go              # Entry point with FX setup
├── internal/
│   ├── models/
│   │   └── models.go            # Data structures
│   ├── queue/
│   │   └── queue.go             # Queue implementation
│   ├── executor/
│   │   ├── executor.go          # Docker execution logic
│   │   └── strategy.go          # Language execution strategies
│   ├── runner/
│   │   └── runner.go            # Worker pool & test case evaluation
│   ├── providers/
│   │   └── providers.go         # FX dependency providers
│   └── api/
│       └── handlers.go          # HTTP API handlers
├── docker/
│   ├── Dockerfile               # Worker service image
│   └── Dockerfile.sandbox       # Code execution sandbox
├── scripts/
│   └── test_api.sh              # API testing script
├── go.mod                       # Go module definition
├── go.sum                       # Go dependencies
├── docker-compose.yml           # Service orchestration
├── Makefile                     # Build commands
└── README.md                    # Documentation
```

### Adding New Languages

1. Create a new strategy struct implementing `LanguageStrategy` in `internal/executor/strategy.go`
2. Register the strategy in `NewStrategyRegistry()` function
3. Install compiler/runtime in `docker/Dockerfile.sandbox`
4. Test with example code

**Example:**
```go
type RustStrategy struct{}

func (s *RustStrategy) GetFilename() string {
    return "solution.rs"
}

func (s *RustStrategy) BuildRunCommand(code, input string) string {
    return fmt.Sprintf("echo '%s' > %s && rustc %s && echo '%s' | ./solution",
        escapeString(code), s.GetFilename(), s.GetFilename(), escapeString(input))
}

// Register in NewStrategyRegistry()
"rust": &RustStrategy{},
```

