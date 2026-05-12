plugins {
    id("com.android.application")
    id("kotlin-android")
    // The Flutter Gradle Plugin must be applied after the Android and Kotlin Gradle plugins.
    id("dev.flutter.flutter-gradle-plugin")
}

android {
    namespace = "com.ego.ego"
    compileSdk = flutter.compileSdkVersion
    ndkVersion = flutter.ndkVersion

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions {
        jvmTarget = JavaVersion.VERSION_17.toString()
    }

    defaultConfig {
        // TODO: Specify your own unique Application ID (https://developer.android.com/studio/build/application-id.html).
        applicationId = "com.ego.ego"
        // You can update the following values to match your application needs.
        // For more information, see: https://flutter.dev/to/review-gradle-config.
        minSdk = flutter.minSdkVersion
        targetSdk = flutter.targetSdkVersion
        versionCode = flutter.versionCode
        versionName = flutter.versionName
    }

    signingConfigs {
        create("release") {
            // 手动解析 key.properties（避免 java.util.Properties 在 KTS 中的兼容问题）
            val keystoreProps = mutableMapOf<String, String>()
            rootProject.file("key.properties").let { file ->
                if (file.exists()) {
                    file.readLines().forEach { line ->
                        val parts = line.split("=", limit = 2)
                        if (parts.size == 2) {
                            keystoreProps[parts[0].trim()] = parts[1].trim()
                        }
                    }
                }
            }

            // CI 环境变量优先；本地开发回退到 key.properties
            storeFile = file(
                System.getenv("ANDROID_KEYSTORE_PATH")
                    ?: keystoreProps["storeFile"]
                    ?: "upload-keystore.jks"
            )
            storePassword = System.getenv("ANDROID_KEYSTORE_PASSWORD")
                ?: keystoreProps["storePassword"] ?: ""
            keyAlias = System.getenv("ANDROID_KEY_ALIAS")
                ?: keystoreProps["keyAlias"] ?: ""
            keyPassword = System.getenv("ANDROID_KEY_PASSWORD")
                ?: keystoreProps["keyPassword"] ?: ""
        }
    }

    buildTypes {
        release {
            signingConfig = signingConfigs.getByName("release")
        }
    }
}

flutter {
    source = "../.."
}
