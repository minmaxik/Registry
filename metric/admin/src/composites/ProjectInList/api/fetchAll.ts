import { ProjectInList } from "../types";

export const fetchAll = async (): Promise<ProjectInList[]> => {
  const response = await fetch(import.meta.env.VITE_SERVER_URL + "project", {
    credentials: "include",
  });

  if (!response.ok) throw new Error("Failed to fetch projects");

  return response.json();
};
