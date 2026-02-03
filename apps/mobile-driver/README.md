# RideHail Driver (Expo)

React Native (Expo) driver app — registration (role=driver), profile verification stub, available rides, place bid, my rides, start/complete, navigation stub.

## Setup

1. From repo root: `pnpm install`
2. Add assets: `icon.png`, `splash.png`, `adaptive-icon.png` in `assets/`
3. Set API URLs (optional): same as passenger app (defaults: localhost 8080, 8083, 8081)
4. Run Auth + User + Ride services; driver must register with role=driver (or create driver profile in User service)
5. `pnpm dev` (or `npx expo start`)

## Features

- **Auth:** Login, Register (driver), token in AsyncStorage, logout
- **Available rides:** List open rides (GET /api/v1/rides/available), tap → ride detail
- **Place bid:** Ride detail — quick buttons 500/700/1000 ₽ or custom price, submit bid
- **My rides:** List rides where I'm driver; tap → detail
- **Ride detail (driver):** If matched — Start ride, Open navigation (Google Maps), Complete, Cancel
- **Profile:** Create driver profile (license number), verification stub (doc upload later)
