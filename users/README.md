# Users service

Users service provides an HTTP API for managing users. Through this API clients
are able to do the following actions:

- register new accounts
- obtain access tokens
- verify access tokens

## Configuration

The service is configured using the environment variables presented in the
following table. Note that any unset variables will be replaced with their
default values.

| Variable                  | Description                                                             | Default        |
|---------------------------|-------------------------------------------------------------------------|----------------|
| ATHENA_USERS_LOG_LEVEL        | Log level for Users (debug, info, warn, error)                          | error          |
| ATHENA_USERS_DB_HOST          | Database host address                                                   | localhost      |
| ATHENA_USERS_DB_PORT          | Database host port                                                      | 5432           |
| ATHENA_USERS_DB_USER          | Database user                                                           | athena       |
| ATHENA_USERS_DB_PASSWORD      | Database password                                                       | athena       |
| ATHENA_USERS_DB               | Name of the database used by the service                                | users          |
| ATHENA_USERS_DB_SSL_MODE      | Database connection SSL mode (disable, require, verify-ca, verify-full) | disable        |
| ATHENA_USERS_DB_SSL_CERT      | Path to the PEM encoded certificate file                                |                |
| ATHENA_USERS_DB_SSL_KEY       | Path to the PEM encoded key file                                        |                |
| ATHENA_USERS_DB_SSL_ROOT_CERT | Path to the PEM encoded root certificate file                           |                |
| ATHENA_USERS_HTTP_PORT        | Users service HTTP port                                                 | 8180           |
| ATHENA_USERS_GRPC_PORT        | Users service gRPC port                                                 | 8181           |
| ATHENA_USERS_SERVER_CERT      | Path to server certificate in pem format                                |                |
| ATHENA_USERS_SERVER_KEY       | Path to server key in pem format                                        |                |
| ATHENA_USERS_SECRET           | String used for signing tokens                                          | users          |
| ATHENA_JAEGER_URL             | Jaeger server URL                                                       | localhost:6831 |
| ATHENA_EMAIL_DRIVER           | Mail server driver, mail server for sending reset password token        | smtp           |
| ATHENA_EMAIL_HOST             | Mail server host                                                        | localhost      |
| ATHENA_EMAIL_PORT             | Mail server port                                                        | 25             |
| ATHENA_EMAIL_USERNAME         | Mail server username                                                    |                |
| ATHENA_EMAIL_PASSWORD         | Mail server password                                                    |                |
| ATHENA_EMAIL_FROM_ADDRESS     | Email "from" address                                                    |                |
| ATHENA_EMAIL_FROM_NAME        | Email "from" name                                                       |                |
| ATHENA_EMAIL_TEMPLATE         | Email template for sending emails with password reset link              | email.tmpl     |
| ATHENA_TOKEN_SECRET           | Password reset token signing secret                                     |                |
| ATHENA_TOKEN_DURATION         | Token duration in minutes                                               | 5              |
| ATHENA_TOKEN_RESET_ENDPOINT   | Password request reset endpoint, for constructing link                  | /reset-request |

## Deployment

The service itself is distributed as Docker container. The following snippet
provides a compose file template that can be used to deploy the service container
locally:

