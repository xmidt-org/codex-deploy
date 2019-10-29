# Deploying Codex

## Docker

In order to deploy into Docker, make sure [Docker is installed](https://docs.docker.com/install/).

1. Clone this repository.

2. Run `deploy/docker-compose/deploy.sh`
   
    This will build `fenrir` locally. It will then run `docker-compose up` which uses images of `gungnir` and `svalinn1` from dockerhub.

    To pull specific versions of the images, just set the `<SERVICE>_VERSION` environment variables before running the script.

    ```bash
    export SVALINN_VERSION=x.x.x
    export GUNGNIR_VERSION=x.x.x
    export FENRIR_VERSION=x.x.x
    ```
    If you don't want to set environment variables, set them inline when you run the script.

    ```
    SVALINN_VERSION=x.x.x deploy/docker-compose/deploy.sh
    ```

3. To bring the containers down:
    ```bash
    docker-compose -f deploy/docker-compose/docker-compose.yml down
    ```
