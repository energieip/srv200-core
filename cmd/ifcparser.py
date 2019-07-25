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
import qrcode
import tempfile
import shutil

from PIL import Image, ImageDraw, ImageFont

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
        "modbusID": modbusID,
        "protocol": driver['properties'].get("APIType", "mqtts"),
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
        "cluster": cluster,
        "protocol": driver['properties'].get("APIType", "rest"),
        "ip": driver['properties'].get("IP", "0"),
        "profil": driver['properties'].get("Profil", "none")
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
    freq = driver['properties'].get("DumpFrequency", 1000)
    apiType = driver['properties'].get("APIType", "modbus")
    res = collections.OrderedDict({
        "label": label,
        "dumpFrequency": freq,
        "friendlyName": friend,
        "slaveID": slaveID,
        "cluster": cluster,
        "apiType": apiType,
        "api": driver['properties'].get("API", {}),
        "ip": driver['properties'].get("IP", "0"),
        "modbusOffset": driver['properties'].get("ModbusOffset", 0),
        "protocol": driver['properties'].get("APIType", "modbus"),
    })
    return res

def buildNanoSense(driver):
    res = {}
    deviceType = getDeviceType(driver)
    if deviceType != "nanosense":
        return res
    label = driver["Label"]
    friend = driver['properties'].get("FriendlyName", label)
    modbusID = driver['properties'].get("ModbusID", 0)
    cluster = driver['properties'].get("Cluster", 0)
    freq = driver['properties'].get("DumpFrequency", 1000)
    apiType = driver['properties'].get("APIType", "modbus")
    res = collections.OrderedDict({
        "label": label,
        "dumpFrequency": freq,
        "friendlyName": friend,
        "modbusID": modbusID,
        "cluster": cluster,
        "apiType": apiType,
        "group": driver['properties'].get("Group", 0),
        "api": driver['properties'].get("API", {}),
        "modbusOffset": driver['properties'].get("ModbusOffset", 0),
        "protocol": driver['properties'].get("APIType", "modbus"),
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
        "cluster": cluster,
        "protocol": driver['properties'].get("APIType", "rest")
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


    ref_folder = os.path.dirname(filepath)
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
        stickers = {}
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
                "modelName": modelName,
                "commissioningDate": ""
            }

            lbl = label.replace("_", "-")
            qr = qrcode.QRCode(
                    version=1,
                    error_correction=qrcode.constants.ERROR_CORRECT_L,
                    box_size=10,
                    border=4,
                    )
            qr.add_data(lbl)
            qr.make(fit=True)

            img_qrcode = qr.make_image(fill_color="black", back_color="transparent")

            stickers[label] = {
                "label": lbl,
                "qrcode": img_qrcode
            }

            if group not in groups:
                groups[group] = {
                    "group": group,
                    "modbusID": group
                }

            if deviceType not in drivers:
                drivers[deviceType] = collections.OrderedDict()

            if deviceType not in ["switch", "frame", "wago", "nanosense"]:
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
            elif deviceType == "nanosense":
                drivers[deviceType][label] = buildNanoSense(instance)
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
        
        folder = tempfile.mkdtemp()
        for label in sorted(stickers):
            sticker = stickers[label]
            lbl = sticker["label"]
            #size in pixels!
            
            color = None
            img = Image.new('RGBA', (1171, 428), color=color)
            width, height = img.size
            d_full = ImageDraw.Draw(img)

            width_small = int(width /2)
            height_small = height
            left_img = Image.new('RGBA', (width_small, height_small), color=color)

            fnt = ImageFont.truetype('/usr/local/share/fonts/LiberationSans-Bold.ttf', 40)
            fntUrl = ImageFont.truetype('/usr/local/share/fonts/LiberationSans-Bold.ttf', 25)
            d_left = ImageDraw.Draw(left_img)

            middle_x = width /2
            middle_y = height / 2
            line_color = (0, 0, 0)

            gridWidth = 10
            for v in range(0, height, gridWidth):
                    d_full.line( (middle_x , v, middle_x, v+5), fill=line_color, width=2)

            w_sticker, h_sticker = sticker["qrcode"].size
            pos_x = int((middle_x / 2) - (w_sticker /2))
            pos_y = int(middle_y - (h_sticker / 2))
            left_img.paste(sticker["qrcode"], (pos_x , pos_y), sticker["qrcode"])

            # text alignment
            text_size_lbl = d_left.textsize(lbl, font=fnt)
            lbl_pos_x = int((width_small / 2) - (text_size_lbl[0] /2))
            lbl_pos_y = int(pos_y /2 - (text_size_lbl[1] /2))
            d_left.text((lbl_pos_x, lbl_pos_y), lbl, font=fnt, fill=(0,0,0))

            #text alignement
            url_eip = "www.energie-ip.com"
            text_size_eip = d_left.textsize(url_eip, font=fntUrl)
            eip_pos_x = int((width_small / 2) - (text_size_eip[0] /2))
            eip_pos_y = int(height - (( height - (pos_y + h_sticker)) /2) - (text_size_eip[1] /2))

            d_left.text((eip_pos_x, eip_pos_y), url_eip, font=fntUrl, fill=(0,0,0))

            angle = 180
            right_img = left_img.rotate(angle)
            img.paste( right_img, (width_small + 2, 0))

            img.paste(left_img, (0 , 0))

            path = os.path.join(folder, lbl + ".png")
            img.save(path)

        os.system("montage -tile 2x7 -geometry 1171x428+7+10 "+ folder+"/*.png " + ref_folder + "/stickers.pdf")
        shutil.rmtree(folder)
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
        "nanosenses": drivers.get("nanosense", {}),
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
