# Docker Compose Guide

## Start services

```bash
docker compose up --build
```

## Stop services

```bash
docker compose down
```

## View running containers

```bash
docker compose ps
```

## View logs

```bash
docker compose logs
```

## Redis CLI access

```bash
docker compose exec redis redis-cli
```

## Restart Redis

```bash
docker compose restart redis
```

## Health checks

```bash
curl http://localhost:8080/health
curl http://localhost:8081/health
```

## Resource limits

- CPU limit: 0.50
- Memory limit: 256M
- Redis maxmemory: 128mb
- Redis eviction policy: allkeys-lru
