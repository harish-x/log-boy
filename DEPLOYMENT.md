# Logboy Deployment Guide

This guide provides step-by-step instructions for deploying the centralised logging server on Azure infrastructure.

## Architecture Overview

The platform consists of multiple Azure services orchestrated within a Virtual Network:
- **Client Server** (Azure App Services) - Frontend dashboard
- **gRPC Gateway** (Azure VM) - Log/metric ingestion endpoint
- **Main Server** (Azure VM) - Core business logic
- **Kafka Cluster** - Message streaming with Schema Registry
- **Elasticsearch Cluster** - Primary log/metric storage
- **PostgreSQL** - Project and authentication data
- **Redis** - Alert management and pub/sub
- **Azure Functions** - Backup cron and alert monitoring
- **Azure Services** - Data Lake, Synapse, Communication Service

## Prerequisites

### Azure Requirements
- Active Azure subscription with sufficient quotas
- Azure CLI installed and configured
- Terraform or ARM templates (optional but recommended)
- Resource Group creation permissions

### Development Tools
- Docker and Docker Compose
- Node.js 18+ for Azure Functions
- Git for source code management
- kubectl (if using AKS alternative)

### Required Azure Resources Quota
- Virtual Machines: 4 instances 
  - gRPC Gateway: Standard_D4s_v3
  - Main Server: Standard_D4s_v3  
  - Kafka Cluster: Standard_D8s_v3
  - Elasticsearch: Standard_E8s_v3 (memory optimized)
- App Service Plans: 1 instance
- Storage Accounts: 2 instances
- PostgreSQL Database: 1 instance
- Function Apps: 2 instances

## Step-by-Step Deployment

### 1. Setup Azure Active Directory Application

First, create an Azure AD application for MSAL authentication:

```bash
# Login to Azure
az login

# Register the Client Application

az ad app create \
  --display-name "Centralized Logging - Client App" \
  --sign-in-audience "AzureADMyOrg" \
  --web-redirect-uris "https://your-domain.com/auth/callback" \
  --enable-id-token-issuance true

# Register the Server Application (API)

az ad app create \
  --display-name "Centralized Logging - Server App" \
  --sign-in-audience "AzureADMyOrg" \
  --identifier-uris "api://centralized-logging-server" \
  --enable-access-token-issuance true

# Define OAuth Permissions from Client → Server App

az ad app permission add \
  --id <CLIENT_APP_ID> \
  --api <SERVER_APP_ID> \
  --api-permissions <SERVER_APP_ID>/user_impersonation=Scope

# Grant Admin Consent for Permissions

az ad app permission grant \
  --id <CLIENT_APP_ID> \
  --api <SERVER_APP_ID>

az ad app permission admin-consent --id <CLIENT_APP_ID>

```

**Important**: Save the Client ID, Tenant ID, and Client Secret - you'll need these for configuration.

### 2. Create Azure Resource Group and Virtual Network

```bash
# Create resource group
az group create \
  --name "logging-platform-rg" \
  --location "East US"

# Create virtual network
az network vnet create \
  --resource-group "logging-platform-rg" \
  --name "logging-platform-vnet" \
  --address-prefix "10.0.0.0/16" \
  --subnet-name "default" \
  --subnet-prefix "10.0.1.0/24"

# Create additional subnets
az network vnet subnet create \
  --resource-group "logging-platform-rg" \
  --vnet-name "logging-platform-vnet" \
  --name "database-subnet" \
  --address-prefix "10.0.2.0/24"

az network vnet subnet create \
  --resource-group "logging-platform-rg" \
  --vnet-name "logging-platform-vnet" \
  --name "functions-subnet" \
  --address-prefix "10.0.3.0/24"
```

### 3. Deploy Database Services

