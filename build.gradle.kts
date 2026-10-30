// The plugins' versions, for every module; each module applies the ones it
// uses.
plugins {
    kotlin("jvm") version "2.4.21" apply false
    id("org.jlleitschuh.gradle.ktlint") version "14.2.0" apply false
    id("org.jetbrains.intellij.platform") version "2.19.0" apply false
}
