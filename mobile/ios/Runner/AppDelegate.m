#include "AppDelegate.h"
#include "GeneratedPluginRegistrant.h"
#include "TflitePlugin.h"
#include "MuseumTflitePlugin.h"

@implementation AppDelegate

- (BOOL)application:(UIApplication *)application
    didFinishLaunchingWithOptions:(NSDictionary *)launchOptions {
    
    
    [MuseumTflitePlugin registerWithRegistrar:[self registrarForPlugin:@"net.pangolinai.mobile/museum_tflite"]];
    
    [TflitePlugin registerWithRegistrar:[self registrarForPlugin:@"net.pangolinai.mobile/tflite"]];
    
  [GeneratedPluginRegistrant registerWithRegistry:self];
    
  // Override point for customization after application launch.
  return [super application:application didFinishLaunchingWithOptions:launchOptions];
}

@end
