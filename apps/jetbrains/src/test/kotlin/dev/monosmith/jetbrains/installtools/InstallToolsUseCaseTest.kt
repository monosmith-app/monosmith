package dev.monosmith.jetbrains.installtools

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class InstallToolsUseCaseTest {
    private val fake = FakePortsHarness()

    private fun plans(admin: Boolean) =
        fake.ok(
            """[{"name":"node","title":"Node.js","why":"runs node projects","consequence":"node projects cannot run",""" +
                """"download":"https://nodejs.org","state":"missing",""" +
                """"default":"install:winget","options":[""" +
                """{"kind":"page","value":"page","label":"Open the page","hint":"download it yourself"},""" +
                """{"kind":"install","value":"install:winget","manager":"winget",""" +
                """"command":["winget","install","OpenJS.NodeJS"],"admin":$admin,"label":"winget","hint":"installs it","note":""},""" +
                """{"kind":"myself","value":"myself","label":"I will do it","hint":"shows the commands"},""" +
                """{"kind":"skip","value":"skip","label":"Skip","hint":"leave it"}]}]""",
        )

    init {
        fake.answers["tools install --dry-run --json"] = plans(admin = false)
    }

    @Test
    fun theMarkedDefaultIsFirstAndEveryRowShowsItsCommand() {
        fake.ones.add("install:winget")
        InstallToolsUseCase.run(fake.ports)
        val rows = fake.rowsSeen.single()
        assertEquals("install:winget", rows.first().name)
        assertTrue(rows.first().detail.startsWith("Runs: winget install OpenJS.NodeJS."))
    }

    @Test
    fun installsOnlyTheWayThePersonPickedAndNeverAsksForAdminWhenItIsNotNeeded() {
        fake.ones.add("install:winget")
        assertTrue(InstallToolsUseCase.run(fake.ports))
        assertEquals(listOf("tools", "install", "node", "--manager", "winget"), fake.calls.last())
        assertTrue(fake.shown.last())
        assertEquals("Node.js is installed.", fake.infos.single())
        assertEquals(1, fake.changedCount)
    }

    @Test
    fun aCommandThatNeedsAdminRightsIsAskedAboutAgainAndSkippingIsTheFirstAnswer() {
        fake.answers["tools install --dry-run --json"] = plans(admin = true)
        fake.ones.addAll(listOf("install:winget", "run"))
        InstallToolsUseCase.run(fake.ports)
        assertEquals(listOf("skip", "myself", "run"), fake.rowsSeen.last().map { it.name })
        assertEquals(listOf("tools", "install", "node", "--manager", "winget", "--with-admin"), fake.calls.last())
    }

    @Test
    fun runningItMyselfRunsNothing() {
        fake.answers["tools install --dry-run --json"] = plans(admin = true)
        fake.ones.addAll(listOf("install:winget", "myself"))
        assertTrue(InstallToolsUseCase.run(fake.ports))
        assertEquals(1, fake.calls.size)
        assertTrue(fake.infos.single().startsWith("Run this yourself, then restart the IDE terminal: winget install OpenJS.NodeJS"))
    }

    @Test
    fun theOfficialPageOpensInTheBrowserAndNothingIsInstalled() {
        fake.ones.add("page")
        InstallToolsUseCase.run(fake.ports)
        assertEquals(listOf("https://nodejs.org"), fake.opened)
        assertEquals(1, fake.calls.size)
        assertTrue(fake.infos.single().startsWith("Opened https://nodejs.org."))
    }

    @Test
    fun skippingNamesWhatTheWorkspaceLoses() {
        fake.ones.add("skip")
        InstallToolsUseCase.run(fake.ports)
        assertEquals("Skipped Node.js: node projects cannot run.", fake.infos.single())
    }

    @Test
    fun doingItMyselfShowsTheCommandsAndTheDownloadPage() {
        fake.ones.add("myself")
        fake.plans.add(fake.yes())
        InstallToolsUseCase.run(fake.ports)
        assertEquals("winget install OpenJS.NodeJS\nhttps://nodejs.org", fake.plansSeen.single())
    }

    @Test
    fun aFailedInstallIsToldAsPartOfTheAnswer() {
        fake.ones.add("install:winget")
        fake.answers["tools install node --manager winget"] = fake.fail("× winget is not installed\n")
        InstallToolsUseCase.run(fake.ports)
        assertTrue(fake.infos.single().startsWith("The install of Node.js did not finish. winget is not installed"))
    }

    @Test
    fun withEveryToolThereNothingIsAsked() {
        fake.answers["tools install --dry-run --json"] = fake.ok("[]")
        assertFalse(InstallToolsUseCase.run(fake.ports))
        assertEquals("Every tool the workspace needs is installed.", fake.infos.single())
    }

    @Test
    fun cancelAtTheFirstQuestionInstallsNothing() {
        assertFalse(InstallToolsUseCase.run(fake.ports))
        assertEquals(1, fake.calls.size)
        assertEquals(emptyList(), fake.infos)
    }

    @Test
    fun aPlanThatIsNotAPlanIsAWarning() {
        fake.answers["tools install --dry-run --json"] = fake.ok("hello")
        assertFalse(InstallToolsUseCase.run(fake.ports))
        assertEquals("smith tools install printed something that is not a plan.", fake.warnings.single())
    }
}
