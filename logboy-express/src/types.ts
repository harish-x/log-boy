import * as grpc from "@grpc/grpc-js";
import Transport from "winston-transport";

// gRpcLogTransport types

export interface GrpcTransportOptions extends Transport.TransportStreamOptions {
  serverAddress: string;
  metadata: grpc.Metadata;
}

export interface BuildDetails {
  nodeVersion?: string;
  appVersion?: string;
}

export interface LogInfo {
  serviceName?: string;
  buildDetails?: BuildDetails;
  level: string;
  message: string;
  stack?: string;
  timestamp: string | Date;
  requestId?: string;
  requestUrl?: string;
  requestMethod?: string;
  remoteIp?: string;
  userAgent?: string;
  responseStatus?: number;
  responseTime?: number;
}

export interface LogMessage {
  serviceName: string;
  buildDetails: {
    nodeVersion: string;
    appVersion: string;
  };
  level: string;
  message: string;
  stack: string;
  timestamp: {
    seconds: number;
    nanos: number;
  };
  requestId: string;
  requestUrl: string;
  requestMethod: string;
  remoteIp: string;
  userAgent: string;
  responseStatus: number;
  responseTime: number;
}

export interface LogService {
  ReceiveLogsStream(
    metadata: grpc.Metadata,
    callback?: (error: grpc.ServiceError | null, response?: any) => void
  ): grpc.ClientDuplexStream<LogMessage, any>;
}

export interface LogProtoDescriptor {
  logboy: {
    LogService: new (
      address: string,
      credentials: grpc.ChannelCredentials
    ) => LogService;
  };
}

// gRPC metrics transport types

export interface GRPCmetricTransportOptions {
  serverAddress: string;
  token: string;
  servicename: string;
}

export interface MetricsService {
  ReceiveMetrics(
    metadata: grpc.Metadata,
    callback?: (error: grpc.ServiceError | null, response?: any) => void
  ): grpc.ClientDuplexStream<any, any>;
}

export interface MetricsProtoDescriptor {
  logboy: {
    MetricsService: new (
      address: string,
      credentials: grpc.ChannelCredentials
    ) => MetricsService;
  };
}

// custom logger types

export interface MemoryUsage {
  timestamp: number;
  totalMemory: number;
  freeMemory: number;
  usedMemory: number;
  memoryUsagePercentage: number;
}

export interface CpuCore {
  core: number;
  usage: number;
}

export interface CpuUsage {
  timestamp: number;
  average: number;
  cores: CpuCore[];
}

export interface CpuTimes {
  user: number;
  nice: number;
  sys: number;
  idle: number;
  irq: number;
}

export interface MetricsData {
  memoryUsage: {
    timestamp: number;
    totalMemory: number;
    freeMemory: number;
    usedMemory: number;
    memoryUsagePercentage: number;
  };
  cpuUsage: {
    timestamp: number;
    average: number;
    cores: Array<{
      core: number;
      usage: number;
    }>;
  };
  serviceName: string;
}

export interface LogMeta {
  [key: string]: any;
  stack?: string;
  errorMessage?: string;
}

// common types

export type LogLevel =
  | "error"
  | "warn"
  | "info"
  | "debug"
  | "verbose"
  | "silly";

export interface ServiceConfig {
  serverAddress: string;
  token: string;
  serviceName: string;
  appVersion?: string;
  logLevel?: LogLevel;
}
