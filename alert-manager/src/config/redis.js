const redis = require("redis");

// Create Redis client with additional configuration options
const client = redis.createClient({
  url: `redis://${process.env.REDIS_HOST}:${process.env.REDIS_PORT}`,
  socket: {
    reconnectStrategy: (retries) => {
      // Exponential backoff for reconnection
      const delay = Math.min(retries * 100, 5000); // Max 5 seconds delay
      console.log(`Redis reconnecting attempt ${retries}, waiting ${delay}ms`);
      return delay;
    },
    connectTimeout: 5000, // 5 seconds connection timeout
  },
});

// Error handling
client.on("error", (err) => {
  console.error("Redis Client Error:", err);
  // You might want to add additional error handling logic here
});

// Connection logging
client.on("connect", () => console.log("Redis client connected"));
client.on("ready", () => console.log("Redis client ready"));
client.on("reconnecting", () => console.log("Redis client reconnecting"));
client.on("end", () => console.log("Redis client disconnected"));

// Async function to ensure connection is established
const connectRedis = async () => {
  try {
    await client.connect();
    console.log("Redis connection established");
  } catch (err) {
    console.error("Failed to connect to Redis:", err);
    // Depending on your use case, you might want to retry or exit the process
    process.exit(1);
  }
};

// Connect immediately and export the client
connectRedis();

// Add promisified versions of methods for easier async/await usage
const getAsync = async (key) => {
  try {
    return await client.get(key);
  } catch (err) {
    console.error("Redis GET error:", err);
    throw err;
  }
};

const setexAsync = async (key, ttl, value) => {
  try {
    return await client.setEx(key, ttl, value);
  } catch (err) {
    console.error("Redis SETEX error:", err);
    throw err;
  }
};

const publishAsync = async (channel, message) => {
  try {
    return await client.publish(channel, message);
  } catch (err) {
    console.error("Redis PUBLISH error:", err);
    throw err;
  }
};

// Enhanced multi/exec with better error handling
const multiExecAsync = async (commands) => {
  const multi = client.multi();

  commands.forEach(([cmd, ...args]) => {
    multi[cmd](...args);
  });

  try {
    return await multi.exec();
  } catch (err) {
    console.error("Redis MULTI/EXEC error:", err);
    throw err;
  }
};

// Export both the client and promisified methods
module.exports = {
  client,
  getAsync,
  setexAsync,
  publishAsync,
  multiExecAsync,
};
