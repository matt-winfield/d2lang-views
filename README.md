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
- Only includes explicitly referenced entities (implicit parents are filtered out)
- Copies relationships where both endpoints are in the view

## Examples

### Nested Entities

When referencing nested entities, only explicitly listed entities are included. Implicit parent containers are filtered out:

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
    auth_view: {
        api: "API Gateway"
        auth: "Auth Service"
        api -> auth
    }
}
```

Notice that `system` is not included because it wasn't explicitly referenced - only `system.api` and `system.auth` were listed. The nested entities become top-level in the view.

### Including Parent Containers

If you want to preserve the parent container, explicitly include it:

```d2
system: {
    api: "API Gateway"
    auth: "Auth Service"
}

system.api -> system.auth

layers: {
    auth_view: { #view
        system
        system.api
        system.auth
    }
}
```

Generates:

```d2
layers: {
    auth_view: {
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
- Support for nested entities (with implicit parent filtering)
- Extract nested children without their parent containers
- Custom view layer names
- Mix referenced and new entities in views

## Roadmap

- [x] Allow including nested children without the parent (e.g. include `parent.child` without `parent`).
- [x] Bring in other properties of entities (styles, classes, etc) from the base diagram.
- [x] Maintain view keywords such as `direction`, `grid-*`, etc.
- [ ] Allow adding or overriding edge labels in views. (what should the syntax be?)
- [ ] Support `# include class=<class-name>` to include entities of a specific class in the view.
- [ ] Support wildcard references `# include pattern=something.*` to include multiple entities matching a pattern.
- [ ] Include comments around the generated view definitions for clarity.
- [x] Automatically compile the generated view diagrams using the D2 CLI.
- [ ] Watch mode to monitor changes in the source diagram and regenerate views automatically.

## Development

```bash
# Build
go build

# Run without building
go run . ./path/to/diagram.d2 output/directory

# Test
go test ./...
```
