# Logboy

**Open Source All-in-One Logging Solution for Modern Startups**

Say goodbye to complex ELK stacks, expensive Grafana setups, and fragmented Prometheus configurations. Our open source centralized logging server delivers enterprise-grade monitoring capabilities in a single, self-hosted platform designed specifically for startups who need powerful logging without the operational overhead.

**Perfect for startups** who want enterprise-level observability without managing multiple tools
 **Scales with enterprises** managing multiple projects and teams on one unified platform
**Cost-effective alternative** to assembling and maintaining separate Loki, Grafana, and Prometheus stacks
**100% Open Source** - Deploy on your infrastructure with complete control and customization

## Architecture Overview

This system consists of multiple interconnected components that work together to provide a robust logging infrastructure:

-   **Client Server**: Frontend interface for users to view logs, metrics, and manage projects
-   **Main Server**: Core backend service handling business logic and data processing
-   **gRPC Gateway Server**: High-performance API gateway for log/metric ingestion
-   **NPM Package**: Client-side SDK for easy integration and log/metric transmission
-   **Backup-Cron**: Automated data archival based on retention policies
-   **Alert-Manager**: Real-time alert processing and notification system

<img width="3111" height="1416" alt="Screenshot from 2025-07-25 22-43-09" src="https://github.com/user-attachments/assets/ac9c95bf-5a5b-40d9-bda3-05bf0b65eadd" />

## Database Architecture


### Primary Databases

-   **Elasticsearch**: Primary storage for logs and metrics with full-text search capabilities
-   **PostgreSQL**: Project management, user authentication, and configuration data
-   **Redis**: Alert state management and pub/sub messaging system
-   **Apache Kafka**: Asynchronous message streaming with Schema Registry

### Analytics & Backup

-   **Azure Synapse**: Analytics workspace for querying archived data
-   **Azure Data Lake**: Long-term storage for archived logs and metrics

## System Flow

### Data Ingestion Flow

1.  **User Authentication**: Users log in via MSAL within the organization
2.  **Project Setup**: Users create projects and receive unique secret keys for each project
3.  **Client Integration**: Applications integrate the npm package and configure it with the project secret key
4.  **Log/Metric Submission**: Applications send logs and metrics to gRPC Gateway using the npm package
5.  **Secret Key Authentication**: gRPC Gateway validates the secret key and establishes secure stream connection
6.  **Message Queuing**: Authenticated data is serialized to Protobuf format and sent to Kafka topics
7.  **Data Processing**: Main server consumes from Kafka and stores data in Elasticsearch
8.  **Client Access**: Users view real-time and historical data through the client interface

### Backup & Retention Flow

1.  **Retention Check**: Backup-cron function runs monthly to evaluate project retention policies
2.  **Data Archival**: Qualifying data is moved from Elasticsearch to Azure Data Lake
3.  **Analytics Access**: Archived data remains queryable through Azure Synapse

### Alert Management Flow

1.  **Rule Evaluation**: Alert-manager function runs every minute, checking rules against Elasticsearch data
2.  **Alert Triggering**: When conditions are met, alerts are sent via Redis pub/sub to main server
3.  **Notification Delivery**: Alerts are stored in Elasticsearch and delivered via:
    -   Client server notifications
    -   Webhook endpoints (if configured)
    -   Email notifications via Azure Communication Service


### Built for Self-Hosted Simplicity

Traditional logging solutions like ELK Stack, Grafana, and Prometheus require complex configurations and multiple moving parts. Our platform consolidates everything into one deployable solution:

-   **Single Deployment**: One comprehensive platform instead of managing 5+ separate tools
-   **Reduced Complexity**: Pre-configured integrations between all components
-   **Your Infrastructure, Your Control**: Deploy on your own servers with complete data ownership
-   **No Vendor Lock-in**: Open source means you're never trapped by licensing or pricing changes
-   **Community Driven**: Benefit from community contributions and enterprise features without enterprise costs

### From Startup to Scale

