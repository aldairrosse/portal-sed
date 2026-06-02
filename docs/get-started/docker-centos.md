# Docker en CentOS 7 (host) para API Go

## Host CentOS 7

```bash
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
sudo yum install -y docker-ce docker-ce-cli containerd.io
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -aG docker $USER
```

> CentOS 7 está EOL; para producción preferir Alma/Rocky 8+ o Amazon Linux 2023.

## Cuando exista `api/Dockerfile`

```bash
cd api
docker build -t sed-api .
docker run --rm -p 8080:8080 --env-file .env sed-api
```

## Compose (futuro en raíz)

```bash
docker compose up -d
```

Servicios previstos: `db`, `api`, `web`.
