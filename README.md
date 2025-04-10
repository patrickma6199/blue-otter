```{bash}
------------------------------------------------------------------------------
    ____  __    __  ______   ____ _______________  ____     _____ __      ____
   / __ )/ /   / / / / __/  / __ /_  __/_  __/ __/ / __ \   / ___// /    /  _/
  / __  / /   / / / / /_   / / / // /   / / / /_  / /_/ /  / /   / /     / /  
 / /_/ / /___/ /_/ / __/  / /_/ // /   / / / __/ / _, _/  / /_  / /___  / / 
/_____/_____/\____/___/   \____//_/   /_/ /___/ /_/ |_|  /____//_____//___/  

--------------------------------- v0.1.0 -------------------------------------
```

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

```{bash}
blue-otter client --username YourName --room RoomName --port 42069
```

### Bootstrap

Run as a bootstrap node for other Blue Otter instances:

```{bash}
blue-otter bootstrap --port 42069
```

### Add Bootstrap

Add a bootstrap node address to your configuration:

```{bash}
blue-otter add-bootstrap --address "/ip4/127.0.0.1/tcp/42069/p2p/QmHashValue"
```

### Remove Bootstrap

Remove a bootstrap node address from your configuration:

```{bash}
blue-otter remove-bootstrap --address "/ip4/127.0.0.1/tcp/42069/p2p/QmHashValue"
```

### List Bootstrap

List all saved bootstrap node addresses:

```{bash}
blue-otter list-bootstrap
```

### Clean Up

Clean up the Blue Otter configuration directory:

```{bash}
blue-otter clean-up
```

Use `--force` or `-f` to skip confirmation prompt.
