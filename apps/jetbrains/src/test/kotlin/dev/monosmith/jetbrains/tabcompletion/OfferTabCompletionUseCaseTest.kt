package dev.monosmith.jetbrains.tabcompletion

import dev.monosmith.jetbrains.smithrun.SmithRunContract
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNull
import kotlin.test.assertTrue

class OfferTabCompletionUseCaseTest {
    private class FakePorts(
        var memory: String? = null,
        var statusRun: SmithRunContract = SmithRunContract(0, """{"shell":"zsh","wanted":true,"answer":""}""", ""),
    ) : TabCompletionPortsContract {
        var asked = false
        var statusCalls = 0
        var terminals = 0
        var setUp: () -> Unit = {}
        var never: () -> Unit = {}
        var shown = 0

        override fun remembered() = memory

        override fun remember(answer: String) {
            memory = answer
        }

        override fun askedThisSession() = asked

        override fun markAsked() {
            asked = true
        }

        override fun status(): SmithRunContract {
            statusCalls++
            return statusRun
        }

        override fun ask(
            setUp: () -> Unit,
            never: () -> Unit,
        ) {
            shown++
            this.setUp = setUp
            this.never = never
        }

        override fun openTerminal() {
            terminals++
        }
    }

    @Test
    fun theOfferWritesNothingItself_setUpOpensTheTerminalAndIsRemembered() {
        val ports = FakePorts()
        assertTrue(OfferTabCompletionUseCase.offer(ports))
        assertEquals(0, ports.terminals)
        ports.setUp()
        assertEquals(1, ports.terminals)
        assertEquals(TabCompletionOfferPolicy.DONE, ports.memory)
    }

    @Test
    fun dontAskAgainIsRememberedAndOpensNoTerminal() {
        val ports = FakePorts()
        OfferTabCompletionUseCase.offer(ports)
        ports.never()
        assertEquals(TabCompletionOfferPolicy.NEVER, ports.memory)
        assertEquals(0, ports.terminals)
        assertFalse(OfferTabCompletionUseCase.offer(FakePorts(memory = ports.memory)))
    }

    @Test
    fun notNowRemembersNothingButTheOfferIsNotRepeatedInTheSession() {
        val ports = FakePorts()
        assertTrue(OfferTabCompletionUseCase.offer(ports))
        assertNull(ports.memory)
        assertFalse(OfferTabCompletionUseCase.offer(ports))
        assertEquals(1, ports.shown)
    }

    @Test
    fun noOfferWhenSmithSaysItIsNotWantedFailsOrIsTooOldToAnswer() {
        val runs =
            listOf(
                SmithRunContract(0, """{"shell":"zsh","wanted":false,"answer":"installed"}""", ""),
                SmithRunContract(1, "", "unknown command"),
                SmithRunContract(0, "unknown command", ""),
            )
        for (run in runs) {
            val ports = FakePorts(statusRun = run)
            assertFalse(OfferTabCompletionUseCase.offer(ports))
            assertEquals(0, ports.shown)
        }
    }

    @Test
    fun smithIsNotEvenAskedAfterAnAnswerThatIsRemembered() {
        val ports = FakePorts(memory = TabCompletionOfferPolicy.NEVER)
        assertFalse(OfferTabCompletionUseCase.offer(ports))
        assertEquals(0, ports.statusCalls)
    }
}
