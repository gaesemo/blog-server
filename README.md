# GSM Blog Server

A toy project for testing modern Go technologies including ConnectRPC, Testcontainers, and MinIO object storage.

## 🚀 Features

- **ConnectRPC**: Type-safe RPC communication with HTTP/JSON support
- **PostgreSQL**: Database with SQLC for type-safe queries
- **OAuth2 Authentication**: GitHub OAuth integration
- **JWT Tokens**: Secure authentication with JWT
- **MinIO Object Storage**: File upload and management
- **Testcontainers**: Integration testing with real databases
- **Microservices Architecture**: Versioned service structure

## 🏗️ Architecture

```
├── cmd/           # CLI commands (Cobra)
├── db/            # Database schema and queries
├── gen/           # Generated code (SQLC)
├── pkg/           # Shared packages
├── service/       # Microservices
│   ├── auth/v1/   # Authentication service
│   ├── user/v1/   # User management (stub)
│   ├── post/v1/   # Blog posts (stub)
│   └── object/v1/ # File storage (stub)
└── server/        # HTTP server
```

## 🛠️ Tech Stack

- **Language**: Go 1.21+
- **RPC Framework**: ConnectRPC
- **Database**: PostgreSQL with SQLC
- **Authentication**: OAuth2 (GitHub) + JWT
- **Object Storage**: MinIO
- **Testing**: Testcontainers
- **CLI**: Cobra
- **Package Management**: pkgx

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL
- MinIO (optional, for object storage)

### Environment Variables

```bash
# GitHub OAuth
GITHUB_OAUTH2_CLIENT_ID=your_client_id
GITHUB_OAUTH2_CLIENT_SECRET=your_client_secret
GITHUB_OAUTH2_REDIRECT_URL=http://your-application-url

# JWT
JWT_SIGNING_SECRET=your_secret_key
```

### Installation

```bash
# Clone the repository
git clone https://github.com/gaesemo/tech-blog-server.git
cd tech-blog-server

# Generate database code
./script/generate.sh

# Run the server
go run . serve
```

### Development Commands

```bash
# Generate type-safe database code
./script/generate.sh

# Run with custom port and debug logging
go run . serve --port 9000 --log-level debug

# Build the project
go build -o gsm .

# Run tests
go test ./...
```

## 📚 API Documentation

The server provides ConnectRPC services:

- **Auth Service** (`/service.auth.v1.AuthService/`)
  - `GetAuthURL` - Get OAuth authorization URL
  - `Login` - Exchange auth code for JWT token
  - `Logout` - Invalidate session

- **User Service** (`/service.user.v1.UserService/`) - *Coming Soon*
- **Post Service** (`/service.post.v1.PostService/`) - *Coming Soon*
- **Object Service** (`/service.object.v1.ObjectService/`) - *Coming Soon*

## 🧪 Testing

This project uses Testcontainers for integration testing with real PostgreSQL instances:

```bash
# Run all tests
go test ./...

# Run specific test
go test ./server -v
```

## 📁 Database Schema

The project uses PostgreSQL with the following core entities:

- **Users**: OAuth-authenticated users with soft delete support
- **Follows**: Social graph relationships (planned)
- **Subscriptions**: Paid subscription model (planned)

## 🤝 Contributing

This is a toy project for learning purposes. Feel free to explore, fork, and experiment!

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [ConnectRPC](https://connectrpc.com/) for the excellent RPC framework
- [SQLC](https://sqlc.dev/) for type-safe SQL
- [Testcontainers](https://testcontainers.com/) for integration testing
- [MinIO](https://min.io/) for object storage