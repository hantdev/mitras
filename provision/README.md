# Provision service

Provision service provides an HTTP API to interact with [Mitras][mitras].
Provision service is used to setup initial applications configuration i.e. clients, channels, connections and certificates that will be required for the specific use case especially useful for gateway provision.

For gateways to communicate with [Mitras][mitras] configuration is required (mqtt host, client, channels, certificates...). To get the configuration gateway will send a request to [Bootstrap][bootstrap] service providing `<external_id>` and `<external_key>` in request. To make a request to [Bootstrap][bootstrap] service you can use [Agent][agent] service on a gateway.

To create bootstrap configuration you can use [Bootstrap][bootstrap] or `Provision` service. [Mitras UI][mgxui] uses [Bootstrap][bootstrap] service for creating gateway configurations. `Provision` service should provide an easy way of provisioning your gateways i.e creating bootstrap configuration and as many clients and channels that your setup requires.

Also you may use provision service to create certificates for each client. Each service running on gateway may require more than one client and channel for communication. Let's say that you are using services [Agent][agent] and [Export][export] on a gateway you will need two channels for `Agent` (`data` and `control`) and one for `Export` and one client. Additionally if you enabled mtls each service will need its own client and certificate for access to [Mitras][mitras]. Your setup could require any number of clients and channels this kind of setup we can call `provision layout`.

Provision service provides a way of specifying this `provision layout` and creating a setup according to that layout by serving requests on `/mapping` endpoint. Provision layout is configured in [config.toml](configs/config.toml).
