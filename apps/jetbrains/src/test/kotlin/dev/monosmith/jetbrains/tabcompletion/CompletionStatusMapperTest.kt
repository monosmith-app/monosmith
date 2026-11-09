package dev.monosmith.jetbrains.tabcompletion

import kotlin.test.Test
import kotlin.test.assertEquals
import kotlin.test.assertFalse
import kotlin.test.assertNull
import kotlin.test.assertTrue

class CompletionStatusMapperTest {
    @Test
    fun readsWantedFromWhatSmithPrints() {
        assertEquals(true, CompletionStatusMapper.wanted("""{"shell":"zsh","wanted":true,"answer":""}"""))
        assertEquals(false, CompletionStatusMapper.wanted("""{"shell":"zsh","wanted":false,"answer":"no"}"""))
    }

    @Test
    fun sayNothingForTextThatIsNotTheStatus() {
        assertNull(CompletionStatusMapper.wanted(""))
        assertNull(CompletionStatusMapper.wanted("unknown command \"status\""))
        assertNull(CompletionStatusMapper.wanted("""{"shell":"zsh"}"""))
        assertNull(CompletionStatusMapper.wanted("""{"wanted":"yes"}"""))
    }

    @Test
    fun theOfferPolicyNeedsAllThreeFacts() {
        assertTrue(TabCompletionOfferPolicy.shouldOffer(null, askedThisSession = false, wanted = true))
        assertFalse(TabCompletionOfferPolicy.shouldOffer(null, askedThisSession = false, wanted = false))
        assertFalse(TabCompletionOfferPolicy.shouldOffer(TabCompletionOfferPolicy.DONE, askedThisSession = false, wanted = true))
        assertFalse(TabCompletionOfferPolicy.shouldOffer(TabCompletionOfferPolicy.NEVER, askedThisSession = false, wanted = true))
        assertFalse(TabCompletionOfferPolicy.shouldOffer(null, askedThisSession = true, wanted = true))
    }
}
