#!/usr/bin/env python3
# _*_ coding: utf-8

import sys
import argparse
import ifcopenshell
import datetime
import traceback
import os
import json

ifc_types = ['IfcProduct']

def Ifc2gltf(filepath):
    storeys = {}
    tmpStoreys = {}
    
    if not os.path.lexists(filepath):
        print("Filepath " + filepath + " not found")
        return 1

    folder = os.path.dirname(filepath)
    file = ifcopenshell.open(filepath)

    for ifc_type in ifc_types:
        #Get all elements for current type
        elements = file.by_type(ifc_type)

        for element in elements:
            eltType = element.is_a()
            if eltType != "IfcBuildingStorey":
                continue
            tmpStoreys[element.Elevation] = {
                "name": element.Name
            }
    try:
        i = 0
        for s in sorted(tmpStoreys.keys()):
            name = tmpStoreys[s]["name"]
            filename = "map_" + str(i) + ".gltf"
            path = "maps/" + filename
            storeys[str(i)] = {
                "name": name,
                "filepath": path,
                "filename": filename
            }
            i += 1
    except Exception as exc:
        print("exc", exc)
        return 1

    for s in storeys:
        storey = storeys[s]
        gltf = os.path.join(folder, storey["filename"])
        dae = gltf.replace("gltf", "dae")
        name = storey["name"]
        cmd = "IfcConvert "+ filepath + " "+ dae +" --use-element-names --include+=arg Name \""+ name +"\""
        print(cmd)
        os.system(cmd)

        cmd = "COLLADA2GLTF-bin -i " + dae + " -o " + gltf
        print(cmd)
        res = os.system(cmd)
        try:
            os.remove(dae)
        except:
            pass
        if res != 0:
            print("Finished with error " + str(res))
            return 1
        print("File " + gltf + " created")

    with open(os.path.join(folder, "maps.json"), "w") as f:
        json.dump(storeys, f, sort_keys=True, ensure_ascii=False, indent=4)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-i", "--ifc", type=str, help="IFC file path")
    args = parser.parse_args()
    return Ifc2gltf(args.ifc)

if __name__ == "__main__":
    sys.exit(main())
