# Contributing to LogBoy

Thank you for your interest in contributing to **LogBoy** – a platform for centralized logging and metrics transport system.

We welcome community contributions to improve the system, add features, optimize performance, and extend support across multiple languages.

---

## General Guidelines

- All contributions must go through pull requests (PRs).
- Ensure your code is well-documented and adheres to the existing code style.
- Include relevant test cases for any new feature or bug fix.
- Keep commits atomic and descriptive.
- Create issues for major changes or features before starting development.

---

## Contribution Scope

### Main Server

The main LogBoy server is written in **Go** and handles authentication, log routing, and metric processing.

You can contribute by:
1. **Authentication Support**:
   - Improve support for Microsoft Authentication Library (MSAL).
   - Add support for alternative identity providers (AWS IAM, Okta, etc.).
   - Add customizable traditional auth options for private/internal use within organizations.

2. **Performance Optimization**:
   - Profile the codebase and eliminate bottlenecks.
   - Optimize gRPC handling and resource consumption.
   - Improve queueing or backpressure handling for high-throughput pipelines.

3. **Alert Integrations**:
   - Integrate support for alert forwarding to **Microsoft Teams** and **Slack**.
   - Allow configuration of webhook URLs per project or environment.
   - Build alert formatting and throttling logic for better usability.

---

### gRPC Gateway

The gRPC Gateway is the entry point for clients (SDKs) to send logs and metrics.

Contribute to:
1. **Security & Robustness**:
   - Harden authentication of incoming logs/metrics.
   - Add API key validation, rate limiting, and IP allow-listing.
   - Implement audit logging for access and changes.

2. **Protocol Extensions**:
   - Add versioning support for schema evolution.
   - Optimize protobuf definitions for log/metric payloads.

---

### SDK Packages

The Node.js SDK is already available via `logboy-express`.

We welcome contributions for:
1. **New Language SDKs**:
   - Implement log & metric transport clients for **Go**, **Java**, **Python**, **Rust**, etc.
   - Follow the existing SDK architecture (gRPC-based, lightweight, non-blocking).
   - Ensure language idioms are followed in each implementation.

2. **Enhancements**:
   - Add batching, retries, or circuit breakers.
   - Improve configuration flexibility.
   - Add example apps and usage guides.

---

### Fraud Detection System

A major upcoming feature is a **Fraud Detection System** on top of the LogBoy infrastructure.

Interested contributors can:
- Design and build real-time anomaly detection modules using logs and metrics.
- Work on pluggable detection algorithms (rule-based, ML-based).
- Integrate with existing alerting or SIEM systems.
- Create alerting triggers that feed into Teams/Slack integrations.

If you're interested in this area, open a discussion issue to collaborate on architecture and planning.

---

## Testing

- Write unit tests for all new logic.
- Use mock or test gRPC servers where applicable.
- Include example usage in SDKs for integration testing.

---

## 📄 License & CLA

By contributing to this project, you agree that your contributions will be licensed under the [MIT License](LICENSE). We may require signing a Contributor License Agreement (CLA) for substantial changes.

---

We appreciate your support in building LogBoy into a reliable and developer-friendly logging platform!
