/** User / passenger / driver shared types */
export type UserRole = "passenger" | "driver" | "admin";

export interface User {
  id: string;
  email: string;
  role: UserRole;
  createdAt: string;
}

export interface PassengerProfile extends User {
  role: "passenger";
}

export interface DriverProfile extends User {
  role: "driver";
  verified: boolean;
}
