package dev.monosmith.jetbrains.addproject

import kotlin.test.Test
import kotlin.test.assertEquals

class OutcomesJsonMapperTest {
    private val catalogue =
        """
        {"layers":[{"id":"framework","title":"Framework"},{"id":"shell","title":"Desktop shell"}],
         "items":[{"id":"language/node","kind":"language","name":"node","title":"TypeScript / JavaScript","description":"web and tools"},
                  {"id":"node/react","kind":"choice","name":"react"}],
         "outcomes":[
          {"id":"web","title":"A website or web app","kind":"app","typical":["React","Vue"],"consequence":"People open it.",
           "languages":["node","go"],"words":["website"],
           "items":[{"name":"react","title":"React","description":"a React app","language":"node","layer":"framework","aliases":["jsx"]},
                    {"name":"gin","language":"go","layer":"framework"}]},
          {"id":"library","title":"A library","kind":"lib","languages":["node"],"items":[]}]}
        """.trimIndent()

    @Test
    fun readsTheOutcomesTheLanguagesAndTheLayerTitles() {
        val (outcomes, languages, layers) = OutcomesJsonMapper.toCatalogue(catalogue)
        assertEquals(listOf("web", "library"), outcomes.map { it.id })
        assertEquals("React, Vue", outcomes.first().typical)
        assertEquals("TypeScript / JavaScript", languages.single().title)
        assertEquals(mapOf("framework" to "Framework", "shell" to "Desktop shell"), layers)
        assertEquals(
            "gin",
            outcomes
                .first()
                .items
                .last()
                .title,
        )
    }

    @Test
    fun theLanguageOfAKindIsTheIdInFrontOfItsRole() {
        val kinds =
            OutcomesJsonMapper.toKinds(
                """[{"kind":"node-app","language":"typescript","label":"Node app","description":"d"},{"kind":"","language":"x"}]""",
            )
        assertEquals(listOf("node"), kinds.map { it.language })
        assertEquals("app", kinds.single().role)
    }

    @Test
    fun keepsOnlyTheLanguagesThatHaveAKindOfTheRoleAndDropsAnEmptyOutcome() {
        val (outcomes, _, _) = OutcomesJsonMapper.toCatalogue(catalogue)
        val kinds =
            listOf(
                KindOptionContract("node-app", "node", "Node app", ""),
                KindOptionContract("go-lib", "go", "Go lib", ""),
            )
        val fitted = OutcomesJsonMapper.fitOutcomes(outcomes, kinds)
        assertEquals(listOf("web"), fitted.map { it.id })
        assertEquals(listOf("node"), fitted.single().languages)
        assertEquals(listOf("react"), fitted.single().items.map { it.name })
    }
}
