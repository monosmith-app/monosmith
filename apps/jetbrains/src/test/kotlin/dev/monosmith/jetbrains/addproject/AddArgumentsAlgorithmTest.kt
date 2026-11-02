package dev.monosmith.jetbrains.addproject

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertTrue

class AddArgumentsAlgorithmTest {
    private val request = AddRequestContract("app", "web", "node", with = listOf("react", "jest"))

    @Test
    fun aPlanIsADryRunAndChangesNothing() {
        assertEquals(
            listOf("add", "app", "web", "--lang", "node", "--with", "react,jest", "--dry-run"),
            AddArgumentsAlgorithm.arguments(request, apply = false),
        )
    }

    @Test
    fun theRealRunPassesYesAndNeverCommitsUnasked() {
        val args = AddArgumentsAlgorithm.arguments(request, apply = true)
        assertEquals("--yes", args.last())
        assertFalse("--commit" in args)
    }

    @Test
    fun commitIsAddedOnlyWhenTheBoxWasTicked() {
        assertEquals(listOf("--yes", "--commit"), AddArgumentsAlgorithm.arguments(request, apply = true, commit = true).takeLast(2))
        assertFalse("--commit" in AddArgumentsAlgorithm.arguments(request, apply = false, commit = true))
    }

    @Test
    fun freshWinsOverLikeAndStackAndEnvironmentsAreCarried() {
        val full =
            request.copy(
                stack = "web",
                like = "api",
                environments = EnvironmentsContract(listOf("lcl", "prd"), "prd", "workspace", "commit"),
            )
        val like = AddArgumentsAlgorithm.arguments(full, apply = false)
        assertTrue("--like" in like && "--fresh" !in like)
        assertEquals(listOf("--stack", "web"), like.slice(like.indexOf("--stack")..like.indexOf("--stack") + 1))
        assertTrue("--environments" in like && "lcl,prd" in like)
        val fresh = AddArgumentsAlgorithm.arguments(full, apply = false, fresh = true)
        assertTrue("--fresh" in fresh && "--like" !in fresh)
    }

    @Test
    fun readsWhetherThePlanFollowsTheSiblings() {
        assertTrue(AddArgumentsAlgorithm.followsSiblings("test: vitest (follows api)"))
        assertTrue(AddArgumentsAlgorithm.followsSiblings("The 3 siblings agree on jest"))
        assertFalse(AddArgumentsAlgorithm.followsSiblings("test: jest"))
    }
}
