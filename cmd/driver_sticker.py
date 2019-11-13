#!/usr/bin/env python3
# _*_ coding: utf-8

import sys
import traceback
import qrcode
import os
import subprocess
import json
import argparse

from PIL import Image, ImageDraw, ImageFont


print("======== STARTING =========")

parser = argparse.ArgumentParser()
parser.add_argument("type", help="device type, eg: led, blind, sensor, hvac", type=str)
parser.add_argument("mac", help="mac address in format 00:00:00:00:00:00", type=str)
args = parser.parse_args()

print("type=" + args.type)
print("mac=" + args.mac)

# params
#lbl = "LED"
lbl=args.type.upper()
color = None
#mac="00:00:00:00:00:00"
mac=args.mac
product=lbl+"-200"

# generate json
data_json = {"mac":mac, "device":lbl}
json_string = json.dumps(data_json)

# fonts
fnt = ImageFont.truetype('/usr/local/share/fonts/LiberationSans-Bold.ttf', 50)

# ----------------- Creating images
img_width=1420 # size*10
img_height=460 # size*10      
img = Image.new('RGBA', (img_width, img_height), color=color)   # 50mmx16mm or 142px.46px @72pix/inch, multiplicated by 10 for comfort
width, height = img.size

# QR code
qr = qrcode.QRCode(
    version=1,
    error_correction=qrcode.constants.ERROR_CORRECT_L,
    box_size=11, # 9 for complex version
    border=0,
)

qr.add_data(json_string)
qr.make(fit=True)

img_qrcode = qr.make_image(fill_color="black", back_color="transparent")
img.paste(img_qrcode, (img_width-440,20))

# load and paste energieIP logo
img_logo= Image.open("/usr/local/share/images/logo.png")
img.paste(img_logo, (10,10))

# load and paste bin picture
img_bin= Image.open("/usr/local/share/images/bin.png")
img.paste(img_bin, (768,15))

# load and paste CE picture
img_ce= Image.open("/usr/local/share/images/ce.png")
img.paste(img_ce, (760,164))

# load and paste selv picture
img_selv= Image.open("/usr/local/share/images/selv.png")
img.paste(img_selv, (757,305))

# load and paste madeinfrance picture
img_madeinfrance= Image.open("/usr/local/share/images/madeinfrance.png")
img.paste(img_madeinfrance, (546,403))

# load and paste url picture
img_url= Image.open("/usr/local/share/images/energieip.png")
img.paste(img_url, (17,403))

# write line1
txt_line1 = "PoE " + lbl + " driver"
empty_img = Image.new('RGBA', (650, 60), color=color)
draw = ImageDraw.Draw(empty_img)
draw.text((0,0), txt_line1, font=fnt, fill=(0,0,0))
img.paste(empty_img, (17,155))

# write line2
txt_line2 = product
empty_img = Image.new('RGBA', (650, 60), color=color)
draw = ImageDraw.Draw(empty_img)
draw.text((0,0), txt_line2, font=fnt, fill=(0,0,0))
img.paste(empty_img, (17,233))

# write line3
txt_line3 = mac
empty_img = Image.new('RGBA', (650, 60), color=color)
draw = ImageDraw.Draw(empty_img)
draw.text((0,0), txt_line3, font=fnt, fill=(0,0,0))
img.paste(empty_img, (17,316))


# ----------------- Building image
img.save("/tmp/sticker.png")

os.system("convert /tmp/sticker.png /tmp/sticker.pdf")

print("======== COMPLETE =========")

sys.exit()

