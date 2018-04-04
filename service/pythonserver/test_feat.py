import cv2
import numpy as np
import time
import os

import _init_path
from features.rmacdescriptor import RMACDescriptor
from imagesearch.searcher import Searcher, cosine_distance
from features.feature_io import load_model

_ONE_DAY_IN_SECONDS = 60 * 60 * 24
VISUAL_FEATURE_DIR = '../../../../../../../work/data/visual_features'

rmac_model_proto = './features/deploy_resnet101_normpython.prototxt'
rmac_model_weights = './features/rmac_model.caffemodel'
print rmac_model_proto
fDesc = RMACDescriptor(rmac_model_proto, rmac_model_weights)
index_feat_path = os.path.join(VISUAL_FEATURE_DIR, 'finearts-annotated', 'rmac_index.csv')
searcher = Searcher(index_feat_path, cosine_distance)
    
img_file = '/home/pangolins/go/src/github.com/aitour/scene/web/assets/Images/museum/FineArtsQueryHighResolution/painting/587_small.JPG'
t0 = time.time()
features = fDesc.extract_feature(img_file)
print 'feature extraction took %.3f'% (time.time() - t0)
