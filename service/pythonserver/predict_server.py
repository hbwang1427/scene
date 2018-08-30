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
import os
import time
import grpc
import toml
import logging


import rpc_pb2
import rpc_pb2_grpc

import _init_path
from features.rmacdescriptor import RMACDescriptor
from imagesearch.searcher import Searcher, cosine_distance
from features.feature_io import load_model

_ONE_DAY_IN_SECONDS = 60 * 60 * 24
VISUAL_FEATURE_DIR = '../../../../../../../work/data/visual_features'

webhost = ""


#PredictServicer implements rpc_pb2_grpc.PredictServicer
class PredictServicer(rpc_pb2_grpc.PredictServicer):
  def __init__(self):
#      logging.debug("create PredictServer. loading model")
#      rmac_model_proto = './features/deploy_resnet101_normpython.prototxt'
#      rmac_model_weights = './features/rmac_model.caffemodel'
#      print rmac_model_proto
#      self._fDesc = RMACDescriptor(rmac_model_proto, rmac_model_weights)
      self._fDesc = None
      index_feat_path = os.path.join(VISUAL_FEATURE_DIR, 'finearts', 'rmac_index.csv')
      self._searcher = Searcher(index_feat_path, cosine_distance)
      logging.debug("model load successfully")
 
  def getDescriptor(self):
    if self._fDesc is None: 
      logging.debug("create PredictServer. loading model")
      rmac_model_proto = './features/deploy_resnet101_normpython.prototxt'
      rmac_model_weights = './features/rmac_model.caffemodel'
      print rmac_model_proto
      self._fDesc = RMACDescriptor(rmac_model_proto, rmac_model_weights)
    return self._fDesc

  def PredictPhoto(self, request, context):
    #decode image from request
    img_array = np.asarray(bytearray(request.data), dtype=np.uint8)
    img = cv2.imdecode(img_array, cv2.IMREAD_COLOR)
  
    t0 = time.time()
    fdesc = self.getDescriptor()
    features = fdesc.describe(img)
    print ('feature extraction took {:.3f}s').format(time.time() - t0)
    t0 = time.time()
    if features is not None:
        results = self._searcher.search(features, 5)
    print ('Matching took {:.3f}s').format(time.time() - t0)

    #print results
    #process image
    photo_info = "image shape:%s, size:%d, dtype:%s" % (img.shape, img.size, img.dtype)
  #  print("predict photo request received. " + photo_info)

    #write response
    return rpc_pb2.PhotoPredictResponse(results=[
      rpc_pb2.PhotoPredictResponse.Result(text=photo_info, image_url= "%s/assets/%s" % (webhost, results[0][0]), audio_url="/assets/audio/sample_0.4mb.mp3"),
      rpc_pb2.PhotoPredictResponse.Result(text=photo_info, image_url="%s/assets/%s" % (webhost, results[1][0])),
      rpc_pb2.PhotoPredictResponse.Result(text=photo_info, image_url="%s/assets/%s" % (webhost, results[2][0]))
    ])


def serve():
  conf = toml.load("../service.toml")
  webhost = conf['web']['host']
  bindaddr = conf['grpc']['bind']
  with open("../" + conf['grpc']['key']) as f:
    #private_key = bytes(f.read(), "ascii")
    private_key = bytes(f.read())
  with open("../" + conf['grpc']['cert']) as f:
    #certificate_chain = bytes(f.read(), "ascii")
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
  serve()
