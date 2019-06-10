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
            if element.is_a() not in filters:
                continue
            instance_properties = {}
            prop_sets = element.IsDefinedBy
            instance['Label'] = element.Name
            instance['ModbusID'] = element.Tag
            # TODO parse group
            instance['group'] = 0

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
            try:
                modbusID = int(instance.get("ModbusID", "0"))
            except:
                modbusID = 0
            # label = label.replace("_", "-")

            projects[label] = {
                "label": label,
                "modbusID": modbusID,
                "modelName": modelName
            }

            if instance['group'] not in groups:
                groups[instance['group']] = {
                    "group": instance['group']
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
                "group": instance["group"],
            }

            if deviceType == "led":
                pmax = product.replace("led", "")
                pmax = pmax.replace("w", "")
                drivers[deviceType][label]["pMax"] = int(pmax)

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
        "leds": drivers["led"],
        "sensors": drivers["sensor"],
        "hvacs": drivers["hvac"],
        "models": models,
        "switchs": drivers["switch"],
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
