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
VISUAL_FEATURE_DIR = '../../../../../../../work/data/visual_features'
WEB_HOST = "http://216.15.112.63:8081"
#WEB_HOST = "http://146.115.70.196:8081"

#PredictServicer implements rpc_pb2_grpc.PredictServicer
class PredictServicer(rpc_pb2_grpc.PredictServicer):
  def __init__(self):
      rmac_model_proto = './features/deploy_resnet101_normpython.prototxt'
      rmac_model_weights = './features/rmac_model.caffemodel'
 #     print rmac_model_proto
      logging.debug("loading model")
      self._fDesc = RMACDescriptor(rmac_model_proto, rmac_model_weights)
      logging.debug("model load successfully")

      index_feat_path = os.path.join(VISUAL_FEATURE_DIR, 'finearts-highlights', 'rmac_index.csv')
      self._searcher = Searcher(index_feat_path, cosine_distance)
 
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

    features = self._fDesc.describe(img)
    logging.debug('feature extraction took %.3f'% (time.time() - t0))
    t0 = time.time()
    if features is not None:
        results = self._searcher.search(features, 5)
        logging.debug('Matching took %.3f' % (time.time() - t0))
        #print results
        response = [rpc_pb2.PhotoPredictResponse.Result(text="score: %6.2f" % (res[1]), \
                    image_url= "%s/assets/%s" % (WEB_HOST, res[0]), \
                    audio_url="%s/assets/%s.mp3" % (WEB_HOST, res[0].split('_small.')[0])) \
                    for res in results]
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
