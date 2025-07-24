export { LogBoy } from "./logger.js";

export const logFormat = `
{
    "requestMethod": ":method",
    "requestUrl": ":url",
    "responseStatus": ":status",
    "responseTime": ":response-time ms",
    "userAgent": ":user-agent",
    "remoteIp": ":remote-addr"
}`;
