package dev.monosmith.jetbrains.addproject

/**
 * How a web app is built once per environment.
 *
 * @property names the environments (`lcl`, `dev`, `uat`, `prd`)
 * @property default the one `smith run build` builds
 * @property output `workspace` or `project`: where the builds go
 * @property envFiles `commit` or `example`: whether the .env files are committed
 */
data class EnvironmentsContract(
    val names: List<String>,
    val default: String,
    val output: String,
    val envFiles: String,
)

/**
 * What the person asked to add.
 *
 * @property role `app`, `lib` or `package`
 * @property stack a named stack (`--stack`), or null
 * @property with the names of the choices (`--with`)
 * @property fresh ignore the siblings and use the stack defaults (`--fresh`)
 * @property like the project whose shape the new one copies (`--like`), or null
 */
data class AddRequestContract(
    val role: String,
    val name: String,
    val language: String,
    val stack: String? = null,
    val with: List<String> = emptyList(),
    val fresh: Boolean = false,
    val like: String? = null,
    val environments: EnvironmentsContract? = null,
)

/** What the questions of Add project ended in. */
sealed interface AddAnswerContract {
    /** The person answered every question. */
    data class Request(
        val request: AddRequestContract,
        val choicesInWords: String,
    ) : AddAnswerContract

    /** The person chose to compose a stack instead. */
    data object Compose : AddAnswerContract
}

/** How `smith add` ended, for the notice. */
sealed interface AddOutcomeContract {
    /** The project is added; [committed] tells whether the person asked for the commit. */
    data class Added(
        val name: String,
        val role: String,
        val operation: String,
        val committed: Boolean,
    ) : AddOutcomeContract

    /** The project is on disk and fails `smith doctor`. */
    data class Failing(
        val name: String,
        val operation: String,
        val failure: DoctorFailureContract,
    ) : AddOutcomeContract

    /** The project is on disk and could not be checked because a program is missing. */
    data class Unchecked(
        val name: String,
        val tool: MissingToolContract,
    ) : AddOutcomeContract

    /** smith refused; [message] has its words and the fix. */
    data class Refused(
        val message: String,
    ) : AddOutcomeContract
}

/**
 * A project that smith wrote and then found failing its own `smith doctor` checks.
 *
 * @property files how many files smith wrote; 0 when unknown
 * @property reasons what could not be checked or failed, one short sentence each
 * @property committed whether the commit that was asked for happened
 */
data class DoctorFailureContract(
    val name: String,
    val files: Int,
    val reasons: List<String>,
    val committed: Boolean,
)

/**
 * A program the check of a new project needs and the computer does not have.
 *
 * @property title what a person calls it ("Node.js")
 * @property url the page that says how to get it, or null
 * @property check the first check of the new project, as the words after `smith`, or null
 */
data class MissingToolContract(
    val title: String,
    val files: Int,
    val url: String?,
    val dir: String?,
    val check: List<String>?,
)
