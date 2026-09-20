# aibot-vps

```bash
cd dotfiles/aibot-vps

mkdir -p data/claude workspace secrets
touch ssh/known_hosts
ssh-keygen -t ed25519 -N "" -C claude-vps -f secrets/raspberry_ed25519
ssh-copy-id -i secrets/raspberry_ed25519.pub -p <port> teadove@<home-endpoint>
 printf '%s' '<raspberry-sudo-password>' > secrets/raspberry_sudo_pass
 printf '%s' '<github-token>' > secrets/github_token
chmod 600 secrets/raspberry_sudo_pass secrets/github_token

cp ssh/config.example ssh/config
ssh-keyscan -p <port> <home-endpoint> > ssh/known_hosts
chown -R 1000:1000 data workspace secrets ssh/known_hosts

docker compose run --rm claude claude # Login once
docker compose run --rm claude claude rc # Enable RC once
docker compose up -d
```
