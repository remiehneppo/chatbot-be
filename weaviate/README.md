# Portable Weaviate Setup

## Requirements

- Docker
- Docker Compose
- Ollama running on host machine

## Installation Steps

1. Extract the zip file:

```bash
unzip weaviate-portable.zip
```

2.Load the Docker image:

```bash
docker load < weaviate-offline.tar
```

3.Start Weaviate:

```bash
docker compose up -d
```

## Configuration

- Weaviate runs on port 8080
- Make sure Ollama is running on the host machine (port 11434)
- Data is persisted in a Docker volume
