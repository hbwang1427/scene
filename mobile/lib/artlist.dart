import 'dart:convert';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:http/http.dart' as http;
import 'package:cached_network_image/cached_network_image.dart';
import 'package:audioplayers/audioplayers.dart';
import './model.dart';
import './global.dart' as globals;

class ArtListView extends StatefulWidget {
  final List<ArtPredict> predicts;
  ArtListView({Key key, this.predicts}) : super(key: key);

  @override
  _ArtListViewState createState() => new _ArtListViewState();
}

class _ArtListViewState extends State<ArtListView> {
  final List<ArtInfo> artInfoList = new List<ArtInfo>();
  @override
  void initState() {
    super.initState();
    asyncInit();
  }

  void asyncInit() async {
    for (var item in this.widget.predicts) {
      http.get('${globals.host}/art/${item.id}').then((response) {
        if (response.statusCode == 200) {
          var obj = json.decode(response.body);
          if (obj["error"] == null) {
            var artInfo =
                ArtInfo.fromJson(json.decode(response.body)['results']);
            artInfo.score = item.score;
            setState(() {
              artInfoList.add(artInfo);
              artInfoList.sort((a, b) => a.score > b.score ? -1 : 1);
            });
          } else {
            print("get art error:${obj['error']}");
          }
        }
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(title: Text('predict results')),
      body: Container(
        child: ListView.builder(
          itemCount: artInfoList.length,
          padding: const EdgeInsets.only(top: 20.0),
          itemBuilder: (context, index) {
            return ArtInfoCard(artInfoList[index]);
          },
        ),
      ),
    );
  }
}

class ArtInfoCard extends StatefulWidget {
  final ArtInfo info;

  ArtInfoCard(this.info);

  @override
  _ArtInfoCardState createState() {
    return _ArtInfoCardState();
  }
}

class _ArtInfoCardState extends State<ArtInfoCard> {
  String renderUrl;

  Widget get _card {
    ArtInfo art = this.widget.info;
    List<Widget> widgets = new List<Widget>();

    widgets.add(ListTile(
      leading: const Icon(Icons.album),
      title: Text('The ${art.title}'),
      subtitle: Text('${art.museumName} / ${art.category}'),
    ));

    art.images.forEach((url) {
      widgets.add(CachedNetworkImage(
        //imageUrl: "http://via.placeholder.com/350x150",
        imageUrl: art.images[0].startsWith("/")
            ? "${globals.host}${art.images[0]}"
            : "${globals.host}/${art.images[0]}",
        placeholder: new CircularProgressIndicator(),
        errorWidget: new Icon(Icons.broken_image),
      ));
    });

    art.audios.forEach((url) {
      widgets.add(AudioPlayerWidget(
          url: url.startsWith("/")
              ? "${globals.host}$url"
              : "${globals.host}/$url"));
    });

    widgets.add(Padding(
      child: Text(art.text),
      padding: new EdgeInsets.all(15.0),
    ));

    widgets.add(new ButtonTheme.bar(
        // make buttons use the appropriate styles for cards
        child: new ButtonBar(children: <Widget>[
      new FlatButton(
        child: Icon(Icons.favorite_border, size: 16),
        onPressed: () {/* ... */},
      ),
      new FlatButton(
        child: Icon(Icons.comment, size: 16),
        onPressed: () {/* ... */},
      )
    ])));

    return new Card(
      child: Padding(
        child: Column(mainAxisSize: MainAxisSize.min, children: widgets),
        padding: EdgeInsets.only(top: 10.0),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return new Container(
      child: _card,
    );
  }
}

class AudioPlayerWidget extends StatefulWidget {
  final String url;

  AudioPlayerWidget({Key key, this.url}) : super(key: key);

  @override
  _AudioPlayerWidgetState createState() => _AudioPlayerWidgetState();
}

class _AudioPlayerWidgetState extends State<AudioPlayerWidget> {
  bool hasError = false;
  Duration duration = new Duration(hours: 0, minutes: 0, seconds: 0),
      position = new Duration(hours: 0, minutes: 0, seconds: 0);
  AudioPlayer audioPlayer = new AudioPlayer();

  @override
  void initState() {
    super.initState();

    audioPlayer.audioPlayerStateChangeHandler = (state) {
      setState(() {});
    };

    audioPlayer.durationHandler = (Duration d) {
      print('Max duration: $d');
      setState(() {
        duration = d;
      });
    };

    audioPlayer.positionHandler = (Duration p) {
      print('Current position: $p');
      setState(() {
        position = p;
      });
    };

    audioPlayer.errorHandler = (msg) {
      print('audioPlayer error : $msg');
      setState(() {
        duration = new Duration(seconds: 0);
        position = new Duration(seconds: 0);
      });
    };
  }

  void buttonPressed() async {
    try {
      if (audioPlayer.state == null || audioPlayer.state == AudioPlayerState.STOPPED) {
        await this.audioPlayer.play(this.widget.url);
      } else if (audioPlayer.state == AudioPlayerState.PLAYING) {
        await this.audioPlayer.pause();
      } else if (audioPlayer.state == AudioPlayerState.PAUSED) {
        await this.audioPlayer.resume();
      } else if (audioPlayer.state == AudioPlayerState.COMPLETED) {
        await this.audioPlayer.seek(Duration());
        await this.audioPlayer.resume();
      }
    } on PlatformException catch (e) {
      print("play ${this.widget.url} error: $e");
      hasError = true;
    }
    setState(() {});
  }

  String prossText() {
    if (position.inSeconds == 0 || duration.inSeconds == 0) {
      return "--/--";
    }

    var timeLeft = duration.inSeconds - position.inSeconds;
    var timeLeftMinutes = timeLeft ~/ 60;
    var timeLeftSeconds = timeLeft - timeLeftMinutes * 60;
    var durseconds = duration.inSeconds - duration.inMinutes * 60;
    return "${timeLeftMinutes}''${timeLeftSeconds}'/${duration.inMinutes}''${durseconds}'";
  }

  @override
  Widget build(BuildContext context) {
    Color iconColor = hasError ? Colors.grey : Colors.black;
    return Padding(
      padding: EdgeInsets.all(5.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceEvenly,
        children: <Widget>[
          Expanded(
              child: Slider(
            min: 0.0,
            max: 1.0,
            value: position.inSeconds == 0 || duration.inSeconds == 0
                ? 0
                : position.inSeconds / duration.inSeconds,
                onChanged: audioPlayer == null ? null : (double val) => audioPlayer.seek(Duration(hours:0, minutes:0, seconds: (val * duration.inSeconds).toInt())),
          )),
          Text(prossText()),
          FlatButton(
            child: audioPlayer.state != AudioPlayerState.PLAYING
                ? Icon(Icons.play_arrow, color: iconColor)
                : Icon(Icons.pause, color: iconColor),
            onPressed: buttonPressed,
          ),
        ],
      ),
    );
  }
}
