import 'dart:io';
import 'package:path_provider/path_provider.dart';
import 'package:http/http.dart' as http;

import './global.dart' as globals;

class ModelInfo {
  final String name;
  final String md5Hash;
  final int size;
  final String downloadPath;

  ModelInfo({this.name, this.md5Hash, this.size, this.downloadPath});

  factory ModelInfo.fromJson(Map<String, dynamic> json) {
    return ModelInfo(
      name: json['Name'],
      md5Hash: json["Md5Hash"],
      size: json["FileSizeInBytes"],
      downloadPath: json["DownloadPath"],
    );
  }

  Future<File> get localFile async {
    var appDoc = (await getApplicationDocumentsDirectory()).path;
    var f = File('$appDoc/models/$name');
    if (f.existsSync()) {
      return f;
    }

    //download from server
    await globals.modelDownloadLock.synchronized(() async {
      var response = await http.get(downloadPath.startsWith("/")
          ? '${globals.host}$downloadPath'
          : '${globals.host}/$downloadPath');
      if (response.statusCode == 200) {
        var fw = await new File(f.path).open(mode: FileMode.write);
        fw.truncateSync(0);
        fw.writeFromSync(response.bodyBytes);
        fw.closeSync();
      }
    });

    return f;
  }
}

class ArtPredict {
  final int id;
  final double score;

  ArtPredict({this.id, this.score});

  factory ArtPredict.fromJson(Map<String, dynamic> json) {
    return ArtPredict(id: json['ArtID'], score: json['Score']);
  }
}

class ArtInfo {
  final int id;
  final int museumId;
  final int artistId;
  final int displayNumber;
  final String creationYear;
  final int price;
  final String title;
  final String category;
  final String location;
  final List<String> images;
  final List<String> audios;
  final String text;
  final String material;
  final String museumName;
  final String museumCity;
  final String museumCountry;

  double _score;

  set score(double v) {
    _score = v;
  }

  double get score => _score;

  ArtInfo(
      {this.id,
      this.museumId,
      this.artistId,
      this.displayNumber,
      this.creationYear,
      this.price,
      this.title,
      this.category,
      this.location,
      this.images,
      this.audios,
      this.text,
      this.material,
      this.museumName,
      this.museumCity,
      this.museumCountry});

  factory ArtInfo.fromJson(Map<String, dynamic> json) {
    return ArtInfo(
      id: json["ArtID"],
      museumId: json["MuseumID"],
      artistId: json["ArtistID"],
      displayNumber: json["DisplayNumber"],
      creationYear: json["CreationYear"],
      price: json["Price"],
      title: json["Title"],
      category: json["Category"],
      location: json["Location"],
      images: json["Images"] == null ? null : json["Images"].cast<String>(),
      audios: json["Audios"] == null ? null : json["Audios"].cast<String>(),
      text: json["Text"],
      material: json["Material"],
      museumName: json["MuseumName"],
      museumCity: json["MuseumCity"],
      museumCountry: json["MuseumCountry"],
    );
  }

  Map<String, dynamic> toJson() =>
    {
      'id': id,
      'title': title,
    };
}
