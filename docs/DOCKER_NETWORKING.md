# Day 3 - Docker Networking

## Completed Tasks

- Created custom Docker network
- Connected Redis container to network
- Tested container communication using ping
- Connected using redis-cli
- Stored and retrieved Redis values
- Removed container and network successfully

## Commands Used

```bash
docker network create sei-network
docker run -d --name sei-redis --network sei-network redis:7-alpine
docker run -it --rm --network sei-network alpine sh
ping sei-redis
docker run -it --rm --network sei-network redis:7-alpine redis-cli -h sei-redis
