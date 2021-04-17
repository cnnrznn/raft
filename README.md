# raft

Raft is a consensus protocol for a cluster of peers to maintain a replicated log.
Raft is a crash fault tolerant (CFT) protocol; i.e. machines can fail by crashing
or network partition, but are NOT malicious (byzantine). Peers follow the protocol.

## Usage

### httpraft

1. Clone https://github.com/cnnrznn/raft-postman for a postman collection
    designed for use with `/cmd/httpraft`
2. Build the `httpraft` binary
    ```
    cd cmd/httpraft
    go build
    ```
3. Create a config file defining the network, such as
    ```json
    {
        "peers": [
            "localhost:8000",
            "localhost:8001",
            "localhost:8002"
        ],
        "apis": [
            "localhost:9000",
            "localhost:9001",
            "localhost:9002"
        ]
    }
    ```
    Place this file in the working directory from which you'll run `./httpraft`
4. Run `./httpraft <id>` where `id` is the index in `peers` and `apis` the node will listen.
5. Point the `leader` environment variable in the postman collection to one of of the
    `apis` addresses.

## Current Status

I am currently in the process of implementing the protocol. Here is the list of tasks:

- [x] Leader election protocol
- [x] Log append protocol
- [x] Module interface for Go clients
- [x] Http REST interface for Web clients

I am implementing the protocol as a go module. I will write instructions later for
installing and using the module in your go program. I am also implementing a
runnable binary that will accept requests (appends) and reads from the log via
HTTP REST endpoints.

### Out of scope

- Membership changes
- Persistence
- Log compaction
