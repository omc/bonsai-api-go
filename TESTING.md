# Testing

## Test Suites

This project makes use of the `_test` idiom as mentioned in 
[pkg.go.dev/testing](https://pkg.go.dev/testing@master). 

All external API testing of the Bonsai API Go Client is done via unit tests
and integration tests in the `bonsai_test` package.

Certain internal API functions and data structures are tested within the
`bonsai` package.

For all suites, we use the [testify toolkit](https://github.com/stretchr/testify).

### Running tests

Run all tests:

```shell
go test ./...
```

#### Re-recording Integration test HTTP interactions

By default, all integration tests will run in `REPLAY_ONLY` mode, and
don't require any API credentials in order to test against the existing
recorded Bonsai API HTTP responses and marshaled results in 
[bonsai/fixtures](bonsai/fixtures).

To re-record these interactions, set these environment variables:

- `BONSAI_REC_MODE`: Determines the recording strategy.
  - One of: 
    `REC_ONCE`, `REC_ONLY`, `REPLAY_ONLY` (default), `REPLAY_WITH_NEW`, or `PASS_THROUGH`.
  - See [go-vcr's godoc](https://pkg.go.dev/gopkg.in/dnaeon/go-vcr.v3/recorder#Mode)
  for more details on recording strategies.
- `BONSAI_API_KEY`: Must be present
- `BONSAI_API_TOKEN`: Must be present

### Adding tests

Ensure that new tests are associated with a Testify `suite`; there are existing
suites for both internal and external API testing which should fit most use cases.

#### Internal API

- Where possible, name test files with the `$filename_impl_test.go` convention.
- The `ClientImplTestSuite` should satisfy all internal testing needs.
- Example @ [bonsai/client_impl_test.go](bonsai/client_impl_test.go)


#### External API

- External API testing should be the primary location for unit and integration tests.
- Name files with the standard `$filename_test.go` convention.

##### Unit Tests

- Unit tests shouldn't interact with any external API endpoints, 
and are expected to mock all HTTP responses if needed. 
- The `ClientMockTestSuite` should be preferred for all unit tests in 
the external API.
- Example @ [bonsai/cluster_test.go](bonsai/cluster_test.go)

##### Integration Tests

- Integration tests rely on the [go-vcr](https://github.com/dnaeon/go-vcr)
package to enable recording and replaying of all HTTP interactions in the
test suite.
- The `ClientVCRTestSuite` should be preferred for all integration tests - 
that is, any external HTTP interaction - in the external API.
- Example @ [bonsai/cluster_test.go](bonsai/cluster_test.go)
