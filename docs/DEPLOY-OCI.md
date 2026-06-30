# Deploying Paint War to Oracle Cloud (Always Free, Arm A1)

A step-by-step runbook for deploying the production stack to a single OCI
Always Free VM. Each step includes a short **Why** note so it doubles as a
learning guide.

## Architecture

```
                Internet (HTTPS/WSS, 443)
                        |
              +---------v----------+   one OCI A1.Flex VM
              |       Caddy        |   (Ubuntu 22.04 arm64)
              |  auto Let's Encrypt|
              |  / , /assets  -> static client files
              |  /ws , /healthz -> reverse_proxy server:8080
              +---+------------+---+
                  |            |  (docker bridge network)
            +-----v----+  +----v----------+
            | server   |  |  database     |
            | Go :8080 |--| postgres:16   |  (named volume)
            +----------+  +---------------+
```

The browser loads `https://you.duckdns.org` and the client opens
`wss://you.duckdns.org/ws` automatically (single origin) -- no per-environment
rebuild. This works because `client/src/lib/net/socket.svelte.js` derives the
socket URL from the page origin when `VITE_WS_URL` is unset.

The production images are built and verified in the repo:
- `server/Dockerfile` target `production` -> ~15 MB scratch binary
- `deploy/Dockerfile.caddy` -> builds the static SPA and bakes it into Caddy
- `compose.prod.yaml` -> Caddy + server + Postgres (no dev cruft)

---

## Part A -- Provision the OCI infrastructure (Console)

### 1. Account and region
Sign up for an OCI account and confirm you are on the **Always Free** resources
(not just the 30-day trial). Pick a home region that has Arm A1 capacity.

**Why:** Always Free includes up to 4 OCPU / 24 GB RAM of Ampere A1 (Arm)
compute that never expires -- ideal for this stack.

### 2. Create the network (VCN)
In **Networking -> Virtual Cloud Networks**, use the **"Create VCN with Internet
Connectivity"** wizard. It provisions a VCN, a **public subnet**, an **Internet
Gateway**, and a **route table** with `0.0.0.0/0 -> Internet Gateway`.

**Why:** a VCN is your isolated private network in the cloud. The Internet
Gateway plus the route rule are what make instances in the public subnet
reachable from the internet.

### 3. Open the cloud firewall (Security List)
On the public subnet's **Security List**, add **ingress** rules:

| Port | Source | Purpose |
|---|---|---|
| 22/tcp | your IP/32 (preferred) or 0.0.0.0/0 | SSH |
| 80/tcp | 0.0.0.0/0 | HTTP (Let's Encrypt challenge + redirect) |
| 443/tcp | 0.0.0.0/0 | HTTPS / WSS |

**Why:** the Security List is a stateful L4 firewall at the subnet edge. Ports
80 and 443 must be open publicly for both gameplay and the ACME certificate
challenge.

### 4. Launch the instance
**Compute -> Instances -> Create**:
- Shape: **VM.Standard.A1.Flex** (e.g. 2 OCPU / 12 GB -- leaves free-tier headroom).
- Image: **Canonical Ubuntu 22.04 (aarch64)**.
- Networking: the VCN and public subnet from step 2; **assign a public IPv4**.
- Add your **SSH public key**.

After it boots, go to the instance's VNIC and **reserve** the public IP.

**Why:** a *reserved* (static) public IP survives stop/start, so your DNS record
stays valid. Arm A1 capacity can be scarce -- if you see "out of capacity", retry
or try another availability domain/region.

---

## Part B -- Configure the host (SSH)

### 5. Open the OS firewall (the #1 OCI gotcha)
Oracle's Ubuntu image ships `iptables` rules that block everything except SSH.
SSH in and open 80/443:

```bash
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 80  -j ACCEPT
sudo iptables -I INPUT 6 -m state --state NEW -p tcp --dport 443 -j ACCEPT
sudo netfilter-persistent save
```

**Why:** OCI has **two** firewall layers -- the subnet Security List (step 3)
*and* the instance's own OS firewall. Both must allow a port or traffic is
silently dropped. This is the single most common reason an OCI deploy "works"
but is unreachable.

### 6. Install Docker Engine + Compose plugin
```bash
sudo apt-get update
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo $VERSION_CODENAME) stable" \
  | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo usermod -aG docker $USER
newgrp docker   # or log out/in
```

**Why:** the official apt repo serves an `arm64` build automatically
(`dpkg --print-architecture` resolves it). Adding your user to the `docker`
group lets you run compose without `sudo`.

---

## Part C -- DNS, deploy, verify

### 7. Point DNS at the VM
At [duckdns.org](https://www.duckdns.org): sign in, create a subdomain, and set
its IP to the VM's **reserved public IP**.

**Why:** Let's Encrypt validates ownership over this hostname (HTTP-01 challenge
on port 80). Because the IP is reserved, a one-time set is enough. As a learning
extra you can install the DuckDNS cron updater, but it is not required for a
static IP.

### 8. Deploy
```bash
git clone <your-repo-url> paint-war && cd paint-war
cp .env.production.example .env.production
# Edit .env.production: set SITE_ADDRESS to your DuckDNS host, set ALLOWED_ORIGINS
# to https://<that host>, and generate a strong DB password (openssl rand -base64 24)
# -- use the SAME password in POSTGRES_PASSWORD and DATABASE_URL.

docker compose -f compose.prod.yaml up -d --build
docker compose -f compose.prod.yaml logs -f caddy   # watch the cert get issued
```

**Why:** on first boot Caddy performs the ACME HTTP-01 challenge over ports
80/443 -- this is why both firewall layers (steps 3 and 5) must allow them. Do
**not** set `VITE_WS_URL` anywhere; leaving it unset is what enables the
origin-derived `wss://host/ws` connection.

### 9. Verify end-to-end
```bash
# TLS + proxy reach the Go server:
curl https://you.duckdns.org/healthz        # -> {"status":"ok"}
```
- Open `https://you.duckdns.org` -> game loads, no mixed-content warnings.
- DevTools -> Network: confirm a `wss://you.duckdns.org/ws` request upgrades
  (status 101) and stays open.
- Play a full match in two browser tabs. After it ends, confirm persistence:
  ```bash
  docker compose -f compose.prod.yaml exec database \
    psql -U paintwar -d paintwar -c 'select * from matches;'
  ```
  A row proves the optional DB path is live.
- Fault-tolerance check: `docker compose -f compose.prod.yaml restart server`
  mid-lobby and confirm the client's exponential-backoff reconnect rejoins.

---

## Operations cheatsheet

```bash
# Logs / status
docker compose -f compose.prod.yaml ps
docker compose -f compose.prod.yaml logs -f server

# Update to latest code
git pull && docker compose -f compose.prod.yaml up -d --build

# Stop / start
docker compose -f compose.prod.yaml down
docker compose -f compose.prod.yaml up -d

# Back up the database
docker compose -f compose.prod.yaml exec database \
  pg_dump -U paintwar paintwar > backup-$(date +%F).sql
```

## Gotchas
- Two firewall layers (Security List **and** OS iptables) -- both must open 80/443.
- Reserve the public IP before pointing DNS at it.
- Build the client **without** `VITE_WS_URL` so origin-derivation is used.
- Set `ALLOWED_ORIGINS=https://you.duckdns.org`; don't ship the `*` default.
- A1 capacity can be scarce; retry or switch availability domain if it fails.
