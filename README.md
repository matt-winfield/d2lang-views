# d2lang-views

A generator built on top of https://d2lang.com/ to allow showing sub-views of existing diagrams.

It scans a D2 diagram file for `layers` that are marked with a special comment `#view`. These view layers may reference entities from the main diagram or other layers, in addition to defining new entities and relationships specific to that view. This referencing is not supported natively by D2 at this time.

Any referenced entities and the relationships between them are automatically included in the generated view diagram, allowing for focused visualizations of specific parts of a larger diagram.

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
