package mitras

// File topics.go contains all NATS subjects that are shared between services.

// InputChannels represents subject all messages will be published to.
// Messages received on this topic are Protobuf serialized Messages.
const InputChannels = "channel.>"
