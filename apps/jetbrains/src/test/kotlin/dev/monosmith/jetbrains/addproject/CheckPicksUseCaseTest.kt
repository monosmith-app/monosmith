package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.addchoices.ChoiceOptionContract
import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals

class CheckPicksUseCaseTest {
    private val fake = FakePortsHarness()

    private fun choice(
        name: String,
        layer: String,
        kinds: List<String> = emptyList(),
    ) = ChoiceOptionContract(layer, name, "", "node", emptyList(), false, false, emptyList(), false, emptyList(), "", layer, kinds)

    private val options =
        AddOptionsContract(
            kinds = emptyList(),
            choices =
                listOf(
                    choice("html", "framework", listOf("app", "package")),
                    choice("react", "framework", listOf("app")),
                    choice("svelte", "framework", listOf("app")),
                    choice("vue", "framework", listOf("app")),
                    choice("eslint", "lint"),
                    choice("jest", "test"),
                ),
            outcomes =
                listOf(
                    OutcomeContract(
                        "web",
                        "A website",
                        "app",
                        listOf("node"),
                        "",
                        "",
                        emptyList(),
                        listOf(OutcomeItemContract("react", "React", "", "node", "framework", emptyList())),
                    ),
                ),
            languages = emptyList(),
            layers = emptyMap(),
        )

    private fun says(check: String) {
        fake.answers["compose --json"] = fake.ok("""{"check":$check}""")
    }

    private fun check(
        picks: List<String>,
        outcomeId: String? = null,
        role: String? = null,
    ) = CheckPicksUseCase.check(fake.ports.smith, options, PicksAskedContract("node", picks, outcomeId, role))

    @Test
    fun asksSmithWithTheSharedFlagsAndStillAsksAWebsiteForItsFramework() {
        says("""{"complete":true}""")
        assertEquals(
            "this kind of project needs a framework: tick one of html, react, svelte, vue, or untick everything",
            check(listOf("eslint", "jest"), "web", "app"),
        )
        assertEquals(listOf(listOf("compose", "--json", "--lang", "node", "--with", "eslint,jest")), fake.calls)
    }

    @Test
    fun lintAndTestChoicesAloneAreFineForAPackage() {
        says("""{"complete":true}""")
        assertEquals("", check(listOf("eslint", "jest"), null, "package"))
    }

    @Test
    fun namesOnlyTheChoicesThatFitTheKind() {
        says("""{"complete":true}""")
        assertEquals(
            "this kind of project needs a framework: tick one of html, or untick everything",
            check(listOf("eslint"), "web", "package"),
        )
    }

    @Test
    fun aTickedChoiceOfTheLayerIsFine() {
        says("""{"complete":true}""")
        assertEquals("", check(listOf("react", "eslint"), "web", "app"))
    }

    @Test
    fun saysAConflictAndWhatIsMissingAsSmithWordsItWithoutFlags() {
        says("""{"conflict":"react and vue are both frameworks"}""")
        assertEquals("react and vue are both frameworks", check(listOf("react", "vue"), null, "app"))
        says("""{"missing":"add chakra to --with, or give no --with to start"}""")
        assertEquals("add chakra, or untick everything to start", check(listOf("chakra"), null, "app"))
    }

    @Test
    fun anOpenQuestionNothingTickedAndAFailingSmithAreLeftToTheFormAndToSmithAdd() {
        says("""{"question":{"axis":"trigger"}}""")
        assertEquals("", check(listOf("react"), "web", "app"))
        assertEquals("", check(emptyList()))
        fake.answers["compose --json"] = fake.fail("boom")
        assertEquals("", check(listOf("eslint"), "web", "app"))
    }
}
