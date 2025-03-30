# Blue Otter

## A peer-to-peer messaging CLI

Blue Otter is a simple peer-to-peer messaging application built with libp2p. It allows users to communicate with each other in chatrooms without relying on centralized servers.

## Features

- Start a client to join chat rooms
- Run as a bootstrap node to help others join the network
- Manage bootstrap node addresses
- Clean up configuration

## Commands

### Client

Start the Blue Otter client to join a chat room:

```
blue-otter client --username YourName --room RoomName --port 42069
```

### Bootstrap

Run as a bootstrap node for other Blue Otter instances:

```
blue-otter bootstrap --port 42069
```

### Add Bootstrap

Add a bootstrap node address to your configuration:

```
blue-otter add-bootstrap --address "/ip4/127.0.0.1/tcp/42069/p2p/QmHashValue"
```

### Remove Bootstrap

Remove a bootstrap node address from your configuration:

```
blue-otter remove-bootstrap --address "/ip4/127.0.0.1/tcp/42069/p2p/QmHashValue"
```

### List Bootstrap

List all saved bootstrap node addresses:

```
blue-otter list-bootstrap
```

### Clean Up

Clean up the Blue Otter configuration directory:

```
blue-otter clean-up
```

Use `--force` or `-f` to skip confirmation prompt.
