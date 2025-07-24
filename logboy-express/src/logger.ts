import winston from "winston";
import * as grpc from "@grpc/grpc-js";
import GrpcLogTransport from "./GrpcLogTransport.js";
import GRPCmetricTransport from "./GrpcMetricTransport.js";
import * as os from "os";
import type { MemoryUsage, CpuUsage, CpuCore, CpuTimes, MetricsData, LogMeta, LogLevel, } from "./types.js";

class LogBoy {
  private serviceName: string;
  private serverAddress: string;
  private logger: winston.Logger;
  private metricTransport: GRPCmetricTransport | null = null;
  private previousCpuTimes: CpuTimes[] | null = null;
  private monitoringInterval: NodeJS.Timeout | null = null;
  private enableMetrics: boolean;

  constructor(
    serverAddress: string,
    token: string,
    serviceName: string,
    options: {
      appVersion?: string;
      logLevel?: LogLevel;
      enableMetrics?: boolean;
    } = {}
  ) {
    if (!serverAddress || !token || !serviceName) {
      throw new Error(
        "serverAddress, token, and serviceName are required for LogBoy."
      );
    }

    this.serviceName = serviceName;

    console.log(
      "Custom logger initialized with server address:",
      serverAddress
    );

    if (options.enableMetrics === undefined) {
      this.enableMetrics = true;
    } else {
      this.enableMetrics = options.enableMetrics;
    }
    this.serverAddress = serverAddress;

    const metadata = new grpc.Metadata();
    metadata.add("authorization", `Bearer ${token}`);
    metadata.add("servicename", this.serviceName);

    this.logger = winston.createLogger({
      level: options.logLevel || "info",
      defaultMeta: {
        serviceName,
        buildDetails: {
          nodeVersion: process.version,
          appVersion: options.appVersion || "1.0.0",
        },
      },
      format: winston.format.combine(
        winston.format.timestamp(),
        winston.format.json()
      ),
      transports: [
        new GrpcLogTransport({
          serverAddress,
          metadata,
        }),
      ],
    });

    if (this.enableMetrics) {
      this.startMonitoring(15000, token);
    }
  }

  public log(level: string, message: string, meta: LogMeta = {}): void {
    this.logger.log(level, message, meta);
  }

  public info(message: string, meta: LogMeta = {}): void {
    this.logger.info(message, meta);
  }

  public error(
    message: string,
    error?: Error | string,
    meta: LogMeta = {}
  ): void {
    if (error instanceof Error) {
      meta.stack = error.stack;
    } else if (typeof error === "string") {
      meta.errorMessage = error;
    }
    this.logger.error(message, meta);
  }

  public warn(message: string, meta: LogMeta = {}): void {
    this.logger.warn(message, meta);
  }

  public verbose(message: string, meta: LogMeta = {}): void {
    this.logger.verbose(message, meta);
  }

  public debug(message: string, meta: LogMeta = {}): void {
    this.logger.debug(message, meta);
  }

  public silly(message: string, meta: LogMeta = {}): void {
    this.logger.silly(message, meta);
  }

  public getInstance(): winston.Logger {
    return this.logger;
  }

  private getMemoryUsage(): MemoryUsage {
    const totalMemory = os.totalmem();
    const freeMemory = os.freemem();
    const nowUtc = new Date();
    const epochMillis = nowUtc.getTime();
    return {
      timestamp: epochMillis,
      totalMemory,
      freeMemory,
      usedMemory: totalMemory - freeMemory,
      memoryUsagePercentage: ((totalMemory - freeMemory) / totalMemory) * 100,
    };
  }

  private getCpuUsage(): CpuUsage | null {
    const cpus = os.cpus();
    const currentCpuTimes: CpuTimes[] = cpus.map((cpu) => cpu.times);

    if (!this.previousCpuTimes) {
      this.previousCpuTimes = currentCpuTimes;
      return null;
    }

    const cpuUsages: CpuCore[] = currentCpuTimes.map((current, index) => {
      const previous = this.previousCpuTimes![index];

      const userDelta = current.user - previous.user;
      const niceDelta = current.nice - previous.nice;
      const sysDelta = current.sys - previous.sys;
      const idleDelta = current.idle - previous.idle;
      const irqDelta = current.irq - previous.irq;

      const totalDelta =
        userDelta + niceDelta + sysDelta + idleDelta + irqDelta;
      const activeDelta = totalDelta - idleDelta;

      const usage = totalDelta > 0 ? (activeDelta / totalDelta) * 100 : 0;

      return {
        core: index,
        usage: Math.round(usage * 100) / 100,
      };
    });

    const avgUsage =
      cpuUsages.reduce((sum, cpu) => sum + cpu.usage, 0) / cpuUsages.length;

    this.previousCpuTimes = currentCpuTimes;
    const nowUtc = new Date();
    const epochMillis = nowUtc.getTime();
    return {
      timestamp: epochMillis,
      average: Math.round(avgUsage * 100) / 100,
      cores: cpuUsages,
    };
  }

  private startMonitoring(intervalMs: number = 15000, token: string): void {
    this.metricTransport = new GRPCmetricTransport({
      serverAddress: this.serverAddress,
      token,
      servicename: this.serviceName,
    });

    this.monitoringInterval = setInterval(() => {
      const memoryData = this.getMemoryUsage();
      const cpuData = this.getCpuUsage();

      if (!cpuData) return;

      const metrics: MetricsData = {
        memoryUsage: {
          timestamp: memoryData.timestamp,
          totalMemory: memoryData.totalMemory,
          freeMemory: memoryData.freeMemory,
          usedMemory: memoryData.usedMemory,
          memoryUsagePercentage:
            Math.round(memoryData.memoryUsagePercentage * 100) / 100,
        },
        cpuUsage: {
          timestamp: cpuData.timestamp,
          average: cpuData.average,
          cores: cpuData.cores.map((core) => ({
            core: core.core,
            usage: core.usage,
          })),
        },
        serviceName: this.serviceName,
      };

      if (this.metricTransport) {
        this.metricTransport.sendMetrics(metrics);
      }
    }, intervalMs);
  }

  public stopMonitoring(): void {
    if (this.monitoringInterval) {
      clearInterval(this.monitoringInterval);
      this.monitoringInterval = null;
    }
    if (this.metricTransport) {
      this.metricTransport.close();
      this.metricTransport = null;
    }
  }

  public close(): void {
    this.stopMonitoring();
    this.logger.close();
  }
}

export { LogBoy };
