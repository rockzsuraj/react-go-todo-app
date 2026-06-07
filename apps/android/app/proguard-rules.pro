# --- INDUSTRY STANDARD PROGUARD RULES ---

# 1. Attributes required for de-obfuscation and generic type preservation
# Preserves line numbers for useful stack traces in Crashlytics
-keepattributes SourceFile,LineNumberTable
# Preserves generic types for Retrofit/Gson to function correctly
-keepattributes Signature
# Preserves annotations like @SerializedName and @Keep
-keepattributes *Annotation*

# 2. Gson: Keep fields with @SerializedName
-keepclassmembers class * {
    @com.google.gson.annotations.SerializedName <fields>;
}

# 3. Retrofit: Keep service interfaces and methods
-keep interface com.build.todoapplearn.data.remote.** { *; }
-dontwarn retrofit2.**

# 4. Coroutines: Standard rules for Kotlin Coroutines
-keepnames class kotlinx.coroutines.internal.MainDispatcherFactory {}
-keepnames class kotlinx.coroutines.CoroutineExceptionHandler {}
-keepclassmembernames class kotlinx.coroutines.android.HandlerContext {
    private final android.os.Handler handler;
}

# 5. Tink / ErrorProne: Suppress warnings for missing compile-time annotations
-dontwarn com.google.errorprone.annotations.**
-dontwarn javax.annotation.**

# 6. OkHttp: Standard rules
-dontwarn okhttp3.**
-dontwarn okio.**
-keep class okhttp3.** { *; }

# Note: Individual DTOs and Domain Models are handled via @Keep annotation in code.