#### PostgreSQL Database
```bash
# Create PostgreSQL server
az postgres flexible-server create \
  --resource-group "logging-platform-rg" \
  --name "logging-platform-postgres" \
  --location "East US" \
  --admin-user "logadmin" \
  --admin-password "SecurePassword123!" \
  --sku-name "Standard_D2s_v3" \
  --tier "GeneralPurpose" \
  --storage-size 128 \
  --version 14

# Create database
az postgres flexible-server db create \
  --resource-group "logging-platform-rg" \
  --server-name "logging-platform-postgres" \
  --database-name "logging_platform"

# Configure firewall (allow Azure services)
az postgres flexible-server firewall-rule create \
  --resource-group "logging-platform-rg" \
  --name "logging-platform-postgres" \
  --rule-name "AllowAzureServices" \
  --start-ip-address "0.0.0.0" \
  --end-ip-address "0.0.0.0"
```

#### Redis Cache
```bash
# Create Redis cache
az redis create \
  --resource-group "logging-platform-rg" \
  --name "logging-platform-redis" \
  --location "East US" \
  --sku "Premium" \
  --vm-size "P1"
```

### 4. Deploy Virtual Machines

#### gRPC Gateway Server
```bash
# Create VM for gRPC Gateway
az vm create \
  --resource-group "logging-platform-rg" \
  --name "grpc-gateway-vm" \
  --image "Ubuntu2204" \
  --size "Standard_D4s_v3" \
  --vnet-name "logging-platform-vnet" \
  --subnet "default" \
  --admin-username "azureuser" \
  --ssh-key-values ~/.ssh/id_rsa.pub \
  --public-ip-sku "Standard" \
  --nsg-rule "SSH"

# Open port 8080 for gRPC
az vm open-port \
  --resource-group "logging-platform-rg" \
  --name "grpc-gateway-vm" \
  --port "8080"
```

#### Main Server
```bash
# Create VM for Main Server
az vm create \
  --resource-group "logging-platform-rg" \
  --name "main-server-vm" \
  --image "Ubuntu2204" \
  --size "Standard_D4s_v3" \
  --vnet-name "logging-platform-vnet" \
  --subnet "default" \
  --admin-username "azureuser" \
  --ssh-key-values ~/.ssh/id_rsa.pub \
  --public-ip-sku "Standard" \
  --nsg-rule "SSH"

# Open port 3000 for API
az vm open-port \
  --resource-group "logging-platform-rg" \
  --name "main-server-vm" \
  --port "3000"
```

#### Kafka Cluster VM
```bash
# Create VM for Kafka Cluster
az vm create \
  --resource-group "logging-platform-rg" \
  --name "kafka-cluster-vm" \
  --image "Ubuntu2204" \
  --size "Standard_D8s_v3" \
  --vnet-name "logging-platform-vnet" \
  --subnet "default" \
  --admin-username "azureuser" \
  --ssh-key-values ~/.ssh/id_rsa.pub \
  --public-ip-sku "Standard" \
  --nsg-rule "SSH"

# Open Kafka ports
az vm open-port \
  --resource-group "logging-platform-rg" \
  --name "kafka-cluster-vm" \
  --port "9092"

az vm open-port \
  --resource-group "logging-platform-rg" \
  --name "kafka-cluster-vm" \
  --port "9093"

az vm open-port \
  --resource-group "logging-platform-rg" \
  --name "kafka-cluster-vm" \
  --port "8081"  # Schema Registry

az vm open-port \
  --resource-group "logging-platform-rg" \
  --name "kafka-cluster-vm" \
  --port "2181"  # Zookeeper
```

#### Elasticsearch Cluster VM
```bash
# Create VM for Elasticsearch Cluster (requires more resources)
az vm create \
  --resource-group "logging-platform-rg" \
  --name "elasticsearch-vm" \
  --image "Ubuntu2204" \
  --size "Standard_E8s_v3" \
  --vnet-name "logging-platform-vnet" \
  --subnet "default" \
  --admin-username "azureuser" \
  --ssh-key-values ~/.ssh/id_rsa.pub \
  --public-ip-sku "Standard" \
  --nsg-rule "SSH"

# Open Elasticsearch ports
az vm open-port \
  --resource-group "logging-platform-rg" \
  --name "elasticsearch-vm" \
  --port "9200"

az vm open-port \
  --resource-group "logging-platform-rg" \
  --name "elasticsearch-vm" \
  --port "9300"
```

