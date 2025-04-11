# BOOTSTRAP SERVICE

New devices need to be configured properly and connected to the Sina. Bootstrap service is used in order to accomplish that. This service provides the following features:

1. Creating new Sina Clients
2. Providing basic configuration for the newly created Clients
3. Enabling/disabling Clients

Pre-provisioning a new Client is as simple as sending Configuration data to the Bootstrap service. Once the Client is online, it sends a request for initial config to Bootstrap service. Bootstrap service provides an API for enabling and disabling Clients. Only enabled Clients can exchange messages over Sina. Bootstrapping does not implicitly enable Clients, it has to be done manually.

In order to bootstrap successfully, the Client needs to send bootstrapping request to the specific URL, as well as a secret key. This key and URL are pre-provisioned during the manufacturing process. If the Client is provisioned on the Bootstrap service side, the corresponding configuration will be sent as a response. Otherwise, the Client will be saved so that it can be provisioned later.

## Client Configuration Entity

Client Configuration consists of two logical parts: the custom configuration that can be interpreted by the Client itself and Sina-related configuration. Sina config contains:

1. corresponding Sina Client ID
2. corresponding Sina Client key
3. list of the Sina channels the Client is connected to

> Note: list of channels contains IDs of the Sina channels. These channels are _pre-provisioned_ on the Sina side and, unlike corresponding Sina Client, Bootstrap service is not able to create Sina Channels.

Enabling and disabling Client (adding Client to/from whitelist) is as simple as connecting corresponding Sina Client to the given list of Channels. Configuration keeps _state_ of the Client:

| State    | What it means                                  |
| -------- | ---------------------------------------------- |
| Inactive | Client is created, but isn't enabled           |
| Active   | Client is able to communicate using Sina |

Switching between states `Active` and `Inactive` enables and disables Client, respectively.

Client configuration also contains the so-called `external ID` and `external key`. An external ID is a unique identifier of corresponding Client. For example, a device MAC address is a good choice for external ID. External key is a secret key that is used for authentication during the bootstrapping procedure.
