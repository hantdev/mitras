# Writers

Writers provide an implementation of various `message writers`.
Message writers are services that normalize (in `SenML` format)
mitras messages and store them in specific data store.

Writers are optional services and are treated as plugins. In order to
run writer services, core services must be up and running.
