#!/bin/bash

export PYTHONPATH=/home/pangolins/caffe-fast-rcnn/python:$PYTHONPATH
python /home/pangolins/go/src/github.com/aitour/scene/service/pythonserver/predict_server_mod.py
