const timewindowFormater = (time_window) => {
  if (!time_window || typeof time_window !== "string") {
    throw new Error("Invalid time_window format");
  }

  const queryTime = time_window.trim().split(/\s+/);

  if (queryTime.length !== 2) {
    throw new Error('time_window must be in format "number unit" (e.g., "5 minutes")');
  }

  const time = parseInt(queryTime[0], 10);
  const unit = queryTime[1].toLowerCase();

  if (isNaN(time) || time <= 0) {
    throw new Error("time_window number must be a positive integer");
  }

  let timeWindow;

  if (unit === "days" || unit === "day") {
    timeWindow = `now-${time}d/d`;
  } else if (unit === "hours" || unit === "hour") {
    timeWindow = `now-${time}h/h`;
  } else if (unit === "minutes" || unit === "minute") {
    timeWindow = `now-${time}m/m`;
  } else if (unit === "seconds" || unit === "second") {
    timeWindow = `now-${time}s/s`;
  } else {
    throw new Error(`Unsupported time unit: ${unit}. Supported units: days, hours, minutes, seconds`);
  }

  return timeWindow;
};

const operators = {
  ">": (a, b) => {
    if (a == null || b == null) return false;
    return parseFloat(a) > parseFloat(b);
  },
  "<": (a, b) => {
    if (a == null || b == null) return false;
    return parseFloat(a) < parseFloat(b);
  },
  "==": (a, b) => {
    if (a == null && b == null) return true;
    if (a == null || b == null) return false;
    return parseFloat(a) === parseFloat(b);
  },
  ">=": (a, b) => {
    if (a == null || b == null) return false;
    return parseFloat(a) >= parseFloat(b);
  },
  "<=": (a, b) => {
    if (a == null || b == null) return false;
    return parseFloat(a) <= parseFloat(b);
  },
  "!=": (a, b) => {
    if (a == null && b == null) return false;
    if (a == null || b == null) return true;
    return parseFloat(a) !== parseFloat(b);
  },
};


// Format alert methods
function formatAlertMethods(rows) {
  if (!Array.isArray(rows)) {
    return [];
  }

  const methods = [];

  for (const alertmethod of rows) {
    if (!alertmethod || !alertmethod.method) {
      continue;
    }

    const { method, value } = alertmethod;

    const validMethods = ["email", "slack", "webhook", "sms", "discord"];
    if (!validMethods.includes(method.toLowerCase())) {
      console.warn(`Unknown alert method: ${method}`);
    }

    if (!value || value.trim() === "") {
      console.warn(`Empty value for alert method: ${method}`);
      continue;
    }

    methods.push({
      method: method.toLowerCase(),
      value: value.trim(),
    });
  }

  return methods;
}

// Validate alert rule
function validateAlertRule(alert) {
  const errors = [];

  if (!alert.project_name || alert.project_name.trim() === "") {
    errors.push("project_name is required");
  }

  if (!alert.rule_type || !["metric_avg", "log_count", "event_count"].includes(alert.rule_type)) {
    errors.push("rule_type must be one of: metric_avg, log_count, event_count");
  }

  if (alert.threshold == null || isNaN(parseFloat(alert.threshold))) {
    errors.push("threshold must be a valid number");
  }

  if (!alert.operator || !Object.keys(operators).includes(alert.operator)) {
    errors.push(`operator must be one of: ${Object.keys(operators).join(", ")}`);
  }

  if (!alert.time_window) {
    errors.push("time_window is required");
  } else {
    try {
      timewindowFormater(alert.time_window);
    } catch (error) {
      errors.push(`Invalid time_window format: ${error.message}`);
    }
  }

  if (alert.rule_type === "metric_avg") {
    if (!alert.metric_name || !["cpu_usage", "memory_usage"].includes(alert.metric_name)) {
      errors.push("metric_name must be one of: cpu_usage, memory_usage");
    }
  } else if (alert.rule_type === "log_count") {
    if (!alert.log_field || !["level", "status_code", "ip_address"].includes(alert.log_field)) {
      errors.push("log_field must be one of: level, status_code, ip_address");
    }

    if (!alert.log_field_value) {
      errors.push("log_field_value is required for log_count alerts");
    }

    if (alert.log_field === "status_code" && !["4xx", "5xx"].includes(alert.log_field_value)) {
      errors.push('log_field_value for status_code must be "4xx" or "5xx"');
    }
  }

  return errors;
}


