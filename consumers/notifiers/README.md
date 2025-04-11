# Notifiers service

Notifiers service provides a service for sending notifications using Notifiers.
Notifiers service can be configured to use different types of Notifiers to send
different types of notifications such as SMS messages, emails, or push notifications.
Service is extensible so that new implementations of Notifiers can be easily added.
Notifiers **are not standalone services** but rather dependencies used by Notifiers service
for sending notifications over specific protocols.

## Configuration

The service is configured using the environment variables.
The environment variables needed for service configuration depend on the underlying Notifier.
