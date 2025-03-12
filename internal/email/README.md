# Athena Email Agent

Athena Email Agent is used for sending emails. It wraps basic SMTP features and 
provides a simple API that Athena services can use to send email notifications.

## Configuration

Athena Email Agent is configured using the following configuration parameters:

| Parameter                           | Description                                                             |
| ----------------------------------- | ----------------------------------------------------------------------- |
| ATHENA_EMAIL_HOST                       | Mail server host                                                        |
| ATHENA_EMAIL_PORT                       | Mail server port                                                        |
| ATHENA_EMAIL_USERNAME                   | Mail server username                                                    |
| ATHENA_EMAIL_PASSWORD                   | Mail server password                                                    |
| ATHENA_EMAIL_FROM_ADDRESS               | Email "from" address                                                    |
| ATHENA_EMAIL_FROM_NAME                  | Email "from" name                                                       |
| ATHENA_EMAIL_TEMPLATE                   | Email template for sending notification emails                          |

There are two authentication methods supported: Basic Auth and CRAM-MD5.
If `ATHENA_EMAIL_USERNAME` is empty, no authentication will be used.