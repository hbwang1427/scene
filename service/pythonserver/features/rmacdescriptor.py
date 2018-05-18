# import the necessary packages
import numpy as np
from numpy import linalg as LA
import cv2
import caffe
import urllib, cStringIO
from PIL import Image

class ImageHelper:
    def __init__(self, S, L, means):
        self.S = S
        self.L = L
        self.means = means

   # def prepare_image_and_grid_regions_for_network(self, fname, roi=None):
   #     # Extract image, resize at desired size, and extract roi region if
   #     # available. Then compute the rmac grid in the net format: ID X Y W H
   #     I, im_resized = self.load_and_prepare_image(fname, roi)
   #     if self.L == 0:
   #         # Encode query in mac format instead of rmac, so only one region
   #         # Regions are in ID X Y W H format
   #         R = np.zeros((1, 5), dtype=np.float32)
   #         R[0, 3] = im_resized.shape[1] - 1
   #         R[0, 4] = im_resized.shape[0] - 1
   #     else:
   #         # Get the region coordinates and feed them to the network.
   #         all_regions = []
   #         all_regions.append(self.get_rmac_region_coordinates(im_resized.shape[0], im_resized.shape[1], self.L))
   #         R = self.pack_regions_for_network(all_regions)
   #     return I, R
    
    def prepare_image_and_grid_regions_for_network(self, img, roi=None):
        # Extract image, resize at desired size, and extract roi region if
        # available. Then compute the rmac grid in the net format: ID X Y W H
        I, im_resized = self.prepare_image(img, roi)
        if self.L == 0:
            # Encode query in mac format instead of rmac, so only one region
            # Regions are in ID X Y W H format
            R = np.zeros((1, 5), dtype=np.float32)
            R[0, 3] = im_resized.shape[1] - 1
            R[0, 4] = im_resized.shape[0] - 1
        else:
            # Get the region coordinates and feed them to the network.
            all_regions = []
            all_regions.append(self.get_rmac_region_coordinates(im_resized.shape[0], im_resized.shape[1], self.L))
            R = self.pack_regions_for_network(all_regions)
        return I, R

    def get_rmac_features(self, I, R, net):
        net.blobs['data'].reshape(I.shape[0], 3, int(I.shape[2]), int(I.shape[3]))
        net.blobs['data'].data[:] = I
        net.blobs['rois'].reshape(R.shape[0], R.shape[1])
        net.blobs['rois'].data[:] = R.astype(np.float32)
        net.forward(end='rmac/normalized')
        return np.squeeze(net.blobs['rmac/normalized'].data)

   # def load_and_prepare_image(self, fname, roi=None):
   #     # Read image, get aspect ratio, and resize such as the largest side equals S
   #     im = cv2.imread(fname)
   #     im_size_hw = np.array(im.shape[0:2])
   #     ratio = float(self.S)/np.max(im_size_hw)
   #     new_size = tuple(np.round(im_size_hw * ratio).astype(np.int32))
   #     im_resized = cv2.resize(im, (new_size[1], new_size[0]))
   #     # If there is a roi, adapt the roi to the new size and crop. Do not rescale
   #     # the image once again
   #     if roi is not None:
   #         roi = np.round(roi * ratio).astype(np.int32)
   #         im_resized = im_resized[roi[1]:roi[3], roi[0]:roi[2], :]
   #     # Transpose for network and subtract mean
   #     I = im_resized.transpose(2, 0, 1) - self.means
   #     return I, im_resized
    
    def prepare_image(self, im, roi=None):
        # Read image, get aspect ratio, and resize such as the largest side equals S
        im_size_hw = np.array(im.shape[0:2])
        ratio = float(self.S)/np.max(im_size_hw)
        new_size = tuple(np.round(im_size_hw * ratio).astype(np.int32))
        im_resized = cv2.resize(im, (new_size[1], new_size[0]))
        # If there is a roi, adapt the roi to the new size and crop. Do not rescale
        # the image once again
        if roi is not None:
            roi = np.round(roi * ratio).astype(np.int32)
            im_resized = im_resized[roi[1]:roi[3], roi[0]:roi[2], :]
        # Transpose for network and subtract mean
        I = im_resized.transpose(2, 0, 1) - self.means
        return I, im_resized


    def pack_regions_for_network(self, all_regions):
        n_regs = np.sum([len(e) for e in all_regions])
        R = np.zeros((n_regs, 5), dtype=np.float32)
        cnt = 0
        # There should be a check of overflow...
        for i, r in enumerate(all_regions):
            try:
                R[cnt:cnt + r.shape[0], 0] = i
                R[cnt:cnt + r.shape[0], 1:] = r
                cnt += r.shape[0]
            except:
                continue
        assert cnt == n_regs
        R = R[:n_regs]
        # regs where in xywh format. R is in xyxy format, where the last coordinate is included. Therefore...
        R[:n_regs, 3] = R[:n_regs, 1] + R[:n_regs, 3] - 1
        R[:n_regs, 4] = R[:n_regs, 2] + R[:n_regs, 4] - 1
        return R

    def get_rmac_region_coordinates(self, H, W, L):
        # Almost verbatim from Tolias et al Matlab implementation.
        # Could be heavily pythonized, but really not worth it...
        # Desired overlap of neighboring regions
        ovr = 0.4
        # Possible regions for the long dimension
        steps = np.array((2, 3, 4, 5, 6, 7), dtype=np.float32)
        w = np.minimum(H, W)

        b = (np.maximum(H, W) - w) / (steps - 1)
        # steps(idx) regions for long dimension. The +1 comes from Matlab
        # 1-indexing...
        idx = np.argmin(np.abs(((w**2 - w * b) / w**2) - ovr)) + 1

        # Region overplus per dimension
        Wd = 0
        Hd = 0
        if H < W:
            Wd = idx
        elif H > W:
            Hd = idx

        regions_xywh = []
        for l in range(1, L+1):
            wl = np.floor(2 * w / (l + 1))
            wl2 = np.floor(wl / 2 - 1)
            # Center coordinates
            if l + Wd - 1 > 0:
                b = (W - wl) / (l + Wd - 1)
            else:
                b = 0
            cenW = np.floor(wl2 + b * np.arange(l - 1 + Wd + 1)) - wl2
            # Center coordinates
            if l + Hd - 1 > 0:
                b = (H - wl) / (l + Hd - 1)
            else:
                b = 0
            cenH = np.floor(wl2 + b * np.arange(l - 1 + Hd + 1)) - wl2

            for i_ in cenH:
                for j_ in cenW:
                    regions_xywh.append([j_, i_, wl, wl])

        # Round the regions. Careful with the borders!
        for i in range(len(regions_xywh)):
            for j in range(4):
                regions_xywh[i][j] = int(round(regions_xywh[i][j]))
            if regions_xywh[i][0] + regions_xywh[i][2] > W:
                regions_xywh[i][0] -= ((regions_xywh[i][0] + regions_xywh[i][2]) - W)
            if regions_xywh[i][1] + regions_xywh[i][3] > H:
                regions_xywh[i][1] -= ((regions_xywh[i][1] + regions_xywh[i][3]) - H)
        return np.array(regions_xywh).astype(np.float32)

