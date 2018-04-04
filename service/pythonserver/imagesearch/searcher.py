# import the necessary packages
import numpy as np
import csv
from operator import itemgetter
#access modules under the same parent directory
from os import sys, path
sys.path.append(path.dirname(path.dirname(path.abspath(__file__))))
from features.feature_io import load_feature
from sklearn.metrics.pairwise import cosine_similarity
from numpy import linalg as LA

def chi2_similarity(histA, histB, eps = 1e-10):
    # compute the chi-squared distance
    d = 0.5 * np.sum([((a - b) ** 2) / (a + b + eps)
        	for (a, b) in zip(histA, histB)])

    # return the chi-squared distance
    return d

def chi2_distance(qFeat, indexFeat):
    # compute the chi-squared distance
    d = [chi2_similarity(feat, qFeat) for feat in indexFeat]
    return d

def cosine_distance(qFeat, indexFeat):
    #normalize query features
    qFeat = qFeat / LA.norm(qFeat)

    d = 1.0 - cosine_similarity(qFeat, indexFeat).flatten()
    #d = 1.0 - np.dot(qFeat, indexFeat)
    return d.tolist()

class Searcher:
	def __init__(self, indexPath, distFunc='chi2_distance'):
		# store our index path
		self._indexPath = indexPath
                self._indexFeat = None
                self._indexFiles = None
                self._distFunc = distFunc
       
    
	def search(self, queryFeatures, limit = 10):
		# initialize our dictionary of results
		results = {}

		# open the index file for reading
                if self._indexFeat is None:
                    self._indexFiles, self._indexFeat = load_feature(self._indexPath)

                d = self._distFunc(queryFeatures, self._indexFeat)

    #            print d
		# sort our results, so that the smaller distances (i.e. the
		# more relevant images are at the front of the list)
		results = sorted(zip(self._indexFiles, d), key=itemgetter(1))

		return results[0:min(limit,len(results))]

#        def chi2_distance(self, histA, histB, eps = 1e-10):
#	    # compute the chi-squared distance
#	    d = 0.5 * np.sum([((a - b) ** 2) / (a + b + eps)
#	        	for (a, b) in zip(histA, histB)])

#	    # return the chi-squared distance
#	    return d


if __name__ == '__main__':
    import numpy as np
    s = Searcher('index.csv')

    from features.colordescriptor import ColorDescriptor
    cd = ColorDescriptor((8, 12, 3))

    from skimage import io
    imagePath='/home/pangolins/work/flask-image-search/app/static/test.jpg'
    image = io.imread(imagePath)
    features = cd.describe(image)
    print s.search(features)
