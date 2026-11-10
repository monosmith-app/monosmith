package dev.monosmith.jetbrains.toolwindow

import dev.monosmith.jetbrains.dependencies.ToolUpdateContract
import dev.monosmith.jetbrains.dependencies.UpdateContract
import dev.monosmith.jetbrains.dependencies.UpdatesReportContract
import dev.monosmith.jetbrains.health.DoctorFindingContract
import dev.monosmith.jetbrains.health.DoctorFixContract
import dev.monosmith.jetbrains.releaseflow.ReleasePreviewContract
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

    @Test
    fun affectedProjectsAreMarkedAndTheFilterAndTheAffectedOnlyBoxNarrowTheList() {
        val api = SmithProjectContract("api", "apps/api", "app", "go", listOf("test"))
        val all = ToolWindowRowsMapper.projectRows(true, Result.success(listOf(web, api)), setOf("web"))
        assertEquals(listOf("● web (node app) · affected", "api (go app)", "New project…"), all.map { it.text })
        assertEquals(
            listOf(
                "api (go app)",
                "New project…",
            ),
            ToolWindowRowsMapper.projectRows(true, Result.success(listOf(web, api)), setOf("web"), "go").map {
                it.text
            },
        )
        assertEquals(
            listOf(
                "● web (node app) · affected",
                "New project…",
            ),
            ToolWindowRowsMapper.projectRows(true, Result.success(listOf(web, api)), setOf("web"), "", true).map {
                it.text
            },
        )
        assertEquals(
            "No project is affected by the changes.",
            ToolWindowRowsMapper.projectRows(true, Result.success(listOf(api)), setOf("web"), "", true).first().text,
        )
        assertEquals(
            "No project matches the filter.",
            ToolWindowRowsMapper.projectRows(true, Result.success(listOf(api)), emptySet(), "zzz").first().text,
        )
    }

    @Test
    fun theFilterMatchesTheTasksToo() {
        assertTrue(ToolWindowRowsMapper.matches(web, "build"))
        assertTrue(ToolWindowRowsMapper.matches(web, "node app"))
        assertTrue(!ToolWindowRowsMapper.matches(web, "go"))
    }

    @Test
    fun theDependenciesViewGroupsTheUpdatesByRiskAndListsTheTools() {
        val report =
            UpdatesReportContract(
                listOf(
                    UpdateContract("node", "zod", "3.22.0", "3.23.1", "minor", listOf("web")),
                    UpdateContract("node", "react", "18.2.0", "19.0.0", "major", listOf("web", "docs")),
                ),
                listOf(ToolUpdateContract("Go", "1.26", "1.27")),
            )
        val rows = ToolWindowRowsMapper.depsRows(Result.success(report))
        assertEquals(listOf("Update dependencies…", "Align versions…"), rows.take(2).map { it.text })
        assertEquals("Major: may break your code; read the changelog (1)", rows[2].text)
        assertEquals(listOf("react 18.2.0 → 19.0.0 (2 projects)"), (rows[2] as ToolWindowRowContract.Group).children)
        assertEquals("Tools with a newer version (1)", rows.last().text)
        assertEquals(
            "Every dependency is up to date.",
            ToolWindowRowsMapper.depsRows(Result.success(UpdatesReportContract(emptyList(), emptyList()))).last().text,
        )
    }

    @Test
    fun theReleaseViewStartsWithTheReleaseAction() {
        val preview = ReleasePreviewContract("", emptyList(), listOf("docs"), "")
        val rows = ToolWindowRowsMapper.releaseRows(Result.success(preview))
        assertIs<ToolWindowRowContract.Link>(rows.first())
        assertEquals("release", (rows.first() as ToolWindowRowContract.Link).action)
        assertTrue(rows.any { it.text.startsWith("Nothing to release") })
        assertEquals("smith failed: boom", ToolWindowRowsMapper.releaseRows(Result.failure(IllegalStateException("boom"))).single().text)
    }
}
