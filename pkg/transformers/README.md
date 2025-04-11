# Message Transformers

A transformer service consumes events published by mitras adapters (such as MQTT and HTTP adapters) and transforms them to an arbitrary message format. A transformer can be imported as a standalone package and used for message transformation on the consumer side.

mitras [SenML transformer](transformer) is an example of Transformer service for SenML messages.

mitras [writers](writers) are using a standalone SenML transformer to preprocess messages before storing them.

[transformers]: https://github.com/hantdev/mitras/tree/main/transformers/senml
[writers]: https://github.com/hantdev/mitras/tree/main/writers
