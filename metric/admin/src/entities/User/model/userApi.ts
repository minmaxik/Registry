import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";

export interface LoginParams {
  username: string;
  password: string;
  remember: boolean;
}

export const userApi = createApi({
  reducerPath: "userApi",
  baseQuery: fetchBaseQuery({
    baseUrl: import.meta.env.VITE_SERVER_URL,
    credentials: "include",
  }),
  endpoints: (build) => ({
    login: build.mutation<void, LoginParams>({
      query: (params) => ({
        url: `auth/login`,
        method: "POST",
        body: params,
      }),
    }),
    logout: build.mutation<void, void>({
      query: () => ({
        url: `auth/logout`,
        method: "GET",
      }),
    }),
  }),
});

export const { useLoginMutation, useLogoutMutation } = userApi;
