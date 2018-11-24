'''
install grpc:  
	python -m pip install grpcio

install grpc-tools:  
	python -m pip install grpcio-tools

to recompile proto files:  
	python -m grpc_tools.protoc -I../../serverpb --python_out=. --grpc_python_out=. ../../serverpb/rpc.proto

details refer to:
	https://grpc.io/docs/quickstart/python.html#update-and-run-the-application
'''

from concurrent import futures
import cv2
import numpy as np
import time
import grpc
import toml
import os
import logging
import caffe

import rpc_pb2
import rpc_pb2_grpc

import _init_path
from features.rmacdescriptor import RMACDescriptor
from imagesearch.searcher import Searcher, cosine_distance
from features.feature_io import load_model

_ONE_DAY_IN_SECONDS = 60 * 60 * 24
MUSEUM_BASE_DIR = '../../web/assets/Images/Museum'
MUSEUM_DATA_DIR= {'fine-arts':'Boston-FineArts', 'met':'MET', 'boston-ica':'Boston-ICA'}

#WEB_HOST = "http://aitour.ml:8081"
#WEB_HOST = "http://216.15.112.63:8081"
#WEB_HOST = "http://146.115.70.196:8081"
WEB_HOST = "http://pangolinai.net"

#PredictServicer implements rpc_pb2_grpc.PredictServicer
class PredictServicer(rpc_pb2_grpc.PredictServicer):
  def __init__(self):
      rmac_model_proto = './features/deploy_resnet101_normpython.prototxt'
      rmac_model_weights = './features/rmac_model.caffemodel'
 #     print rmac_model_proto
      logging.debug("loading model")
      self._fDesc = RMACDescriptor(rmac_model_proto, rmac_model_weights)
      logging.debug("model loaded successfully")

      #index_feat_path = os.path.join(MUSEUM_DATA_DIR, MUSEUM, 'visual_features', 'rmac_index.csv')
      #self._searcher = Searcher(index_feat_path, cosine_distance)
      self._searcher = {}
      logging.debug("features loaded successfully")
 
  def _get_searcher(self, museum):
      if museum not in self._searcher:
          index_feat_path = os.path.join(MUSEUM_BASE_DIR, MUSEUM_DATA_DIR[museum], 'visual_features', 'rmac_index.csv')
          self._searcher[museum] = Searcher(index_feat_path, cosine_distance)

      return self._searcher[museum]  

  def PredictPhoto(self, request, context):
    logging.info("PredictPhoto request received")
    #decode image from request
    img_array = np.asarray(bytearray(request.data), dtype=np.uint8)
    img = cv2.imdecode(img_array, cv2.IMREAD_COLOR)
    height, width = img.shape[:2]
    if height > 400 or width > 400:
	if height > width: 
		sw, sh = 400 * width/height, 400 
	else:
		sw, sh = 400, 400 * height/width
        img = cv2.resize(img, (sw, sh), interpolation = cv2.INTER_CUBIC)
#        img = img[:, :, ::-1].copy() 
	logging.debug("img width=%d, height=%d, resize to width=%d, height=%d" % (width, height, sw, sh))
    
    logging.debug("decode request image ok")

    # postpone model loadint untile seeing the first image
    photo_info = ''
    response = []
    t0 = time.time()
    # a possible BUG in caffe. reattach gpu
    caffe.set_mode_gpu()
    caffe.set_device(0)

    # which site
    museum = request.site
    # what language
    language = request.language
    print 'Museum: %s Language %s' % (museum, language)
    

    features = self._fDesc.describe(img)
    logging.debug('feature extraction took %.3f'% (time.time() - t0))
    t0 = time.time()
    if features is not None:
        results = self._get_searcher(museum).search(features, 5)
        logging.debug('Matching took %.3f' % (time.time() - t0))
        #print results
        response = []
        for res in results:
            img_path, file_name = os.path.split(res[0])
            [img_name, img_ext] = file_name.split('.')
            img_url = "%s/assets/Images/%s/%s" % (WEB_HOST, img_path, img_name + '_small.' + img_ext)
            lang_suffix = '_en'
            if language in ['zh', 'zh-hans', 'zh-Hans','zh_hans', 'zh_Hans', 'zh_hant', 'zh_Hant']:
                lang_suffix = '_zh' 
            audio_path = os.path.split(img_path)[0] + '/Audio'
            audio_url = "%s/assets/Images/%s/%s.mp3" % (WEB_HOST, audio_path, img_name + lang_suffix)
            desc_file = os.path.join('../../web/assets/Images', os.path.split(img_path)[0], 'Description', img_name+lang_suffix+'.txt')
            #print desc_file
            text = ''
            if os.path.isfile(desc_file):
                with open(desc_file, 'rt') as fid:
                    text = fid.read()
            #print text
            response.append(rpc_pb2.PhotoPredictResponse.Result(text=text, image_url=img_url, audio_url=audio_url))

        logging.debug(response)
#        for res in results:
#            if photo_info == '': # first one
#                photo_info = res[0]
#            else:
#                photo_info = photo_info + ';' + res[0]
    #logging.debug("predict ok, photo_info:%s" % (photo_info))

    #process image
    #photo_info = "image shape:%s, size:%d, dtype:%s" % (img.shape, img.size, img.dtype)
    #print("predict photo request received. " + photo_info)

    #write response
    #two-way return: first return N matched images; 
    #only after user clicks the right image, return the corresponding text and audio 
    #return rpc_pb2.PhotoPredictResponse(text=photo_info, audio_url="/assets/audio/sample_0.4mb.mp3")
    #return rpc_pb2.PhotoPredictResponse(results=[
    #  rpc_pb2.PhotoPredictResponse.Result(text=photo_info, image_url= "%s/assets/%s" % (webhost, results[0][0]), audio_url="/assets/audio/sample_0.4mb.mp3"),
     # rpc_pb2.PhotoPredictResponse.Result(text=photo_info, image_url="%s/assets/%s" % (webhost, results[1][0])),
     # rpc_pb2.PhotoPredictResponse.Result(text=photo_info, image_url="%s/assets/%s" % (webhost, results[2][0]))
   # ]
    return rpc_pb2.PhotoPredictResponse(results=response)


def serve():
  conf = toml.load("../service.toml")
  bindaddr = conf['grpc']['bind']
  with open("../" + conf['grpc']['key']) as f:
    #private_key = bytes(f.read(), "ascii")
    private_key = bytes(f.read())
  with open("../" + conf['grpc']['cert']) as f:
    #certificate_chain = bytes(f.read(), 'ascii')
    certificate_chain = bytes(f.read())
  server_credentials = grpc.ssl_server_credentials(
      ((private_key, certificate_chain,),))

  server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
  rpc_pb2_grpc.add_PredictServicer_to_server(PredictServicer(), server)
  #server.add_insecure_port('[::]:50051')
  server.add_secure_port(bindaddr, server_credentials)
  server.start()
  print("service runing on " + bindaddr)
  try:
    while True:
      time.sleep(_ONE_DAY_IN_SECONDS)
  except KeyboardInterrupt:
    server.stop(0)

if __name__ == '__main__':
    #config log 
    logging.basicConfig(format='%(asctime)s %(message)s', level=logging.DEBUG)
    serve()
