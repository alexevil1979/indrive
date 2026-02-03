# RideHail Passenger (Expo)

React Native (Expo) passenger app — registration, login, ride request, bidding UI, map stub.

## Setup

1. From repo root: `pnpm install`
2. Add assets: `icon.png`, `splash.png`, `adaptive-icon.png` in `assets/`
3. Set API URLs (optional): `EXPO_PUBLIC_API_URL`, `EXPO_PUBLIC_RIDE_API_URL`, `EXPO_PUBLIC_USER_API_URL` (defaults: localhost 8080, 8083, 8081)
4. Run Auth + User + Ride services (see services/README)
5. `pnpm dev` (or `npx expo start`)

## Features

- **Auth:** Login, Register (passenger), token in AsyncStorage, logout
- **Home:** Map stub (OpenStreetMap placeholder) + create ride (from/to, submit)
- **My rides:** List rides, tap → ride detail
- **Ride detail:** Status, from/to, price; list bids; accept bid; update status (in_progress, completed, cancelled)
