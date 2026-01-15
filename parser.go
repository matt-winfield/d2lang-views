package main

import (
	"io"

	"oss.terrastruct.com/d2/d2ast"
	"oss.terrastruct.com/d2/d2parser"
)

// parseD2 reads from the provided io.Reader and returns the parsed D2 map.
//
// path is the file path used for error reporting
// reader is the input source containing D2 content
func parseD2(path string, reader io.Reader) (*d2ast.Map, error) {
	opts := &d2parser.ParseOptions{}
	return d2parser.Parse(path, reader, opts)
}

type LayerInfo struct {
	Name  string
	Items []d2ast.Node
}

// getLayers extracts and returns the layers from the given D2 map.
func getLayers(d2map *d2ast.Map) []LayerInfo {
	var layers []LayerInfo
	for _, node := range d2map.Children() {
		layers = append(layers, LayerInfo{
			Name:  node.Type(),
			Items: node.Children(),
		})
	}
	return layers
}
