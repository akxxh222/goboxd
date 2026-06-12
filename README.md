# goboxd

A small, secure Go service for running untrusted code inside sandboxed environments and returning structured results over HTTP.

Features

- Runs user-submitted code in isolated `nsjail` sandboxes.
- Supported languages

- Python 3 (`py3`)
- C (`c`)
- C++ (`cpp`)
- Java (`java`)
- Bash (`bash`)
- Node.js / JavaScript (`node`)
- Verilog (`verilog`)
- Request validation, resource limits, and output truncation.
- Unit and integration tests with a Docker-based local workflow.

Quick start

Prerequisites: Docker, Make, and a Go toolchain.

1. Build the Docker image:

	 ```sh
	 make build
	 ```

2. Run the service locally:

	 ```sh
	 make run
	 ```

3. Check health:

	 ```sh
	 curl http://localhost:8080/healthz
	 ```

	 Expected response: `{"status":"ok"}`

Development

- Run unit tests:

	```sh
	make test
	```

- Run integration tests:

	```sh
	make integration
	```

- Run Go tests directly:

	```sh
	go test ./...
	```

API and docs

See the docs folder for detailed documentation and API specs: [docs/api.md](docs/api.md), [docs/architecture.md](docs/architecture.md).

Repository layout

- `cmd/` — service entrypoint
- `internal/` — application code (runner, config, httpapi, security)
- `docs/` — design and API documentation
- `tests/` — integration and end-to-end tests

Contributing

Contributions are welcome. Please open issues or pull requests and follow existing code style and tests.

License

This project is licensed under the terms in the repository `LICENSE` file.

Contact

For questions or support, open an issue on the repository.

----

Updated for clarity and streamlined developer onboarding.
