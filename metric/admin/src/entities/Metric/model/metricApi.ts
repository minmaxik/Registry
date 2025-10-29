import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { IMetric } from "..";
import { IAbstractMetricDetailed } from "../types";

export interface MetricParams {
  project: string;
  resource: string;
  name: string;
}

export const metricApi = createApi({
  reducerPath: "metricApi",
  baseQuery: fetchBaseQuery({ baseUrl: import.meta.env.VITE_SERVER_URL, credentials: 'include' }),
  endpoints: (build) => ({
    getMetricInfo: build.query<IAbstractMetricDetailed[], void>({
      query: () => "metric",
    }),
    createMetric: build.mutation<IMetric[], MetricParams>({
      query: (metricParams) => ({
        url: "metric",
        method: "POST",
        body: metricParams,
      }),
    }),
    deleteMetric: build.mutation<void, string>({
      query: (id) => ({
        url: `metric/${id}`,
        method: "DELETE",
      }),
    }),
    updateMetric: build.mutation<void, IMetric>({
      query: (metric) => ({
        url: `metric/${metric.id}`,
        method: "PUT",
        body: { metric: { ...metric, data: [] } },
      }),
    }),
    executeMetric: build.mutation<void, IMetric>({
      query: (metric) => ({
        url: `metric/${metric.id}/execute`,
        method: "POST",
        body: {
          metric: { ...metric, data: [] },
        },
      }),
    }),
    startMetric: build.mutation<void, IMetric>({
      query: (metric) => ({
        url: `metric/${metric.id}/start`,
        method: "POST",
        body: {
          metric: { ...metric, data: [] },
        },
      }),
    }),
    stopMetric: build.mutation<void, IMetric>({
      query: (metric) => ({
        url: `metric/${metric.id}/stop`,
        method: "POST",
        body: {
          metric: { ...metric, data: [] },
        },
      }),
    }),
  }),
});

export const {
  useGetMetricInfoQuery,
  useCreateMetricMutation,
  useDeleteMetricMutation,
  useUpdateMetricMutation,
  useStartMetricMutation,
  useStopMetricMutation,
  useExecuteMetricMutation,
} = metricApi;
