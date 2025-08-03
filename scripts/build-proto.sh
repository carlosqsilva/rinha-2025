#!/bin/sh -e

echo "Compiling Cap'n Proto files..."

for f in internal/models/*.capnp; do
  capnp compile -ogo:. -I/go/pkg/mod/capnproto.org/go/capnp/v3@v3.1.0-alpha.1/std "$f"
done

echo "Compilation finished."
