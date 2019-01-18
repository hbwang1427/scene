package net.pangolinai.mobile;

import android.content.res.AssetFileDescriptor;
import android.content.res.AssetManager;
import android.graphics.Bitmap;
import android.graphics.Canvas;
import android.graphics.Matrix;
import android.graphics.BitmapFactory;
import android.util.Log;

import io.flutter.plugin.common.MethodCall;
import io.flutter.plugin.common.MethodChannel;
import io.flutter.plugin.common.PluginRegistry;

import org.tensorflow.lite.Interpreter;

import java.io.BufferedReader;
import java.io.File;
import java.io.FileDescriptor;
import java.io.FileInputStream;
import java.io.IOException;
import java.io.InputStream;
import java.io.InputStreamReader;
import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.MappedByteBuffer;
import java.nio.channels.FileChannel;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Comparator;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.PriorityQueue;
import java.util.Vector;

public class MuseumTfliteHandler implements MethodChannel.MethodCallHandler {
    public static final String CHANNEL = "net.pangolinai.mobile/museum_tflite";
    private final PluginRegistry.Registrar mRegistrar;
    private Interpreter tfLite;
    float[][] labelProb;
    final static int CLASS_NUM = 1280;
    final static int IMAGE_SIZE = 224;
    final static int BYTES_PER_CHANNEL = 4;
    final static String OUTPUT_LAYER = "global_average_pooling2d_1";


    public static void registerWith(PluginRegistry.Registrar registrar) {
        final MethodChannel channel = new MethodChannel(registrar.messenger(), CHANNEL);
        channel.setMethodCallHandler(new MuseumTfliteHandler(registrar));
    }

    private MuseumTfliteHandler(PluginRegistry.Registrar registrar) {
        this.mRegistrar = registrar;
    }

    @Override
    public void onMethodCall(MethodCall call, MethodChannel.Result result) {
        if (call.method.equals("loadModel")) {
            try {
                String res = loadModel((HashMap) call.arguments);
                result.success(res);
            }
            catch (Exception e) {
                result.error("Failed to load model" , e.getMessage(), e);
            }
        } else if (call.method.equals("runModelOnImage")) {
            try {
                float[] feature  = runModelOnImage((HashMap) call.arguments);
                ByteBuffer buf = ByteBuffer.allocate(feature.length * 4).order(ByteOrder.LITTLE_ENDIAN);
                for (int i = 0; i < feature.length; i++) {
                    buf.putFloat(feature[i]);
                }
                result.success(buf.array());
            }
            catch (Exception e) {
                result.error("Failed to run model" , e.getMessage(), e);
            }
        } else if (call.method.equals("runModelOnBinary")) {
            try {
                float[] feature = runModelOnBinary((HashMap) call.arguments);
                ByteBuffer buf = ByteBuffer.allocate(feature.length * 4).order(ByteOrder.LITTLE_ENDIAN);
                for (int i = 0; i < feature.length; i++) {
                    buf.putFloat(feature[i]);
                }
                result.success(buf.array());
            }
            catch (Exception e) {
                result.error("Failed to run model" , e.getMessage(), e);
            }
        } else if (call.method.equals("close")) {
            close();
        }
    }

    private String loadModel(HashMap args) throws IOException {
        String model = args.get("model").toString();
        //AssetManager assetManager = mRegistrar.context().getAssets();
        //String key = mRegistrar.lookupKeyForAsset(model);

        //AssetFileDescriptor fileDescriptor = assetManager.openFd(key);
        File f = new File(model);
        FileInputStream inputStream = new FileInputStream(f);
        FileChannel fileChannel = inputStream.getChannel();

        long startOffset = 0;
        long declaredLength = f.length();
        MappedByteBuffer buffer = fileChannel.map(FileChannel.MapMode.READ_ONLY, startOffset, declaredLength);
        tfLite = new Interpreter(buffer);

        return "success";
    }

