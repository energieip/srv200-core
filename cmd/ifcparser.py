#!/usr/bin/env python3
# _*_ coding: utf-8

import sys
import argparse
import ifcopenshell
import datetime
import traceback
import os
import json
import collections

#To install ifcopenshell refers to https://github.com/IfcOpenShell/IfcOpenShell

#Set wich element we want to get
ifc_types = ['IfcProduct']

filters = ["IfcBuildingElementProxy", "IfcFurnishingElement"]


def isDriver(deviceType):
    if deviceType in ["led", "blind", "hvac", "sensor"]:
        return True
    return False

def getDeviceType(driver):
    return driver['properties'].get("Type", "").lower()

def buildDriver(driver):
    res = {}
    deviceType = getDeviceType(driver)
    if not isDriver(deviceType):
        return res
    label = driver["Label"]
    group = driver['properties'].get("Group", 0)
    friend = driver['properties'].get("FriendlyName", label)
    freq = driver['properties'].get("DumpFrequency", 1000)
    modbusID = driver['properties'].get("ModbusID", 0)

    res = collections.OrderedDict({
        "label": label,
        "group": group,
        "friendlyName": friend,
        "dumpFrequency": freq,
        "modbusID": modbusID
    })

    if deviceType != "hvac":
        res["isBleEnabled"] = driver['properties'].get("ActivateBluetooth", False)
        res["bleMode"] = driver['properties'].get("BluetoothMode", "service")
        res["iBeaconUUID"] = driver['properties'].get("IBeaconUUID", "")
        res["iBeaconMajor"] = driver['properties'].get("IBeaconMajor", 0)
        res["iBeaconMinor"] = driver['properties'].get("IBeaconMinor", 0)
        res["iBeaconTxPower"] = driver['properties'].get("IBeaconTxPower", 0)

    if deviceType == "led":
        res.update(buildLed(driver))
    elif deviceType == "sensor":
        res.update(buildSensor(driver))
    return res

def buildLed(driver):
    deviceType = getDeviceType(driver)
    if deviceType != "led":
        return {}
    return collections.OrderedDict({
        "pMax": driver['properties'].get("Power", 0),
        "defaultSetpoint": driver['properties'].get("DefaultSetpoint", 5),
        "firstDay": driver['properties'].get("FirstDay", False)
    })

def buildSensor(driver):
    res = {}
    deviceType = getDeviceType(driver)
    if deviceType != "sensor":
        return res
    return collections.OrderedDict({
        "thresoldPresence": driver['properties'].get("ThresoldPresence", 10)
    })

def buildSwitch(driver):
    res = {}
    deviceType = getDeviceType(driver)
    if deviceType != "switch":
        return res
    label = driver["Label"]
    friend = driver['properties'].get("FriendlyName", label)
    freq = driver['properties'].get("DumpFrequency", 1000)
    modbusID = driver['properties'].get("ModbusID", 0)
    cluster = driver['properties'].get("Cluster", 0)
    res = collections.OrderedDict({
        "label": label,
        "friendlyName": friend,
        "dumpFrequency": freq,
        "modbusID": modbusID,
        "cluster": cluster
    })
    return res

def buildWago(driver):
    res = {}
    deviceType = getDeviceType(driver)
    if deviceType != "wago":
        return res
    label = driver["Label"]
    friend = driver['properties'].get("FriendlyName", label)
    slaveID = driver['properties'].get("SlaveID", 0)
    cluster = driver['properties'].get("Cluster", 0)
    apiType = driver['properties'].get("APIType", "modbus")
    res = collections.OrderedDict({
        "label": label,
        "friendlyName": friend,
        "slaveID": slaveID,
        "cluster": cluster,
        "apiType": apiType,
        "api": driver['properties'].get("valeur de tableau", {})
    })
    return res

def buildFrame(driver):
    res = {}
    deviceType = getDeviceType(driver)
    if deviceType != "frame":
        return res
    label = driver["Label"]
    friend = driver['properties'].get("FriendlyName", label)
    cluster = driver['properties'].get("Cluster", 0)
    modbusID = driver['properties'].get("ModbusID", 0)
    res = {
        "label": label,
        "friendlyName": friend,
        "modbusID": modbusID,
        "cluster": cluster
    }
    return res

