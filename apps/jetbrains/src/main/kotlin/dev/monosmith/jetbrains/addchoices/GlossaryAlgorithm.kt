package dev.monosmith.jetbrains.addchoices

/** The one-line "what is this" that sits under a question: the layer's own words from the catalogue of `smith compose --json`. */
object GlossaryAlgorithm {
    /**
     * @param layer the layer or axis id (`shell`, `host`, `test`)
     * @param titles the title of each layer ("Desktop shell")
     * @param about what each layer means, from the catalogue
     * @return "Desktop shell = what wraps the app as a program you install on a computer.", or an empty text when the catalogue says nothing
     */
    fun line(
        layer: String,
        titles: Map<String, String>,
        about: Map<String, String>,
    ): String {
        val meaning = about[layer]?.trim().orEmpty()
        if (meaning.isEmpty()) return ""
        return "${(titles[layer] ?: layer).lowercase()} = ${meaning.removeSuffix(".")}."
    }

    /** Puts the glossary line after [header], on its own line, when there is one. */
    fun withLine(
        header: String,
        glossary: String,
    ): String = if (glossary.isEmpty()) header else "$header\n$glossary"
}
