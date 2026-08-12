# Risk-Based Internal Audit System - Frontend

A modern web application built with Nuxt.js for managing enterprise risk and internal audit processes.

## Features

- 🔐 **Authentication System** - Complete login, register, and password reset flow
- 🌍 **Multi-language Support** - English and Indonesian translations
- 🎨 **Modern UI** - Built with Tailwind CSS and Nuxt UI
- 📊 **Risk Management** - Risk profiles, priority lists, and mitigation tracking
- 📝 **Audit Planning** - Strategic plans, annual plans, and audit activities
- 📈 **Dashboard & Analytics** - Real-time monitoring and reporting
- 🧪 **Testing** - Unit, integration, and E2E tests with Vitest and Playwright
- 🐳 **Docker Support** - Development and production Docker configurations

## Tech Stack

- **Framework:** Nuxt.js 3 with TypeScript
- **State Management:** Pinia
- **UI Components:** Nuxt UI + Tailwind CSS
- **Charts:** Chart.js with vue-chartjs
- **Date Handling:** date-fns
- **Testing:** Vitest, Playwright, Vue Test Utils
- **Validation:** Zod

## Prerequisites

- Node.js 20.x or higher
- npm or yarn
- Docker (optional, for containerized development)

## Getting Started

### 1. Install Dependencies

\`\`\`bash
npm install
# or
make install
\`\`\`

### 2. Environment Setup

Copy the example environment file and configure your settings:

\`\`\`bash
cp .env.example .env
\`\`\`

Edit `.env` with your configuration:

\`\`\`env
API_BASE_URL=http://localhost:3001/api
NODE_ENV=development
\`\`\`

### 3. Development Server

Start the development server:

\`\`\`bash
npm run dev
# or
make dev
\`\`\`

The application will be available at http://localhost:3000

## Project Structure

\`\`\`
frontend/
├── assets/              # Static assets (CSS, images, fonts)
├── components/          # Vue components organized by feature
│   ├── auth/
│   ├── audit-charter/
│   ├── risk-profile/
│   └── ...
├── composables/         # Vue composables (useI18n, etc.)
├── layouts/             # Layout components
├── locales/             # Translation files (en, id)
├── middleware/          # Route middleware (auth, guest)
├── pages/               # File-based routing pages
│   ├── auth/
│   ├── audit-charter/
│   └── ...
├── plugins/             # Nuxt plugins
├── public/              # Public static files
├── server/              # Server-side API routes and middleware
├── stores/              # Pinia stores
├── tests/               # Test files
│   ├── unit/
│   ├── integration/
│   └── e2e/
├── types/               # TypeScript type definitions
├── utils/               # Utility functions
├── app.vue              # Root application component
├── nuxt.config.ts       # Nuxt configuration
└── tailwind.config.ts   # Tailwind CSS configuration
\`\`\`

## Available Scripts

### Development

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run generate` - Generate static site

### Testing

- `npm run test` - Run all tests
- `npm run test:unit` - Run unit tests
- `npm run test:e2e` - Run E2E tests
- `npm run test:coverage` - Generate test coverage report

### Code Quality

- `npm run lint` - Run ESLint
- `npm run lint:fix` - Fix ESLint issues
- `npm run type-check` - Run TypeScript type checking

## Docker

### Development with Docker

\`\`\`bash
# Build and start development container
make docker-up

# View logs
make docker-logs

# Stop containers
make docker-down
\`\`\`

### Production with Docker

\`\`\`bash
# Build production image
make docker-build

# Start production containers
make docker-up-prod
\`\`\`

## Make Commands

Run `make help` to see all available commands:

\`\`\`bash
make help
\`\`\`

Common commands:
- `make install` - Install dependencies
- `make dev` - Start development server
- `make build` - Build for production
- `make test` - Run tests
- `make lint` - Run linter
- `make docker-up` - Start Docker containers
- `make ci` - Run all CI checks

## Key Features

### Authentication

Full authentication system with:
- Login / Register
- Password reset flow
- Protected routes with middleware
- Token-based authentication
- User session management

### Translation System

Multi-language support using a custom i18n composable:

\`\`\`typescript
const { t, locale, setLocale } = useI18n()

// Use in templates
{{ t('auth.login.title') }}

// Switch language
setLocale('id') // Switch to Indonesian
\`\`\`

### Type Utilities

Comprehensive type conversion utilities:

\`\`\`typescript
import { toNumber, toBoolean, formatDate } from '~/utils'

const num = toNumber('123') // 123
const bool = toBoolean('true') // true
const date = formatDate(new Date()) // 'yyyy-MM-dd'
\`\`\`

### State Management

Centralized state management with Pinia:

\`\`\`typescript
const authStore = useAuthStore()
await authStore.login({ email, password })
\`\`\`

## Testing

### Unit Tests

Located in `tests/unit/`. Example:

\`\`\`bash
npm run test:unit
\`\`\`

### E2E Tests

Located in `tests/e2e/`. Uses Playwright:

\`\`\`bash
npm run test:e2e
\`\`\`

## Deployment

### Building for Production

\`\`\`bash
npm run build
\`\`\`

### Using Docker

\`\`\`bash
docker build -t rbia-frontend .
docker run -p 3000:3000 rbia-frontend
\`\`\`

## Contributing

1. Follow the existing code structure
2. Write tests for new features
3. Run linting and type checking before committing
4. Use conventional commit messages


## Features
- Risk Management
- Audit Planning
- Audit Execution
- Audit Reporting

## License

Private - All rights reserved

