package dev.monosmith.jetbrains.adoptfolder

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNotNull
import kotlin.test.assertNull
import kotlin.test.assertTrue

class AdoptFolderTest {
    private val fake = FakePortsHarness()
    private val program = RootProjectContract("package.json", "node", "billing-api", true)
    private val library = RootProjectContract("go.mod", "go", "tools", false)
    private val report =
        """{"name":"repo","providers":["github"],"projects":[],"blockers":[],"rootProject":{"manifest":"package.json","ecosystem":"node","name":"billing-api","app":true}}"""

    init {
        fake.answers["adopt --dry-run --json"] = fake.ok(report)
        fake.answers["new --questions --json"] =
            fake.ok(
                """[{"key":"ci","prompt":"Which CI?","default":"github","many":true,"options":[{"value":"github","label":"GitHub Actions"},{"value":"azure","label":"Azure"}]}]""",
            )
        fake.answers["adopt --init --move apps/billing-api --gitattributes add --agent add --ci github --ci-existing keep --dry-run"] =
            fake.ok("would write smith.toml\n+ diff\n")
    }

    @Test
    fun readsWhatSmithFoundAndTheBlockers() {
        val found = AdoptReportMapper.toReport(report)
        assertEquals(program, found.rootProject)
        assertEquals(listOf("github"), found.providers)
        val blocked =
            AdoptReportMapper.toReport(
                """{"blockers":[{"message":"this is already a smith workspace","step":"run smith doctor"}]}""",
            )
        assertEquals(listOf("this is already a smith workspace. run smith doctor"), blocked.blockers)
        assertNull(blocked.rootProject)
    }

    @Test
    fun theSuggestedFolderDependsOnWhetherTheProjectIsAProgram() {
        assertEquals("apps", AdoptReportMapper.moveChoices(program).first { it.suggested }.key)
        assertEquals("packages", AdoptReportMapper.moveChoices(library).first { it.suggested }.key)
        assertEquals(5, AdoptReportMapper.moveChoices(program).size)
        assertEquals("apps", AdoptReportMapper.moveRows(program).first().name)
        assertEquals("(suggested)", AdoptReportMapper.moveRows(program).first().description)
    }

    @Test
    fun checksTheNameAndTheFolderWithTheFix() {
        assertNull(AdoptReportMapper.nameProblem("web-app"))
        assertNotNull(AdoptReportMapper.nameProblem("WebApp"))
        assertEquals("tools/x" to null, AdoptReportMapper.folderOrProblem(" tools\\x "))
        assertTrue(AdoptReportMapper.folderOrProblem(".").second!!.contains("repository root"))
        assertTrue(AdoptReportMapper.folderOrProblem("../x").second!!.contains("leaves the repository"))
        assertTrue(AdoptReportMapper.folderOrProblem("/x").second!!.contains("relative"))
        assertTrue(AdoptReportMapper.folderOrProblem("a/.b").second!!.contains("dot"))
        assertTrue(AdoptReportMapper.folderOrProblem("a b").second!!.contains("letters"))
    }

    @Test
    fun everyAnswerBecomesAFlagSoNothingIsAskedTwice() {
        assertEquals(
            listOf("adopt", "--init", "--move", "none", "--gitattributes", "skip", "--agent", "never", "--ci", "none", "--dry-run"),
            AdoptReportMapper.initArguments("none", "skip", "never", emptyList(), null, apply = false),
        )
        assertEquals(
            "--yes",
            AdoptReportMapper.initArguments(null, "add", "add", listOf("github", "azure"), "replace", apply = true).last(),
        )
    }

    @Test
    fun asksTheMoveTheFourQuestionsAndShowsThePlanBeforeWriting() {
        fake.ones.addAll(listOf("apps", "add", "add", "keep"))
        fake.texts.add("billing-api")
        fake.manies.add(listOf("github"))
        fake.plans.add(fake.yes())
        assertTrue(AdoptFolderUseCase.run(fake.ports) {})
        assertEquals("would write smith.toml\n+ diff", fake.plansSeen.single())
        assertEquals(
            listOf(
                "adopt",
                "--init",
                "--move",
                "apps/billing-api",
                "--gitattributes",
                "add",
                "--agent",
                "add",
                "--ci",
                "github",
                "--ci-existing",
                "keep",
                "--yes",
            ),
            fake.calls.last(),
        )
    }

    @Test
    fun theSuggestedMoveIsFirstAndLeavingItIsAnAnswerNotAnAssumption() {
        fake.ones.addAll(listOf("apps", "add", "add", "keep"))
        fake.texts.add("billing-api")
        fake.manies.add(listOf("github"))
        AdoptFolderUseCase.run(fake.ports) {}
        assertEquals(listOf("apps", "packages", "libs", "other", "none"), fake.rowsSeen.first().map { it.name })
    }

    @Test
    fun neverCommitsUnlessTheBoxIsTicked() {
        fake.ones.addAll(listOf("apps", "add", "add", "keep"))
        fake.texts.add("billing-api")
        fake.manies.add(listOf("github"))
        fake.plans.add(fake.yes(commit = true))
        AdoptFolderUseCase.run(fake.ports) {}
        assertEquals("--commit", fake.calls.last().last())
    }

    @Test
    fun aFolderThatIsAlreadyAWorkspaceIsRefusedWithTheStepSmithGives() {
        fake.answers["adopt --dry-run --json"] =
            fake.ok("""{"blockers":[{"message":"this is already a smith workspace","step":"run smith doctor"}]}""")
        assertFalse(AdoptFolderUseCase.run(fake.ports) {})
        assertEquals("Nothing was adopted. this is already a smith workspace. run smith doctor", fake.warnings.single())
    }

    @Test
    fun cancelAtTheMoveQuestionWritesNothing() {
        assertFalse(AdoptFolderUseCase.run(fake.ports) {})
        assertTrue(fake.calls.none { "--init" in it })
    }
}
