#!/usr/bin/python

'''
translate from en.json to mulitple languages with google translator

pip3 install googletrans
'''

# coding: utf-8

# In[1]:


import os
import json
import copy
from googletrans import Translator


#os.environ['HTTP_PROXY']="http://127.0.0.1:8089"
#os.environ['HTTPS_PROXY']="http://127.0.0.1:8089"

translator = Translator()
def transMap(lang, m):
    for k in m:
        print(k)
        if isinstance(m[k], str):
            tr = translator.translate(m[k], src="en", dest=lang)
            m[k] = tr.text
        elif isinstance(m[k], dict):
            transMap(lang, m[k])
            


# In[2]:


fp = open("en.json")
en = json.loads(fp.read())
trlanguages = ["zh-cn", "zh-tw", "fr", "es", "ja", "ko", "ar", "ru"]
trfiles = ["zh-Hans", "zh-Hant", "fr", "es", "ja", "ko", "ar", "ru"]


for idx, trlan in enumerate(trlanguages):
        tr2 = copy.deepcopy(en)
        transMap(trlan, tr2)
        print("trans complete:", tr2)
        v = json.dumps(tr2, ensure_ascii=False)
        fp = open(trfiles[idx]+".json", mode='w', encoding="utf-8")
        fp.write(v)
        fp.close()
        

