/** Ride / trip domain types */
export type RideStatus =
  | "requested"
  | "bidding"
  | "matched"
  | "in_progress"
  | "completed"
  | "cancelled";

export interface Ride {
  id: string;
  passengerId: string;
  driverId?: string;
  status: RideStatus;
  from: { lat: number; lng: number; address?: string };
  to: { lat: number; lng: number; address?: string };
  price?: number;
  createdAt: string;
}