### 5. Install and Configure Services on VMs

#### SSH into each VM and install Docker:
```bash
# SSH into VM
ssh azureuser@<VM_PUBLIC_IP>

# Install Docker
sudo apt update
sudo apt install -y docker.io docker-compose
sudo usermod -aG docker $USER
sudo systemctl enable docker
sudo systemctl start docker

# Logout and login again to apply group changes
exit
ssh azureuser@<VM_PUBLIC_IP>
```

#### Deploy Kafka Cluster (on Kafka Cluster VM)
```bash
# SSH into Kafka VM
ssh azureuser@<KAFKA_VM_PUBLIC_IP>

# Create docker-compose.yml for Kafka
cat > kafka-compose.yml << 'EOF'
version: '3.8'
services:
  zookeeper:
    image: confluentinc/cp-zookeeper:7.4.0
    environment:
      ZOOKEEPER_CLIENT_PORT: 2181
      ZOOKEEPER_TICK_TIME: 2000
    ports:
      - "2181:2181"
    volumes:
      - zookeeper-data:/var/lib/zookeeper/data
      - zookeeper-logs:/var/lib/zookeeper/log

  kafka-broker-1:
    image: confluentinc/cp-kafka:7.4.0
    depends_on:
      - zookeeper
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://10.0.1.X:9092  # Use internal IP
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 2
    volumes:
      - kafka-broker-1-data:/var/lib/kafka/data

  kafka-broker-2:
    image: confluentinc/cp-kafka:7.4.0
    depends_on:
      - zookeeper
    ports:
      - "9093:9093"
    environment:
      KAFKA_BROKER_ID: 2
      KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://10.0.1.X:9093  # Use internal IP
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 2
    volumes:
      - kafka-broker-2-data:/var/lib/kafka/data

  schema-registry:
    image: confluentinc/cp-schema-registry:7.4.0
    depends_on:
      - kafka-broker-1
      - kafka-broker-2
    ports:
      - "8081:8081"
    environment:
      SCHEMA_REGISTRY_HOST_NAME: schema-registry
      SCHEMA_REGISTRY_KAFKASTORE_BOOTSTRAP_SERVERS: kafka-broker-1:9092,kafka-broker-2:9093
      SCHEMA_REGISTRY_LISTENERS: http://0.0.0.0:8081

volumes:
  zookeeper-data:
  zookeeper-logs:
  kafka-broker-1-data:
  kafka-broker-2-data:
EOF

# Start Kafka cluster
docker-compose -f kafka-compose.yml up -d
```

#### Deploy Elasticsearch Cluster (on Elasticsearch VM)
```bash
# SSH into Elasticsearch VM
ssh azureuser@<ELASTICSEARCH_VM_PUBLIC_IP>

# Increase virtual memory for Elasticsearch
sudo sysctl -w vm.max_map_count=262144
echo 'vm.max_map_count=262144' | sudo tee -a /etc/sysctl.conf

# Create docker-compose.yml for Elasticsearch
cat > elasticsearch-compose.yml << 'EOF'
version: '3.8'
services:
  elasticsearch-primary:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.9.0
    environment:
      - node.name=primary-node
      - cluster.name=logging-cluster
      - discovery.seed_hosts=secondary-node-1,secondary-node-2
      - cluster.initial_master_nodes=primary-node
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms4g -Xmx4g"
      - xpack.security.enabled=false
      - network.host=0.0.0.0
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - elasticsearch-primary-data:/usr/share/elasticsearch/data
    ports:
      - "9200:9200"
      - "9300:9300"

  elasticsearch-secondary-1:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.9.0
    environment:
      - node.name=secondary-node-1
      - cluster.name=logging-cluster
      - discovery.seed_hosts=primary-node,secondary-node-2
      - cluster.initial_master_nodes=primary-node
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms4g -Xmx4g"
      - xpack.security.enabled=false
      - network.host=0.0.0.0
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - elasticsearch-secondary-1-data:/usr/share/elasticsearch/data
    ports:
      - "9201:9200"
      - "9301:9300"

  elasticsearch-secondary-2:
    image: docker.elastic.co/elasticsearch/elasticsearch:8.9.0
    environment:
      - node.name=secondary-node-2
      - cluster.name=logging-cluster
      - discovery.seed_hosts=primary-node,secondary-node-1
      - cluster.initial_master_nodes=primary-node
      - bootstrap.memory_lock=true
      - "ES_JAVA_OPTS=-Xms4g -Xmx4g"
      - xpack.security.enabled=false
      - network.host=0.0.0.0
    ulimits:
      memlock:
        soft: -1
        hard: -1
    volumes:
      - elasticsearch-secondary-2-data:/usr/share/elasticsearch/data
    ports:
      - "9202:9200"
      - "9302:9300"

volumes:
  elasticsearch-primary-data:
  elasticsearch-secondary-1-data:
  elasticsearch-secondary-2-data:
EOF

# Start Elasticsearch cluster
docker-compose -f elasticsearch-compose.yml up -d
```

