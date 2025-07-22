import { reAuthQuery } from "./BaseQuery";

export const AlertServices = reAuthQuery.injectEndpoints({
  endpoints: (builder) => ({
    getVerifiedEmails: builder.query({
      query: (data) => `alerts/email/${data}`,
    }),
    getAllAlertRules: builder.query({
      query: (data) => `alerts/${data}/all`,
    }),
    createAlertEmail: builder.mutation({
      query: (data) => ({
        url: `alerts/email`,
        method: "POST",
        body: data,
      }),
    }),
    verifyAlertEmail: builder.mutation({
      query: (data) => ({
        url: `alerts/email/verify`,
        method: "PATCH",
        body: data,
      }),
    }),
    createAlertRule: builder.mutation({
      query: (data) => ({
        url: `alerts/new`,
        method: "POST",
        body: data,
      }),
    }),
    getAlerts: builder.query({
      query: (data) => `alerts/${data}/old_alerts`,
    }),
  }),
});

export const {
  useLazyGetAllAlertRulesQuery,
  useLazyGetVerifiedEmailsQuery,
  useCreateAlertEmailMutation,
  useVerifyAlertEmailMutation,
  useCreateAlertRuleMutation,
  useLazyGetAlertsQuery,
} = AlertServices;
