import numpy as np
import csv
#from sklearn.decomposition import PCA

def load_features(filename):
    feat = []
    feat_fname = []
    with open(filename) as f:
	# initialize the CSV reader
	reader = csv.reader(f)
	for row in reader:
	    feat.append([float(x) for x in row[1:]])
            feat_fname.append(row[0])
    return  feat_fname, np.array(feat)


if __name__ == '__main__':
    fname, feat = load_features('../index_all.csv')
    print fname
    print feat
    #pca = PCA(256)
    #pca.fit(feat)
    #print pca.transform(feat)
