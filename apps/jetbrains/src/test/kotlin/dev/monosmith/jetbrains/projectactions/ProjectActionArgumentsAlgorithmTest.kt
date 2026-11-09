package dev.monosmith.jetbrains.projectactions

import kotlin.test.Test
import kotlin.test.assertEquals

class ProjectActionArgumentsAlgorithmTest {
    @Test
    fun removeAlwaysNamesKeepCopyOrDelete() {
        assertEquals(listOf("remove", "web", "--dry-run"), ProjectActionArgumentsAlgorithm.removePlan("web"))
        assertEquals(listOf("remove", "web", "--keep-copy", "--yes"), ProjectActionArgumentsAlgorithm.remove("web", delete = false))
        assertEquals(listOf("remove", "web", "--delete", "--yes"), ProjectActionArgumentsAlgorithm.remove("web", delete = true))
    }

    @Test
    fun kindHasAPlanAndAChange() {
        assertEquals(listOf("kind", "web", "lib", "--dry-run"), ProjectActionArgumentsAlgorithm.kindPlan("web", "lib"))
        assertEquals(listOf("kind", "web", "lib", "--yes"), ProjectActionArgumentsAlgorithm.kind("web", "lib"))
    }

    @Test
    fun aProjectCanBecomeAnyKindButItsOwn() {
        assertEquals(listOf("lib", "package"), ProjectActionArgumentsAlgorithm.otherKinds("app"))
        assertEquals(listOf("app", "lib"), ProjectActionArgumentsAlgorithm.otherKinds("package"))
        assertEquals(listOf("lib", "package"), ProjectActionArgumentsAlgorithm.kindRows("app").map { it.name })
    }

    @Test
    fun saysAKindWithItsArticle() {
        assertEquals("an app", ProjectActionArgumentsAlgorithm.withArticle("app"))
        assertEquals("a lib", ProjectActionArgumentsAlgorithm.withArticle("lib"))
    }
}
