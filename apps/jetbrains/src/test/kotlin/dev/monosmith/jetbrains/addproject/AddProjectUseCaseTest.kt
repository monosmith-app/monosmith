package dev.monosmith.jetbrains.addproject

import dev.monosmith.jetbrains.flowports.testing.FakePortsHarness
import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class AddProjectUseCaseTest {
    private val fake = FakePortsHarness()
    private var undone = 0
    private var committed = ""
    private var tools = 0
    private val follow =
        AddFollowUpsContract(undo = {
            undone++
        }, commit = { committed = it }, useInApp = {}, installTools = { tools++ }, openHealth = {}, openPage = {})

    init {
        fake.answers["kinds --json"] =
            fake.ok(
                """[{"kind":"node-app","language":"typescript","label":"Node app","description":"d"},{"kind":"go-lib","language":"go","label":"Go lib"}]""",
            )
        fake.answers["choices --json"] =
            fake.ok(
                """[{"axis":"test","name":"jest","default":true,"language":"node"},{"axis":"test","name":"vitest","language":"node"}]""",
            )
        fake.answers["compose --json"] =
            fake.ok(
                """{"layers":[{"id":"framework","title":"Framework","about":"how the app is written"},{"id":"test","title":"Test","about":"the tool that runs the tests"}],
                   "items":[{"kind":"language","name":"node","title":"TypeScript"}],
                   "outcomes":[{"id":"web","title":"A website","kind":"app","languages":["node"],"typical":["React"],"consequence":"People open it.",
                     "items":[{"name":"react","title":"React","language":"node","layer":"framework"}]}]}""",
            )
        fake.answers["add app web --lang node --with react,jest"] = fake.ok("would write 12 files\n")
    }

    private fun answerEveryQuestion(commit: Boolean) {
        fake.ones.addAll(listOf("web", "react", "jest", "prd", "workspace", "commit"))
        fake.texts.addAll(listOf("web", "lcl, prd"))
        fake.plans.add(fake.yes(commit))
    }

    private fun addCalls() = fake.calls.filter { it.first() == "add" }

    @Test
    fun asksTheOutcomeFirstThenTheAxesShowsThePlanAndRunsWithYes() {
        answerEveryQuestion(commit = false)
        assertTrue(AddProjectUseCase.run(fake.ports, emptyList(), follow) {})
        assertEquals("add app web --lang node --with react,jest", addCalls().first().take(7).joinToString(" "))
        assertEquals("--dry-run", addCalls().first().last())
        assertEquals("--yes", addCalls().last().last())
        assertTrue(fake.plansSeen.single().contains("would write 12 files"))
        assertTrue(fake.infos.single().contains("web is added"))
    }

    @Test
    fun theTestAxisStartsOnItsDefaultAndTheGlossaryLineIsInTheQuestion() {
        answerEveryQuestion(commit = false)
        AddProjectUseCase.run(fake.ports, emptyList(), follow) {}
        val testRows = fake.rowsSeen.first { rows -> rows.any { it.name == "jest" } }
        assertEquals(listOf("jest", "vitest"), testRows.map { it.name })
        assertEquals("(default)", testRows.first().description)
        assertTrue(fake.headersSeen.any { it.endsWith("test = the tool that runs the tests.") })
        assertTrue(fake.headersSeen.any { it.endsWith("framework = how the app is written.") })
    }

    @Test
    fun neverCommitsUnlessTheBoxIsTicked() {
        answerEveryQuestion(commit = false)
        AddProjectUseCase.run(fake.ports, emptyList(), follow) {}
        assertFalse("--commit" in addCalls().last())
        assertTrue(fake.infos.single().contains("Not committed"))
        assertEquals(
            "Commit now",
            fake.buttons
                .single()
                .first()
                .label,
        )
    }

    @Test
    fun aTickedBoxPassesCommitAndOffersUndo() {
        answerEveryQuestion(commit = true)
        AddProjectUseCase.run(fake.ports, emptyList(), follow) {}
        assertEquals(listOf("--yes", "--commit"), addCalls().last().takeLast(2))
        fake.buttons
            .single()
            .first { it.label == "Undo" }
            .action()
        assertEquals(1, undone)
    }

    @Test
    fun thePlanTheDialogShowsDoesNotEndWithTheDryRunLine() {
        assertEquals(
            "Plan: add app web",
            AddProjectUseCase.withoutDryRunLine("Plan: add app web\nDry run: nothing was changed.\n"),
        )
    }

    @Test
    fun aCommitThatSmithSaysItDidNotMakeIsNotOfferedAnUndo() {
        answerEveryQuestion(commit = true)
        val environments = "--environments lcl,prd --default-environment prd --environment-output workspace --env-files commit"
        fake.answers["add app web --lang node --with react,jest $environments --yes --commit"] =
            fake.ok("Done: app web is ready.\nNot committed. Your work tree had other changes.\n")
        AddProjectUseCase.run(fake.ports, emptyList(), follow) {}
        assertTrue(fake.infos.single().contains("Not committed"))
        assertEquals(
            "Commit now",
            fake.buttons
                .single()
                .first()
                .label,
        )
    }

    @Test
    fun theCommitNowButtonStartsTheCommitOfThisOperation() {
        answerEveryQuestion(commit = false)
        AddProjectUseCase.run(fake.ports, emptyList(), follow) {}
        fake.buttons
            .single()
            .first()
            .action()
        assertEquals("add app web", committed)
    }

    @Test
    fun aProjectThatFailsDoctorIsCreatedNotRefusedAndHasThreeLabelsToChooseFrom() {
        answerEveryQuestion(commit = false)
        val environments = "--environments lcl,prd --default-environment prd --environment-output workspace --env-files commit"
        fake.answers["add app web --lang node --with react,jest $environments --yes"] =
            fake.fail(
                "Wrote 12 files\nsmith: the new project web was written but fails smith doctor\nThe new project web fails smith doctor:\n  ✖ node is old\n",
            )
        assertFalse(AddProjectUseCase.run(fake.ports, emptyList(), follow) {})
        assertTrue(fake.warnings.single().startsWith("The project web was created (12 files)"))
        assertEquals(listOf("Install tools", "Open Health", "Commit anyway", "Keep uncommitted"), fake.buttons.single().map { it.label })
        fake.buttons
            .single()
            .first()
            .action()
        assertEquals(1, tools)
    }

    @Test
    fun composeAsksForTheComposerAndAddsNothing() {
        fake.ones.add("\u0000compose")
        var composed = false
        assertFalse(AddProjectUseCase.run(fake.ports, emptyList(), follow) { composed = true })
        assertTrue(composed)
        assertEquals(emptyList(), addCalls())
    }

    @Test
    fun anOldSmithThatCannotListItsChoicesAddsNothing() {
        fake.answers["choices --json"] = fake.fail("✖ unknown command")
        assertFalse(AddProjectUseCase.run(fake.ports, emptyList(), follow) {})
        assertTrue(fake.warnings.single().startsWith("Nothing was added."))
    }

    @Test
    fun cancelAtTheFirstQuestionChangesNothing() {
        assertFalse(AddProjectUseCase.run(fake.ports, emptyList(), follow) {})
        assertEquals(0, fake.changedCount)
    }
}
