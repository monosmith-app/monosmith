package dev.monosmith.jetbrains.tauriapp

import dev.monosmith.jetbrains.flowports.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class TauriAppUseCaseTest {
    private val fake = FakePortsHarness()
    private var committed = ""

    @Test
    fun theIconsAreMadeFromOnePngAfterTheFilesTheyOverwriteAreShown() {
        fake.answers["tauri icon logo.png --dry-run"] = fake.ok("would overwrite src-tauri/icons/icon.png\n")
        fake.texts.addAll(listOf("logo.png", ""))
        fake.plans.add(fake.yes())
        assertTrue(TauriAppUseCase.icons(fake.ports))
        assertEquals("would overwrite src-tauri/icons/icon.png", fake.plansSeen.single())
        assertEquals(listOf("tauri", "icon", "logo.png", "--yes"), fake.calls.last())
    }

    @Test
    fun theAppIsNamedWithDashPOnlyWhenTheAnswerIsNotEmpty() {
        fake.texts.addAll(listOf("art/logo.PNG", "desktop"))
        fake.plans.add(fake.yes())
        TauriAppUseCase.icons(fake.ports)
        assertEquals(listOf("tauri", "icon", "art/logo.PNG", "-p", "desktop", "--yes"), fake.calls.last())
    }

    @Test
    fun cancelAtTheAppQuestionMakesNoIcons() {
        fake.texts.add("logo.png")
        assertFalse(TauriAppUseCase.icons(fake.ports))
        assertEquals(emptyList(), fake.calls)
    }

    @Test
    fun makingAnAppATauriAppAsksTheBundleIdentifierAndShowsThePlan() {
        fake.answers["tauri add web --identifier com.acme.web --dry-run"] = fake.ok("would write src-tauri/Cargo.toml\n")
        fake.texts.addAll(listOf("web", "com.acme.web"))
        fake.plans.add(fake.yes())
        assertTrue(TauriAppUseCase.add(fake.ports) { committed = it })
        assertEquals("would write src-tauri/Cargo.toml", fake.plansSeen.single())
        assertEquals(listOf("tauri", "add", "web", "--identifier", "com.acme.web", "--yes"), fake.calls.last())
        assertTrue(fake.infos.single().startsWith("web is a Tauri app. Not committed"))
    }

    @Test
    fun anEmptyIdentifierLeavesTheChoiceToSmithsDefault() {
        fake.texts.addAll(listOf("web", ""))
        fake.plans.add(fake.yes(commit = true))
        TauriAppUseCase.add(fake.ports) {}
        assertEquals(listOf("tauri", "add", "web", "--yes", "--commit"), fake.calls.last())
    }

    @Test
    fun anUntickedCommitBoxOffersTheCommitButton() {
        fake.texts.addAll(listOf("web", ""))
        fake.plans.add(fake.yes())
        TauriAppUseCase.add(fake.ports) { committed = it }
        fake.pressLastButton()
        assertEquals("tauri add web", committed)
    }

    @Test
    fun cancelChangesNothing() {
        fake.texts.addAll(listOf("web", ""))
        assertFalse(TauriAppUseCase.add(fake.ports) {})
        assertTrue(fake.calls.none { "--yes" in it })
    }
}
