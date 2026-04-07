<p align="center">
  <img src="https://raw.githubusercontent.com/kube-rca/.github/main/img/Kube-RCA-Logo-NoBG.png" alt="KubeRCA Logo" width="120"/>
</p>

<h1 align="center">KubeRCA Frontend</h1>

<p align="center">
  <strong>React Web Dashboard for Kubernetes Incident Management</strong>
</p>

<p align="center">
  <img src="https://img.shields.io/badge/React-18-61DAFB?style=flat-square&logo=react&logoColor=black" alt="React">
  <img src="https://img.shields.io/badge/TypeScript-5-3178C6?style=flat-square&logo=typescript&logoColor=white" alt="TypeScript">
  <img src="https://img.shields.io/badge/Vite-5-646CFF?style=flat-square&logo=vite&logoColor=white" alt="Vite">
  <img src="https://img.shields.io/badge/Tailwind_CSS-3-06B6D4?style=flat-square&logo=tailwindcss&logoColor=white" alt="Tailwind CSS">
</p>

---

## Overview

The KubeRCA Frontend is a React-based web dashboard that provides a user interface for viewing and managing Kubernetes incidents. It displays RCA (Root Cause Analysis) results, alerts, and allows users to search for similar past incidents.

### Key Features

- **Incident Dashboard** - View and manage Kubernetes incidents
- **RCA Viewer** - Display AI-generated root cause analysis
- **Alert Management** - Browse, filter, and manually resolve alerts (single & bulk)
- **Split Alert Analysis History** - View firing and resolved analyses separately in alert details
- **Similar Incident Search** - Find related past incidents via vector similarity
- **Muted Incidents** - Manage hidden/muted incidents
- **Analysis Dashboard** - Explore incident/alert trends, severity distribution, and top namespaces
- **Dark Mode** - Toggle between light and dark themes
- **Authentication** - Login/signup with JWT tokens
- **Operator Feedback** - Vote and comment on incident/alert analysis results
- **In-App AI Chat** - Context-aware floating chat panel for incident queries
- **Webhook Settings** - CRUD management for outbound webhook integrations
- **App Settings UI** - Configure notification, flapping detection, AI provider, and analysis mode
- **SSE Realtime Sync** - Server-Sent Events with polling fallback for live updates
- **OIDC Login** - One-click Google authentication with email allowlist

---

## Screenshots

The frontend currently provides views for:
- Incident list, detail, and resolve workflows
- Alert list, detail, manual resolve, and analysis history views
- Hidden/muted incident management
- Analysis dashboard with trend and breakdown widgets
- Settings pages for webhooks, notifications, flapping, AI provider, and analysis mode
- Floating chat panel for context-aware incident queries

---

## Tech Stack

| Category | Technology |
|----------|------------|
| **Framework** | React 18 |
| **Language** | TypeScript |
| **Build Tool** | Vite |
| **Styling** | Tailwind CSS |
| **Linting** | ESLint |
| **Container** | Docker + Nginx |

---

## Quick Start

### Prerequisites

- Node.js 18+ (LTS recommended)
- npm 9+

### Installation

```bash
# Run in repository root
# (monorepo layout: cd frontend)
npm ci
```

### Run Development Server

```bash
npm run dev
```

Open `http://localhost:5173` in your browser.

### Build for Production

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

---

## Project Structure

```text
frontend/
├── src/
│   ├── App.tsx            # Main app routes and dashboard orchestration
│   ├── components/        # Incident, alert, analytics, settings, and chat UI
│   ├── hooks/             # Realtime hooks such as SSE integration
│   ├── context/           # Theme and shared UI state
│   ├── utils/api.ts       # Backend API client and typed helpers
│   └── types.ts           # Shared frontend data types
├── public/                # Static assets
├── Dockerfile             # Production build (Nginx)
└── package.json           # Scripts and dependencies
```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | Backend API URL | `''` (same-origin) |

### Local Development

Create a `.env` file:

```bash
VITE_API_BASE_URL=http://localhost:8080
```

### Production (Kubernetes)

In Kubernetes deployments, set `VITE_API_BASE_URL` when frontend and backend are served from different origins.

---

## Available Scripts

| Command | Description |
|---------|-------------|
| `npm run dev` | Start development server |
| `npm run build` | Build for production |
| `npm run preview` | Preview production build |
| `npm run lint` | Run ESLint |

### Full Verification

```bash
npm run lint && npm run build
```

---

## API Integration

The frontend communicates with the KubeRCA Backend API:

### Endpoints Used

| API | Description |
|-----|-------------|
| `/api/v1/auth/*` | Authentication (login, register, refresh) |
| `/api/v1/incidents` | Incident list and details |
| `/api/v1/alerts` | Alert list and details |
| `/api/v1/embeddings/search` | Similar incident search |
| `/api/v1/events` | SSE realtime updates |
| `/api/v1/chat` | Context-aware chat |
| `/api/v1/analytics/dashboard` | Analysis dashboard data |
| `/api/v1/settings/webhooks*` | Webhook management |
| `/api/v1/settings/app/:key` | Notification, flapping, AI provider, and analysis settings |

### API Client

Located at `src/utils/api.ts`, the API client handles:
- Request/response formatting
- Authentication token management
- Error handling
- SSE/auth-aware request flows for realtime and settings pages

---

## Styling

### Tailwind CSS

The project uses Tailwind CSS for styling:

```tsx
// Example component with Tailwind classes
<div className="bg-white dark:bg-gray-800 p-4 rounded-lg shadow">
  <h1 className="text-xl font-bold text-gray-900 dark:text-white">
    Incident Details
  </h1>
</div>
```

### Dark Mode

Dark mode is implemented using Tailwind's `dark:` prefix:

```tsx
<div className="bg-white dark:bg-gray-900">
  {/* Content adapts to theme */}
</div>
```

Toggle is available in the header component.

---

## Docker

### Build Image

```bash
docker build -t kube-rca-frontend .
```

### Run Container

```bash
docker run -d -p 80:80 kube-rca-frontend
```

Access at `http://localhost`.

### Dockerfile Overview

Multi-stage build:
1. **Build stage** - Node.js builds the React app
2. **Production stage** - Nginx serves static files

---

## Development Guidelines

### TypeScript

- Use strict mode (`tsconfig.json`)
- Define shared types in `src/types.ts`
- Use interface for component props

### Components

- Functional components with hooks
- Keep components focused and reusable
- Use React Context for global state

### API Calls

```tsx
// Example API call pattern
const [data, setData] = useState<Incident[]>([]);
const [loading, setLoading] = useState(true);
const [error, setError] = useState<string | null>(null);

useEffect(() => {
  const fetchData = async () => {
    try {
      const response = await api.getIncidents();
      setData(response);
    } catch (err) {
      setError('Failed to fetch incidents');
    } finally {
      setLoading(false);
    }
  };
  fetchData();
}, []);
```

---

## Merge Policy (release-please)

- Avoid merge commits when merging PRs; they can create duplicate changelog entries.
- Prefer "Squash and merge" or "Rebase and merge" and ensure the final commit message is conventional.

---

## Related Components

- [KubeRCA Backend](https://github.com/kube-rca/backend) - Go REST API server
- [KubeRCA Agent](https://github.com/kube-rca/agent) - Python analysis service
- [Helm Charts](https://github.com/kube-rca/helm-charts) - Kubernetes deployment

---

## License

This project is part of KubeRCA, licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
