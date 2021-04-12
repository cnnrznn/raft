# raft

Raft is a consensus protocol for a cluster of peers to maintain a replicated log.
Raft is a crash fault tolerant (CFT) protocol; i.e. machines can fail by crashing
or network partition, but are NOT malicious (byzantine). Peers follow the protocol.

## Current Status

I am currently in the process of implementing the protocol. Here is the list of tasks:

- [x] Leader election protocol
- [x] Log append protocol
- [ ] Module interface for Go clients
- [ ] Http REST interface for Web clients

I am implementing the protocol as a go module. I will write instructions later for
installing and using the module in your go program. I am also implementing a
runnable binary that will accept requests (appends) and reads from the log via
HTTP REST endpoints.

### Out of scope

- Membership changes
- Persistence (we'll see in the future)
- Log compaction
