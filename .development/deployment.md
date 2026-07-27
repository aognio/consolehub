# Deployment Policy & Operational Boundaries

> [!IMPORTANT]
> **Deployment Management**: All application deployment operations are handled exclusively out-of-band by a separate dedicated agent/pipeline.

## Coding Agent Rules:
1. Do NOT execute deployment commands (e.g. systemctl restart, remote ssh deployment scripts, docker push, k8s apply).
2. Limit all agent activities strictly to code, tests, documentation, and local verification builds (`make build`, `make test`).
