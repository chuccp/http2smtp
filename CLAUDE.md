# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 1. Common Commands

### Backend Commands

#### Build Commands
- **Build for Linux**: `make release-linux`
- **Build for macOS**: `make release-osx`
- **Build directly**: `go build -o http2smtp ./`
- **Build Docker image**: `docker build -t http2smtp .`

#### Run the Application
- **Direct run**: `./http2smtp`
- **Run with custom ports**: `./http2smtp -web_port 12566 -api_port 12567`
- **Run with custom storage root**: `./http2smtp -storage_root /path/to/storage`
- **Run with Docker**: `docker run -p 12566:12566 -p 12567:12567 -it --rm http2smtp`
- **Run with Docker Compose**: `cd docker && docker compose up -d`

#### Testing
- **Run all tests**: `go test ./...`
- **Run specific test**: `go test ./util -run TestSchedule`

### Frontend Commands (React + Next.js + shadcn UI)

#### Directory
```
view/
```

#### Development
- **Install dependencies**: `cd view && npm install`
- **Run development server**: `cd view && npm run dev` (runs on http://localhost:3000)
- **Build for production**: `cd view && npm run build`
- **Run production server**: `cd view && npm start`
- **Lint code**: `cd view && npm run lint`

## 2. High-Level Architecture

HTTP2SMTP is a full-stack application that provides an HTTP-to-SMTP gateway service, allowing users to send emails via HTTP requests. The application consists of:
1. **Backend**: Go-based API server with custom web framework
2. **Frontend**: React + Next.js + shadcn UI admin panel

### Backend Architecture

Built with Go and a custom web framework (`go-web-frame`), following a layered architecture pattern.

#### Main Entry Point
- `main.go`: Initializes configuration, sets up web server, and registers REST endpoints

#### Key Backend Packages

- **`rest/`**: Contains REST API endpoints for management and operations:
  - `Set.go`: System settings
  - `User.go`: User management
  - `Token.go`: API token management
  - `Mail.go`: Email sending and templates
  - `Smtp.go`: SMTP server configuration
  - `API.go`: Public API for email sending

- **`model/`**: Data models and configuration structs
  - Contains the main `Config` struct that maps to `config.ini`

- **`service2/`**: Business logic services:
  - `TokenService.go`: Token validation and management
  - `ScheduleService.go`: Scheduled task management
  - `LogService.go`: Log management

- **`auth/`**: Authentication middleware
  - `authentication.go`: Cookie-based session authentication
  - `token.go`: API token authentication

- **`smtp/`**: SMTP client implementation
  - Handles actual email sending via SMTP

- **`schedule/`**: Scheduled task system
  - Uses cron expressions to schedule periodic email sending
  - Can fetch data from URLs and use templates to generate emails

- **`util/`**: Utility functions
  - Includes string handling, template rendering, mail utilities, schedule parsing, etc.

- **`config/`**: Configuration management

### Frontend Architecture

Modern web application built with Next.js 16, React 18, TypeScript, shadcn UI components, and Tailwind CSS.

#### Key Frontend Directories

- **`view/app/`**: Main application routes and components
  - **`/(dashboard)/`**: Authenticated dashboard pages
    - **`page.tsx`**: Dashboard homepage with system overview
    - **`smtp/`**: SMTP server management page
  - **`auth/`**: Authentication pages
    - **`login/`**: User login page
  - **`setup/`**: Initial system setup page
  - **`Providers.tsx`**: Authentication and state management providers

- **`view/components/`**: Reusable UI components
  - **`ui/`**: shadcn UI components (card, button, input, table, etc.)

- **`view/lib/`**: Utility functions and business logic
  - **`auth.ts`**: Authentication utilities and API client
  - **`utils.ts`**: General utility functions

- **`view/types/`**: TypeScript type definitions
  - **`auth.ts`**: Authentication-related types
  - **`smtp.ts`**: SMTP server types

### Architectural Patterns

1. **Layered Architecture**: Presentation (REST/UI) → Service (Business Logic) → Data Access (Model/DB) → External Services (SMTP)

2. **Dependency Injection**: Uses `go-web-frame` for dependency injection of services and models

3. **Configuration-Driven**: All settings are loaded from `config.ini` or environment variables

4. **Dual Server Setup** (Backend):
   - Management server (port 12566) with web UI and admin APIs
   - Public API server (port 12567) for email sending

5. **Extensible Authentication**: Uses filter pattern for authentication middleware with both cookie-based sessions and API tokens

6. **JAMstack Frontend**: Next.js with server-side rendering and static generation

### Architectural Patterns

1. **Layered Architecture**: Presentation (REST) → Service (Business Logic) → Data Access (Model/DB) → External Services (SMTP)

2. **Dependency Injection**: Uses `go-web-frame` for dependency injection of services and models

3. **Configuration-Driven**: All settings are loaded from `config.ini` or environment variables

4. **Dual Server Setup**:
   - Management server (port 12566) with web UI and admin APIs
   - Public API server (port 12567) for email sending

5. **Extensible Authentication**: Uses filter pattern for authentication middleware

### Key External Dependencies

#### Backend Dependencies
- **gin-gonic/gin**: HTTP web framework
- **wneessen/go-mail**: SMTP client library
- **robfig/cron/v3**: Scheduled task management
- **gorm.io/gorm**: ORM for database operations
- **zap**: Structured logging
- **go-web-frame**: Custom web framework for dependency injection and REST routing
- **gopkg.in/ini.v1**: INI configuration file parsing
- **github.com/google/uuid**: UUID generation

#### Frontend Dependencies
- **Next.js 16**: React framework with App Router
- **React 18**: UI library with hooks
- **TypeScript**: Type-safe JavaScript
- **shadcn UI**: Beautiful, reusable UI components
- **Tailwind CSS**: Utility-first CSS framework
- **Lucide React**: Beautiful open-source icons
- **Axios**: HTTP client for API calls

## 3. Application Features

### Core Features
- 🚀 Configure SMTP servers and recipient emails via Web UI
- 📦 Support for GET/POST/JSON request formats
- 🔒 Token-based API access control
- 📎 Multi-file attachment support (Base64 encoding/form upload)
- 🐳 Ready-to-use Docker image
- 📊 Send record query and statistics
- 📅 Scheduled task support for requesting links and sending emails
- 📧 Email template support

## 4. Important Workflow Notes

### Configuration
- First run generates `config.ini` with default settings
- Supports SQLite and MySQL databases
- Configuration can be overridden via command line flags or environment variables
- Frontend reads API endpoint from `NEXT_PUBLIC_API_URL` environment variable

### Authentication Flow

#### Cookie-Based Authentication (Web UI)
1. User accesses frontend at http://localhost:3000
2. Frontend checks if user is logged in by checking for authentication cookie
3. If not logged in, user is redirected to login page
4. User submits credentials
5. Frontend calculates signature using MD5 hashing
6. Backend validates signature and sets authentication cookie
7. User is redirected to dashboard

#### API Token Authentication (API Access)
1. User generates API token via web UI
2. Client includes token in request headers or query parameters
3. Backend validates token using AES encryption
4. Request is processed if token is valid

### Email Sending Flow
1. Client sends HTTP request to `/sendMail` endpoint
2. API validates authentication token (either cookie or API token)
3. Request is routed to appropriate handler
4. Email is rendered using templates (if configured)
5. SMTP client sends email via configured SMTP server
6. Transaction is logged to database

### Scheduled Tasks
1. Tasks configured via web UI with cron expressions
2. Schedule service runs periodic tasks
3. Tasks can fetch data from external URLs
4. Email content is generated using templates with fetched data
5. Email is sent via configured SMTP server

### Development Workflow

#### Backend Development
1. Make changes to Go code
2. Build and run the server: `go build -o http2smtp ./ && ./http2smtp`
3. Test endpoints using curl or frontend

#### Frontend Development
1. Navigate to view directory: `cd view`
2. Install dependencies: `npm install`
3. Run development server: `npm run dev`
4. Access frontend at http://localhost:3000
5. Changes are hot-reloaded in development


#### Full Stack Development
1. Start backend server on port 12567
2. Start frontend development server on port 3000
3. Frontend automatically proxies API requests to backend
4. Test full end-to-end flow

前端功能参考/Users/cao/Documents/GitHub/d-mail-view
接口功能参考 https://github.com/chuccp/http2smtp/tree/v0.2.10
go-web-frame 本地路径是 /Users/cao/Documents/GitHub/go-web-frame

lsof -ti :12567 | xargs -I {} kill -9 {}