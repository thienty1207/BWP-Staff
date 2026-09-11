# BWP SonaSea frontend

The frontend uses SvelteKit, Svelte 5, TypeScript, Vite, and Bun.

## Developing

Install the locked dependencies and start the development server:

```bash
bun install --frozen-lockfile
bun run dev
```

The local Vite server proxies `/api/*` to the Go backend at
`http://127.0.0.1:3000`. Start the backend separately, then open
`http://localhost:5173/login`.

## Building

Run the frontend checks and create a production build with:

```bash
bun run check
bun run build
```

You can preview the production build with `bun run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.
