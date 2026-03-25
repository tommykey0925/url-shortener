Start the local development environment for the URL shortener.

1. Activate flox: `flox activate`
2. Start Go API in background: `cd api && go run . &`
3. Start SvelteKit dev server: `cd web && pnpm dev`

The Vite proxy in web/vite.config.ts forwards /api/* and /r/* to localhost:8080 (Go API).

Note: DynamoDB access requires valid AWS credentials. For local dev without AWS, suggest using DynamoDB Local.
