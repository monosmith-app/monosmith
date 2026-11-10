package dev.monosmith.jetbrains.addchoices

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertNotNull
import kotlin.test.assertNull

class PickChoicesUseCaseTest {
    private val fake = FakePortsHarness()

    private fun choice(
        name: String,
        axis: String,
        kinds: List<String> = emptyList(),
        isDefault: Boolean = false,
    ) = ChoiceOptionContract(axis, name, "", "node", emptyList(), isDefault, false, emptyList(), false, emptyList(), "", "", kinds)

    private val choices =
        listOf(
            choice("aws-lambda", "framework", listOf("app")),
            choice("esm", "framework", listOf("package")),
            choice("eslint", "lint", isDefault = true),
        )

    @Test
    fun aPackageIsNotOfferedAwsLambda() {
        fake.ones.addAll(listOf("", "eslint"))
        PickChoicesUseCase.pick(fake.ports.prompt, choices, "node", emptyList(), null, emptyMap(), role = "package")
        assertEquals(listOf("", "esm"), fake.rowsSeen.first().map { it.name })
    }

    @Test
    fun saysTheProblemAndStartsTheQuestionsAgainWhenThePersonAgrees() {
        val problem = "this kind of project needs a framework: tick one of esm, or untick everything"
        val problems = ArrayDeque(listOf(problem, ""))
        fake.ones.addAll(listOf("", "eslint", "again", "esm", "eslint"))
        val picked =
            PickChoicesUseCase.pick(fake.ports.prompt, choices, "node", emptyList(), null, emptyMap(), role = "package") {
                problems.removeFirst()
            }
        assertEquals(listOf("esm", "eslint"), picked)
        assertEquals(problem, fake.headersSeen.single { it == problem })
    }

    @Test
    fun cancelsWhenThePersonDoesNotWantToPickAgain() {
        fake.ones.addAll(listOf("", "eslint", "cancel"))
        assertNull(
            PickChoicesUseCase.pick(fake.ports.prompt, choices, "node", emptyList(), null, emptyMap(), role = "package") { "a problem" },
        )
        assertNotNull(fake.headersSeen.firstOrNull { it == "a problem" })
    }
}
