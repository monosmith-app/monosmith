import type { CommandHost } from '../command-host/index.ts'
import { newRecipe } from './new-recipe.use-case.ts'
import { askChoice, askRecipeBasics, confirmRecipe, decideLevel, pickLevel, tellRecipeOutcome } from './recipe-prompts.client.ts'
import { say } from '../words/index.ts'

/**
 * Asks the name, then a pick list or tick boxes for each level, shows the `smith recipe new` command and the file, and writes on a yes.
 * @param host - the folders and the Output channel of the extension
 * @param argument - what the editor passed to the command
 */
export async function newRecipeHere (host: CommandHost, argument: unknown): Promise<void> {
  const state = await host.folderFor(argument)
  if (state?.location === undefined || !host.mayWrite(say('vscode.new-recipe.writing-a-recipe'))) return
  host.showOutput()
  const outcome = await newRecipe({
    run:       host.smithIn(state),
    apply:     host.smithIn(state, true),
    askBasics: askRecipeBasics,
    pick:      pickLevel,
    decide:    decideLevel,
    askChoice,
    confirm:   confirmRecipe,
  })
  host.report('recipe', state, outcome)
  tellRecipeOutcome(outcome, state.root, host.note)
  if (outcome.kind === 'written') await host.refreshFolder(state.root, 'full', false)
}
