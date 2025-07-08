import { createApi, fetchBaseQuery } from "@reduxjs/toolkit/query/react";
import { msalInstance, apiTokenRequest } from "@/authConfig";

const baseQuery = fetchBaseQuery({
  baseUrl: "http://localhost:8080/api/v1/",
  prepareHeaders: async (headers) => {
    const accounts = msalInstance.getAllAccounts();
    if (accounts.length > 0) {
      const tokenResponse = await msalInstance.acquireTokenSilent({
        ...apiTokenRequest,
        account: accounts[0],
      });

      headers.set("Authorization", `Bearer ${tokenResponse.accessToken}`);
    }

    return headers;
  },
});

const baseQueryWithReauth = async (args, api, extraOptions) => {
  let result = await baseQuery(args, api, extraOptions);

  if (result.error && result.error.status === 403) {
    // Optional: Handle silent token renewal or logout
    try {
      const accounts = msalInstance.getAllAccounts();
      if (accounts.length > 0) {
        const tokenResponse = await msalInstance.acquireTokenSilent({
          ...apiTokenRequest,
          account: accounts[0],
          forceRefresh: true,
        });

        if (tokenResponse.accessToken) {
          result = await baseQuery(args, api, extraOptions);
        }
      } else {
        sessionStorage.clear();
        throw new Error("No accounts found");
      }
    } catch (err) {
      sessionStorage.clear();
    }
  }

  return result;
};

export const reAuthQuery = createApi({
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Projects"],
  endpoints: (builder) => ({}),
});
