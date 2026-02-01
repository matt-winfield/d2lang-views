# d2lang-views

A generator built on top of https://d2lang.com/ to create sub-views of larger diagrams.

D2 doesn't natively support referencing entities from the base diagram in layers. This tool fixes that by allowing you to mark layers with `#view` and reference base entities by name. The tool automatically expands references to include full entity definitions with labels and relationships.

## Installation

```bash
go install github.com/matt-winfield/d2lang-views@latest
```

## Usage

```bash
d2lang-views ./path/to/diagram.d2 output/diagram.svg
```

### Options

- `--layout, -l` - Layout engine to use (e.g., dagre, elk)
- `--debug, -d` - Enable debug output (intermediate AST files)
- `--views-only` - Only output SVGs for view layers (marked with `#view`)
- `--watch, -w` - Watch source file and imports for changes, automatically recompile

### Watch Mode

Use `--watch` or `-w` to automatically recompile when the source file or any of its imports change:

```bash
d2lang-views --watch ./path/to/diagram.d2 output/diagram.svg
```

The watcher:

- Monitors the source D2 file for changes
- Automatically detects and watches imported files (`@import` syntax)
- Recompiles when any watched file changes
- Updates watched files after each compilation (picks up new imports)
- Continues running even if compilation fails (shows error and waits for next change)

Press `Ctrl+C` to stop watching.

### Views-Only Mode

Use `--views-only` to skip rendering non-view layers:

```bash
d2lang-views --views-only ./path/to/diagram.d2 output/diagram.svg
```

This is useful when you only need the focused view diagrams and don't need the full diagram or regular layers rendered.

### Output

This generates:

- `<source-dir>/<filename>-compiled.d2` - Compiled D2 file in the same directory as the source (preserves relative import paths)
- `output/diagram/` - SVG files including `index.svg` and individual layer SVGs

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

By default, when you reference a nested entity like `system.api`, only the leaf entity is included (the parent `system` is filtered out). You have two options to include parent containers:

**Option 1: Explicitly list all containers**

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

**Option 2: Use `# include-parents` comment**

Add `# include-parents` after any entity reference to automatically include all its parent containers:

```d2
system: {
    api: "API Gateway" {
        auth: "Auth Service"
    }
}

system.api.auth -> database

layers: {
    auth_view: { #view
        system.api.auth # include-parents
        database
    }
}
```

Generates:

```d2
layers: {
    auth_view: {
        system: "System"
        system.api: "API Gateway"
        system.api.auth: "Auth Service"
        database
        system.api.auth -> database
    }
}
```

The `# include-parents` comment is particularly useful for deeply nested entities where listing all ancestors would be tedious.

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

### Overriding Edge Properties

Use `#override` to override edge properties (labels, styles, classes, etc.) from the base diagram with view-specific values:

```d2
client: "Web Client"
server: "API Server"
database: "PostgreSQL"

client -> server: "HTTP requests" {
    style: {
        stroke: "#ff0000"
    }
}
server -> database: "SQL queries"

layers: {
    client_view: { #view
        client
        server
        client -> server: "REST API calls" {
            style: {
                stroke: "#00ff00"
            }
        } #override
    }
}
```

Generates:

```d2
layers: {
    client_view: {
        client: "Web Client"
        server: "API Server"
        client -> server: "REST API calls" {
            style: {
                stroke: "#00ff00"
            }
        }
    }
}
```

The `#override` comment:

- Overrides properties (label, style, classes, tooltip, link) of the first matching edge from the base diagram
- If no matching edge exists, adds the edge as a new connection
- View properties take precedence; base properties are used as defaults
- Works with all edge types (`->`, `<-`, `<->`, `--`)
- Matches edges case-insensitively

Edge styles, classes, and other properties from the base diagram are automatically preserved in views, even without `#override`.

### Removing Edges

Use `#remove` to exclude specific edges from the base diagram in a view:

```d2
client: "Web Client"
server: "API Server"
database: "PostgreSQL"

client -> server: "HTTP requests"
server -> database: "SQL queries"

layers: {
    client_view: { #view
        client
        server
        database
        server -> database #remove
    }
}
```

Generates:

```d2
layers: {
    client_view: {
        client: "Web Client"
        server: "API Server"
        database: "PostgreSQL"
        client -> server: "HTTP requests"
    }
}
```

The `#remove` comment:

- Removes **all** matching edges from the base diagram (not just the first)
- Matches by source, destination, and arrow direction
- The removed edge is not added as a new edge
- Works with all edge types (`->`, `<-`, `<->`, `--`)
- Matches edges case-insensitively
- Can be combined with `#override` on other edges in the same view

