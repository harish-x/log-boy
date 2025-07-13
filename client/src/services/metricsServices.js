import { reAuthQuery } from "./BaseQuery";

export const MetricsServices = reAuthQuery.injectEndpoints({
  endpoints: (builder) => ({
    getCpuUsage: builder.query({
      query: (data) => `metrics/${data.project}/cpu?from=${data.from}&to=${data.to}&points=${data.groupBy}`,
    }),
    getMemoryUsage: builder.query({
      query: (data) => `metrics/${data.project}/memory?from=${data.from}&to=${data.to}&points=${data.groupBy}`,
    }),
  }),
});

export const { useLazyGetCpuUsageQuery, useLazyGetMemoryUsageQuery } = MetricsServices;
