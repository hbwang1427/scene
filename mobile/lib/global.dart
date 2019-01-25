library aitour.globals;

import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:connectivity/connectivity.dart';
import 'package:http/http.dart' as http;
import 'package:synchronized/synchronized.dart';


import './model.dart';


ConnectivityResult gConnectivityResult = ConnectivityResult.none;
bool gIsLoggedIn = false;
bool gIsModelChecked = false;
List<ModelInfo> gModelInfos = new List<ModelInfo>();
Future<bool> gModelChecked;
String gAppDocDir;
String gModelListFileBody;
Lock modelDownloadLock = new Lock();
Locale gMyLocale = null;
//const host = 'http://192.168.0.220:8081';
const host = 'http://pangolinai.net';

Future<List<ModelInfo>> getModelInfo({String listFileBody: ""}) async {
  if (gModelInfos.length > 0) {
    return gModelInfos;
  }

  gModelInfos = await modelDownloadLock.synchronized(() async {
    if (listFileBody == null || listFileBody.length == 0) {
      listFileBody = await getModelListFile();
    }

    var array = json.decode(listFileBody);
    var models = array['models'] as List;
    if (models != null) {
      return models.map((m) => ModelInfo.fromJson(m)).toList();
    }
    return new List<ModelInfo>();
  });

  return gModelInfos;
}

Future<String> getModelListFile() async {
  if (gModelListFileBody != null && gModelListFileBody.length > 0) {
    return gModelListFileBody;
  }

  var response = await http.get("$host/model/list");
  if (response.statusCode == 200) {
    gModelListFileBody = response.body;
  }

  return gModelListFileBody;
}
