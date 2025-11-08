# FieldFuze EC2 Deployment

Simple Docker Compose deployment for AWS EC2 with automatic GitHub Actions CI/CD.

## 🚀 Quick Setup

### 1. EC2 Instance Setup

**Launch EC2 Instance:**
- Instance type: t3.medium or larger
- AMI: Amazon Linux 2023 or Ubuntu 22.04
- Security Group: Allow ports 22 (SSH), 80 (HTTP), 8081 (API)

**Install Docker & Docker Compose:**
```bash
# Amazon Linux 2023
sudo yum update -y
sudo yum install -y docker
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -a -G docker $USER

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Logout and login again for group changes
exit
```

### 2. GitHub Secrets Configuration

Add these secrets in your GitHub repository settings:

```
AWS_REGION=us-east-1
AWS_ACCESS_KEY_ID=your-aws-access-key-id
AWS_SECRET_ACCESS_KEY=your-aws-secret-access-key
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
EC2_HOST=your-ec2-public-ip-or-domain
EC2_USER=ec2-user
EC2_SSH_PRIVATE_KEY=your-private-key-content
```

### 3. Manual Deployment (Optional)

```bash
# Clone repository
git clone <your-repo-url>
cd field_fuze_backend/deployment

# Copy environment file
cp .env.example .env
# Edit .env with your actual values

# Deploy
docker-compose up -d --build

# Check logs
docker-compose logs -f
```

## 📱 Access URLs

After successful deployment:

- **API Health:** `http://your-ec2-ip:8081/api/v1/auth/health`
- **Swagger UI:** `http://your-ec2-ip:8081/swagger`
- **Main API:** `http://your-ec2-ip:8081/api/v1/auth`

## 🔧 Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AWS_REGION` | AWS region | `us-east-1` |
| `AWS_ACCESS_KEY_ID` | AWS access key | Required |
| `AWS_SECRET_ACCESS_KEY` | AWS secret key | Required |
| `JWT_SECRET` | JWT signing secret | Required |
| `DYNAMODB_TABLE_PREFIX` | DynamoDB table prefix | `prod` |
| `LOG_LEVEL` | Logging level | `info` |
| `LOG_FORMAT` | Log format | `json` |

### Port Mapping

- Container port `8081` → Host ports `80` and `8081`
- Access via both `http://server/` and `http://server:8081/`

## 🐳 Docker Commands

```bash
# View running containers
docker-compose ps

# View logs
docker-compose logs fieldfuze-backend

# Restart application
docker-compose restart fieldfuze-backend

# Update and redeploy
git pull
docker-compose down
docker-compose up -d --build

# Clean up old images
docker image prune -f
```

## 🚀 CI/CD Workflow

The GitHub Actions workflow automatically:

1. **Tests** - Runs Go tests, vet, and formatting checks
2. **Deploys** - Copies files to EC2 and runs Docker Compose
3. **Verifies** - Checks application health after deployment

**Triggered on:**
- Push to `master` branch
- Manual workflow dispatch

## 🔍 Troubleshooting

### Application won't start
```bash
# Check logs
docker-compose logs fieldfuze-backend

# Check if DynamoDB tables exist
aws dynamodb list-tables --region us-east-1
```

### Port conflicts
```bash
# Check what's using port 8081
sudo lsof -i :8081

# Stop conflicting services
sudo systemctl stop service-name
```

### GitHub Actions deployment fails
- Verify all secrets are set correctly
- Check EC2 security group allows SSH (port 22)
- Ensure EC2 user has Docker permissions

## 💡 Production Considerations

1. **Security:**
   - Use IAM roles instead of access keys
   - Rotate JWT secrets regularly
   - Enable HTTPS with SSL certificates

2. **Monitoring:**
   - Set up CloudWatch logs
   - Configure health checks
   - Monitor resource usage

3. **Backup:**
   - Regular DynamoDB backups
   - Application data backup strategy

## 🔗 Architecture

```
GitHub → GitHub Actions → EC2 Instance
                            ↓
                        Docker Compose
                            ↓
                      FieldFuze Container (Port 8081)
                            ↓
                        AWS DynamoDB
```

**Cost:** ~$25-50/month for t3.medium EC2 instance + DynamoDB usage