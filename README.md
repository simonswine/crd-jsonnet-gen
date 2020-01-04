# crd-jsonnet-gen

This is a PoC for generating Jsonnet from CRD specs

## Example

```
# Build binary
go build

# Generate Jsonnet from cert-manager
./crd-jsonnet-gen github.com/jetstack/cert-manager/pkg/apis > examples/cert-manager.libsonnet

# Generate Jsonnet from contour
./crd-jsonnet-gen github.com/projectcontour/contour/apis/contour/v1beta1 > examples/contour.libsonnet
```