#### Deploy gRPC Gateway Service (on gRPC Gateway VM)
```bash
# SSH into gRPC Gateway VM
ssh azureuser@<GRPC_GATEWAY_VM_PUBLIC_IP>

# Clone your application repository
git clone <YOUR_GRPC_GATEWAY_REPO>
cd <REPO_NAME>

# Create .env file
cat > .env << 'EOF'
POSTGRESQL_HOST=<POSTGRES_SERVER_NAME>.postgres.database.azure.com
POSTGRESQL_PORT=5432
POSTGRESQL_DATABASE=logging_platform
POSTGRESQL_USER=logadmin
POSTGRESQL_PASSWORD=SecurePassword123!
KAFKA_BROKERS=<KAFKA_VM_INTERNAL_IP>:9092,<KAFKA_VM_INTERNAL_IP>:9093
SCHEMA_REGISTRY_URL=http://<KAFKA_VM_INTERNAL_IP>:8081
GRPC_PORT=8080
EOF

# Build and run the service
docker build -t grpc-gateway .
docker run -d -p 8080:8080 --env-file .env --name grpc-gateway grpc-gateway
```

#### Deploy Main Server Service (on Main Server VM)
```bash
# SSH into Main Server VM
ssh azureuser@<MAIN_SERVER_VM_PUBLIC_IP>

# Clone your application repository
git clone <YOUR_MAIN_SERVER_REPO>
cd <REPO_NAME>

# Create .env file
cat > .env << 'EOF'
POSTGRESQL_HOST=<POSTGRES_SERVER_NAME>.postgres.database.azure.com
POSTGRESQL_PORT=5432
POSTGRESQL_DATABASE=logging_platform
POSTGRESQL_USER=logadmin
POSTGRESQL_PASSWORD=SecurePassword123!
ELASTICSEARCH_URL=http://<ELASTICSEARCH_VM_INTERNAL_IP>:9200
REDIS_HOST=<REDIS_NAME>.redis.cache.windows.net
REDIS_PORT=6380
REDIS_PASSWORD=<REDIS_ACCESS_KEY>
KAFKA_BROKERS=<KAFKA_VM_INTERNAL_IP>:9092,<KAFKA_VM_INTERNAL_IP>:9093
AZURE_COMMUNICATION_CONNECTION_STRING=<COMMUNICATION_SERVICE_CONNECTION_STRING>
API_PORT=3000
EOF

# Build and run the service
docker build -t main-server .
docker run -d -p 3000:3000 --env-file .env --name main-server main-server
```

### 6. Deploy Client Server (Azure App Service)

