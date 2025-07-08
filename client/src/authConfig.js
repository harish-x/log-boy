import { PublicClientApplication } from "@azure/msal-browser";

export const msalConfig = {
  auth: {
    clientId: import.meta.env.VITE_MSAL_CLIENT_ID,
    authority: `https://login.microsoftonline.com/${import.meta.env.VITE_MSAL_TENANT_ID}/v2.0`,
    redirectUri: "http://localhost:5173/dashboard/projects",
  },
  cache: {
    cacheLocation: "sessionStorage",
    storeAuthStateInCookie: false,
  },
};

export const loginRequest = {
  scopes: ["openid", "profile", "User.Read"],
};

export const apiTokenRequest = {
  scopes: [`api://${import.meta.env.VITE_MSAL_SERVER_CLIENT_ID}/Data.Read`],
};

export const msalInstance = new PublicClientApplication(msalConfig);
