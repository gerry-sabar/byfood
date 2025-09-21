# Books Atomic Next (Frontend)

Next.js + Tailwind (Atomic Design) frontend for your Go Books API.

## Run

1. Install deps: `pnpm i` (or `npm i` / `yarn`)
2. Create `.env.local`:

```code
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

3.Dev server: `pnpm dev` → `http://localhost:3000`

## Endpoints

This project expects your Go backend to expose:

- GET    /books/
- POST   /books/
- GET    /books/{id}
- PUT    /books/{id}
- DELETE /books/{id}

Edit `lib/api.ts` if your paths differ.