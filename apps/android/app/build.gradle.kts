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

        // Inject Google Client ID into BuildConfig
        val googleClientId = getLocalProperty("GOOGLE_CLIENT_ID", "")
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
    implementation(libs.androidx.credentials)
    implementation(libs.androidx.lifecycle.runtime.ktx)
    implementation(libs.androidx.activity.compose)
    implementation(platform(libs.androidx.compose.bom))
    implementation(libs.androidx.compose.ui)
    implementation(libs.androidx.compose.ui.graphics)
    implementation(libs.androidx.compose.ui.tooling.preview)
    implementation(libs.androidx.compose.material3)
    implementation(libs.androidx.compose.material.icons.core)
    implementation(libs.androidx.compose.material.icons.extended)
    implementation(libs.androidx.browser)
    testImplementation(libs.junit)
    androidTestImplementation(libs.androidx.junit)
    androidTestImplementation(libs.androidx.espresso.core)
    androidTestImplementation(platform(libs.androidx.compose.bom))
    androidTestImplementation(libs.androidx.compose.ui.test.junit4)
    debugImplementation(libs.androidx.compose.ui.tooling)
    debugImplementation(libs.androidx.compose.ui.test.manifest)
    implementation(libs.androidx.appcompat)

    implementation(libs.androidx.lifecycle.viewmodel.ktx)
    implementation(libs.androidx.lifecycle.livedata.ktx)

    implementation(libs.androidx.room.runtime)
    ksp(libs.androidx.room.compiler)

    implementation(libs.androidx.room.ktx)

    implementation(libs.androidx.recyclerview)
    implementation(libs.javax.inject)

    // Hilt
    implementation(libs.hilt.android)
    ksp(libs.hilt.compiler)
    ksp(libs.kotlin.metadata)
    
    implementation(libs.androidx.activity.ktx)
    implementation(libs.google.gson)
    implementation(libs.retrofit.core)
    implementation(libs.retrofit.converter.gson)
    implementation(libs.androidx.security.crypto)
    implementation(libs.androidx.navigation.compose)
    implementation(libs.androidx.datastore.preferences)
    implementation(libs.tink.android)
    compileOnly(libs.google.errorprone.annotations)

    // Logging & Debugging
    implementation(libs.timber)
    implementation(libs.okhttpLogging)
    debugImplementation(libs.chuckerDebug)
    "stagingImplementation"(libs.chuckerDebug)
    releaseImplementation(libs.chuckerRelease)

    // Android Credential Manager libraries
}