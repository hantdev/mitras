# Consumers

Consumers provide an abstraction of various `Mitras consumers`.
Mitras consumer is a generic service that can handle received messages - consume them.
The message is not necessarily a Mitras message - before consuming, Mitras message can
be transformed into any valid format that specific consumer can understand. For example,
writers are consumers that can take a SenML or JSON message and store it.

Consumers are optional services and are treated as plugins. In order to
run consumer services, core services must be up and running.
