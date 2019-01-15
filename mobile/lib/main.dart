import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'widgets/placehold_widget.dart';
import "package:latlngconv/latlngconv.dart";

void main() {
  runApp(MaterialApp(
    home: Scaffold(
      appBar: AppBar(title: const Text('Maps demo')),
      body: MapsDemo(),
    ),
  ));
}

class MapsDemo extends StatefulWidget {
  @override
  State createState() => MapsDemoState();
}

class MapsDemoState extends State<MapsDemo> {
  LatLng _currentPoint = new LatLng(51.5, -0.09);
  bool _outOfChina = true;
  bool _baiduMapBigFont = false;
  MapController _mapController = new MapController();
  
  @override
  void initState() {
    super.initState();
    _mapController = new MapController();
    _initPlatformState();
  }

 // Platform messages are asynchronous, so we initialize in an async method.
  Future<void> _initPlatformState() async {
    Position position;
    // Platform messages may fail, so we use a try/catch PlatformException.
    try {
      final Geolocator geolocator = Geolocator();
        //..forceAndroidLocationManager = true;
      position = await geolocator.getCurrentPosition(
          desiredAccuracy: LocationAccuracy.bestForNavigation);
    } on PlatformException {
      position = null;
    }

    // If the widget was removed from the tree while the asynchronous platform
    // message was in flight, we want to discard the reply rather than calling
    // setState to update our non-existent appearance.
    if (!mounted) {
      return;
    }

    if (position != null) {
      LatLng wgs84 = new LatLng(position.latitude, position.longitude);
      LatLng gcj02 = LatLngConvert(wgs84, LatLngType.WGS84, LatLngType.GCJ02);
      setState(() {
        _outOfChina = OutofChina(wgs84);
        _currentPoint = gcj02;
      });
      //_currentPoint = new LatLng(position.latitude, position.longitude);
      _mapController.move(_currentPoint, _mapController.zoom);
    }
  }

  @override
  Widget build(BuildContext context) {
    return FutureBuilder<GeolocationStatus>(
      future: Geolocator().checkGeolocationPermissionStatus(),
      builder: (BuildContext context, AsyncSnapshot<GeolocationStatus> snapshot) {
        if (!snapshot.hasData) {
            return const Center(child: CircularProgressIndicator());
          }

          if (snapshot.data == GeolocationStatus.disabled) {
            return const PlaceholderWidget('Location services disabled',
                'Enable location services for this App using the device settings.');
          }

          if (snapshot.data == GeolocationStatus.denied) {
            return const PlaceholderWidget('Access to location denied',
                'Allow access to the location services for this App using the device settings.');
          }

          return new FlutterMap(
            mapController: _mapController,
            options: new MapOptions(
              center: new LatLng(31.1695941, 121.3926092),
              zoom: 15.0,
            ),
            layers: [
              new TileLayerOptions(
                // urlTemplate: "https://api.tiles.mapbox.com/v4/"
                //     "{id}/{z}/{x}/{y}@2x.png?access_token={accessToken}",

                // urlTemplate: "https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png",
                // subdomains: ['a', 'b', 'c'],

                urlTemplate: "http://mt1.google.cn/vt/lyrs=m@207000000&hl=zh-CN&gl=CN&src=app&x={x}&y={y}&z={z}&s=Galile",
                //urlTemplate: "http://mt1.google.cn/vt/lyrs=s&hl=zh-CN&gl=CN&x={x}&y={y}&z={z}&s=Gali",

                //urlTemplate: "http://online{s}.map.bdimg.com/onlinelabel/?qt=tile&x={x}&y={y}&z={z}&styles=pl&scaler=1&p=1",
                //subdomains: ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9'],

                // urlTemplate: "http://webrd0{s}.is.autonavi.com/appmaptile?lang=zh_cn&size=1&scale=1&style=8&x={x}&y={y}&z={z}",
                // subdomains: ['1', '2', '3', '4'],

                // urlTemplate: 'http://t{s}.tianditu.cn/DataServer?T=vec_w&X={x}&Y={y}&L={z}',
                // subdomains: ['0', '1', '2', '3', '4', '5', '6', '7'],
                
                additionalOptions: {
                  'accessToken': 'pk.eyJ1IjoibHVja3lrdyIsImEiOiJjanE2NWUxNTMyNTJmM3htdW9xbDNnb3lmIn0.05-Y4v23YalvWHFK21hXMQ',
                  'id': 'mapbox.streets',
                },
              ),
              // new TileLayerOptions(
              //   urlTemplate: "http://t{s}.tianditu.cn/DataServer?T=cva_w&X={x}&Y={y}&L={z}",
              //   subdomains: ['0', '1', '2', '3', '4', '5', '6', '7'],
              //   backgroundColor: Color.fromARGB(0, 0, 0, 0),
              // ),
              new MarkerLayerOptions(
                markers: [
                  new Marker(
                    width: 20.0,
                    height: 35.0,
                    point: _currentPoint,
                    builder: (ctx) =>
                    new Container(
                      child: new Image.asset("assets/mappin.png"),
                    ),
                  ),
                ],
              ),
            ],
          );
      },
    );
  }
}