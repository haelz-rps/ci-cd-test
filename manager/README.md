# Manager

## Running the project

1. To spin up all the components that manager relies on run:
```bash
docker-compose -f docker-compose.yaml -f docker-compose.airbyte.yaml up
```
2. Now you can run
```bash
go run ./cmd/rest-api
```
3. Run the test
```bash
go test ./...
go test ./... -tags e2e
```
## Contribution Guidelines

### Adding or changing enpoints

We have an OpenAPI centric process for development, which means that any change or addition to the system must first go through the OpenAPI contract located [here](./docs/open-api-schema.yaml).

Once you have reflected your change to the OpenAPI schema, submit a draft PR for review. Once the team agrees on the schema, follow on with the implementation.

### Implementation

We have adopted the [Gokit](https://gokit.io/) pattern as a standard, and we use [Domain Driven Design](https://en.wikipedia.org/wiki/Domain-driven_design) as our designing pattern. Make sure that your contributions respect them.

### Testing your code

We have adopted the strategy to test all endpoints end to end, and we use our OpenAPI schema as the source of truth for our tests. All rest-api related tests should live inside `./test/e2e/rest_api_test.go`. 

Our tests use [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen/) to generate an http client that lives in `./test/e2e/rest-api-client/`, then on our tests we use it to trigger requests to a live rest-api instance. You can configure the rest-api host through the environment variable `REST_API_HOST`. To generate the client run `go generate ./...`

We verify the API responses using a JSON schema validator that was implemented under `./test/e2e/schema-validator/`. This package takes the schemas defined under the OpenAPI definition and validates if the response given by the rest-api respects the schema.

### OpenAPI contributions

As it has been described on the Testing your code section, our testing strategy heavily relies in the OpenAPI definition, there fore this document must be maintained carefully and respecting some rules.

1. The Input and Output schemas must be defined under `components/schemas` section of OpenAPI. This is the place our schema validator package grabs the schemas to validate server responses.

2. The Schemas must be strict, meaning that each property should have precise types and in the case of strings strict formats should be defined. Furthermore the required properties must be defined in the schema, as otherwise the schema validation will be too loose and will be error prone.

Here is an example of a well defined schema:
```yaml
Source:
  type: object
  properties:
    sourceDefinitionId:
      type: string
      format: uuid
    name:
      type: string
    dockerRepository:
      type: string
    dockerImageTag:
      type: string
    documentationUrl:
      type: string
    icon:
      type: string
  required:
    - sourceDefinitionId
    - name
    - icon
```

The property `sourceDefinitionId` is a string with a proper format set, all required fields are marked under the required field.

## Go commands

```bash
go run cmd/rest-api/main.go
go fmt ./...
go test ./...
go test ./test/e2e/
go generate ./...
```

## Airbyte Integration

To integrate with Airbyte we utilized a code generated tool called [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen). This will auto generate the client so we don't need to worry testing the http layer.

To use this tool we shall interact with the open api schema that we maintain inside the repository folder.

Once changes have been done you can generate the new code by running:
```
go generate ./repositories/data-extractor
```
## Entity Mapping

To make our entities to be more idiomatic with the current Ubix domain we named our api entities differently from Airbyte. Here how our entities map to Airbyte entities:

| Dripper          | Airbyte          |
| ---------------- | ---------------- |
| SourceDefinition | SourceDefinition |
| Source           | Source           |
| Connector        | Connection       |
| Table            | Stream           |

Also there are some Airbyte entities that we abstracted away because our use case is a bit simpler then Airbyte:
- Destination: We assume a single destination which is a kafka broker deployed for each Ubix tenant.
- Workspace: We assume a single workspace that comes by default created in each Airbyte tenant.

They should be considered in a future where we make Dripper multi tenant 

## References

- [Airbyte server APIs](https://github.com/airbytehq/airbyte-platform/tree/ed7bf3e27317ba5c7cba077653c8956d11529631/airbyte-server/src/main/java/io/airbyte/server/apis)

TODO:
- switch get table to get connector
- switch update table to update connector
