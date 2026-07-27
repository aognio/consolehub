# UI Design Notes

## Tech Stack & Architecture
- Server-side Go `html/template` with HTMX for partial page swaps and Alpine.js for local DOM state.
- Tailwind CSS styling with responsive design (Desktop, Tablet, Mobile).
- Light/Dark mode state stored in browser cookie (`consolehub_theme`), toggleable prior to login.

## Key Screens & Components
- **Console Viewer (`/runs/:id/console`)**: Monospace terminal font, auto-scroll, tail mode, pause/resume, search buffer, timestamp jump, raw download, text & JSON expandable view.
- **Dashboard (`/dashboard`)**: Metric cards for running processes, online/offline host count, recent runs, failure indicators, and live activity feed.
