# Easy CLI

A production-grade CLI tool for automated SaaS project deployments. This tool provisions and configures infrastructure across AWS S3, PostgreSQL databases, DigitalOcean Apps, and Vercel projects in a single command.

## Features

- **Multi-cloud deployment**: Provisions resources across AWS, DigitalOcean, and Vercel
- **Structured logging**: Comprehensive logging with contextual information  
- **Error handling**: Built-in retry logic and rollback mechanisms
- **Input validation**: Validates all input parameters before execution
- **Secure configuration**: Environment-based configuration management
- **Production-ready**: Organized codebase with clear separation of concerns

## Prerequisites

- Go 1.24.4 or later
- AWS Access Key ID and Secret Access Key
- DigitalOcean API token
- Vercel API token and team ID
- PostgreSQL database access

## Quick Start

Install Easy CLI to your system:

```bash
curl -sSL https://raw.githubusercontent.com/CaioDGallo/easy-cli/main/install.sh | bash
```

This installs the `easy-cli` binary system-wide and creates a template environment file at `~/.easy-cli.env`.

## Installation

### One-line Install (Recommended)

Install Easy CLI to `/usr/local/bin` (requires sudo for installation):

```bash
curl -sSL https://raw.githubusercontent.com/CaioDGallo/easy-cli/main/install.sh | bash
```

Or install to a custom directory:

```bash
curl -sSL https://raw.githubusercontent.com/CaioDGallo/easy-cli/main/install.sh | bash -s -- --install-dir /usr/local/bin
```

### Manual Installation

Download the latest release from [GitHub Releases](https://github.com/CaioDGallo/easy-cli/releases):

```bash
# Download for Linux/amd64
curl -fsSL https://github.com/CaioDGallo/easy-cli/releases/latest/download/easy-cli_linux_amd64.tar.gz | tar -xz
sudo mv easy-cli /usr/local/bin/
sudo chmod +x /usr/local/bin/easy-cli
```

### Development Setup

For development:

```bash
git clone https://github.com/CaioDGallo/easy-cli
cd easy-cli
make dev-setup  # Creates .env from .env.example
make build
```

## Configuration

### Environment Variables

| Variable | Description | Required |
|----------|-------------|----------|
| `DB_PASSWORD` | PostgreSQL database password | ✅ |
| `VERCEL_TOKEN` | Vercel API token | ✅ |
| `VERCEL_TEAM_ID` | Vercel team ID | ✅ |
| `VERCEL_FRONTEND_REPO_UUID` | Frontend repository UUID | ✅ |
| `DO_TOKEN` | DigitalOcean API token | ✅ |
| `AWS_ACCESS_KEY_ID` | AWS access key ID | ✅ |
| `AWS_SECRET_ACCESS_KEY` | AWS secret access key | ✅ |
| `DB_HOST` | Database host | ❌ (default provided) |
| `DB_USER` | Database username | ❌ (default: postgres) |
| `AWS_REGION` | AWS region | ❌ (default: us-east-1) |

### Environment File Locations

Easy CLI looks for environment configuration in the following order:

1. **`~/.easy-cli.env`** (recommended - created by installer)
2. **Same directory as binary** (e.g., `/usr/local/bin/.env`)
3. **Current working directory** (`.env`)

### Setup Environment

Edit the environment file created by the installer:

```bash
nano ~/.easy-cli.env
```

Or create it manually:

```bash
# Database Configuration
DB_PASSWORD=your_database_password_here

# Vercel Configuration  
VERCEL_TOKEN=your_vercel_token_here
VERCEL_TEAM_ID=your_vercel_team_id_here
VERCEL_FRONTEND_REPO_UUID={e4839c5c-d412-4c9d-88f7-c6209fef4b6a}

# AWS Configuration
AWS_ACCESS_KEY_ID=your_aws_access_key_id_here
AWS_SECRET_ACCESS_KEY=your_aws_secret_access_key_here

# DigitalOcean Configuration
DO_TOKEN=your_digitalocean_token_here
```

## Usage

After installation and environment setup, deploy a new client:

```bash
easy-cli fresh-install --client-name "My Client"
```

The CLI will automatically find and load your environment configuration from `~/.easy-cli.env`.

### Command Options

```bash
easy-cli fresh-install --help
```

All SMTP parameters have sensible defaults and are optional. You can override them:

```bash
easy-cli fresh-install --client-name "My Client" \
  --smtp-server "mail.example.com" \
  --smtp-username "user@example.com"
```

This command will:

1. **Validate** all input parameters
2. **Create AWS S3 bucket** with proper encryption and public access configuration
3. **Set up PostgreSQL databases** (main and Hangfire) from templates
4. **Deploy DigitalOcean app** with backend service and environment variables
5. **Create Vercel project** with frontend configuration and environment variables

### Command Options

| Flag | Short | Description | Default |
|------|-------|-------------|---------|
| `--client-name` | `-c` | Client name (required) | - |
| `--smtp-server` | `-s` | SMTP server | Configured via `SMTP_SERVER` env var |
| `--smtp-port` | `-P` | SMTP port | `587` |
| `--smtp-username` | `-u` | SMTP username | Configured via `SMTP_USERNAME` env var |
| `--smtp-password` | `-p` | SMTP password | `your-smtp-password` |
| `--smtp-donotreplyname` | `-r` | Do not reply name | `Do Not Reply` |
| `--smtp-donotreplyemail` | `-m` | Do not reply email | Configured via `SMTP_DO_NOT_REPLY_EMAIL` env var |
| `--smtp-devemail` | `-e` | Developer email | Configured via `SMTP_DEV_EMAIL` env var |

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Install to GOPATH
make install
```

### Testing

```bash
# Run tests
make test

# Format code
make fmt

# Run vet
make vet

# Run linter (requires golangci-lint)
make lint
```

### Development Workflow

```bash
# Quick build and run
make dev
```

## Project Structure

```
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   └── fresh-install.go   # Fresh install command
├── internal/              # Internal packages
│   ├── aws/               # AWS S3 service
│   ├── config/            # Configuration management
│   ├── database/          # PostgreSQL service
│   ├── digitalocean/      # DigitalOcean app service
│   ├── envvars/           # Environment variable generation
│   ├── interfaces/        # Service interfaces
│   ├── logger/            # Structured logging
│   ├── retry/             # Retry logic utilities
│   ├── rollback/          # Rollback mechanisms
│   ├── types/             # Type definitions
│   ├── utils/             # Utility functions
│   ├── validation/        # Input validation
│   └── vercel/            # Vercel project service
├── .env.example           # Example environment file
├── .gitignore            # Git ignore rules
├── Makefile              # Build automation
├── README.md             # This file
├── go.mod                # Go module definition
└── main.go               # Application entry point
```

## Error Handling

The CLI includes comprehensive error handling:

- **Retry Logic**: Automatic retries for transient failures with exponential backoff
- **Rollback Mechanisms**: Automatic cleanup of partially created resources on failure
- **Input Validation**: Pre-flight validation of all input parameters
- **Structured Logging**: Detailed logging with context for debugging

## Security

- **No Hardcoded Credentials**: All sensitive data loaded from environment variables
- **Input Validation**: All user inputs are validated and sanitized
- **Least Privilege**: Services only have access to required resources
- **Secure Defaults**: Security-first configuration for all cloud resources

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and linting
5. Submit a pull request

## License

This project is licensed under the terms specified in the LICENSE file.

## Support

For issues and questions, please use the GitHub issue tracker.