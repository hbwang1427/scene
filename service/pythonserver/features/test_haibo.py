import cv2
import numpy as np
import os

from rmacdescriptor import RMACDescriptor

VISUAL_FEATURE_DIR = '../../../../../../../work/data/visual_features'

#PredictServicer implements rpc_pb2_grpc.PredictServicer
class PredictServicer:
  def __init__(self):
      rmac_model_proto = 'deploy_resnet101_normpython.prototxt'
      rmac_model_weights = 'rmac_model.caffemodel'
      self._fDesc = RMACDescriptor(rmac_model_proto, rmac_model_weights)
 
if __name__ == '__main__':
   p = PredictServicer()
