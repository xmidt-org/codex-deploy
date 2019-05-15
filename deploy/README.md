# Deploying Codex

## Docker

In order to deploy into Docker, make sure [Docker is installed](https://docs.docker.com/install/).

1. Have the services you want to bring up built (Svalinn, Gungnir, and Fenrir).

2. Set an environment variables relevant for the services you are deploying. If
   you aren't using locally built images, replace `local` with the correct tag:
   ```bash
   export SVALINN_VERSION=local
   export GUNGNIR_VERSION=local
   export FENRIR_VERSION=local
   ```
   If you don't want to set environment variables, set them inline with each 
   `docker-compose` command below.

3. To bring the containers up run:
   ```bash
   docker-compose up -d
   ```
   If you only want to bring up, for example, the database and Svalinn, run:
   ```bash
   docker-compose up -d db db-init svalinn
   ```
   This can be done with any combination of services and the database.

4. To bring the containers down:
   ```bash
   docker-compose down
   ```