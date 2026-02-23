# Production (Kamal) - Remaining Steps

## 1) First deploy

```bash
mise run kamal setup
```

## 2) Verify app

```bash
curl -i https://bandcash.app/health
mise run kamal app details
mise run kamal app logs
```

## 3) Boot Dozzle log viewer (optional)

```bash
mise run dozzle-boot
```

Open via SSH tunnel (kept private on server):

```bash
ssh -N -L 8088:127.0.0.1:8088 peti@bandcash
```

Then browse `http://localhost:8088`.

View Dozzle logs:

```bash
mise run dozzle-logs
```

## 4) Fallback/history log archive

```bash
ssh peti@bandcash "sudo tar -czf - /var/lib/docker/volumes/bandcash_data/_data/logs" > bandcash-logs-$(date +%Y%m%d-%H%M%S).tar.gz
```

## 5) Ongoing deploy

```bash
mise run kamal deploy
mise run kamal rollback
mise run kamal prune all
```
