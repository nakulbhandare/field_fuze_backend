# FieldFuze Backend - AWS Architecture Diagram

This diagram shows the complete AWS architecture for the FieldFuze backend application, including CI/CD pipeline, multi-AZ deployment, and monitoring.

```mermaid
graph TB
    %% Users and External Access
    subgraph "Internet"
        Users[👥 Users]
        Developers[👨‍💻 Developers]
    end

    %% DNS and CDN Layer
    subgraph "Global Edge Services"
        Route53[🌐 Route 53<br/>DNS Management]
        CloudFront[⚡ CloudFront<br/>Global CDN]
    end

    %% Load Balancing and Entry Point
    subgraph "AWS Region: us-east-1"
        subgraph "Public Subnets"
            ALB[⚖️ Application Load Balancer<br/>Port 80/443]
        end

        %% Multi-AZ Application Deployment
        subgraph "Availability Zone 1a"
            subgraph "Private Subnet 1a"
                ASG1a[🔄 Auto Scaling Group]
                EC2_1a[🖥️ EC2 Instance<br/>FieldFuze Backend<br/>Port 8081]
                EC2_2a[🖥️ EC2 Instance<br/>FieldFuze Backend<br/>Port 8081]
            end
        end

        subgraph "Availability Zone 1b"
            subgraph "Private Subnet 1b"
                ASG1b[🔄 Auto Scaling Group]
                EC2_1b[🖥️ EC2 Instance<br/>FieldFuze Backend<br/>Port 8081]
                EC2_2b[🖥️ EC2 Instance<br/>FieldFuze Backend<br/>Port 8081]
            end
        end

        %% Database Layer
        subgraph "Database Services"
            DynamoDB[(📊 DynamoDB<br/>Tables: users1, role<br/>Auto-scaling Enabled)]
        end

        %% Monitoring and Observability
        subgraph "Monitoring & Logging"
            CloudWatch[📈 CloudWatch<br/>Logs & Metrics]
            XRay[🔍 X-Ray<br/>Distributed Tracing]
        end
    end

    %% CI/CD Pipeline
    subgraph "DevOps Pipeline"
        GitHub[📁 GitHub Repository]
        GitHubActions[⚙️ GitHub Actions<br/>CI/CD Pipeline]
        ECR[📦 Amazon ECR<br/>Container Registry]
        CodeDeploy[🚀 AWS CodeDeploy<br/>Blue/Green Deployment]
    end

    %% External Services
    subgraph "External APIs"
        Telnyx[📞 Telnyx API<br/>Communication Services]
    end

    %% User Traffic Flow
    Users --> Route53
    Route53 --> CloudFront
    CloudFront --> ALB
    ALB --> EC2_1a
    ALB --> EC2_2a
    ALB --> EC2_1b
    ALB --> EC2_2b

    %% Auto Scaling Configuration
    ASG1a --> EC2_1a
    ASG1a --> EC2_2a
    ASG1b --> EC2_1b
    ASG1b --> EC2_2b

    %% Database Connections
    EC2_1a --> DynamoDB
    EC2_2a --> DynamoDB
    EC2_1b --> DynamoDB
    EC2_2b --> DynamoDB

    %% External API Connections
    EC2_1a --> Telnyx
    EC2_2a --> Telnyx
    EC2_1b --> Telnyx
    EC2_2b --> Telnyx

    %% Monitoring Connections
    EC2_1a --> CloudWatch
    EC2_2a --> CloudWatch
    EC2_1b --> CloudWatch
    EC2_2b --> CloudWatch
    EC2_1a --> XRay
    EC2_2a --> XRay
    EC2_1b --> XRay
    EC2_2b --> XRay
    ALB --> CloudWatch
    DynamoDB --> CloudWatch

    %% CI/CD Flow
    Developers --> GitHub
    GitHub --> GitHubActions
    GitHubActions --> ECR
    GitHubActions --> CodeDeploy
    ECR --> CodeDeploy
    CodeDeploy --> EC2_1a
    CodeDeploy --> EC2_2a
    CodeDeploy --> EC2_1b
    CodeDeploy --> EC2_2b

    %% Styling
    classDef userClass fill:#e1f5fe,stroke:#01579b,stroke-width:2px
    classDef awsService fill:#ff9800,stroke:#e65100,stroke-width:2px
    classDef compute fill:#4caf50,stroke:#1b5e20,stroke-width:2px
    classDef database fill:#9c27b0,stroke:#4a148c,stroke-width:2px
    classDef monitoring fill:#2196f3,stroke:#0d47a1,stroke-width:2px
    classDef cicd fill:#ff5722,stroke:#bf360c,stroke-width:2px
    classDef external fill:#607d8b,stroke:#263238,stroke-width:2px

    class Users,Developers userClass
    class Route53,CloudFront,ALB awsService
    class ASG1a,ASG1b,EC2_1a,EC2_2a,EC2_1b,EC2_2b compute
    class DynamoDB database
    class CloudWatch,XRay monitoring
    class GitHub,GitHubActions,ECR,CodeDeploy cicd
    class Telnyx external
```

## Architecture Components

### 1. **Global Edge Services**
- **Route 53**: DNS management and health checks
- **CloudFront**: Global CDN for content delivery and DDoS protection

### 2. **Load Balancing & Traffic Distribution**
- **Application Load Balancer (ALB)**: 
  - Distributes incoming traffic across multiple AZs
  - Supports HTTP/HTTPS termination
  - Health checks for backend instances

### 3. **Compute Layer (Multi-AZ)**
- **Auto Scaling Groups**: 
  - Automatic scaling based on CPU/memory utilization
  - Maintains desired capacity across availability zones
- **EC2 Instances**: 
  - Run FieldFuze Backend application on port 8081
  - Deployed across us-east-1a and us-east-1b for high availability

### 4. **Database Layer**
- **DynamoDB**: 
  - NoSQL database with tables: users1, role
  - Auto-scaling enabled for read/write capacity
  - Multi-AZ replication for high availability

### 5. **Monitoring & Observability**
- **CloudWatch**: 
  - Application and infrastructure metrics
  - Log aggregation and alerting
  - Custom dashboards for system health
- **X-Ray**: 
  - Distributed tracing for request flows
  - Performance analysis and debugging

### 6. **CI/CD Pipeline**
- **GitHub**: Source code repository
- **GitHub Actions**: 
  - Automated testing and building
  - Container image creation
- **Amazon ECR**: Container registry for Docker images
- **AWS CodeDeploy**: 
  - Blue/green deployment strategy
  - Zero-downtime deployments

### 7. **External Integrations**
- **Telnyx API**: Communication services integration

## Security Features

- **VPC**: Isolated network environment
- **Private Subnets**: EC2 instances in private subnets
- **Security Groups**: Network-level access controls
- **IAM Roles**: Service-specific permissions
- **JWT Authentication**: Stateless authentication tokens
- **HTTPS/TLS**: Encrypted communication

## High Availability & Scalability

- **Multi-AZ Deployment**: Instances across multiple availability zones
- **Auto Scaling**: Automatic capacity adjustment based on demand
- **Load Balancing**: Traffic distribution across healthy instances
- **Database Replication**: DynamoDB's built-in multi-AZ replication

## Deployment Strategy

- **Blue/Green Deployment**: Zero-downtime deployments using CodeDeploy
- **Container-based**: Docker containers for consistent deployments
- **Infrastructure as Code**: Automated infrastructure provisioning
