# ARTEX-Superversion on Parrot Security OS

This guide covers a local, single-machine ARTEX-Superversion deployment on **Parrot Security OS**. Parrot is Debian-based, so the host preparation uses `apt`. The documented container path uses Docker Engine with Docker Compose v2.

> **Status:** These instructions describe the repository's Compose deployment. They are not a claim that every Parrot release, desktop environment, VM, or third-party Compose provider has been tested. Podman notes are included separately.

## 1. Prepare Parrot

Update package metadata and install the basic tools:

```bash
sudo apt update
sudo apt install -y git ca-certificates curl openssl
```

Install Docker Engine and the **Compose v2** plugin using the official [Docker instructions for Debian](https://docs.docker.com/engine/install/debian/). Parrot tracks Debian, but check that the Docker instructions support the Debian base version used by your Parrot release.

Verify the tools are available:

```bash
docker --version
docker compose version
```

If Docker requires elevated privileges on your installation, use `sudo docker compose ...` for the commands below. Do not change socket permissions broadly to work around access errors.

## 2. Get ARTEX-Superversion

Clone this repository (not the upstream ARTEX repository):

```bash
git clone https://github.com/lockmeilluminati/ARTEX-Superversion.git
cd ARTEX-Superversion
```

## 3. Configure the local environment

Create a private environment file:

```bash
cp .env.example .env
```

Edit `.env` and replace the example `POSTGRES_PASSWORD` with a unique password. You can generate a strong random value with:

```bash
openssl rand -hex 32
```

Paste the generated value after `POSTGRES_PASSWORD=` in `.env`. Keep `.env` private: it contains credentials and must not be committed or uploaded. The optional LLM API keys can be configured now or later in the ARTEX interface.

The default bind address is `127.0.0.1`, which keeps the web interface available only from the Parrot machine. Keep this setting unless you intentionally need network access and understand the exposure.

## 4. Build and start

From the repository root:

```bash
docker compose up -d --build
docker compose ps
```

The application waits for PostgreSQL's health check before starting. Inspect startup output if a service is not healthy:

```bash
docker compose logs --tail=100 postgres artex
```

Open [http://127.0.0.1:8787](http://127.0.0.1:8787) in the Parrot browser. Follow the first-run setup flow; the one-time setup token is printed in the ARTEX service logs:

```bash
docker compose logs artex | grep "Setup token"
```

## 5. Stop and restart safely

Stop the services without deleting the database volume:

```bash
docker compose down
```

Start them again:

```bash
docker compose up -d
```

**Do not add `-v` to `docker compose down` unless you deliberately intend to remove the PostgreSQL volume and its data.** The project also keeps runtime data in `data/` and persistent application keys in `keys/`; preserve these when backing up or moving the installation.

## 6. Podman users

The Compose file follows the Compose specification, but Podman Compose support depends on the external Compose provider installed on the host. Check whether a provider is available:

```bash
podman --version
podman compose version
```

If `podman compose version` succeeds, you can try the equivalent commands:

```bash
podman compose up -d --build
podman compose ps
podman compose logs --tail=100 postgres artex
```

Provider behavior, networking, health checks, and volume handling can differ. Docker Compose v2 remains the documented baseline; treat Podman deployment as provider-dependent until verified on your Parrot release. Never remove a database volume as a troubleshooting shortcut.

## 7. Yakit integration status

ARTEX-Superversion's direction is to connect external security tools through a documented, permission-aware tool layer. Yakit is represented in the project catalogue as an integration candidate. A catalogue entry does **not** mean that every Yakit workflow is already connected or safe to execute automatically. Check the current project documentation and tests for the implemented capabilities, and keep discovery separate from approval to install or run a tool.

## 8. Security and operational notes

- Use ARTEX only against systems you own or are explicitly authorized to assess.
- Keep the web interface bound to localhost unless remote access is intentionally configured and protected.
- Never commit `.env`, `config.json`, `data/`, `keys/`, database files, generated binaries, or local backups.
- Back up the PostgreSQL database together with the persistent `keys/` material needed to decrypt stored secrets. See the main README's **Operations → Backups** section.
- Before upgrades, back up the database, `data/`, and `keys/`; do not delete the named database volume to resolve a startup error.

For architecture, development, and the full feature set, continue with the [main README](../README.md).
