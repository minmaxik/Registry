import { IPlatform } from "..";

export const fetchAll = async (): Promise<IPlatform[]> => {
  const response = await fetch(import.meta.env.VITE_SERVER_URL + "platform", {
    credentials: "include",
  });

  if (!response.ok) throw new Error("Failed to fetch platforms");

  return response.json();
};
