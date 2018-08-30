# import the necessary packages
from colordescriptor import ColorDescriptor
import argparse
import glob
import cv2
import os
import numpy as np
from skimage import io

# construct the argument parser and parse the arguments
ap = argparse.ArgumentParser()
ap.add_argument("-d", "--dataset", required=True,
                help="Path to the directory that contains the images to be indexed")
ap.add_argument("-i", "--index", required=True,
                help="Path to where the computed index will be stored")
args = vars(ap.parse_args())

with open(args["dataset"]) as f:
    img_list = f.readlines()
img_list = [x.strip() for x in img_list]

# initialize the color descriptor
cd = ColorDescriptor((8, 12, 3))

# open the output index file for writing
output = open(args["index"], "w")

cnt = 0
for img_file in img_list:
   if img_file.find(',') >= 0:
        continue
#        if cnt > 12000:
#            continue

        # extract the image ID (i.e. the unique filename) from the image
        # path and load the image itself
    imagePath = os.path.join(
        '/home/pangolins/work/flask-image-search/app/static/' + img_file)
        '/home/pangolins/work/flask-image-search/app/static/' + img_file)
    print imagePath

    image = io.imread(imagePath)

    # write the features to file
    # write the features to file
    features = [str(f) for f in features]
    output.write("%s,%s\n" % (img_file, ",".join(features)))
    cnt = cnt + 1

# close the index file
output.close()
