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
- **Alert Management** - Browse and filter alerts
- **Similar Incident Search** - Find related past incidents via vector similarity
- **Muted Incidents** - Manage hidden/muted incidents
- **Dark Mode** - Toggle between light and dark themes
- **Authentication** - Login/signup with JWT tokens

---

## Screenshots

The dashboard provides views for:
- Incident list with filtering and pagination
- Incident detail with RCA analysis results
- Alert list and detail views
- Hidden/muted incident management

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
cd frontend
npm install
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

```
frontend/
├── src/
│   ├── main.tsx              # React entrypoint
│   ├── App.tsx               # Main app with routing
│   ├── types.ts              # TypeScript type definitions
│   ├── index.css             # Tailwind CSS entrypoint
│   ├── components/           # React components
│   │   ├── AlertTable.tsx         # Alerts list table
│   │   ├── AlertDetailView.tsx    # Alert detail view
│   │   ├── RCATable.tsx           # Incidents/RCA list table
│   │   ├── RCADetailView.tsx      # Incident detail with RCA
│   │   ├── MuteTable.tsx          # Muted incidents table
│   │   ├── MuteDetailView.tsx     # Muted incident detail
│   │   ├── Header.tsx             # App header (theme, auth)
│   │   ├── AuthPanel.tsx          # Login/signup panel
│   │   ├── Pagination.tsx         # Pagination component
│   │   └── TimeRangeSelector.tsx  # Time range filter
│   ├── context/              # React Context providers
│   │   ├── ThemeContext.ts        # Theme context definition
│   │   └── ThemeProvider.tsx      # Theme provider component
│   ├── constants/            # Application constants
│   │   └── index.ts
│   └── utils/                # Utility functions
│       ├── api.ts                 # Backend API client
│       ├── auth.ts                # Authentication utilities
│       ├── config.ts              # Configuration utilities
│       └── filterAlerts.ts        # Alert filtering logic
├── public/                   # Static assets
├── Dockerfile               # Production build (Nginx)
├── package.json             # Dependencies and scripts
├── tsconfig.json            # TypeScript configuration
├── vite.config.ts           # Vite configuration
├── tailwind.config.js       # Tailwind CSS configuration
└── eslint.config.js         # ESLint configuration
```

---

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | Backend API URL | `http://kube-rca-backend:8080` |

### Local Development

Create a `.env` file:

```bash
VITE_API_BASE_URL=http://localhost:8080
```

### Production (Kubernetes)

In Kubernetes deployments, the default `http://kube-rca-backend:8080` is used for internal cluster communication.

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
| `/api/v1/incidents/:id/alerts` | Alerts for an incident |
| `/api/v1/alerts` | Alert list and details |
| `/api/v1/embeddings/search` | Similar incident search |

### API Client

Located at `src/utils/api.ts`, the API client handles:
- Request/response formatting
- Authentication token management
- Error handling

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

## Related Components

- [KubeRCA Backend](https://github.com/kube-rca/backend) - Go REST API server
- [KubeRCA Agent](https://github.com/kube-rca/agent) - Python analysis service
- [Helm Charts](https://github.com/kube-rca/helm-charts) - Kubernetes deployment

---

## License

This project is part of KubeRCA, licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
