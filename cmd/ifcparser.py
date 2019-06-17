#!/usr/bin/env python3
# _*_ coding: utf-8

import sys
import argparse
import ifcopenshell
import datetime
import traceback
import os
import json

#To install ifcopenshell refers to https://github.com/IfcOpenShell/IfcOpenShell

#Set wich element we want to get
ifc_types = ['IfcProduct']

filters = ["IfcBuildingElementProxy", "IfcFurnishingElement"]

def parseIfc(filepath):
    #Define lists for storing unique property names and unique type properties name
    prop_keys = []
    instances = []

    projects = {}
    models = {}
    drivers = {}
    groups = {}
    dump = {}

    if not os.path.lexists(filepath):
        print("Filepath " + filepath + " not found")
        return 1

    file = ifcopenshell.open(filepath)

    for ifc_type in ifc_types:
        #Get all elements for current type
        elements = file.by_type(ifc_type)

        #Define a dictionary for storing current element
        for element in elements:
            instance = {}
            eltType = element.is_a()
            if eltType not in filters:
                continue
            instance_properties = {}
            prop_sets = element.IsDefinedBy
            instance['Label'] = element.Name

            for prop_set in prop_sets:
                if not prop_set.is_a('IfcRelDefinesByProperties'):
                    continue
                properties = prop_set.RelatingPropertyDefinition.HasProperties
                for prop in properties:
                    try:
                        instance_properties[prop.Name] = prop.NominalValue.wrappedValue
                        if prop.Name not in prop_keys:
                            prop_keys.append(prop.Name)
                    except Exception as exc:
                        pass

                if not prop_set.is_a('IfcRelDefinesByType'):
                    continue
                if prop_set.RelatingType.HasPropertySets is None:
                    continue
                type_prop_sets = prop_set.RelatingType.HasPropertySets
                for type_prop_set in type_prop_sets:
                    if (type_prop_set.is_a('IfcPropertySet')):
                        properties = type_prop_set.HasProperties
                        for prop in properties:
                            try:
                                instance_properties[prop.Name] = prop.NominalValue.wrappedValue
                                if prop.Name not in prop_keys:
                                    prop_keys.append(prop.Name)
                            except Exception as exc:
                                    pass

            instance['properties'] = instance_properties
            instances.append(instance)

    try:
        for instance in instances:
            infos = instance["Label"].split("_", 1)
            if len(infos) < 2:
                continue
            product = infos[0].lower()
            label = infos[1]
            if "mobilier" in product:
                continue
            modelName = instance['properties'].get("SKU (BO_prodsku)", product)
            modbusID = instance['properties'].get("modbusID", 0)
            group = instance['properties'].get("Group", 0)

            projects[label] = {
                "label": label,
                "modbusID": modbusID,
                "modelName": modelName
            }

            if group not in groups:
                groups[group] = {
                    "group": group
                }

            deviceType = ""
            if "led" in product:
                deviceType = "led"
            elif "bld" in product:
                deviceType = "blind"
            elif "mca" in product:
                deviceType = "sensor"
            elif "hvac" in product:
                deviceType = "hvac"
            elif "swh" in product:
                deviceType = "switch"
            elif "fra" in product:
                deviceType = "frame"

            if deviceType not in drivers:
                drivers[deviceType] = {}

            drivers[deviceType][label] = {
                "label": label,
                "group": group,
            }

            if deviceType == "led":
                drivers[deviceType][label]["pMax"] = instance['properties'].get("Power", 0)

            if modelName in models:
                continue
            
            vendor = ""
            if "Manufacturer" in instance['properties']:
                vendor = instance['properties']["Manufacturer"]
            elif "ManufacturName (BO_Manufac)" in instance['properties']:
                vendor = instance['properties']["ManufacturName (BO_Manufac)"]
            url = ""
            if "TechnicalDescription (BO_techcert)" in instance['properties']:
                url = instance['properties']["TechnicalDescription (BO_techcert)"]
            elif "ProductUrl (BO_producturl)" in instance['properties']:
                url = instance['properties']["ProductUrl (BO_producturl)"]

            models[modelName] = {
                "vendor": vendor,
                "name": modelName,
                "url": url,
                "deviceType": deviceType
            }
    except Exception as exc:
        print("exc", exc)
        return 1

    dump = {
        "groups": groups,
        "leds": drivers.get("led", {}),
        "sensors": drivers.get("sensor", {}),
        "hvacs": drivers.get("hvac", {}),
        "models": models,
        "switchs": drivers.get("switch", {}),
        "projects": projects
    }

    print(json.dumps(dump, indent=4, sort_keys=True))
    return 0


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-i", "--ifc", type=str, help="IFC file path")
    args = parser.parse_args()
    return parseIfc(args.ifc)

if __name__ == "__main__":
    sys.exit(main())
