import { AppDispatch } from "@/app/store";
import { projectSlice } from "@/entities/Project";
import { fetchOne } from "../api/fetchOne";
import { extractProject } from "../utils/extractProject";
import { resourceSlice } from "@/entities/Resource";
import { extractResources } from "../utils/extractResources";
import { metricSlice } from "@/entities/Metric";
import { extractMetrics } from "../utils/extractMetrics";
import { memberSlice } from "@/entities/Member";
import { extractMembers } from "../utils/extractMembers";
import { addUsersToResources } from "../utils/addUsersToResources";

const slices = [projectSlice, resourceSlice, metricSlice, memberSlice];

const setLoadingStates = (dispatch: AppDispatch, isLoading: boolean) => {
  slices.forEach((slice) => dispatch(slice.actions.setLoading(isLoading)));
};

const setErrorStates = (dispatch: AppDispatch, error: string) => {
  slices.forEach((slice) => dispatch(slice.actions.setError(error)));
};

export const initializeProjectDetailed = async (
  dispatch: AppDispatch,
  id: string,
) => {
  setLoadingStates(dispatch, true);
  const result = await fetchOne(id);

  if (!result) {
    setErrorStates(dispatch, "Failed to fetch project data");
    setLoadingStates(dispatch, false);

    return;
  }

  const metrics = extractMetrics(result);

  dispatch(projectSlice.actions.setProject(extractProject(result)));
  dispatch(resourceSlice.actions.setResources(extractResources(result)));
  dispatch(metricSlice.actions.setMetrics(metrics));
  dispatch(memberSlice.actions.setMembers(extractMembers(result)));

  addUsersToResources(dispatch, metrics);

  setLoadingStates(dispatch, false);
};
