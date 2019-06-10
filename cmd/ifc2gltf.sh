#!/bin/bash

if [ $# -ne 2 ]; then
    echo "ifc2gltf <source.ifc> <destination.gltf>"
    exit 1
fi

src=$1
dst=$2

if [ ! -f "$src" ]; then
    echo "File  $src not found"
    exit 1
fi

dae="${src%.ifc}.dae"
optimized="${dst%.gltf}.optimized.gltf"

IfcConvert --use-element-names $src $dae
res=$?
if [ "$res" != "0" ]; then
    echo "Ifc convert failed with status $res"
    exit $res
fi

COLLADA2GLTF-bin -i $dae -o $dst
res=$?
if [ "$res" != "0" ]; then
    echo "Ifc convert failed with status $res"
    exit $res
fi

rm -f $dae

echo "Normal conversion $dst"

exit 0