#!/bin/bash
# Diagnostico nginx pid pwrite EPERM en CentOS 7
# Ejecutar como: ./diagnose-nginx.sh

set +e

echo "=== Test A: nginx:stable-alpine fresco en este CentOS 7 ==="
echo "(replica la falla exacta con la imagen pura)"
docker run --rm --name nginx-fresh nginx:stable-alpine timeout 4 sh -c 'nginx -g "daemon off;" 2>&1 | head -30; echo "---FILES IN /run---"; ls -laZ /run/ 2>&1; echo "---DONE---"' 2>&1
echo

echo "=== Test B: replica EXACTA de la secuencia nginx (open sin O_TRUNC + lseek + pwrite) ==="
docker run --rm --name nginx-replica nginx:stable-alpine sh -c '
apk add --no-cache gcc musl-dev >/dev/null 2>&1
cat > /tmp/n.c <<CEOF
#include <unistd.h>
#include <fcntl.h>
#include <stdio.h>
#include <errno.h>
#include <string.h>
int main() {
    int fd = open("/run/nginx.pid", O_CREAT|O_RDWR, 0644);
    if (fd < 0) { printf("open fail: %s (errno=%d)\n", strerror(errno), errno); return 1; }
    printf("open OK fd=%d\n", fd);
    off_t r = lseek(fd, 0, SEEK_SET);
    printf("lseek = %ld\n", (long)r);
    int n = pwrite(fd, "1234\n", 5, 0);
    printf("pwrite = %d (errno=%d %s)\n", n, n<0?errno:0, n<0?strerror(errno):"OK");
    close(fd);
    return 0;
}
CEOF
gcc -o /tmp/n /tmp/n.c -static 2>&1 | tail -3
echo "--- ejecucion ---"
/tmp/n
echo "Exit: $?"
' 2>&1
echo

echo "=== Test C: estado del contenedor web que falla ==="
docker ps -a 2>&1 | grep -E "portal-sed-web|NAME"
echo "--- inspect ---"
docker inspect portal-sed-web-1 --format='Status={{.State.Status}} ExitCode={{.State.ExitCode}} Error={{.State.Error}} StartedAt={{.State.StartedAt}}' 2>&1
echo "--- logs ultima corrida ---"
docker logs --tail=10 portal-sed-web-1 2>&1
echo "--- /run dentro del contenedor (si esta corriendo) ---"
docker exec portal-sed-web-1 sh -c 'echo "ls /run:"; ls -la /run/ 2>&1; echo "ls -laZ /run:"; ls -laZ /run/ 2>&1; echo "id:"; id; echo "caps:"; grep ^Cap /proc/1/status 2>&1; echo "domain:"; cat /proc/1/attr/current 2>&1' 2>&1
echo

echo "=== Test D: arrancar web CON --privileged ==="
docker run -d --name web-priv --privileged portal-sed-web 2>&1
sleep 2
docker logs --tail=10 web-priv 2>&1
echo "---"
docker exec web-priv ls -la /run/nginx.pid 2>&1
docker rm -f web-priv 2>&1
echo

echo "=== Test E: arrancar web CON cap_add explicito ==="
docker run -d --name web-caps --cap-add=DAC_OVERRIDE --cap-add=SYS_ADMIN portal-sed-web 2>&1
sleep 2
docker logs --tail=10 web-caps 2>&1
docker exec web-caps ls -la /run/nginx.pid 2>&1
docker rm -f web-caps 2>&1
echo

echo "=== Test F: arrancar web SIN security_opt (default) ==="
docker run -d --name web-default portal-sed-web 2>&1
sleep 2
docker logs --tail=10 web-default 2>&1
docker exec web-default ls -la /run/nginx.pid 2>&1
docker rm -f web-default 2>&1
echo

echo "=== DONE ==="