def parseIfc(filepath):
    instances = []

    projects = collections.OrderedDict()
    models = collections.OrderedDict()
    drivers = collections.OrderedDict()
    groups = collections.OrderedDict()

    if not os.path.lexists(filepath):
        print("Filepath " + filepath + " not found")
        return 1

    file = ifcopenshell.open(filepath)

    for ifc_type in ifc_types:
        #Get all elements for current type
        elements = file.by_type(ifc_type)

        #Define a dictionary for storing current element
        for element in elements:
            instance = collections.OrderedDict()
            eltType = element.is_a()
            if eltType not in filters:
                continue
            instance_properties = collections.OrderedDict()
            prop_sets = element.IsDefinedBy
            instance['Label'] = element.Name

            for prop_set in prop_sets:
                if not prop_set.is_a('IfcRelDefinesByProperties'):
                    continue
                properties = prop_set.RelatingPropertyDefinition.HasProperties
                for prop in properties:
                    try:
                        if getattr(prop, "NominalValue", None):
                            instance_properties[prop.Name] = prop.NominalValue.wrappedValue
                        elif prop.is_a("IfcPropertyTableValue"):
                            sub_dict = collections.OrderedDict()
                            for k, v in enumerate(prop.DefiningValues):
                                sub_dict[v.wrappedValue] = prop.DefinedValues[k].wrappedValue
                            instance_properties[prop.Name] = sub_dict
                        elif prop.is_a("IfcPropertyEnumeratedValue"):
                            t = ()
                            for v in prop.EnumerationValues:
                                t = (*t, v.wrappedValue)
                            instance_properties[prop.Name] = t
                        else:
                            instance_properties[prop.Name] = prop
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
                            except Exception as exc:
                                pass

            instance['properties'] = instance_properties
            instances.append(instance)

    try:
        for instance in instances:
            label = instance["Label"]
            if "mobilier" in label.lower():
                continue
            deviceType = instance['properties'].get("Type", "").lower()
            if deviceType == "":
                continue
            modelName = instance['properties'].get("ModelLabel", deviceType)
            group = instance['properties'].get("Group", 0)

            projects[label] = {
                "label": label,
                "modelName": modelName
            }

            if group not in groups:
                groups[group] = {
                    "group": group,
                    "modbusID": group
                }

            if deviceType not in drivers:
                drivers[deviceType] = collections.OrderedDict()

            if deviceType not in ["switch", "frame", "wago"]:
                modbusID = instance['properties'].get("ModbusID", 0)
                projects[label]["modbusID"] = modbusID
                drivers[deviceType][label] = buildDriver(instance)
            elif deviceType == "switch":
                modbusID = instance['properties'].get("ModbusID", 0)
                projects[label]["modbusID"] = modbusID
                drivers[deviceType][label] = buildSwitch(instance)
            elif deviceType == "wago":
                drivers[deviceType][label] = buildWago(instance)
                slaveID = instance['properties'].get("SlaveID", 0)
                projects[label]["slaveID"] = slaveID
            elif deviceType == "frame":
                modbusID = instance['properties'].get("ModbusID", 0)
                projects[label]["modbusID"] = modbusID
                drivers[deviceType][label] = buildFrame(instance)
            else:
                drivers[deviceType][label] = {
                    "label": label
                }

            if modelName in models:
                continue

            models[modelName] = {
                "vendor": instance['properties'].get("Manufacturer", ""),
                "name": modelName,
                "url": instance['properties'].get("ModelReference", ""),
                "productionYear": instance['properties'].get("ProductionYear", ""),
                "deviceType": deviceType
            }

    except Exception as exc:
        print("exc", exc)
        return 1

    dump = collections.OrderedDict({
        "groups": groups,
        "leds": drivers.get("led", {}),
        "blinds": drivers.get("blind", {}),
        "sensors": drivers.get("sensor", {}),
        "hvacs": drivers.get("hvac", {}),
        "frames": drivers.get("frame", {}),
        "models": models,
        "switchs": drivers.get("switch", {}),
        "wagos": drivers.get("wago", {}),
        "projects": projects
    })
    print(json.dumps(dump, indent=4))
    return 0


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("-i", "--ifc", type=str, help="IFC file path")
    args = parser.parse_args()
    return parseIfc(args.ifc)

if __name__ == "__main__":
    sys.exit(main())
