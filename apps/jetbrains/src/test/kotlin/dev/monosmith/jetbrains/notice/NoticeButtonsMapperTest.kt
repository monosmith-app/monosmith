package dev.monosmith.jetbrains.notice

import kotlin.test.Test
import kotlin.test.assertEquals

class NoticeButtonsMapperTest {
    @Test
    fun showsThreeButtonsOrFewerAsTheyAre() {
        assertEquals(listOf("a", "b", "c") to emptyList(), NoticeButtonsMapper.split(listOf("a", "b", "c")))
        assertEquals(listOf("a") to emptyList(), NoticeButtonsMapper.split(listOf("a")))
    }

    @Test
    fun putsTheRestBehindMoreWhenThereAreMoreThanThree() {
        assertEquals(listOf("a", "b") to listOf("c", "d", "e"), NoticeButtonsMapper.split(listOf("a", "b", "c", "d", "e")))
    }

    @Test
    fun neverShowsMoreThanTheLimitWithTheMoreButton() {
        val (shown, more) = NoticeButtonsMapper.split((1..9).toList())
        assertEquals(NoticeButtonsMapper.LIMIT - 1, shown.size)
        assertEquals(9, shown.size + more.size)
    }
}
