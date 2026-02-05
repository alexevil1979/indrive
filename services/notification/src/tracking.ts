/**
 * Real-time driver tracking via WebSocket
 * Passengers subscribe to driver location updates for their ride
 */
import type { WebSocket } from "ws";
import logger from "./logger.js";

type Location = {
  lat: number;
  lng: number;
  heading?: number; // direction in degrees
  speed?: number;   // km/h
  timestamp: number;
};

type DriverClient = {
  ws: WebSocket;
  driverId: string;
  rideId: string;
  lastLocation?: Location;
};

type PassengerClient = {
  ws: WebSocket;
  passengerId: string;
  rideId: string;
};

// Active driver connections per ride
const driversByRide = new Map<string, DriverClient>();

// Passenger subscribers per ride
const passengersByRide = new Map<string, Set<PassengerClient>>();

/**
 * Register a driver's WebSocket connection for location broadcasting
 */
export function registerDriver(
  ws: WebSocket,
  rideId: string,
  driverId: string
): void {
  // Close previous connection if exists
  const existing = driversByRide.get(rideId);
  if (existing && existing.ws !== ws) {
    existing.ws.close(4001, "New connection from same driver");
  }

  const client: DriverClient = { ws, driverId, rideId };
  driversByRide.set(rideId, client);

  logger.info({ msg: "driver_tracking_registered", rideId, driverId });

  ws.on("close", () => {
    if (driversByRide.get(rideId)?.ws === ws) {
      driversByRide.delete(rideId);
      // Notify passengers that driver disconnected
      broadcastToPassengers(rideId, {
        type: "driver_offline",
        rideId,
      });
      logger.info({ msg: "driver_tracking_disconnected", rideId, driverId });
    }
  });
}

/**
 * Subscribe a passenger to driver location updates
 */
export function subscribePassenger(
  ws: WebSocket,
  rideId: string,
  passengerId: string
): void {
  let passengers = passengersByRide.get(rideId);
  if (!passengers) {
    passengers = new Set();
    passengersByRide.set(rideId, passengers);
  }

  const client: PassengerClient = { ws, passengerId, rideId };
  passengers.add(client);

  logger.info({ msg: "passenger_subscribed_tracking", rideId, passengerId });

  // Send current driver location if available
  const driver = driversByRide.get(rideId);
  if (driver?.lastLocation) {
    sendToClient(ws, {
      type: "driver_location",
      rideId,
      driverId: driver.driverId,
      location: driver.lastLocation,
    });
  } else if (driver) {
    // Driver connected but no location yet
    sendToClient(ws, {
      type: "driver_online",
      rideId,
      driverId: driver.driverId,
    });
  }

  ws.on("close", () => {
    passengers?.delete(client);
    if (passengers?.size === 0) {
      passengersByRide.delete(rideId);
    }
    logger.info({ msg: "passenger_unsubscribed_tracking", rideId, passengerId });
  });
}

/**
 * Update driver location and broadcast to subscribers
 */
export function updateDriverLocation(
  rideId: string,
  driverId: string,
  location: Location
): void {
  const driver = driversByRide.get(rideId);
  
  // Verify it's the correct driver
  if (!driver || driver.driverId !== driverId) {
    logger.warn({ msg: "location_update_from_unknown_driver", rideId, driverId });
    return;
  }

  driver.lastLocation = location;

  // Broadcast to all subscribed passengers
  broadcastToPassengers(rideId, {
    type: "driver_location",
    rideId,
    driverId,
    location,
  });
}

/**
 * Get current driver location for a ride (if available)
 */
export function getDriverLocation(rideId: string): Location | null {
  return driversByRide.get(rideId)?.lastLocation ?? null;
}

/**
 * Check if driver is connected for a ride
 */
export function isDriverOnline(rideId: string): boolean {
  const driver = driversByRide.get(rideId);
  return !!driver && driver.ws.readyState === 1; // OPEN
}

/**
 * Broadcast message to all passenger subscribers for a ride
 */
function broadcastToPassengers(rideId: string, payload: object): void {
  const passengers = passengersByRide.get(rideId);
  if (!passengers) return;

  const msg = JSON.stringify(payload);
  for (const client of passengers) {
    sendToClient(client.ws, payload, msg);
  }
}

/**
 * Send message to a single WebSocket client
 */
function sendToClient(ws: WebSocket, payload: object, serialized?: string): void {
  if (ws.readyState === 1) { // OPEN
    ws.send(serialized ?? JSON.stringify(payload));
  }
}

/**
 * Clean up tracking for a ride (when ride completes/cancels)
 */
export function cleanupRide(rideId: string): void {
  const driver = driversByRide.get(rideId);
  if (driver) {
    driver.ws.close(4002, "Ride ended");
    driversByRide.delete(rideId);
  }

  const passengers = passengersByRide.get(rideId);
  if (passengers) {
    for (const client of passengers) {
      client.ws.close(4002, "Ride ended");
    }
    passengersByRide.delete(rideId);
  }

  logger.info({ msg: "tracking_cleanup", rideId });
}

/**
 * Get tracking stats (for monitoring)
 */
export function getStats(): { activeDrivers: number; activePassengers: number } {
  let activePassengers = 0;
  for (const passengers of passengersByRide.values()) {
    activePassengers += passengers.size;
  }
  return {
    activeDrivers: driversByRide.size,
    activePassengers,
  };
}
