# Alibaba Cloud ECS Deployment

This guide deploys RentNestHub with Docker Compose on a single ECS instance.
It assumes a Linux instance with Docker Engine and the Docker Compose plugin.

## 1. Prepare the Instance

- Use an instance with at least 2 vCPU, 4 GB RAM, and 30 GB free disk.
- Install Docker Engine and Docker Compose.
- Clone the repository and switch to the release branch or a tagged revision.
- Do not place `.env` in source control.

```bash
git clone git@github.com:JuhaoChen666/RentNestHub.git
cd RentNestHub
git switch main
cp .env.example .env
```

## 2. Configure Environment Variables

Set strong, unique database passwords in `.env` before starting the stack.
Set `PUBLIC_BASE_URL` to the public HTTPS origin used by tenants and landlords.

```dotenv
APP_PORT=80
MYSQL_DATABASE=rentnesthub
MYSQL_USER=rentnest
MYSQL_PASSWORD=replace-with-a-long-random-password
MYSQL_ROOT_PASSWORD=replace-with-another-long-random-password
PUBLIC_BASE_URL=https://rent.example.com
AI_API_URL=https://provider.example.com/v1/chat/completions
AI_API_KEY=replace-with-provider-secret
AI_MODEL=provider-model-name
```

Leave the three `AI_*` values empty to use deterministic local recommendations.
Never commit AI keys or database passwords.

## 3. Security Groups and Network Exposure

Allow inbound TCP 80 from the required public networks. Allow TCP 443 when a
TLS reverse proxy or load balancer is configured. Do not add a public rule for
TCP 3306.

The default Compose file maps MySQL for local development. On ECS, remove the
`mysql` service `ports` block before deployment, or restrict port 3306 to a
private VPC security group only.

Use an Alibaba Cloud Application Load Balancer, Nginx, or Caddy for TLS. The
proxy should forward traffic to `web` on the host `APP_PORT` and terminate TLS
before it reaches the container stack.

## 4. Start and Verify

```bash
docker compose pull
docker compose up -d --build
docker compose ps
curl -f http://127.0.0.1:${APP_PORT:-8080}/api/v1/health
```

The `mysql_data` volume stores database files. The `house_uploads` volume
stores property images. Both volumes must remain attached across releases.

## 5. Backup

Back up the database before application or schema changes. Keep the backup
outside the ECS instance when possible.

```bash
docker compose exec -T mysql \
  mysqldump -u root -p"$MYSQL_ROOT_PASSWORD" "$MYSQL_DATABASE" \
  > rentnesthub-backup.sql
```

List persistent volumes with:

```bash
docker volume ls
```

Back up uploaded images by archiving the `house_uploads` volume or copying it
to object storage on a schedule.

## 6. Upgrade and Rollback

```bash
git fetch origin
git switch main
git pull --ff-only
docker compose up -d --build
docker compose ps
```

To roll back, switch to the previous known-good tag or commit and rebuild. Do
not run `docker compose down -v` in production: the `-v` flag removes database
and upload volumes.

```bash
git switch <previous-tag-or-commit>
docker compose up -d --build
```

## 7. Operations

Inspect logs with `docker compose logs -f api web mysql`. The Compose restart
policy restarts containers after a process failure. Configure host-level reboot
startup for Docker so the stack is restored after ECS maintenance or restart.
