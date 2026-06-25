import java.util.Properties

plugins {
    alias(libs.plugins.android.application)
    alias(libs.plugins.kotlin.compose)
    alias(libs.plugins.ksp)
    alias(libs.plugins.hilt.android)
}

// Industry Standard: Load local.properties to avoid hardcoding local IPs in Git
val localProperties = Properties().apply {
    val file = rootProject.file("local.properties")
    if (file.exists()) {
        load(file.inputStream())
    }
}

fun getLocalProperty(key: String, defaultValue: String): String {
    return localProperties.getProperty(key) ?: defaultValue
}

android {
    namespace = "com.build.todoapplearn"
    compileSdk = 37

    defaultConfig {
        applicationId = "com.build.todoapplearn"
        minSdk = 33
        targetSdk = 37
        versionCode = 1
        versionName = "1.0"

        testInstrumentationRunner = "androidx.test.runner.AndroidJUnitRunner"

        // Inject Google Auth Endpoint and Redirect URI into BuildConfig
        val googleAuthEndpoint = getLocalProperty("GOOGLE_AUTH_ENDPOINT", "https://accounts.google.com/o/oauth2/v2/auth")
        val redirectUri = getLocalProperty("REDIRECT_URI", "todoapp://oauth/callback")
        val googleClientId = getLocalProperty("GOOGLE_CLIENT_ID", "")
        buildConfigField("String", "GOOGLE_AUTH_ENDPOINT", "\"$googleAuthEndpoint\"")
        buildConfigField("String", "REDIRECT_URI", "\"$redirectUri\"")
        buildConfigField("String", "GOOGLE_CLIENT_ID", "\"$googleClientId\"")
    }

    buildTypes {
        release {
            isMinifyEnabled = true
            isShrinkResources = true
            proguardFiles(getDefaultProguardFile("proguard-android-optimize.txt"), "proguard-rules.pro")

            val apiUrl = getLocalProperty("API_URL_RELEASE", "https://react-springboot-full-stack.onrender.com/")
            buildConfigField("String", "BASE_URL", "\"$apiUrl\"")
            buildConfigField("boolean", "LOGS_ENABLED", "false")
            signingConfig = signingConfigs.getByName("debug")
        }

        create("staging") {
            initWith(getByName("release"))
            isDebuggable = false
            applicationIdSuffix = ".staging"

            val apiUrl = getLocalProperty("API_URL_STAGING", "https://react-springboot-full-stack.onrender.com/")
            buildConfigField("String", "BASE_URL", "\"$apiUrl\"")
            buildConfigField("boolean", "LOGS_ENABLED", "true")
        }

        debug {
            // Priority: local.properties > default IP
            val apiUrl = getLocalProperty("API_URL_DEBUG", "http://10.0.2.2:3000/")
            buildConfigField("String", "BASE_URL", "\"$apiUrl\"")
            buildConfigField("boolean", "LOGS_ENABLED", "true")
        }
    }

    buildFeatures {
        buildConfig = true
        compose = true
    }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }
}

kotlin {
    compilerOptions {
        jvmTarget.set(org.jetbrains.kotlin.gradle.dsl.JvmTarget.JVM_17)
    }
}

dependencies {
    // UI & Compose
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.material3)
    implementation(libs.androidx.compose.material)
    implementation(libs.androidx.compose.ui.graphics)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.compose.material.icons.core)
    implementation(libs.androidx.compose.material.icons.extended)
    implementation(libs.androidx.activity.compose)
    implementation(libs.androidx.navigation.compose)
    implementation(libs.coil.compose)

    // AndroidX & Lifecycle
    implementation(libs.androidx.core.ktx)
    implementation(libs.androidx.lifecycle.viewmodel.ktx)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.browser)
    implementation(libs.androidx.credentials)
    implementation(libs.androidx.credentials.auth)
    implementation(libs.googleid)

    // DI - Hilt
    implementation(libs.hilt.android)
    implementation(libs.hilt.navigation.compose)
    ksp(libs.hilt.compiler)

    // Networking
    implementation(libs.retrofit.core)
    implementation(libs.retrofit.converter.gson)
    implementation(libs.okhttpLogging)
    implementation(libs.google.gson)
    debugImplementation(libs.chuckerDebug)
    releaseImplementation(libs.chuckerRelease)

    // Local Data
    implementation(libs.androidx.room.runtime)
    implementation(libs.androidx.room.ktx)
    ksp(libs.androidx.room.compiler)
    implementation(libs.androidx.datastore.preferences)
    implementation(libs.androidx.security.crypto)
    implementation(libs.tink.android)

    // Utilities
    implementation(libs.timber)
  // ── Secure storage ─────────────────────────────────────────────────────
  implementation(libs.androidx.security.crypto.v110alpha06)

  // ── Networking ─────────────────────────────────────────────────────────
  implementation(libs.retrofit.v2110)
  implementation(libs.converter.moshi)
  implementation(libs.moshi.kotlin)
  ksp(libs.moshi.kotlin.codegen)
  implementation(libs.okhttp)
  implementation(libs.okhttp.sse)
  implementation(libs.logging.interceptor.v4120)

}