//  Create alert cache key
function createAlertCacheKey(alert) {
  const baseKey = `${alert.project_name}_${alert.rule_type}`;

  if (alert.rule_type === "metric_avg") {
    return `${baseKey}_${alert.metric_name}_${alert.threshold}_${alert.operator}`;
  } else if (alert.rule_type === "log_count") {
    return `${baseKey}_${alert.log_field}_${alert.log_field_value}_${alert.threshold}_${alert.operator}`;
  } else if (alert.rule_type === "event_count") {
    return `${baseKey}_${alert.threshold}_${alert.operator}`;
  }

  return baseKey;
}

// Format alert message
function formatAlertMessage(alert) {
  const timestamp = new Date(alert.timestamp).toISOString();
  let message = `[ALERT] ${alert.project_name} - ${alert.rule_type.toUpperCase()}`;

  if (alert.rule_type === "metric_avg") {
    message += ` - ${alert.metric_name}: ${alert.current_value}% ${alert.operator} ${alert.threshold}%`;
  } else if (alert.rule_type === "log_count") {
    message += ` - ${alert.log_field}(${alert.log_field_value}): ${alert.current_value}% ${alert.operator} ${alert.threshold}%`;
    if (alert.error_count) {
      message += ` (${alert.error_count}/${alert.total_count})`;
    }
  } else if (alert.rule_type === "event_count") {
    message += ` - Events: ${alert.current_value}% ${alert.operator} ${alert.threshold}%`;
  }

  message += ` | Time Window: ${alert.time_window} | Triggered: ${timestamp}`;

  return message;
}


// Group alerts by type and metric/field
function groupAlertsByType(alerts) {
  const grouped = {};

  for (const alert of alerts) {
    let groupKey;

    if (alert.rule_type === "metric_avg") {
      groupKey = `${alert.project_name}_${alert.rule_type}_${alert.metric_name}`;
    } else if (alert.rule_type === "log_count") {
      groupKey = `${alert.project_name}_${alert.rule_type}_${alert.log_field}`;
    } else if (alert.rule_type === "event_count") {
      groupKey = `${alert.project_name}_${alert.rule_type}`;
    } else {
      groupKey = `${alert.project_name}_${alert.rule_type}`;
    }

    if (!grouped[groupKey]) {
      grouped[groupKey] = [];
    }

    grouped[groupKey].push(alert);
  }

  return grouped;
}

// Select highest priority alert from each group
function selectHighestPriorityAlerts(groupedAlerts) {
  const selectedAlerts = [];

  for (const [groupKey, alerts] of Object.entries(groupedAlerts)) {
    const sortedAlerts = alerts.sort((a, b) => b.priority - a.priority); // Sort by priority
    selectedAlerts.push(sortedAlerts[0]);
  }

  return selectedAlerts;
}

// Calculate priority based on threshold and operator
function calculatePriority(alert) {
  let basePriority = alert.threshold;

  if (alert.operator === ">" || alert.operator === ">=") {
    return basePriority;
  }

  if (alert.operator === "<" || alert.operator === "<=") {
    return 100 - basePriority;
  }

  return basePriority;
}

module.exports = {
  timewindowFormater,
  operators,
  formatAlertMethods,
  validateAlertRule,
  createAlertCacheKey,
  formatAlertMessage,
  calculatePriority,
  selectHighestPriorityAlerts,
  groupAlertsByType,
};
