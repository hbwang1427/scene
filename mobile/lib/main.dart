import 'dart:async';
import 'dart:io';
import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:flutter/material.dart';
//import 'package:flutter/rendering.dart';
import 'package:connectivity/connectivity.dart';

import 'package:path_provider/path_provider.dart';

import './home.dart';
import './predict.dart';
import './profile.dart';
import './model.dart';
import './global.dart' as globals;

void main() {
  //debugPaintSizeEnabled = true;
  asyncInit();
  runApp(AitourApp());
}

void asyncInit() async {
  globals.gAppDocDir = (await getApplicationDocumentsDirectory()).path;
  await Directory('${globals.gAppDocDir}/models').create(recursive: true);

  var connectivityResult = await (new Connectivity().checkConnectivity());
  if (connectivityResult == ConnectivityResult.wifi) {
    var downloadError = false;

    var modelListFile = await globals.getModelListFile();
    var models = await globals.getModelInfo(listFileBody: modelListFile);
    if (models.length > 0) {
      List<ModelInfo> mlist;
      var f = File('${globals.gAppDocDir}/models/mlist.json');
      if (await f.exists()) {
        mlist = (json.decode(await f.readAsString())['models'] as List)
            .map((m) => ModelInfo.fromJson(m))
            .toList();

        //删除不同步的文件
        for (var mi in mlist) {
          var i = globals.gModelInfos.indexWhere((v) => v.name == mi.name);
          if (i < 0) {
            await File('${globals.gAppDocDir}/models/${mi.name}').delete();
          }
        }
      }

      //如果有必要， 下载模型
      for (var mi in globals.gModelInfos) {
        if (mlist != null &&
            mlist.indexWhere((v) => v.md5Hash == mi.md5Hash) >= 0) {
          continue;
        } else {
          var modelPath = '${globals.gAppDocDir}/models/${mi.name}';
          var f = File(modelPath);
          if (f.existsSync()) {
            var digest = md5.convert(f.readAsBytesSync());
            if ('$digest' == mi.md5Hash) {
              continue;
            }
          }

          //download the file
          f = await mi.localFile;
          if (!f.existsSync()) {
            downloadError = true;
          }
        }
      }
    }

    if (!downloadError) {
      var fw = File('${globals.gAppDocDir}/models/mlist.json')
          .openSync(mode: FileMode.write);
      fw.writeStringSync(modelListFile);
      fw.closeSync();
    }
  }
}

class AitourApp extends StatefulWidget {
  AitourApp({Key key}) : super(key: key);

  @override
  _AitourAppState createState() => _AitourAppState();
}

class _AitourAppState extends State<AitourApp> {
  final Connectivity _connectivity = Connectivity();
  StreamSubscription<ConnectivityResult> _connectivitySubscription;
  int _selectedIndex = 1;
  final _widgetOptions = [
    HomePage(),
    ArtPredictPage(),
    ProfilePage(),
  ];

  @override
  void initState() {
    super.initState();

    _connectivitySubscription =
        _connectivity.onConnectivityChanged.listen((ConnectivityResult result) {
      globals.gConnectivityResult = result;
      if (!globals.gIsModelChecked && result == ConnectivityResult.wifi) {}
      //setState(() => _connectionStatus = result.toString());
    });
  }

  @override
  void dispose() {
    _connectivitySubscription.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      home: Scaffold(
        appBar: AppBar(
          title: Text('Aitour'),
        ),
        body: Center(
          child: _widgetOptions.elementAt(_selectedIndex),
        ),
        bottomNavigationBar: BottomNavigationBar(
          items: <BottomNavigationBarItem>[
            BottomNavigationBarItem(
                icon: Icon(Icons.home), title: Text('Home')),
            BottomNavigationBarItem(
                icon: Icon(Icons.business), title: Text('Tour')),
            BottomNavigationBarItem(
                icon: Icon(Icons.school), title: Text('Profile')),
          ],
          currentIndex: _selectedIndex,
          fixedColor: Colors.deepPurple,
          onTap: _onItemTapped,
        ),
      ),
    );
  }

  void _onItemTapped(int index) {
    setState(() {
      _selectedIndex = index;
    });
  }
}
