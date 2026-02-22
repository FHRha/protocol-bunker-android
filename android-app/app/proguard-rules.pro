# Keep entry points referenced from the manifest.
-keep class com.protocolbunker.host.MainActivity { *; }
-keep class com.protocolbunker.host.WebViewActivity { *; }
-keep class com.protocolbunker.host.server.ServerForegroundService { *; }

# Keep potential JS bridge methods if they are added later.
-keepclassmembers class * {
    @android.webkit.JavascriptInterface <methods>;
}
