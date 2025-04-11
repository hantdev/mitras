# Mitras Email Agent

Mitras Email Agent is used for sending emails. It wraps basic SMTP features and 
provides a simple API that Mitras services can use to send email notifications.

## Configuration

Mitras Email Agent is configured using the following configuration parameters:

| Parameter                           | Description                                                             |
| ----------------------------------- | ----------------------------------------------------------------------- |
| MITRAS_EMAIL_HOST                       | Mail server host                                                        |
| MITRAS_EMAIL_PORT                       | Mail server port                                                        |
| MITRAS_EMAIL_USERNAME                   | Mail server username                                                    |
| MITRAS_EMAIL_PASSWORD                   | Mail server password                                                    |
| MITRAS_EMAIL_FROM_ADDRESS               | Email "from" address                                                    |
| MITRAS_EMAIL_FROM_NAME                  | Email "from" name                                                       |
| MITRAS_EMAIL_TEMPLATE                   | Email template for sending notification emails                          |

There are two authentication methods supported: Basic Auth and CRAM-MD5.
If `MITRAS_EMAIL_USERNAME` is empty, no authentication will be used.
