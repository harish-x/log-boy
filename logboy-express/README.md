# logboy-express

**logboy-express** is the official Node.js SDK for integrating [LogBoy](https://github.com/your-project/logboy) with Express-based applications. LogBoy enables **centralized logging and metrics collection** with minimal overhead, helping teams gain observability across distributed systems effortlessly.

## ✨ Features

- Lightweight integration with Express.js
- Centralized log and metrics transport via gRPC
- Supports five log levels: `info`, `warn`, `debug`, `error`, `silly`
- Compatible with [Morgan](https://www.npmjs.com/package/morgan) for automatic HTTP request logging
- Custom metadata attachment for detailed request insights
- Smart defaults, easy configuration


```javascript
const express = require("express");
const { LogBoy } = require("logboy-express");

const app = express();

// Attach request metadata
const attachMetadata = (req, res, next) => {
  req.Metadata = {
    requestMethod: req.method,
    requestUrl: req.originalUrl || req.url,
    responseStatus: res?.statusCode ?? null,
    userAgent: req.headers["user-agent"],
    remoteIp: req.ip,
  };
  next();
};

app.use(attachMetadata);

// Initialize LogBoy
const log = new LogBoy(
  "localhost:50051",     // gRPC endpoint
  "private key",         // project private key
  "project_name",        // project name
  {
    appVersion: "1.0.0",
    logLevel: "debug",
    enableMetrics: true,
  }
);

// Sample route
app.get("/", (req, res) => {
  log.info("HTTP request received", req.Metadata);
  res.status(401).json({ message: "Hello World!" });
});
```

#### Log Methods

```javascript
log.info("message", metadata?);
log.warn("message", metadata?);
log.debug("message", metadata?);
log.error("message", error, metadata?);
log.silly("message", metadata?);
```

| Option          | Type    | Description                                      |
| --------------- | ------- | ------------------------------------------------ |
| `appVersion`    | String  | Your application version                         |
| `logLevel`      | String  | One of `debug`, `info`, `warn`, `error`, `silly` |
| `enableMetrics` | Boolean | Enable/disable automatic metrics                 |

#### Error Logging Example

```javascript
try {
  // some logic
} catch (error) {
  log.error("Error in authentication", error, req.Metadata);
}
```

#### Morgan Integration (Optional)

```javascript
const { LogBoy, logFormat } = require("logboy-express");
const morgan = require("morgan");

// Define log handler
function logMessageHandler(message) {
  log.info("HTTP request received", JSON.parse(message.trim()));
}

// Middleware for morgan logging
const loggingMiddleware = morgan(logFormat, {
  stream: { write: logMessageHandler },
});

app.use(loggingMiddleware);
```


Let me know if you'd like me to tailor it further for publishing on NPM


