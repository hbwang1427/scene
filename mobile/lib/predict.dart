import 'dart:async';
import 'dart:io';
import 'dart:convert';

import 'package:camera/camera.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;

import './artlist.dart';
import './model.dart';
import './global.dart' as globals;

class ArtPredictPage extends StatefulWidget {
  @override
  _ArtPredictPageState createState() {
    return _ArtPredictPageState();
  }
}

/// Returns a suitable camera icon for [direction].
IconData getCameraLensIcon(CameraLensDirection direction) {
  switch (direction) {
    case CameraLensDirection.back:
      return Icons.camera_rear;
    case CameraLensDirection.front:
      return Icons.camera_front;
    case CameraLensDirection.external:
      return Icons.camera;
  }
  throw ArgumentError('Unknown lens direction');
}

void logError(String code, String message) =>
    print('Error: $code\nError Message: $message');

class _ArtPredictPageState extends State<ArtPredictPage> {
  bool camerasInitialized = false;
  bool tfliteModelLoaded = false;
  bool isPredicting = false;
  CameraController controller;
  List<CameraDescription> cameras;
  MethodChannel tflite;

  String imagePath;

  final GlobalKey<ScaffoldState> _scaffoldKey = GlobalKey<ScaffoldState>();

  void asyncInit() async {
    var models = await globals.getModelInfo();
    //load tflite model if nessary
    if (models.length > 0) {
      //wait for the model to be downloaded
      await models[0].localFile;
      tflite = const MethodChannel('net.pangolinai.mobile/museum_tflite');
      try {
        final String result = await tflite.invokeMethod('loadModel',
            {'model': '${globals.gAppDocDir}/models/${models[0].name}'});
        if (result == "success") {
          tfliteModelLoaded = true;
        }
      } on PlatformException catch (e) {
        print('load tflite model error:${e.message}');
      }
    }

    cameras = await availableCameras();
    for (var cam in cameras) {
      if (cam.lensDirection == CameraLensDirection.back) {
        if (controller != null) {
          await controller.dispose();
        }
        controller = CameraController(cam, ResolutionPreset.high);

        // If the controller is updated then update the UI.
        controller.addListener(() {
          if (mounted) setState(() {});
          if (controller.value.hasError) {
            showInSnackBar('Camera error ${controller.value.errorDescription}');
          }
        });

        try {
          await controller.initialize();
        } on CameraException catch (e) {
          _showCameraException(e);
        }

        if (mounted) {
          setState(() {
            camerasInitialized = true;
          });
        }
      }
    }
  }

  @override
  void initState() {
    // TODO: implement initState
    super.initState();
    asyncInit();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      key: _scaffoldKey,
      // appBar: AppBar(
      //   title: const Text('Camera example'),
      // ),
      body: Column(
        children: <Widget>[
          Expanded(
            child: Container(
              child: Padding(
                padding: const EdgeInsets.all(1.0),
                child: Center(
                  child: camerasInitialized
                      ? _cameraPreviewWidget()
                      : CircularProgressIndicator(
                          semanticsLabel: "waiting for camera..."),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  /// Display the preview from the camera (or a message if the preview is not available).
  Widget _cameraPreviewWidget() {
    if (controller == null || !controller.value.isInitialized) {
      return const Text(
        'Tap a camera',
        style: TextStyle(
          color: Colors.white,
          fontSize: 24.0,
          fontWeight: FontWeight.w900,
        ),
      );
    } else {
      return AspectRatio(
        aspectRatio: controller.value.aspectRatio,
        child: GestureDetector(
            onTap: onCameraPreviewPressed, child: CameraPreview(controller)),
      );
    }
  }

  void onCameraPreviewPressed() {
    if (!tfliteModelLoaded) {
      showInSnackBar('Error: model was not loaded');
      return;
    }

    if (isPredicting) {
      showInSnackBar('Please wait...');
      return;
    }

    takePicture().then((String filePath) async {
      if (mounted) {
        isPredicting = true;
        List<int> feature;
        try {
          feature =
              await tflite.invokeMethod('runModelOnImage', <String, dynamic>{
            'path': filePath,
            'inputSize': 224, // wanted input size, defaults to 224
            'numChannels': 3, // wanted input channels, defaults to 3
            'imageMean': 127.5, // defaults to 117.0
            'imageStd': 127.5, // defaults to 1.0
            'numResults': 6, // defaults to 5
            'threshold': 0.05, // defaults to 0.1
            'numThreads': 1, // defaults to 1
          });
        } on PlatformException catch (e) {
          print("$e");
        }
        //print('$feature');
        var response = await http.post('${globals.host}/predict2?k=5',
            headers: {'Content-Type': 'application/octet-stream'},
            body: feature);
        if (response.statusCode == 200) {
          print("${response.body}");
          var results = json.decode(response.body)['results'] as List;
          List<ArtPredict> artPredicts =
              results.map((item) => ArtPredict.fromJson(item)).toList();
          Navigator.of(context).push(MaterialPageRoute(
              builder: (context) => ArtListView(predicts: artPredicts)));
        } else {
          showInSnackBar('predict error: ${response.statusCode}');
        }
        isPredicting = false;

        // setState(() {
        //   imagePath = filePath;
        // });
        // if (filePath != null) showInSnackBar('Picture saved to $filePath');
      }
    });
  }

  String timestamp() => DateTime.now().millisecondsSinceEpoch.toString();

  void showInSnackBar(String message) {
    _scaffoldKey.currentState.showSnackBar(SnackBar(content: Text(message)));
  }

  void onNewCameraSelected(CameraDescription cameraDescription) async {
    if (controller != null) {
      await controller.dispose();
    }
    controller = CameraController(cameraDescription, ResolutionPreset.high);

    // If the controller is updated then update the UI.
    controller.addListener(() {
      if (mounted) setState(() {});
      if (controller.value.hasError) {
        showInSnackBar('Camera error ${controller.value.errorDescription}');
      }
    });

    try {
      await controller.initialize();
    } on CameraException catch (e) {
      _showCameraException(e);
    }

    if (mounted) {
      setState(() {});
    }
  }

  Future<String> takePicture() async {
    if (!controller.value.isInitialized) {
      showInSnackBar('Error: select a camera first.');
      return null;
    }

    final String dirPath = '${globals.gAppDocDir}/Pictures/flutter_test';
    await Directory(dirPath).create(recursive: true);
    final String filePath = '$dirPath/${timestamp()}.jpg';

    if (controller.value.isTakingPicture) {
      // A capture is already pending, do nothing.
      return null;
    }

    try {
      await controller.takePicture(filePath);
    } on CameraException catch (e) {
      _showCameraException(e);
      return null;
    }
    return filePath;
  }

  void _showCameraException(CameraException e) {
    logError(e.code, e.description);
    showInSnackBar('Error: ${e.code}\n${e.description}');
  }
}
