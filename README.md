[![Go Actions Status](https://github.com/apigee/registry/workflows/Go/badge.svg)](https://github.com/apigee/registry/actions)

# Registry API Reference Implementation

This repository contains a reference implementation of the Registry API.

## The Registry API

The Registry API allows teams to upload and share machine-readable descriptions
of APIs that are in use and in development. These descriptions include API
specifications in standard formats like [OpenAPI](https://www.openapis.org/),
the
[Google API Discovery Service Format](https://developers.google.com/discovery),
and the
[Protocol Buffers Language](https://developers.google.com/protocol-buffers).
These API specifications can be used by tools like linters, browsers,
documentation generators, test runners, proxies, and API client and server
generators. The Registry API itself can be seen as a machine-readable
enterprise API catalog designed to back online directories, portals, and
workflow managers.

The Registry API is formally described by the Protocol Buffer source files in
[google/cloud/apigee/registry/v1](google/cloud/apigee/registry/v1). It closely
follows the Google API Design Guidelines at [aip.dev](https://aip.dev) and
presents a developer experience consistent with production Google APIs. Please
tell us about your experience if you use it.

## This Implementation

This reference implementation is a [gRPC](https://grpc.io) service written in
Go. It can be run locally or deployed in a container using services including
[Google Cloud Run](https://cloud.google.com/run). It stores data using a
configurable relational interface layer that currently supports
[PostgreSQL](https://www.postgresql.org/) and [SQLite](https://www.sqlite.org/)
(see [config](config) for details).

The Registry API service is annotated to support
[gRPC HTTP/JSON transcoding](https://aip.dev/127), which allows it to be
automatically published as a JSON REST API using a proxy. Proxies also enable
[gRPC web](https://github.com/grpc/grpc-web), which allows gRPC calls to be
directly made from browser-based applications. A configuration for the
[Envoy](https://www.envoyproxy.io/) proxy is included
([deployments/envoy/envoy.yaml](deployments/envoy/envoy.yaml)) along with
scripts to build and deploy the Registry API server and a proxy in a single
container on Google Cloud Run.

The Registry API protos also include configuration to support
[generated API clients (GAPICS)](https://googleapis.github.io/gapic-generators/),
which allow idiomatic API usage from a variety of languages. A Go GAPIC library
is generated as part of the build process using
[gapic-generator-go](https://github.com/googleapis/gapic-generator-go). A
sample Go GAPIC-based client is in
[examples/gapic-client](examples/gapic-client).

Two command-line interfaces are included:

- [cmd/apg](cmd/apg), automatically generated from the API description using
  the
  [protoc-gen-go_cli](https://github.com/googleapis/gapic-generator-go/tree/master/cmd/protoc-gen-go_cli)
  tool in
  [gapic-generator-go](https://github.com/googleapis/gapic-generator-go).
- [cmd/registry](cmd/registry), a hand-written command-line tool that uses the
  Go GAPIC library to support additional API management tasks.

The entry point for the Registry API server itself is
[cmd/registry-server](cmd/registry-server).

## Build Instructions

The following tools are needed to build this software:

- Go 1.15 or later.
- protoc, the Protocol Buffer Compiler, version 3.10 or later.
- make, git, and other elements of common unix build environments.

This repository contains a Makefile that downloads all other dependencies and
builds this software (`make all`). With dependencies downloaded, subsequent
builds can be made with `go install ./...` or `make lite`. The Makefile also
includes targets that build and deploy the API on
[Google Cloud Run](https://cloud.google.com/run) (see below).

## Generated Components

Several directories of generated code are produced by the build process (see
the [COMPILE-PROTOS.sh](tools/COMPILE-PROTOS.sh) script for details). These
include:

- [rpc](rpc), containing generated Go Protocol Buffer support code.
- [gapic](gapic), containing the Go GAPIC (generated API client) library.
- [cmd/apg](cmd/apg), containing a generated command-line interface.

## Quickstart

The easiest way to try the Registry API is to run `registry-server` locally
using the SQLite backend.

`registry-server -c config/sqlite.yaml`

Next, in a separate terminal, configure your environment to point to this
server with the following:

`source auth/LOCAL.sh`

Now you can check your server and configuration with the
automatically-generated `apg` client:

`apg registry get-status`

Next run a suite of tests with `make test` and see a corresponding walkthrough
of API features in [tests/demo/walkthrough.sh](tests/demo/walkthrough.sh). For
more demonstrations, see the [demos](demos) directory.

## Running the Registry API server locally

### Running the Registry API server

Running `source auth/LOCAL.sh` will configure your environment to run the
Registry API server locally and for the included clients to call your local
instance. Start the server by running `registry-server`. By default a SQLite
backend will be used (this is equivalent to running
`registry-server -c config/sqlite.yaml`).

### Optional: Running with a PostgreSQL backend

The `config` directory contains examples of files that can be used to configure
the `registry-server`. [config/postgres.yaml](config/postgres.yaml) contains a
sample configuration; you'll likely need to customize this for your own
`postgresql` instance. Use it with:

`registry-server -c config/postgres.yaml`

### Optional: Running with a PostgreSQL backend on Google CloudSQL

[config/cloudsql-postgres.yaml](config/cloudsql-postgres.yaml) contains the
configuration to connect to a PostgreSQL database hosted on CloudSQL. If you
don't have an existing PostgreSQL instance, you can follow
[these instructions](https://cloud.google.com/sql/docs/postgres/quickstart) to
setup one. Please make sure to update
[config/cloudsql-postgres.yaml](config/cloudsql-postgres.yaml) with the correct
host configuration. As previously noted, you can start the server with the
following:

`registry-server -c config/cloudsql-postgres.yaml`

### Optional: Proxying a local service with Envoy

`registry-server` provides a gRPC service only. For a transcoded HTTP/JSON
interface, run the [Envoy](https://www.envoyproxy.io) proxy locally using the
configuration in the [deployments/envoy](deployments/envoy) directory. With a
local installation of Envoy, this can be done by running the following inside
the [deployments/envoy](deployments/envoy) directory.

`envoy -c envoy.yaml`

### Optional: Local authorization with authz-server

The included Envoy configuration uses Envoy's
[ext_authz_filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/ext_authz_filter)
to validate requests using a simple authorization server in
[cmd/authz-server](cmd/authz-server). You can start this server with the
following:

`authz-server -c cmd/authz-server/authz.yaml`

### Optional: Running the registry-graphql proxy

[cmd/registry-graphql](cmd/registry-graphql) contains a simple proxy that
provides a read-only GraphQL interface to the Registry API. It can be run with
a local or remote `registry-server`.

## Running the Registry API server in a container

The `containers` directory contains Dockerfiles and other configurations to
allow `registry-server` to be run in containers. Containers can be built that
run `registry-server` standalone (recommended) or in a bundled container that
includes `envoy` and a simple authorization server (mainly for running secured
instances on Cloud Run). x64 and arm64 platforms are currently supported.

To build a container that runs `registry-server` standalone, use the following:

```
docker build -f containers/registry-server/Dockerfile -t registry-server .
```

To run the image with docker, you'll need to expose the default port (8080)
that the server uses in the container. Your `docker run` invocation will look
like this:

```
docker run -p 8080:8080 registry-server:latest
```

Since this is using the default configuration, you'll get an error message
similar to this:

```
Failed to start: sqlite3 is unavailable, please recompile with CGO_ENABLED=1 or configure registry-server to use a different database
```

This is because container builds exclude `CGO`, which is required by the
default database (sqlite3). To resolve this, you could rebuild your container
with a modified `registry.yaml` (this is the default configuration used by the
build) or, more simply, specify a configuration using the environment variables
referenced in [config/registry.yaml](config/registry.yaml). Following those,
your `docker run` invocation might look like this:

```
docker run \
  -p 8080:8080 \
  -e REGISTRY_DATABASE=postgres \
  -e REGISTRY_DBCONFIG="host=${PGHOST} port=5432 user=registry dbname=registry password=iloveapis sslmode=disable" \
  registry-server:latest
```

Be sure to replace `${PGHOST}` with the address of your Postgres server (either
directly or by setting `PGHOST` with another `-e` argument to `docker run`),
check all the other REGISTRY_DBCONFIG parameters, and verify that your server
is configured to accept remote connections (in `postgres.conf` and
`pg_hba.conf`).

## Running the Registry API server with Google Cloud Run

The Registry API server can be deployed on
[Google Cloud Run](https://cloud.google.com/run). To support this, the
[Makefile](Makefile) contains targets that build a Docker image and that deploy
it to Google Cloud Run. Both use the `gcloud` command, which should be
authenticated and configured for the project where the services should be run.

Requirements:

- Both targets require the [gcloud](https://cloud.google.com/sdk/gcloud)
  command, which is part of the
  [Google Cloud SDK](https://cloud.google.com/sdk).

- If not already done, `gcloud auth login` gets user credentials for subsequent
  `gcloud` operations and `gcloud config set project PROJECT_ID` can be used to
  set your project ID to the one where you plan to host your servce.

- The Makefile gets your project ID from the `REGISTRY_PROJECT_IDENTIFIER`
  environment variable. This can be set automatically by running
  `source auth/CLOUDRUN.sh`.

`make build` uses [Google Cloud Build](https://cloud.google.com/cloud-build) to
build a container containing the API server. The container is stored in
[Google Container Registry](https://cloud.google.com/container-registry). This
uses the `Dockerfile` at the top level of the repo, which is a link to
[containers/registry-server/Dockerfile](containers/registry-server/Dockerfile).
A second Dockerfile is available in
[containers/registry-bundle/Dockerfile](containers/registry-bundle/Dockerfile);
this contains the `registry-server`, `envoy`, and a simple authorization server
(`authz-server`). To use it, just change the top level link to point to this
Dockerfile.

`make deploy` deploys the built container on
[Google Cloud Run](https://cloud.google.com/run).

When deploying to Cloud Run for the first time, you will be asked a few
questions, including this one:

`Allow unauthenticated invocations to [registry-backend] (y/N)?`

If you answer "y", you will be able to make calls without authentication. This
is the easiest way to test the API, but it's not necessary - running
`source auth/CLOUDRUN.sh` configures your environment so that the Registry CLI
and other tools will authenticate using your user ID.

Important note: If you answer "N" to the above question, Cloud Run will require
an auth token for all requests to the server. `source auth/CLOUDRUN.sh` adds
this token to your environment, but there two possible pitfalls:

1. CORS requests will fail if your backend requires authentication
   ([details](https://groups.google.com/g/gce-discussion/c/WQUxKhZORjo)).
2. Cloud Run removes signatures from accepted JWT tokens, replacing them with
   "SIGNATURE_REMOVED_BY_GOOGLE"
   ([details](https://cloud.google.com/run/docs/troubleshooting#signature-removed)).
   If your deployment includes the Envoy proxy and
   [authz-server](cmd/authz-server), then the authz-server configuration will
   need to be updated to trust the JWT tokens that are passed through, since
   they've already been verified and further checking is impossible. You can do
   that by setting `trustJWTs: true` in
   [authz.yaml](cmd/authz-server/authz.yaml).

If you initially answer "N" and change your mind, you can enable
unauthenticated calls by going to the Permissions view in the Cloud Run console
and adding the "Cloud Run Invoker" role to the special username "allUsers".
(Changes take a few seconds to propagate.)

Now you can call the API with your generated CLI.

`apg registry get-status`

You can also verify your installation by running `make test`. This will run
tests against the same service that your CLI is configured to use via the
environment variables set by the `auth/*.sh` scripts.

Auth tokens are short-lived. When your token expires, your calls will return a
message like this:
`rpc error: code = Unauthenticated desc = Unauthorized: HTTP status code 401`.
To generate a new token, rerun `source auth/CLOUDRUN.sh`.

## Running the Registry API server on GKE

The [Makefile](Makefile) contains targets that build a Docker image
(`make build`) and that deploy it to GKE (`make deploy-gke`).

Requirements:

- Ensure you have [gcloud](https://cloud.google.com/sdk/gcloud) and
  [kubectl](https://cloud.google.com/kubernetes-engine/docs/quickstart)
  installed.

- If not already done, `gcloud auth login` gets user credentials for subsequent
  `gcloud` operations and `gcloud config set project PROJECT_ID` can be used to
  set your project ID to the one where you plan to host your servce.

- The Makefile gets your project ID from the `REGISTRY_PROJECT_IDENTIFIER`
  environment variable. This can be set automatically by running
  `source auth/GKE.sh`.

For detailed steps on how to deploy to GKE, please refer to
[deployments/gke/README.md](deployments/gke/README.md).

```

## License

This software is licensed under the Apache License, Version 2.0. See
[LICENSE](LICENSE) for the full license text.

## Disclaimer

This is not an official Google product. Issues filed on Github are not subject
to service level agreements (SLAs) and responses should be assumed to be on an
ad-hoc volunteer basis.

## Contributing

Contributions are welcome! Please see [CONTRIBUTING](CONTRIBUTING.md) for notes
on how to contribute to this project.
```
