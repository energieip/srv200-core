#!/usr/bin/env python3
# _*_ coding: utf-8

import sys
import argparse
import ifcopenshell
import datetime
import traceback
import os
import json
import threading

ifc_types = ['IfcProduct']

class ExtractStorey(threading.Thread):
    def __init__(self, filepath, gltf, dae, name):
        threading.Thread.__init__(self)
        self.gltf = gltf
        self.dae = dae
        self.name = name
        self.filepath = filepath

    def run(self):
        cmd = "IfcConvert \""+ self.filepath + "\" \""+ self.dae +"\" -y --center-model --use-element-names --include+=arg Name \""+ self.name +"\""
        print(cmd)
        res = os.system(cmd)
        if res != 0:
            print("Finished with error " + str(res))
            sys.exit(1)

        cmd = "COLLADA2GLTF-bin -i " + self.dae + " -o " + self.gltf
        print(cmd)
        res = os.system(cmd)
        try:
            os.remove(self.dae)
        except:
            pass
        if res != 0:
            print("Finished with error " + str(res))
            sys.exit(1)
        print("File " + self.gltf + " created")

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
                "filename": filename,
                "default": False,
            }
            if s == 0:
                storeys[str(i)]["default"] = True
            i += 1
    except Exception as exc:
        print("exc", exc)
        return 1

    threads = []
    for s in storeys:
        storey = storeys[s]
        gltf = os.path.join(folder, storey["filename"])
        dae = gltf.replace("gltf", "dae")
        name = storey["name"]
        t = ExtractStorey(filepath, gltf, dae, name)
        t.start()
        threads.append(t)

    for s in threads:
        s.join()

    with open(os.path.join(folder, "maps.json"), "w") as f:
        json.dump(storeys, f, sort_keys=True, ensure_ascii=False, indent=4)

def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-i", "--ifc", type=str, help="IFC file path")
    args = parser.parse_args()
    return Ifc2gltf(args.ifc)

if __name__ == "__main__":
    sys.exit(main())
