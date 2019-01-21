package net.pangolinai.mobile;

import android.os.Bundle;
import io.flutter.app.FlutterActivity;
import io.flutter.plugins.GeneratedPluginRegistrant;


public class MainActivity extends FlutterActivity {

  @Override
  protected void onCreate(Bundle savedInstanceState) {
    super.onCreate(savedInstanceState);
    GeneratedPluginRegistrant.registerWith(this);

    //setup tflite method channel
    TfliteHandler.registerWith(this.registrarFor(TfliteHandler.CHANNEL));

    //a special museum handler
    MuseumTfliteHandler.registerWith(this.registrarFor(MuseumTfliteHandler.CHANNEL));
  }
}
