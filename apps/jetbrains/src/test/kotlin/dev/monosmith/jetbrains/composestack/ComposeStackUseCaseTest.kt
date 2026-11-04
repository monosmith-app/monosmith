package dev.monosmith.jetbrains.composestack

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import dev.monosmith.jetbrains.prompt.PlanAnswerContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class ComposeStackUseCaseTest {
    private val fake = FakePortsHarness()
    private val picks = ArrayDeque<List<String>?>()
    private val messages = mutableListOf<String>()
    private var toolsStarted = 0
    private var committed = ""

    private val catalogue =
        """{"layers":[{"id":"framework","title":"Framework","about":"how it is written"}],"items":[""" +
            """{"id":"node","kind":"language","name":"node","title":"TypeScript","layer":"language","language":"node"},""" +
            """{"id":"node/react","kind":"choice","name":"react","title":"React","axis":"framework",""" +
            """"layer":"framework","language":"node","default":true},""" +
            """{"id":"node/http","kind":"choice","name":"http","title":"HTTP","axis":"trigger","layer":"host","language":"node"}]}"""

    private fun checked(body: String) = fake.ok("""{"layers":[],"items":[],"check":$body}""")

    private fun plan(tools: String = "[]") =
        fake.ok(
            """{"layers":[],"items":[],"plan":{"command":"smith add app web --lang node --with react",""" +
                """"argv":["smith","add","app","web","--lang","node","--with","react"],""" +
                """"files":[{"action":"create","path":"apps/web/package.json","note":""}],"tools":$tools,"notes":[],""" +
                """"recipeArgv":["smith","recipe","new","--name","web"]}}""",
        )

    private fun compose(): Boolean {
        val pick =
            ComposeStackUseCase.PickStack { _, _, message ->
                messages.add(message)
                picks.removeFirstOrNull()
            }
        return ComposeStackUseCase.run(fake.ports, pick, { toolsStarted++ }) { committed = it }
    }

    init {
        fake.answers["compose --json"] = fake.ok(catalogue)
        fake.answers["compose --json --lang node --with react"] = checked("""{"language":"node","complete":true,"selection":["react"]}""")
        fake.answers["compose --json --lang node --with react --kind app --name web"] = plan()
        fake.texts.add("web")
    }

    @Test
    fun checksThePicksWithSmithThenNamesAndPlansBeforeAnythingIsWritten() {
        picks.add(listOf("node/react"))
        fake.plans.add(fake.yes())
        assertTrue(compose())
        assertEquals(
            listOf(
                listOf("compose", "--json"),
                listOf("compose", "--json", "--lang", "node", "--with", "react"),
                listOf("projects", "--json"),
                listOf("compose", "--json", "--lang", "node", "--with", "react", "--kind", "app", "--name", "web"),
                listOf("add", "app", "web", "--lang", "node", "--with", "react", "--yes"),
            ),
            fake.calls,
        )
        assertTrue(fake.plansSeen.single().startsWith("Run: smith add app web"))
        assertTrue(fake.infos.single().startsWith("web is added. Not committed"))
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun theCommitBoxAddsCommitAndTheCommitNowButtonStartsTheCommitOtherwise() {
        picks.add(listOf("node/react"))
        fake.plans.add(fake.yes(commit = true))
        assertTrue(compose())
        assertEquals(listOf("--commit", "--yes"), fake.calls.last().takeLast(2))
        assertEquals(emptyList(), fake.buttons.last())
    }

    @Test
    fun anUntickedCommitBoxOffersACommitButton() {
        picks.add(listOf("node/react"))
        fake.plans.add(fake.yes())
        compose()
        fake.pressLastButton()
        assertEquals("add app web", committed)
    }

    @Test
    fun aConflictGoesBackToThePickWithSmithsSentence() {
        fake.answers["compose --json --lang node --with react"] =
            checked("""{"language":"node","complete":false,"conflict":"React and Vue do not go together: tick one","selection":[]}""")
        picks.add(listOf("node/react"))
        assertFalse(compose())
        assertEquals(listOf("", "Conflict: React and Vue do not go together: tick one"), messages)
        assertTrue(fake.calls.none { "--yes" in it })
    }

    @Test
    fun nothingTickedEndsTheFlowWithoutAskingSmithToCheck() {
        picks.add(emptyList())
        assertFalse(compose())
        assertEquals(listOf(listOf("compose", "--json")), fake.calls)
    }

    @Test
    fun anOpenQuestionIsAskedWithTheSuggestedAnswerFirstAndTheAnswerIsTicked() {
        fake.answers["compose --json --lang node --with react"] =
            checked(
                """{"language":"node","complete":false,"question":{"choice":"azure-functions","about":"the trigger","axis":"trigger",""" +
                    """"options":[{"name":"timer"},{"name":"http","suggested":true}]},"selection":[]}""",
            )
        fake.ones.add("http")
        picks.add(listOf("node/react"))
        assertFalse(compose())
        assertEquals(listOf("http", "timer"), fake.rowsSeen.single().map { it.name })
        assertEquals(2, messages.size)
    }

    @Test
    fun keepingThePlanAsARecipeWritesTheRecipeAndNoProject() {
        picks.add(listOf("node/react"))
        fake.plans.add(PlanAnswerContract(false, emptyList(), 0))
        assertTrue(compose())
        assertEquals(listOf("recipe", "new", "--name", "web"), fake.calls.last())
        assertEquals("The recipe is written.", fake.infos.single())
    }

    @Test
    fun installingTheToolsFirstHandsOverAndWritesNothing() {
        fake.answers["compose --json --lang node --with react --kind app --name web"] =
            plan("""[{"choice":"react","name":"node","title":"Node.js","state":"missing","command":"smith tools install node"}]""")
        picks.add(listOf("node/react"))
        fake.plans.add(PlanAnswerContract(false, emptyList(), 0))
        assertFalse(compose())
        assertEquals(1, toolsStarted)
        assertTrue(fake.calls.none { "--yes" in it })
    }

    @Test
    fun cancelAtThePlanWritesNothing() {
        picks.add(listOf("node/react"))
        assertFalse(compose())
        assertTrue(fake.calls.none { "--yes" in it })
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aFailedAddIsAWarningThatNamesTheCommand() {
        picks.add(listOf("node/react"))
        fake.plans.add(fake.yes())
        fake.answers["add app web"] = fake.fail("✖ apps/web exists\n")
        assertFalse(compose())
        assertTrue(fake.warnings.single().contains("apps/web exists"))
        assertTrue(fake.warnings.single().contains("The command was: smith add app web"))
    }

    @Test
    fun aLanguageTickedAloneDoesNotSayThatAChoiceThatWritesTheProjectIsNeeded() {
        picks.add(listOf("node"))
        assertFalse(compose())
        assertEquals(
            listOf("", "Tick at least one choice. To start from the defaults of the stack, pick the defaults in Add project."),
            messages,
        )
        assertTrue(messages.none { "writes the project" in it })
    }
}
