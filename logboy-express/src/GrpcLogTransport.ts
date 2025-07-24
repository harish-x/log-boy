import Transport from "winston-transport";
import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import path from "path";
import * as googleProtoFiles from "google-proto-files";
import crypto from "crypto";
import type {
  GrpcTransportOptions,
  LogInfo,
  LogMessage,
  LogService,
  LogProtoDescriptor,
} from "./types.js";
import { getProtobufPaths } from "./helper.js";

class GrpcTransport extends Transport {
  private serverAddress: string;
  private metadata: grpc.Metadata;
  private stream: grpc.ClientDuplexStream<LogMessage, any> | null = null;
  private client: LogService;
  private reconnectAttempts: number = 0;
  private isConnecting: boolean = false;

  constructor(opts: GrpcTransportOptions) {
    super(opts);
    this.serverAddress = opts.serverAddress;
    this.metadata = opts.metadata;
    const { logsProtoPath, protoDirectory } = getProtobufPaths();
    const packageDefinition = protoLoader.loadSync(logsProtoPath, {
      keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true,
      includeDirs: [
        protoDirectory,
        googleProtoFiles.getProtoPath(protoDirectory),
      ],
    });

    const protoDescriptor = grpc.loadPackageDefinition(
      packageDefinition
    ) as unknown as LogProtoDescriptor;
    this.client = new protoDescriptor.logboy.LogService(
      this.serverAddress,
      grpc.credentials.createInsecure()
    );

    this.initializeStream();
  }

  private initializeStream(): void {
    if (this.stream) {
      try {
        this.stream.end();
      } catch (e: any) {
        this.isConnecting = false;
        this.handleReconnection();
        console.warn("Stream end failed:", e.message);
      }
      this.stream = null;
    }
    this.isConnecting = true;

    this.stream = this.client.ReceiveLogsStream(
      this.metadata,
      (error, response) => {
        this.isConnecting = false;
        if (error) {
          console.error("Stream error:", error.message);
        } else {
          console.log("Stream ended, ack:", response?.ack);
        }
      }
    );
  }

  log(info: LogInfo, callback: () => void): void {
    setImmediate(() => this.emit("logged", info));

    const logMessage: LogMessage = {
      serviceName: info.serviceName || "unknown",
      buildDetails: {
        nodeVersion: info.buildDetails?.nodeVersion || process.version,
        appVersion: info.buildDetails?.appVersion || "1.0.0",
      },
      level: info.level,
      message: info.message,
      stack: info.stack || "",
      timestamp: {
        seconds: Math.floor(new Date(info.timestamp).getTime() / 1000),
        nanos: (new Date(info.timestamp).getTime() % 1000) * 1e6,
      },
      requestId: info.requestId || crypto.randomUUID(),
      requestUrl: info.requestUrl || "",
      requestMethod: info.requestMethod || "",
      remoteIp: info.remoteIp || "",
      userAgent: info.userAgent || "",
      responseStatus: info.responseStatus || 0,
      responseTime: info.responseTime || 0,
    };

    if (this.stream) {
      try {
        this.stream.write(logMessage);
        callback();
      } catch (error: any) {
        console.error("Stream write error:", error.message);
        callback();
      }
    } else {
      console.error("Stream not initialized");
      this.handleReconnection();
      callback();
    }
  }

  private handleReconnection(): void {
    if (this.isConnecting) {
      console.warn("Already connecting, skipping reconnection attempt.");
      return;
    }
    this.reconnectAttempts++;
    if (this.reconnectAttempts > 5) {
      console.error(
        "Max reconnection attempts reached. Stopping reconnection."
      );
      return;
    }
    console.log(`Reconnection attempt #${this.reconnectAttempts}`);

    const delay = 1000 * Math.pow(2, this.reconnectAttempts);
    setTimeout(() => {
      this.initializeStream();
    }, delay);
  }

  public close(): void {
    if (this.stream) {
      try {
        this.stream.end();
      } catch (e: any) {
        console.warn("Stream close failed:", e.message);
      }
      this.stream = null;
    }
  }
}

export default GrpcTransport;
