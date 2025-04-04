# Docker Composition

Configure environment variables and run Mitras Docker Composition.

\*Note\*\*: `docker-compose` uses `.env` file to set all environment variables. Ensure that you run the command from the same location as .env file.

## Installation

Follow the [official documentation](https://docs.docker.com/compose/install/).

## Usage

Run the following commands from the project root directory.

```bash
docker compose -f docker/docker-compose.yaml up
```

```bash
docker compose -f docker/addons/<path>/docker-compose.yaml up
```

To pull docker images from a specific release you need to change the value of `MITRAS_RELEASE_TAG` in `.env` before running these commands.
