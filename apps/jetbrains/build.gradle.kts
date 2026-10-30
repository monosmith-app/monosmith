import org.jetbrains.intellij.platform.gradle.IntelliJPlatformType
import org.jetbrains.intellij.platform.gradle.TestFrameworkType
import org.jetbrains.kotlin.gradle.dsl.JvmTarget

plugins {
    kotlin("jvm")
    id("org.jlleitschuh.gradle.ktlint")
    id("org.jetbrains.intellij.platform")
}

group = "dev.monosmith"
version = "0.0.0"

layout.buildDirectory = rootProject.layout.projectDirectory.dir("dist/gradle/monosmith-jetbrains")

java {
    sourceCompatibility = JavaVersion.VERSION_21
    targetCompatibility = JavaVersion.VERSION_21
}

kotlin {
    compilerOptions {
        jvmTarget = JvmTarget.JVM_21
    }
}

repositories {
    mavenCentral()
    intellijPlatform {
        defaultRepositories()
    }
}

dependencies {
    intellijPlatform {
        create(IntelliJPlatformType.IntellijIdeaCommunity, "2025.1")
        testFramework(TestFrameworkType.Platform)
    }
    testImplementation("junit:junit:4.13.2")
    testImplementation(kotlin("test"))
}

intellijPlatform {
    pluginConfiguration {
        name = "MonoSmith"
        ideaVersion {
            sinceBuild = "251"
        }
    }
    // Publishing to the JetBrains Marketplace is deliberately not configured: see README.md.
}
