# import the necessary packages
import numpy as np
from numpy import linalg as LA

from keras.applications.vgg16 import VGG16
from keras.preprocessing import image
from keras.applications.vgg16 import preprocess_input

import urllib, cStringIO
from PIL import Image

class CNNDescriptor:
    def __init__(self, input_shape=(224,224,3)):
        # store the number of bins for the 3D histogram
        self._input_shape = input_shape
        self._model = VGG16(weights = 'imagenet', input_shape = (input_shape[0], input_shape[1], input_shape[2]), pooling = 'max', include_top = False)

    def describe(self, img):
#        print self._model
        img = image.img_to_array(img)
        img = np.expand_dims(img, axis=0)
        img = preprocess_input(img)
        feat = self._model.predict(img)
        norm_feat = feat[0]/LA.norm(feat[0])
        return norm_feat.reshape(1,-1)
    
    def extract_feature(self, img_path):
#        img = image.load_img(img_path, target_size=(self._input_shape[0], self._input_shape[1]))
        try:
            file = cStringIO.StringIO(urllib.urlopen(img_path).read())
            img = Image.open(file)
        except:
            return None
 
        # fix the rotation issue in iOS (version 5.0 or before)
        try:
            exif = None
            if hasattr(img, '_getexif'):
                exif = img._getexif()

            if exif is not None:
                orientation = 0x0112
                device_pos = exif[orientation]

                if device_pos == 3 :
                    img = img.rotate(180)

                if device_pos == 6 :
                    img = img.rotate(270)

                if device_pos == 8 :
                    img = img.rotate(90)
        except:
            pass

        img = img.resize((self._input_shape[0], self._input_shape[1]))
        return self.describe(img)
