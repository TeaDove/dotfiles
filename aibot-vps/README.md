# aibot-vps

```bash
cd dotfiles/aibot-vps

mkdir -p data/claude workspace secrets
touch ssh/known_hosts
ssh-keygen -t ed25519 -N "" -C claude-vps -f secrets/raspberry_ed25519
ssh-copy-id -i secrets/raspberry_ed25519.pub -p <port> claude@<home-endpoint>

cp ssh/config.example ssh/config
ssh-keyscan -p <port> <home-endpoint> > ssh/known_hosts
chown -R 1000:1000 data workspace secrets ssh/known_hosts

docker compose run --rm claude claude
docker compose up -d
```
