package dev.monosmith.jetbrains.toolwindow

import dev.monosmith.jetbrains.health.DoctorFindingContract
import dev.monosmith.jetbrains.health.DoctorFixContract
import dev.monosmith.jetbrains.workspace.SmithProjectContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertIs
import kotlin.test.assertTrue

class ToolWindowRowsMapperTest {
    private val web = SmithProjectContract("web", "apps/web", "app", "node", listOf("build"))

    @Test
    fun aFolderWithNoSmithTomlOffersANewWorkspace() {
        val rows = ToolWindowRowsMapper.projectRows(false, null)
        assertEquals("No smith.toml in this project", rows.first().text)
        assertIs<ToolWindowRowContract.NewWorkspace>(rows.last())
    }

    @Test
    fun aWorkspaceWithNoProjectOffersAddYourFirstProject() {
        val rows = ToolWindowRowsMapper.projectRows(true, Result.success(emptyList()))
        assertEquals(listOf("Add your first project"), rows.map { it.text })
        assertIs<ToolWindowRowContract.AddProject>(rows.single())
    }

    @Test
    fun projectsAreFollowedByANewProjectRow() {
        val rows = ToolWindowRowsMapper.projectRows(true, Result.success(listOf(web)))
        assertEquals(listOf("web (node app)", "New project…"), rows.map { it.text })
    }

    @Test
    fun anUnreadableSmithTomlShowsSmithsWords() {
        val rows =
            ToolWindowRowsMapper.projectRows(
                true,
                Result.failure(IllegalStateException("smith.toml:3:9 expected a value Try: fix it")),
            )
        assertEquals("smith failed: smith.toml:3:9 expected a value Try: fix it", rows.single().text)
        assertEquals("Reading the workspace…", ToolWindowRowsMapper.projectRows(true, null).single().text)
    }

    @Test
    fun theHealthViewListsFailedChecksFirstThenWarnings() {
        val fixed = DoctorFindingContract("cache", false, "d", "r", false, DoctorFixContract("ignore it", true, emptyList()))
        val warn = DoctorFindingContract("vintage", false, "", "", true, null)
        val ok = DoctorFindingContract("git", true, "", "", false, null)
        val rows = ToolWindowRowsMapper.healthRows(true, Result.success(listOf(warn, ok, fixed)))
        assertEquals(listOf("Doctor: 1 failed, 1 warnings, 1 passed.", "failed: cache", "warning: vintage"), rows.map { it.text })
    }

    @Test
    fun aFindingWithAFixHasAFixRowUnderIt() {
        val fixed = DoctorFindingContract("cache", false, "d", "r", false, DoctorFixContract("ignore it", true, emptyList()))
        val children = ToolWindowRowsMapper.findingChildren(fixed)
        assertEquals(listOf("d", "fix: r", "Fix this check: ignore it"), children.map { it.text })
        assertTrue(children.last() is ToolWindowRowContract.Fix)
        assertEquals(emptyList(), ToolWindowRowsMapper.findingChildren(DoctorFindingContract("x", false, "", "", false, null)))
    }
}
