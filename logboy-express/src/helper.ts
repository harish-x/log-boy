import path from "path";
import { fileURLToPath } from "url";
import { createRequire } from "module";


export function getProtobufPaths(currentFileUrl?: string) {
  let baseDir: string;

  try {
    if (typeof __dirname !== "undefined") {
      baseDir = __dirname;
    } else if (currentFileUrl) {
      baseDir = path.dirname(fileURLToPath(currentFileUrl));
    } else {
      const require = createRequire(import.meta.url);
      const currentPath = require.resolve("./pathResolver.js");
      baseDir = path.dirname(currentPath);
    }
  } catch (error) {
    baseDir = process.cwd();
  }

  return {
    metricsProtoPath: path.resolve(baseDir, "../protobuf/metrics.proto"),
    logsProtoPath: path.resolve(baseDir, "../protobuf/log.proto"),
    protoDirectory: path.resolve(baseDir, "../protobuf"),
  };
}

export function getCurrentDir(fileUrl?: string): string {
  if (typeof __dirname !== "undefined") {
    return __dirname;
  }

  if (fileUrl) {
    return path.dirname(fileURLToPath(fileUrl));
  }

  return process.cwd();
}
