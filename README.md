# d2lang-views

A generator built on top of https://d2lang.com/ to allow showing sub-views of existing diagrams.

It scans a D2 diagram file for `layers` that are marked with a special comment `#view`. These view layers may reference entities from the main diagram or other layers, in addition to defining new entities and relationships specific to that view. This referencing is not supported natively by D2 at this time.

Any referenced entities and the relationships between them are automatically included in the generated view diagram, allowing for focused visualizations of specific parts of a larger diagram.

## Installation

Ensure you have Go installed, then run:

```bash
go install github.com/matt-winfield/d2lang-views@latest
```

## Usage

```bash
d2lang-views ./path/to/diagram.d2 output/directory
```

This will generate separate D2 files in the specified output directory for each view layer found in the input diagram file. Each generated file will contain the necessary entities and relationships to render the view correctly.

## Example

Given a D2 diagram file `diagram.d2` with the following content:

```d2
first: "First Entity" {
    second: "Second Entity"
    third: "Third Entity"
    fourth: "Fourth Entity"
    second -> fourth
    third -> fourth
}
fifth: "Fifth Entity"
forth -> fifth

layers: {
    view1: { #view
        second
        fourth
        fifth

        AnotherEntity -> SomethingElse
    }
    view2: "Custom View Name" { #view
        second -> third
    }
}
```

## Features

-   [x] Parse D2 diagrams to identify view layers marked with `#view`.
-   [x] Extract entities explicitly defined in the base layer
-   [ ] Extract implicitly defined entities from relationships in the base layer.
-   [ ] Identify referenced entities from the base layer within each view layer.
-   [ ] Generate a new D2 diagram file, copying the necessary entities and relationships into the view layer.
-   [ ] Support `# include classes.<class-name` to include entities of a specific class in the view.
-   [ ] Automatically compile the generated view diagrams using the D2 CLI.
-   [ ] Watch mode to monitor changes in the source diagram and regenerate views automatically.

## Development

To build the project locally, clone the repository and run:

```bash
go build
```

To run the project without building, use:

```bash
go run . ./path/to/diagram.d2 output/directory
```

### Testing

Tests can be run using:

```bash
go test ./...
```
