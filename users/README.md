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
| VIOT_USERS_LOG_LEVEL        | Log level for Users (debug, info, warn, error)                          | error          |
| VIOT_USERS_DB_HOST          | Database host address                                                   | localhost      |
| VIOT_USERS_DB_PORT          | Database host port                                                      | 5432           |
| VIOT_USERS_DB_USER          | Database user                                                           | viot       |
| VIOT_USERS_DB_PASSWORD      | Database password                                                       | viot       |
| VIOT_USERS_DB               | Name of the database used by the service                                | users          |
| VIOT_USERS_DB_SSL_MODE      | Database connection SSL mode (disable, require, verify-ca, verify-full) | disable        |
| VIOT_USERS_DB_SSL_CERT      | Path to the PEM encoded certificate file                                |                |
| VIOT_USERS_DB_SSL_KEY       | Path to the PEM encoded key file                                        |                |
| VIOT_USERS_DB_SSL_ROOT_CERT | Path to the PEM encoded root certificate file                           |                |
| VIOT_USERS_HTTP_PORT        | Users service HTTP port                                                 | 8180           |
| VIOT_USERS_GRPC_PORT        | Users service gRPC port                                                 | 8181           |
| VIOT_USERS_SERVER_CERT      | Path to server certificate in pem format                                |                |
| VIOT_USERS_SERVER_KEY       | Path to server key in pem format                                        |                |
| VIOT_USERS_SECRET           | String used for signing tokens                                          | users          |
| VIOT_JAEGER_URL             | Jaeger server URL                                                       | localhost:6831 |
| VIOT_EMAIL_DRIVER           | Mail server driver, mail server for sending reset password token        | smtp           |
| VIOT_EMAIL_HOST             | Mail server host                                                        | localhost      |
| VIOT_EMAIL_PORT             | Mail server port                                                        | 25             |
| VIOT_EMAIL_USERNAME         | Mail server username                                                    |                |
| VIOT_EMAIL_PASSWORD         | Mail server password                                                    |                |
| VIOT_EMAIL_FROM_ADDRESS     | Email "from" address                                                    |                |
| VIOT_EMAIL_FROM_NAME        | Email "from" name                                                       |                |
| VIOT_EMAIL_TEMPLATE         | Email template for sending emails with password reset link              | email.tmpl     |
| VIOT_TOKEN_SECRET           | Password reset token signing secret                                     |                |
| VIOT_TOKEN_DURATION         | Token duration in minutes                                               | 5              |
| VIOT_TOKEN_RESET_ENDPOINT   | Password request reset endpoint, for constructing link                  | /reset-request |

## Deployment

The service itself is distributed as Docker container. The following snippet
provides a compose file template that can be used to deploy the service container
locally:

```yaml
version: "2"
services:
  users:
    image: viot/users:[version]
    container_name: [instance name]
    ports:
      - [host machine port]:[configured HTTP port]
    environment:
      VIOT_USERS_LOG_LEVEL: [Users log level]
      VIOT_USERS_DB_HOST: [Database host address]
      VIOT_USERS_DB_PORT: [Database host port]
      VIOT_USERS_DB_USER: [Database user]
      VIOT_USERS_DB_PASS: [Database password]
      VIOT_USERS_DB: [Name of the database used by the service]
      VIOT_USERS_DB_SSL_MODE: [SSL mode to connect to the database with]
      VIOT_USERS_DB_SSL_CERT: [Path to the PEM encoded certificate file]
      VIOT_USERS_DB_SSL_KEY: [Path to the PEM encoded key file]
      VIOT_USERS_DB_SSL_ROOT_CERT: [Path to the PEM encoded root certificate file]
      VIOT_USERS_HTTP_PORT: [Service HTTP port]
      VIOT_USERS_GRPC_PORT: [Service gRPC port]
      VIOT_USERS_SECRET: [String used for signing tokens]
      VIOT_USERS_SERVER_CERT: [String path to server certificate in pem format]
      VIOT_USERS_SERVER_KEY: [String path to server key in pem format]
      VIOT_JAEGER_URL: [Jaeger server URL]
      VIOT_EMAIL_DRIVER: [Mail server driver smtp]
      VIOT_EMAIL_HOST: [MF_EMAIL_HOST]
      VIOT_EMAIL_PORT: [MF_EMAIL_PORT]
      VIOT_EMAIL_USERNAME: [MF_EMAIL_USERNAME]
      VIOT_EMAIL_PASSWORD: [MF_EMAIL_PASSWORD]
      VIOT_EMAIL_FROM_ADDRESS: [MF_EMAIL_FROM_ADDRESS]
      VIOT_EMAIL_FROM_NAME: [MF_EMAIL_FROM_NAME]
      VIOT_EMAIL_TEMPLATE: [MF_EMAIL_TEMPLATE]
      VIOT_TOKEN_SECRET: [MF_TOKEN_SECRET]
      VIOT_TOKEN_DURATION: [MF_TOKEN_DURATION]
      VIOT_TOKEN_RESET_ENDPOINT: [MF_TOKEN_RESET_ENDPOINT]
```

To start the service outside of the container, execute the following shell script:

```bash
# download the latest version of the service
git clone https://github.com/hantdev/viot.git

cd viot

# compile the service
make users

# copy binary to bin
make install

# set the environment variables and run the service
VIOT_USERS_LOG_LEVEL=[Users log level] VIOT_USERS_DB_HOST=[Database host address] VIOT_USERS_DB_PORT=[Database host port] VIOT_USERS_DB_USER=[Database user] VIOT_USERS_DB_PASS=[Database password] VIOT_USERS_DB=[Name of the database used by the service] VIOT_USERS_DB_SSL_MODE=[SSL mode to connect to the database with] VIOT_USERS_DB_SSL_CERT=[Path to the PEM encoded certificate file] VIOT_USERS_DB_SSL_KEY=[Path to the PEM encoded key file] VIOT_USERS_DB_SSL_ROOT_CERT=[Path to the PEM encoded root certificate file] VIOT_USERS_HTTP_PORT=[Service HTTP port] VIOT_USERS_GRPC_PORT=[Service gRPC port] VIOT_USERS_SECRET=[String used for signing tokens] VIOT_USERS_SERVER_CERT=[Path to server certificate] VIOT_USERS_SERVER_KEY=[Path to server key] VIOT_JAEGER_URL=[Jaeger server URL] VIOT_EMAIL_DRIVER=[Mail server driver smtp] VIOT_EMAIL_HOST=[Mail server host] VIOT_EMAIL_PORT=[Mail server port] VIOT_EMAIL_USERNAME=[Mail server username] VIOT_EMAIL_PASSWORD=[Mail server password] VIOT_EMAIL_FROM_ADDRESS=[Email from address] VIOT_EMAIL_FROM_NAME=[Email from name] VIOT_EMAIL_TEMPLATE=[Email template file] VIOT_TOKEN_SECRET=[Password reset token signing secret] VIOT_TOKEN_DURATION=[Password reset token duration] VIOT_TOKEN_RESET_ENDPOINT=[Password reset token endpoint] $GOBIN/viot-users
```

If `VIOT_EMAIL_TEMPLATE` doesn't point to any file service will function but password reset functionality will not work.

## Usage

For more information about service capabilities and its usage, please check out
the [API documentation](swagger.yaml).
