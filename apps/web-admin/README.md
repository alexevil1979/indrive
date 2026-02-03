# RideHail Admin (Next.js 15)

Admin panel — ride monitoring, user management (stub), analytics basics. App Router, Tailwind, shadcn-style UI.

## Setup

1. From repo root: `pnpm install`
2. Run Ride service (and Auth for admin user).
3. Create admin user in Auth: register with role=admin (or insert in DB), get JWT.
4. Set env: `ADMIN_JWT=<access_token_with_role_admin>` and optionally `NEXT_PUBLIC_RIDE_API_URL=http://localhost:8083`.
5. `pnpm dev` — open http://localhost:3000.

## Features

- **Dashboard:** Cards: total rides, completed today, in progress, revenue (from Ride API).
- **Rides:** Table of rides (ID, status, from/to, price, created). Data from GET /api/v1/admin/rides (admin JWT required).
- **Users:** Stub — connect User/Auth admin list API when available.

## Env

- `ADMIN_JWT` — JWT with role=admin (from Auth) to call Ride admin/rides.
- `NEXT_PUBLIC_RIDE_API_URL` — default http://localhost:8083.
