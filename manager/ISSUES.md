# Manager

### Airbyte returns deleted connections and sources if the client specifically ask for them

They won't be listed in the List endpoints. However, if the client has the `sourceId` or `connectionId`, they can still retrieve the data of the deleted source or connection. For connections, it's straightforward to handle on our end because the `GetConnectionWithResponse` endpoint provides the connection status. Unfortunately, the `GetSourceWithResponse` endpoint does not return the status of the source.