class RMACDescriptor:
    def __init__(self, model_proto, model_weights, input_s=800, input_l=2):
        means = np.array([103.93900299,  116.77899933,  123.68000031], dtype=np.float32)[None, :, None, None]

        # Configure caffe and load the network
        self._input_s = input_s
        self._input_l = input_l
        caffe.set_mode_gpu()
        caffe.set_device(0)
        #print model_proto, model_weights
        self._model = caffe.Net(model_proto, model_weights, caffe.TEST)
        self._image_helper = ImageHelper(input_s, input_l, means)

    def describe(self, img):
#        print self._model
        I, R = self._image_helper.prepare_image_and_grid_regions_for_network(img, roi=None)
        features = self._image_helper.get_rmac_features(I, R, self._model)
        # normalize
        features = features / np.sqrt(np.dot(features, features))
        return features
    
    def extract_feature(self, img_path):
        try:
            file = cStringIO.StringIO(urllib.urlopen(img_path).read())
            img = Image.open(file)
        except:
            print 'load image failed ......!'
            return None

        # fix the rotation issue in iOS (version 5.0 or before)
        try:
            exif = None
        #    if hasattr(img, '_getexif'):
            if True:
                exif = img._getexif()

            if exif is not None:
                orientation = 0x0112
                device_pos = exif[orientation]
                print '============================='
                print "==== device position " + str(device_pos) + "===="
                print '============================='
                if device_pos == 3 :
                    img = img.rotate(180)

                if device_pos == 6 :
                    img = img.rotate(270)

                if device_pos == 8 :
                    img = img.rotate(90)
        except:
            pass

        # convert it to opencv image
        open_cv_img = np.array(img) 
        # Convert RGB to BGR 
        open_cv_img = open_cv_img[:, :, ::-1].copy() 
        return self.describe(open_cv_img)
