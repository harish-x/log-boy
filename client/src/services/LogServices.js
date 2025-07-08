import { reAuthQuery } from "./BaseQuery";

export const LogService = reAuthQuery.injectEndpoints({
  endpoints: (builder) => ({
    // get logs API
    getLogs: builder.query({
      query: (filters) => {
        var query = `logs/${filters.project}?page=${filters.page}&limit=${filters.limit}`;
        if (filters.mode) {
          query += `&level=${filters.mode}`;
        }
        if (filters.fromdate) {
          query += `&from=${filters.fromdate}`;
        }
        if (filters.todate) {
          query += `&to=${filters.todate}`;
        }
        if (filters.SortByDate) {
          query += `&sortByDate=${filters.SortByDate}`;
        }
        return query;
      },
    }),
    // get max and min date of logs API
    getMinMaxDate: builder.query({
      query: (project) => `logs/${project}/date`,
    }),
    // get archived logs files list
    getArchiveLogList: builder.query({
      query: (project) => `logs/${project}/archives`,
    }),
    // get archived log from the file
    getArchivedLog: builder.query({
      query: (filter) => {
        let query = `logs/${filter.project}/archive?file=${filter.logName}&page=${filter.page}&limit=${filter.limit}`;
        if (filter.mode) {
          query += `&level=${filter.mode}`;
        }
        if (filter.fromdate) {
          query += `&from=${filter.fromdate}`;
        }
        if (filter.todate) {
          query += `&to=${filter.todate}`;
        }
        if (filter.SortByDate) {
          query += `&sortByDate=${filter.SortByDate}`;
        }
        return query;
      },
    }),
    getArchivedLogMinMaxDate: builder.query({
      query: (filter) => `logs/${filter.project}/archive/availabledates?file=${filter.logName}`,
    }),
  }),
});

export const { useLazyGetLogsQuery, useLazyGetMinMaxDateQuery, useGetArchiveLogListQuery, useLazyGetArchivedLogQuery, useLazyGetArchivedLogMinMaxDateQuery } = LogService;
