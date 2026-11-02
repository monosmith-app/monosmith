package dev.monosmith.jetbrains.addproject

/** Reads the output of a failed `smith add` to tell a project that stays on disk from a plain refusal. */
object AddFailureMapper {
    private val titles =
        mapOf(
            "dotnet" to "the .NET SDK",
            "go" to "Go",
            "node" to "Node.js",
            "npm" to "Node.js",
            "npx" to "Node.js",
            "python" to "Python",
            "python3" to "Python",
            "pip" to "Python",
            "cargo" to "Rust",
            "rustc" to "Rust",
            "java" to "Java (JDK)",
            "javac" to "Java (JDK)",
            "gradle" to "Gradle",
            "mvn" to "Maven",
            "flutter" to "Flutter",
            "dart" to "Dart",
            "git" to "Git",
            "docker" to "Docker",
            "cmake" to "CMake",
        )

    /**
     * Tells whether the project was written and then failed its doctor checks (a tool or a system library
     * is missing). Such a project stays on disk, and smith holds back the commit.
     *
     * @param output what smith printed, standard output then standard error
     * @return the facts, or null for any other failure
     */
    fun doctorFailure(output: String): DoctorFailureContract? {
        val name = Regex("the new project (\\S+) was written but fails smith doctor").find(output)?.groupValues?.get(1) ?: return null
        val lines = output.lines()
        val start = lines.indexOfFirst { it.startsWith("The new project $name fails smith doctor") }
        val rest = if (start == -1) emptyList() else lines.drop(start)
        val block = mutableListOf<String>()
        for (line in rest) {
            if (block.isNotEmpty() && line.isNotBlank() && !line.first().isWhitespace()) break
            block.add(line.trimEnd())
        }
        return DoctorFailureContract(
            name = name,
            files =
                Regex("Wrote (\\d+) file")
                    .find(output)
                    ?.groupValues
                    ?.get(1)
                    ?.toInt() ?: 0,
            reasons = block.filter { it.trimStart().startsWith("✖") }.map { it.trimStart().removePrefix("✖").trim() },
            committed = Regex("^Committed\\b", RegexOption.MULTILINE).containsMatchIn(output),
        )
    }

    /**
     * Tells whether `smith add` failed because a program is missing, which leaves the new project on disk.
     *
     * @param output what smith printed, standard output then standard error
     * @return the program and what was made, or null for any other failure
     */
    fun missingTool(output: String): MissingToolContract? {
        val program =
            Regex("(\\S+) is not installed or not on PATH")
                .find(output)
                ?.groupValues
                ?.get(1)
                ?.lowercase()
                ?.removeSuffix(".exe")
                ?: return null
        return MissingToolContract(
            title = titles[program] ?: program,
            files =
                Regex("Wrote (\\d+) file")
                    .find(output)
                    ?.groupValues
                    ?.get(1)
                    ?.toInt() ?: 0,
            url = Regex("Install it from (https?://[^\\s,]+)").find(output)?.groupValues?.get(1),
            dir =
                Regex("^Creating \\S+ \\S+ in (\\S+?):", RegexOption.MULTILINE)
                    .find(output)
                    ?.groupValues
                    ?.get(1)
                    ?.removeSuffix("/"),
            check =
                Regex("^Verifying: smith (run .+ -p \\S+)$", RegexOption.MULTILINE)
                    .find(output)
                    ?.groupValues
                    ?.get(1)
                    ?.split(" "),
        )
    }

    /** The text of the notice for a failing project: what was made, what could not be checked and why, and whether the commit happened. */
    fun describe(failure: DoctorFailureContract): String {
        val files = if (failure.files > 0) " (${failure.files} files)" else ""
        val why = if (failure.reasons.isEmpty()) "a check of smith doctor fails" else failure.reasons.joinToString("; ")
        val commit =
            if (failure.committed) {
                "It was committed."
            } else {
                "Nothing was committed: the files are in your work tree, so the Commit window shows them."
            }
        return "The project ${failure.name} was created$files, but smith could not finish checking it: $why. $commit"
    }
}