```bash
# Create App Service Plan
az appservice plan create \
  --resource-group "logging-platform-rg" \
  --name "logging-platform-plan" \
  --location "East US" \
  --sku "P1V2" \
  --is-linux

# Create Web App
az webapp create \
  --resource-group "logging-platform-rg" \
  --plan "logging-platform-plan" \
  --name "logging-platform-client" \
  --runtime "NODE:18-lts"

# Configure app settings
az webapp config appsettings set \
  --resource-group "logging-platform-rg" \
  --name "logging-platform-client" \
  --settings \
    AZURE_CLIENT_ID="<YOUR_CLIENT_ID>" \
    AZURE_TENANT_ID="<YOUR_TENANT_ID>" \
    AZURE_CLIENT_SECRET="<YOUR_CLIENT_SECRET>" \
    MAIN_SERVER_URL="http://<MAIN_SERVER_VM_IP>:3000" \
    SYNAPSE_ENDPOINT="<SYNAPSE_ENDPOINT>"
```

### 7. Deploy Azure Functions

#### Backup Cron Functions
```bash
# Create Function App for backup
az functionapp create \
  --resource-group "logging-platform-rg" \
  --consumption-plan-location "East US" \
  --runtime "node" \
  --runtime-version "18" \
  --functions-version "4" \
  --name "logging-backup-functions" \
  --storage-account "loggingplatformstorage"

# Configure function app settings
az functionapp config appsettings set \
  --resource-group "logging-platform-rg" \
  --name "logging-backup-functions" \
  --settings \
    POSTGRESQL_CONNECTION_STRING="<POSTGRES_CONNECTION_STRING>" \
    ELASTICSEARCH_URL="http://<MAIN_SERVER_VM_IP>:9200" \
    AZURE_STORAGE_CONNECTION_STRING="<STORAGE_CONNECTION_STRING>"
```

#### Alert Manager Functions
```bash
# Create Function App for alerts
az functionapp create \
  --resource-group "logging-platform-rg" \
  --consumption-plan-location "East US" \
  --runtime "node" \
  --runtime-version "18" \
  --functions-version "4" \
  --name "logging-alert-functions" \
  --storage-account "loggingplatformstorage"

# Configure function app settings
az functionapp config appsettings set \
  --resource-group "logging-platform-rg" \
  --name "logging-alert-functions" \
  --settings \
    POSTGRESQL_CONNECTION_STRING="<POSTGRES_CONNECTION_STRING>" \
    ELASTICSEARCH_URL="http://<MAIN_SERVER_VM_IP>:9200" \
    REDIS_CONNECTION_STRING="<REDIS_CONNECTION_STRING>"
```

### 8. Deploy Supporting Azure Services

#### Azure Data Lake Storage
```bash
# Create storage account for Data Lake
az storage account create \
  --resource-group "logging-platform-rg" \
  --name "loggingdatalake" \
  --location "East US" \
  --sku "Standard_LRS" \
  --kind "StorageV2" \
  --hierarchical-namespace true

# Create containers
az storage container create \
  --account-name "loggingdatalake" \
  --name "logs-archive" \
  --auth-mode login

az storage container create \
  --account-name "loggingdatalake" \
  --name "metrics-archive" \
  --auth-mode login
```

#### Azure Synapse Analytics
```bash
# Create Synapse workspace
az synapse workspace create \
  --resource-group "logging-platform-rg" \
  --name "logging-synapse-workspace" \
  --storage-account "loggingdatalake" \
  --file-system "synapsefilesystem" \
  --sql-admin-login-user "synapseadmin" \
  --sql-admin-login-password "SecurePassword123!" \
  --location "East US"

# Create SQL pool
az synapse sql pool create \
  --resource-group "logging-platform-rg" \
  --workspace-name "logging-synapse-workspace" \
  --name "LoggingPool" \
  --performance-level "DW100c"
```

#### Azure Communication Service
```bash
# Create communication service
az communication create \
  --resource-group "logging-platform-rg" \
  --name "logging-communication-service" \
  --location "Global"

# Get connection string
az communication list-key \
  --resource-group "logging-platform-rg" \
  --name "logging-communication-service"
```

## ⚙️ Configuration

### Environment Variables

Create `.env` files for each service with the following variables:

#### gRPC Gateway Server (.env)
```bash
POSTGRES_DB=<POSTGRES_URL>
KAFKA_HOST=<kafka URI>
SCHEMA_REGISTRY_URL=<schema registry URI>
SERVER_PORT=8080
GRPC_SECRET = <secret key to hash grpc tokens>
```

