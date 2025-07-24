import * as grpc from "@grpc/grpc-js";
import * as protoLoader from "@grpc/proto-loader";
import { getProtobufPaths } from "./helper.js";

import * as googleProtoFiles from "google-proto-files";
import type {
  GRPCmetricTransportOptions,
  MetricsService,
  MetricsProtoDescriptor,
  MetricsData,
} from "./types.js";

class GRPCmetricTransport {
  private serverAddress: string;
  private token: string;
  private servicename: string;
  private stream: grpc.ClientDuplexStream<any, any> | null = null;
  private client: MetricsService | null = null;
  private isConnecting: boolean = false;
  private reconnectAttempts: number = 0;

  constructor({
    serverAddress,
    token,
    servicename,
  }: GRPCmetricTransportOptions) {
    this.serverAddress = serverAddress;
    this.token = token;
    this.servicename = servicename;
    const { metricsProtoPath, protoDirectory } = getProtobufPaths();
    const packageDefinition = protoLoader.loadSync(metricsProtoPath, {
      keepCase: true,
      longs: String,
      enums: String,
      defaults: true,
      oneofs: true,
      includeDirs: [protoDirectory, googleProtoFiles.getProtoPath()],
    });

    const protoDescriptor = grpc.loadPackageDefinition(
      packageDefinition
    ) as unknown as MetricsProtoDescriptor;
    this.client = new protoDescriptor.logboy.MetricsService(
      this.serverAddress,
      grpc.credentials.createInsecure()
    );

    this.initializeStream();
  }

  private initializeStream(): void {
    if (this.stream) {
      try {
        this.stream.end();
      } catch (err: any) {
        console.warn("Failed to end previous stream:", err.message);
      }
    }

    if (!this.client) {
      console.error("gRPC client not initialized");
      return;
    }

    const metadata = new grpc.Metadata();
    metadata.add("authorization", `Bearer ${this.token}`);
    metadata.add("servicename", this.servicename);

    this.stream = this.client.ReceiveMetrics(metadata, (error, response) => {
      if (error) {
        console.error("ReceiveMetrics stream error:", error.message);
        this.handleReconnection();
      } else {
        console.log("Metrics stream response:", response);
      }
    });

    this.stream.on("error", (err: Error) => {
      console.error("Stream error:", err.message);
      this.handleReconnection();
    });

    this.stream.on("end", () => {
      console.warn("Stream ended unexpectedly.");
      this.handleReconnection();
    });

    console.log("gRPC metrics stream initialized.");
  }

  public sendMetrics(metrics: MetricsData): void {
    if (!this.stream) {
      console.warn("Metrics stream not ready.");
      return;
    }

    try {
      this.stream.write(metrics);
    } catch (err: any) {
      console.error("Failed to write metrics:", err.message);
      this.handleReconnection();
    }
  }

  private handleReconnection(): void {
    if (this.isConnecting) return;
    this.isConnecting = true;

    const backoff = 1000 * Math.min(5, this.reconnectAttempts + 1);
    console.log(`Reconnecting in ${backoff}ms...`);

    setTimeout(() => {
      this.reconnectAttempts++;
      this.initializeStream();
      this.isConnecting = false;
    }, backoff);
  }

  public close(): void {
    if (this.stream) {
      this.stream.end();
      this.stream = null;
    }
  }
}

export default GRPCmetricTransport;