```yaml
version: "2"
services:
  users:
    image: athena/users:[version]
    container_name: [instance name]
    ports:
      - [host machine port]:[configured HTTP port]
    environment:
      ATHENA_USERS_LOG_LEVEL: [Users log level]
      ATHENA_USERS_DB_HOST: [Database host address]
      ATHENA_USERS_DB_PORT: [Database host port]
      ATHENA_USERS_DB_USER: [Database user]
      ATHENA_USERS_DB_PASS: [Database password]
      ATHENA_USERS_DB: [Name of the database used by the service]
      ATHENA_USERS_DB_SSL_MODE: [SSL mode to connect to the database with]
      ATHENA_USERS_DB_SSL_CERT: [Path to the PEM encoded certificate file]
      ATHENA_USERS_DB_SSL_KEY: [Path to the PEM encoded key file]
      ATHENA_USERS_DB_SSL_ROOT_CERT: [Path to the PEM encoded root certificate file]
      ATHENA_USERS_HTTP_PORT: [Service HTTP port]
      ATHENA_USERS_GRPC_PORT: [Service gRPC port]
      ATHENA_USERS_SECRET: [String used for signing tokens]
      ATHENA_USERS_SERVER_CERT: [String path to server certificate in pem format]
      ATHENA_USERS_SERVER_KEY: [String path to server key in pem format]
      ATHENA_JAEGER_URL: [Jaeger server URL]
      ATHENA_EMAIL_DRIVER: [Mail server driver smtp]
      ATHENA_EMAIL_HOST: [ATHENA_EMAIL_HOST]
      ATHENA_EMAIL_PORT: [ATHENA_EMAIL_PORT]
      ATHENA_EMAIL_USERNAME: [ATHENA_EMAIL_USERNAME]
      ATHENA_EMAIL_PASSWORD: [ATHENA_EMAIL_PASSWORD]
      ATHENA_EMAIL_FROM_ADDRESS: [ATHENA_EMAIL_FROM_ADDRESS]
      ATHENA_EMAIL_FROM_NAME: [ATHENA_EMAIL_FROM_NAME]
      ATHENA_EMAIL_TEMPLATE: [ATHENA_EMAIL_TEMPLATE]
      ATHENA_TOKEN_SECRET: [ATHENA_TOKEN_SECRET]
      ATHENA_TOKEN_DURATION: [ATHENA_TOKEN_DURATION]
      ATHENA_TOKEN_RESET_ENDPOINT: [ATHENA_TOKEN_RESET_ENDPOINT]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/hantdev/athena.git

cd athena

# compile the service
make users

# copy binary to bin
make install

# set the environment variables and run the service
ATHENA_USERS_LOG_LEVEL=[Users log level] ATHENA_USERS_DB_HOST=[Database host address] ATHENA_USERS_DB_PORT=[Database host port] ATHENA_USERS_DB_USER=[Database user] ATHENA_USERS_DB_PASS=[Database password] ATHENA_USERS_DB=[Name of the database used by the service] ATHENA_USERS_DB_SSL_MODE=[SSL mode to connect to the database with] ATHENA_USERS_DB_SSL_CERT=[Path to the PEM encoded certificate file] ATHENA_USERS_DB_SSL_KEY=[Path to the PEM encoded key file] ATHENA_USERS_DB_SSL_ROOT_CERT=[Path to the PEM encoded root certificate file] ATHENA_USERS_HTTP_PORT=[Service HTTP port] ATHENA_USERS_GRPC_PORT=[Service gRPC port] ATHENA_USERS_SECRET=[String used for signing tokens] ATHENA_USERS_SERVER_CERT=[Path to server certificate] ATHENA_USERS_SERVER_KEY=[Path to server key] ATHENA_JAEGER_URL=[Jaeger server URL] ATHENA_EMAIL_DRIVER=[Mail server driver smtp] ATHENA_EMAIL_HOST=[Mail server host] ATHENA_EMAIL_PORT=[Mail server port] ATHENA_EMAIL_USERNAME=[Mail server username] ATHENA_EMAIL_PASSWORD=[Mail server password] ATHENA_EMAIL_FROM_ADDRESS=[Email from address] ATHENA_EMAIL_FROM_NAME=[Email from name] ATHENA_EMAIL_TEMPLATE=[Email template file] ATHENA_TOKEN_SECRET=[Password reset token signing secret] ATHENA_TOKEN_DURATION=[Password reset token duration] ATHENA_TOKEN_RESET_ENDPOINT=[Password reset token endpoint] $GOBIN/athena-users
```

If `ATHENA_EMAIL_TEMPLATE` doesn't point to any file service will function but password reset functionality will not work.

## Usage

For more information about service capabilities and its usage, please check out
the [API documentation](swagger.yaml).
