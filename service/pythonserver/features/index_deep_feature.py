# import the necessary packages
import argparse
import glob
import cv2
import os
import numpy as np
from skimage import io
from numpy import linalg as LA
from keras.applications.vgg16 import VGG16
from keras.preprocessing import image
from keras.applications.vgg16 import preprocess_input
from cnndescriptor import CNNDescriptor

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

# initialize the model
input_shape = (224, 224, 3)
model = VGG16(weights='imagenet', input_shape=(input_shape[0], input_shape[1], 
    input_shape[2]), pooling='max', include_top=False)
        

# open the output index file for writing
output = open(args["index"], "w")

dc = CNNDescriptor()
cnt = 0
for img_file in img_list:
    if img_file.find(',') >= 0:
        continue
#    if cnt > 10:
#        break
        # extract the image ID (i.e. the unique filename) from the image
        # path and load the image itself
    imgPath = os.path.join('/home/pangolins/work/flask-image-search/app/static/' + img_file)
    try:
        print imgPath
#        img = image.load_img(imgPath, target_size=(input_shape[0], input_shape[1]))
#        img = image.img_to_array(img)
#        img = np.expand_dims(img, axis=0)
#        img = preprocess_input(img)
#        feat = model.predict(img)
#        features = feat[0]/LA.norm(feat[0])

        features = dc.extract_feature(imgPath)
        # write the features to file
        features = features.flatten().tolist()
        features = [str(f) for f in features]
        output.write("%s,%s\n" % (img_file, ",".join(features)))
        cnt = cnt + 1
    except:
        continue
# close the index file
output.close()