Whether you're a 3-person startup or a growing enterprise, deploy once and scale on your infrastructure:

-   **Resource Efficient**: Optimized to run on modest hardware initially, scales with your infrastructure
-   **Team Ready**: Built-in multi-tenancy and role-based access from day one
-   **Enterprise Features**: Advanced alerting, data retention policies, and backup strategies included

### Multi-Tenant Architecture

-   **Project Management**: Users can create and manage unlimited projects on a single platform
-   **Isolated Data**: Each project maintains separate log streams and configurations
-   **Role-Based Access**: Granular permissions and access control per project

### Real-Time Monitoring

-   **Live Log Streaming**: Real-time log and metric visualization
-   **Historical Analysis**: Query logs from specific date/time ranges
-   **Backup Data Access**: Seamless access to archived data via Synapse integration

### Advanced Alerting

-   **Custom Alert Rules**: Flexible rule engine for defining alert conditions
-   **Smart Deduplication**: Redis-based alert state management prevents spam
-   **Multiple Channels**: Support for webhooks, email, and in-app notifications

### High Performance & Scalability

-   **gRPC Protocol**: High-performance binary protocol for data ingestion
-   **Kafka Streaming**: Asynchronous processing handles high-throughput scenarios
-   **Schema Registry**: Protobuf serialization ensures data consistency and efficiency

### Easy Integration

-   **NPM Package**: Simple client-side SDK for quick application integration
-   **Secret Key Authentication**: Secure project-based authentication using unique secret keys
-   **gRPC Streaming**: High-performance streaming protocol for real-time log transmission
-   **Auto-Configuration**: NPM package handles connection management and data serialization

## Security & Authentication

### API Authentication

-   **Secret Key Authentication**: Each project has a unique secret key for gRPC Gateway access
-   **Secure Streaming**: Secret keys authenticate and establish secure gRPC stream connections
-   **NPM Package Integration**: Client applications use secret keys through the npm package for seamless authentication

## Data Management

### Retention Policies

-   **Configurable Retention**: Set custom retention periods per project
-   **Automated Archival**: Monthly cron jobs handle data lifecycle management
-   **Cost Optimization**: Move older data to cheaper storage while maintaining accessibility

### Data Formats

-   **Protobuf Serialization**: Efficient binary format for Kafka message streaming
-   **Schema Registry**: Ensures data consistency and enables schema evolution
-   **Elasticsearch Indexing**: Optimized for fast search and aggregation queries

## Technology Stack

### Core Services

-   **Backend**: Main server and gRPC gateway
-   **Frontend**: Client server for user interface
-   **Client SDK**: NPM package for application integration
-   **Message Broker**: Apache Kafka with Schema Registry
-   **Search Engine**: Elasticsearch for log storage and querying

### Cloud Services (Azure)

-   **Azure Functions**: Serverless compute for backup-cron and alert-manager
-   **Azure Data Lake**: Scalable data storage for archival
-   **Azure Synapse**: Analytics platform for archived data queries
-   **Azure Communication Service**: Email delivery for notifications

### Supporting Infrastructure

-   **PostgreSQL**: Relational database for metadata and configuration
-   **Redis**: In-memory store for caching and pub/sub messaging
-   **MSAL**: Microsoft Authentication Library for security

## Benefits

-   **Centralized Management**: Single platform for all logging needs across multiple projects
-   **Scalable Architecture**: Handles growing data volumes with horizontal scaling capabilities
-   **Cost-Effective Storage**: Intelligent data lifecycle management reduces storage costs
-   **Real-Time Insights**: Immediate visibility into system health and performance
-   **Flexible Alerting**: Customizable notification system prevents issues before they impact users
-   **Enterprise Security**: Microsoft authentication integration ensures secure access

We welcome contributions, feedback, and ideas! If you'd like to get involved, check out our CONTRIBUTING.md, open an issue, or submit a pull request. Don't forget to ⭐ the repo if you find it useful!




---

