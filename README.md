# d2lang-views

A generator built on top of https://d2lang.com/ to create sub-views of larger diagrams.

D2 doesn't natively support referencing entities from the base diagram in layers. This tool fixes that by allowing you to mark layers with `#view` and reference base entities by name. The tool automatically expands references to include full entity definitions with labels and relationships.

## Installation

```bash
go install github.com/matt-winfield/d2lang-views@latest
```

## Usage

```bash
d2lang-views ./path/to/diagram.d2 output/directory
```

Generates `<filename>-with-views.d2` with expanded view layers.

## How It Works

Mark a layer with `#view` comment, then reference entities from the base diagram:

**Input:**
```d2
client: "Web Client"
server: "API Server"
database: "PostgreSQL"
cache: "Redis"

client -> server
server -> database
server -> cache

layers: {
    frontend: { #view
        client
        server
    }
}
```

**Output:**
```d2
layers: {
    frontend: { #view
        client: "Web Client"
        server: "API Server"
        client -> server
    }
}
```

The tool:
- Expands entity references to include their labels
- Automatically includes parent entities for nested references
- Copies relationships where both endpoints are in the view

## Examples

### Nested Entities

```d2
system: {
    api: "API Gateway"
    auth: "Auth Service"
}

system.api -> system.auth

layers: {
    auth_view: { #view
        system.api
        system.auth
    }
}
```

Generates:
```d2
layers: {
    auth_view: { #view
        system
        system.api: "API Gateway"
        system.auth: "Auth Service"
        system.api -> system.auth
    }
}
```

### Multiple Views

```d2
frontend: "React App"
backend: "Node.js"
db: "MongoDB"

frontend -> backend
backend -> db

layers: {
    ui: { #view
        frontend
        backend
    }
    data: { #view
        backend
        db
    }
}
```

### Mixed Content

Views can reference base entities and define new ones:

```d2
prod_db: "Production DB"
prod_api: "Production API"

prod_api -> prod_db

layers: {
    deployment: { #view
        prod_api
        prod_db

        dev_env: "Development"
        dev_env -> prod_api: "mirrors"
    }
}
```

## Features

- Automatic entity expansion with labels
- Relationship copying between referenced entities
- Support for nested entities
- Support for all arrow types (`->`, `<-`, `<->`, `--`)
- Custom view layer names
- Mix referenced and new entities in views

## Development

```bash
# Build
go build

# Run without building
go run . ./path/to/diagram.d2 output/directory

# Test
go test ./...
```
