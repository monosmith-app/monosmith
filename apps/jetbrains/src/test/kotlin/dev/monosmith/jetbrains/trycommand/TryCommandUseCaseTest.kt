package dev.monosmith.jetbrains.trycommand

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class TryCommandUseCaseTest {
    private val fake = FakePortsHarness()
    private val passed =
        """{"command":["add","lib","utils"],"commandOk":true,"changed":["libs/utils/package.json"],""" +
            """"checks":[{"name":"build","ok":true}]}"""
    private val failed = """{"command":["add","lib","utils"],"commandOk":true,"checks":[{"name":"test","ok":false,"tail":"1 failed"}]}"""

    @Test
    fun triesInACopyShowsWhatItDidAndRunsForRealOnlyAfterAYes() {
        fake.answers["try add lib utils --json"] = fake.ok(passed)
        fake.texts.add("smith add lib utils")
        fake.plans.add(fake.yes())
        assertTrue(TryCommandUseCase.run(fake.ports))
        assertTrue(fake.plansSeen.single().startsWith("Try: smith add lib utils"))
        assertEquals(listOf("try", "add", "lib", "utils", "--yes"), fake.calls.last())
        assertTrue(fake.shown.last())
        assertEquals("The command ran for real.", fake.infos.single())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun theTryItselfNeverUsesYesSoTheFilesOfThePersonStayAsTheyAre() {
        fake.answers["try add lib utils --json"] = fake.ok(passed)
        fake.texts.add("add lib utils")
        TryCommandUseCase.run(fake.ports)
        assertEquals(listOf(listOf("try", "add", "lib", "utils", "--json")), fake.calls)
    }

    @Test
    fun aFailedCheckIsShownButNotOfferedForReal() {
        fake.answers["try add lib utils --json"] = fake.ok(failed)
        fake.texts.add("add lib utils")
        fake.plans.add(fake.yes())
        assertFalse(TryCommandUseCase.run(fake.ports))
        assertTrue(fake.plansSeen.single().contains("test: failed"))
        assertTrue(fake.calls.none { "--yes" in it })
    }

    @Test
    fun noAnswerAfterAPassedTryRunsNothing() {
        fake.answers["try add lib utils --json"] = fake.ok(passed)
        fake.texts.add("add lib utils")
        assertFalse(TryCommandUseCase.run(fake.ports))
        assertEquals(0, fake.changedCount)
    }

    @Test
    fun aTryThatSmithRefusedIsAWarningWithSmithsWords() {
        fake.answers["try add lib utils --json"] = fake.fail("× this folder has no smith.toml\n")
        fake.texts.add("add lib utils")
        assertFalse(TryCommandUseCase.run(fake.ports))
        assertEquals("Nothing was tried. this folder has no smith.toml", fake.warnings.single())
    }

    @Test
    fun cancelAtTheCommandTriesNothing() {
        assertFalse(TryCommandUseCase.run(fake.ports))
        assertEquals(emptyList(), fake.calls)
    }

    @Test
    fun aFailedRealRunIsAWarning() {
        fake.answers["try add lib utils --json"] = fake.ok(passed)
        fake.answers["try add lib utils --yes"] = fake.fail("× libs/utils exists\n")
        fake.texts.add("add lib utils")
        fake.plans.add(fake.yes())
        assertFalse(TryCommandUseCase.run(fake.ports))
        assertEquals("Nothing was changed. libs/utils exists", fake.warnings.single())
    }
}