### Referencing with Wildcards

You can reference multiple entities using wildcard patterns with `# include pattern=<pattern>`:

```d2
infrastructure: {
    compute: {
        server1: "Server 1"
        server2: "Server 2"
    }
    storage: {
        db1: "Database 1"
        db2: "Database 2"
    }
    cache: {
        redis1: "Redis Cache 1"
        redis2: "Redis Cache 2"
    }
}

infrastructure.compute.server1 -> infrastructure.storage.db1
infrastructure.compute.server2 -> infrastructure.storage.db2
infrastructure.compute.server1 -> infrastructure.cache.redis1
infrastructure.compute.server2 -> infrastructure.cache.redis2

layers: {
    wildcard: { #view
        # include pattern=infrastructure.compute.*
        # include pattern=infrastructure.storage.*
    }
}
```

Generates:

```d2
layers: {
    wildcard: {
        infrastructure.compute.server1: "Server 1"
        infrastructure.compute.server2: "Server 2"
        infrastructure.storage.db1: "Database 1"
        infrastructure.storage.db2: "Database 2"
        infrastructure.compute.server1 -> infrastructure.storage.db1
        infrastructure.compute.server2 -> infrastructure.storage.db2
    }
}
```

### Referencing with Classes (tags)

You can reference entities by class using `# include class=<class-name>`. This can be useful to tag entities and include them in views:

```d2
infrastructure: {
        server1: "Server 1" { class: "compute" }
        server2: "Server 2" { class: "compute" }
        db1: "Database 1" { class: "storage" }
        db2: "Database 2" { class: "storage" }
        redis1: "Redis Cache 1" { class: "cache" }
        redis2: "Redis Cache 2" { class: "cache" }
}

infrastructure.server1 -> infrastructure.db1
infrastructure.server2 -> infrastructure.db2
infrastructure.server1 -> infrastructure.redis1
infrastructure.server2 -> infrastructure.redis2

layers: {
    wildcard: { #view
        # include class=compute
        # include class=storage
    }
}
```

Generates:

```d2
layers: {
    wildcard: {
        server1: "Server 1"
        server2: "Server 2"
        db1: "Database 1"
        db2: "Database 2"
        server1 -> db1
        server2 -> db2
    }
}
```

Note that children of matching entities are NOT automatically included, unless they also match the class.
Parent containers of a matching entity are also NOT automatically included, unless explicitly referenced.
The class may include wildcards, e.g. `# include class=service-*`, `# include class=*-db`. The class matching is case-insensitive.

## Features

- Automatic entity expansion with labels
- Relationship copying between referenced entities
- Support for nested entities (with implicit parent filtering)
- Extract nested children without their parent containers
- Auto-include parent containers with `# include-parents` comment
- Preserve edge styles, classes, and properties from base diagram
- Override edge properties with `#override` for view-specific context
- Remove specific edges from views with `#remove`
- Custom view layer names
- Mix referenced and new entities in views

## Roadmap

- [x] Allow including nested children without the parent (e.g. include `parent.child` without `parent`).
- [x] Bring in other properties of entities (styles, classes, etc) from the base diagram.
- [x] Maintain view keywords such as `direction`, `grid-*`, etc.
- [x] Allow adding or overriding edge properties in views (using `#override` syntax).
- [x] Preserve edge styles/classes/properties from the base diagram.
- [x] Allow removing edges from views (using `#remove` syntax).
- [x] Maintain `title: |md` multi-line labels from the base diagram.
- [x] Output the generated d2 diagram in the same directory as the source (to preserve relative import paths).
- [x] Support CLI option to disable outputting non-view layers as SVGs.
- [ ] Support disabling outputting the compiled D2 file via CLI option.
- [x] Support `# include class=<class-name>` to include entities of a specific class in the view. (tagging)
    - [x] Support class wildcards e.g. `# include class=service-*`, `# include class=*-db`
- [x] Support wildcard references `# include pattern=something.*` to include multiple entities matching a pattern.
    - [x] Support wildcards not at the end e.g. `# include pattern=*-worker`, `# include pattern=service-*-db`
- [x] Support auto-including parent containers via `# include-parents` comment following the entity reference.
- [x] Include comments around the generated view definitions for clarity. Include the version of the tool used to generate it.
- [x] Automatically compile the generated view diagrams using the D2 CLI.
- [x] Watch mode to monitor changes in the source diagram and regenerate views automatically.

## Development

```bash
# Build
go build

# Run without building
go run . ./path/to/diagram.d2 output/diagram.svg

# Test
go test ./...
```

Install [golangci-lint](https://golangci-lint.run/usage/install/) for linting:

```bash
golangci-lint run
```
