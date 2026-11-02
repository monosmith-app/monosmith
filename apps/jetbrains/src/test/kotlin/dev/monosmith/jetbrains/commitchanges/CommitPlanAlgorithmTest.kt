package dev.monosmith.jetbrains.commitchanges

import kotlin.test.Test
import kotlin.test.assertEquals

class CommitPlanAlgorithmTest {
    @Test
    fun readsTheFilesGitShowsThatTheOperationDidNotChange() {
        val text =
            "Operation: add app web (3 files)\n  apps/web/package.json\n" +
                "Git shows 2 more changed files:\n  README.md\n  notes.txt\nPass --yes to commit.\n"
        val plan = CommitPlanAlgorithm.read(text)
        assertEquals(listOf("README.md", "notes.txt"), plan.extras)
        assertEquals(text.trim(), plan.text)
    }

    @Test
    fun noExtrasWhenGitShowsNothingMore() {
        assertEquals(emptyList(), CommitPlanAlgorithm.read("Operation: add app web (3 files)\n  apps/web/package.json\n").extras)
    }

    @Test
    fun theExtrasRunToTheEndWhenNothingFollowsThem() {
        assertEquals(listOf("a", "b"), CommitPlanAlgorithm.read("Git shows 2 more changed files:\n  a\n  b").extras)
    }

    @Test
    fun theThreeCommandLines() {
        assertEquals(listOf("commit", "add", "app", "web", "--dry-run"), CommitPlanAlgorithm.arguments("add app web", null))
        assertEquals(listOf("commit", "--only-operation", "--yes"), CommitPlanAlgorithm.arguments("", "operation"))
        assertEquals(listOf("commit", "--all", "--yes"), CommitPlanAlgorithm.arguments("", "all"))
    }
}