    private ByteBuffer loadImage(String path, int width, int height, int channels, float mean, float std)
            throws IOException {
        InputStream inputStream = new FileInputStream(path.replace("file://",""));
        Bitmap bitmapRaw = BitmapFactory.decodeStream(inputStream);

        Matrix matrix = getTransformationMatrix(
                bitmapRaw.getWidth(), bitmapRaw.getHeight(),
                width, height, false);

        int[] intValues = new int[width * height];
        ByteBuffer imgData = ByteBuffer.allocateDirect(1 * width * height * channels * BYTES_PER_CHANNEL);
        imgData.order(ByteOrder.nativeOrder());

        Bitmap bitmap = Bitmap.createBitmap(width, height, Bitmap.Config.ARGB_8888);
        final Canvas canvas = new Canvas(bitmap);
        canvas.drawBitmap(bitmapRaw, matrix, null);
        bitmap.getPixels(intValues, 0, bitmap.getWidth(), 0, 0, bitmap.getWidth(), bitmap.getHeight());

        int pixel = 0;
        for (int i = 0; i < width; ++i) {
            for (int j = 0; j < height; ++j) {
                int pixelValue = intValues[pixel++];
                imgData.putFloat((((pixelValue >> 16) & 0xFF) - mean) / std);
                imgData.putFloat((((pixelValue >> 8) & 0xFF) - mean) / std);
                imgData.putFloat(((pixelValue & 0xFF) - mean) / std);
            }
        }

        return imgData;
    }

    private float[] runModelOnImage(HashMap args) throws IOException {
        String path = args.get("path").toString();
        int NUM_THREADS = (int)args.get("numThreads");
        int WANTED_WIDTH = (int)args.get("inputSize");
        int WANTED_HEIGHT = (int)args.get("inputSize");
        int WANTED_CHANNELS = (int)args.get("numChannels");
        double mean = (double)(args.get("imageMean"));
        float IMAGE_MEAN = (float)mean;
        double std = (double)(args.get("imageStd"));
        float IMAGE_STD = (float)std;
        int NUM_RESULTS = (int)args.get("numResults");
        double threshold = (double)args.get("threshold");
        float THRESHOLD = (float)threshold;

        ByteBuffer imgData = loadImage(path, WANTED_WIDTH, WANTED_HEIGHT, WANTED_CHANNELS, IMAGE_MEAN, IMAGE_STD);
        tfLite.setNumThreads(NUM_THREADS);

        float[][] feature = new float[1][CLASS_NUM];
        tfLite.run(imgData, feature);
        return feature[0];
    }

    private float[] runModelOnBinary(HashMap args) throws IOException {
        byte[] binary = (byte[])args.get("binary");
        int NUM_THREADS = (int)args.get("numThreads");
        int NUM_RESULTS = (int)args.get("numResults");
        double threshold = (double)args.get("threshold");
        float THRESHOLD = (float)threshold;

        ByteBuffer imgData = ByteBuffer.wrap(binary);
        tfLite.setNumThreads(NUM_THREADS);
        float[][] feature = new float[1][CLASS_NUM];
        tfLite.run(imgData, feature);
        return feature[0];
    }

    private void close() {
        tfLite.close();
        labelProb = null;
    }

    public static Matrix getTransformationMatrix(final int srcWidth,
                                                 final int srcHeight,
                                                 final int dstWidth,
                                                 final int dstHeight,
                                                 final boolean maintainAspectRatio)
    {
        final Matrix matrix = new Matrix();

        if (srcWidth != dstWidth || srcHeight != dstHeight) {
            final float scaleFactorX = dstWidth / (float) srcWidth;
            final float scaleFactorY = dstHeight / (float) srcHeight;

            if (maintainAspectRatio) {
                final float scaleFactor = Math.max(scaleFactorX, scaleFactorY);
                matrix.postScale(scaleFactor, scaleFactor);
            } else {
                matrix.postScale(scaleFactorX, scaleFactorY);
            }
        }

        matrix.invert(new Matrix());
        return matrix;
    }
}