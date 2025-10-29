import { IProjectDetailed } from "..";

export const fetchOne = async (
  id: string,
): Promise<IProjectDetailed | null> => {
  const response = await fetch(
    import.meta.env.VITE_SERVER_URL + `project/${id}`,
    { credentials: "include" },
  );

  if (!response.ok) throw new Error("Failed to fetch project data");

  try {
    return response.json();
  } catch {
    return null;
  }
};
