import numpy as np
import csv
#from sklearn.decomposition import PCA
import pickle

def load_feature(filename):
    feat = []
    feat_fname = []
    with open(filename) as f:
	# initialize the CSV reader
	reader = csv.reader(f)
	for row in reader:
	    feat.append([float(x) for x in row[1:]])
            feat_fname.append(row[0])
    return  feat_fname, np.array(feat).astype(float)

def save_feature(output_fname, img_fnames, feat):
    assert(len(img_fnames) == feat.shape[0])
    output = open(output_fname, "w")
 
    for k in xrange(feat.shape[0]):
        features = [str(f) for f in feat[k,:]]
        output.write("%s,%s\n" % (img_fnames[k], ",".join(features)))
    output.close()
    
def load_model(filename):
    return pickle.load(open(filename, 'rb'))

def save_model(filename, model):
    pickle.dump(model, open(filename, 'wb'))

if __name__ == '__main__':
    fname, feat = load_features('../index_all.csv')
    print fname
    print feat
    #pca = PCA(256)
    #pca.fit(feat)
    #print pca.transform(feat)
