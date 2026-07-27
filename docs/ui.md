# ConsoleHub UI Architecture

## Design & Theme System
- Built with HTML templates (`html/template`), HTMX, Alpine.js, and Tailwind CSS.
- **Theme Switcher**: Supports Light and Dark mode. Theme choice is stored in browser cookie (`consolehub_theme`) and does NOT require login before choosing.

## Flagship Console Viewer
- Monospace font (`JetBrains Mono`).
- Fast real-time rendering.
- Features: Auto-scroll toggle, Pause/Resume stream, Tail mode, Text search filter, Copy line, Jump to timestamp, Download raw stream.
- Content formatting: Supports plain text and collapsible JSON lines rendering.