#### Main Server (.env)
```bash
SERVER_PORT=:8080
GRPC_PORT=<grpc host URI>
GRPC_SECRET=<secret key to hash grpc tokens>
DIRECTORY_TENANT_ID=< azure backend AD tennent ID>
APPLICATION_CLIENT_ID=< azure AD client tennent ID>
AZURE_STORAGE_ACCOUNT_NAME=<azure strorage account name>
COLD_STORAGE_CONTAINER=<azure storage account container>
AZURE_STORAGE_KEY=<azure storage account key>
SYNAPSE_DB=< azure synapse host>
POSTGRES_DB=<POSTGRES_URL>
ELASTIC_SEARCH_DNS=<elasticsearch URI>
KAFKA_BROKERS=<kafka URI>
SCHEMA_REGISTRY_URL=<schema registry URI>
AZURE_EMAIL_SENDER_ADDRESS=<your azure communication service email sender address>
```

##  Security Configuration

### Network Security Groups
```bash
# Create NSG
az network nsg create \
  --resource-group "logging-platform-rg" \
  --name "logging-platform-nsg"

# Add rules
az network nsg rule create \
  --resource-group "logging-platform-rg" \
  --nsg-name "logging-platform-nsg" \
  --name "AllowgRPC" \
  --priority 100 \
  --direction "Inbound" \
  --access "Allow" \
  --protocol "TCP" \
  --destination-port-ranges "8080"

az network nsg rule create \
  --resource-group "logging-platform-rg" \
  --nsg-name "logging-platform-nsg" \
  --name "AllowHTTPS" \
  --priority 200 \
  --direction "Inbound" \
  --access "Allow" \
  --protocol "TCP" \
  --destination-port-ranges "443"
```

### SSL/TLS Configuration
- Configure SSL certificates for all public endpoints
- Use Azure Key Vault for certificate management
- Enable HTTPS redirect on App Services
- Enable mTLS between gRPc gateway and your application server (optional)



##  Deployment Verification

### 1. Test gRPC Gateway
```bash
# Test gRPC endpoint
grpcurl -plaintext <GRPC_GATEWAY_IP>:8080 list
```

### 2. Test Client Server
```bash
# Access web interface
curl https://<CLIENT_SERVER_URL>/health
```

### 3. Test Database Connections
```bash
# Test PostgreSQL connection
psql "host=<POSTGRES_HOST> port=5432 dbname=logging_platform user=logadmin password=SecurePassword123! sslmode=require"

# Test Redis connection
redis-cli -h <REDIS_HOST> -p 6380 -a <REDIS_PASSWORD> ping
```

### 4. Test Kafka Topics
```bash
# List Kafka topics
docker exec -it <KAFKA_CONTAINER> kafka-topics --list --bootstrap-server localhost:9092
```

### Monitoring Commands
```bash
# Check Docker services
docker-compose ps

# View logs
docker-compose logs -f <service-name>

# Check Azure resources
az resource list --resource-group "logging-platform-rg" --output table
```

##  Post-Deployment Steps

1. **Configure Backup Schedules**
   - Set up automated backups for databases
   - Configure retention policies
   - Test restore procedures

2. **Set Up Monitoring Dashboards**
   - Create Azure Monitor dashboards
   - Configure alerting rules
   - Set up notification channels

3. **Performance Tuning**
   - Optimize Elasticsearch indices
   - Configure Kafka partitions
   - Tune VM sizes based on load

4. **Security Hardening**
   - Enable Azure Security Center
   - Configure access policies
   - Set up audit logging

## Next Steps

After successful deployment:
1. Deploy your application code to each service
2. Configure the NPM package for client applications
3. Set up CI/CD pipelines for automated deployments
4. Create user documentation and onboarding guides
5. Implement backup and disaster recovery procedures

---

**Congratulations!** Your centralized logging server is now deployed on Azure. The platform is ready to receive logs and metrics from your applications through the NPM package integration.
