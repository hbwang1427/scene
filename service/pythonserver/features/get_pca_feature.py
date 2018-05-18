# import the necessary packages
import os
import argparse
from feature_io import *
from numpy import linalg as LA
from sklearn.decomposition import PCA
 
# construct the argument parser and parse the arguments
ap = argparse.ArgumentParser()
ap.add_argument("-f", "--feature", required = True,
	help = "which feature?")
ap.add_argument("-d", "--dim", required = True, type = int,
	help = "pca dimensionality")
args = vars(ap.parse_args())

print "loading features..."
feat_dir = os.path.join(os.path.dirname(__file__), args['feature'])
feat_name, feat = load_feature(os.path.join(feat_dir, 'index.csv'))
#print feat
num_feat = args['dim']
pca = PCA(num_feat)
pca.fit(feat)
pca_feat = pca.transform(feat)
# normalize
pca_feat = pca_feat / np.tile(LA.norm(pca_feat,axis=1), (num_feat,1)).transpose()
output_index_name = os.path.join(feat_dir, 'pca_index.csv')
save_feature(output_index_name, feat_name, pca_feat)
output_model_name = os.path.join(feat_dir, 'model.pca')
save_model(output_model_name, pca)
