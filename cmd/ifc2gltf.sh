#!/bin/bash

if [ $# -ne 2 ]; then
    echo "ifc2gltf <source.ifc> <storey1,storey2,etc.>"
    exit 1
fi

src=$1
storey=$2

if [ ! -f "$src" ]; then
    echo "File  $src not found"
    exit 1
fi

IFS="," #separator
for str in $storey
do
    echo "parse $str"

    dae="$str.dae"
    dst="$str.gltf"

    echo "IfcConvert $src $dae --use-element-names --include+=arg Name "$str" --use-element-hierarchy"

    IfcConvert $src $dae --use-element-names --include+=arg Name "$str" --use-element-hierarchy

    COLLADA2GLTF-bin -i $dae -o $dst
    res=$?
    if [ "$res" != "0" ]; then
        echo "Ifc convert failed with status $res"
        exit $res
    fi

    rm -f $dae

    echo "File $dst ready"
done

exit 0