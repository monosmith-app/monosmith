import { describe, expect, it } from 'vitest'
import type { SmithRun } from '../smith-run/index.ts'
import type { LevelDecider, LevelQuestion, RecipeChoice } from './new-recipe.contract.ts'
import { newRecipe, type NewRecipeDeps } from './new-recipe.use-case.ts'
import { checkChoiceAnswer, checkChoiceQuestion, checkRecipeName, checkRecipeVersion, parseRecipeLevels, questionFor, recipeArguments } from './recipe-levels.mapper.ts'

const plugin = (name: string): { name: string, description: string } => ({ name, description: `about ${name}` })
const LEVELS: SmithRun = {
  code:   0,
  stderr: '',
  stdout: JSON.stringify({
    levels: [
      { level: 'language', additive: false, plugins: [plugin('node'), plugin('python')] },
      { level: 'lint', additive: false, plugins: [plugin('eslint'), plugin('ruff')] },
      { level: 'structure', additive: true, plugins: [plugin('layers'), plugin('slices')] },
    ],
    fits: {
      node:   { lint: ['eslint'], structure: ['layers', 'slices'] },
      python: { lint: ['ruff'], structure: ['slices'] },
    },
  }),
}

interface Fake { asked: LevelQuestion[], decided: string[], applied: string[][], calls: string[][], shown: Array<{ command: string, file: string }>, deps: NewRecipeDeps }

/** A fake prompts client: `picks` answers the level questions in order, `yes` answers the confirmation; `deciders` and `choices` answer who decides a level and the question of a chosen one (fixed when none is queued). */
const BASICS = { name: 'team', version: '0.1.0', extends: '' }

function fake (picks: Array<string[] | undefined>, yes = true, basics: typeof BASICS | null = BASICS, run: SmithRun = LEVELS, deciders: LevelDecider[] = [], choices: Array<RecipeChoice | undefined> = []): Fake {
  const f: Fake = { asked: [], decided: [], applied: [], calls: [], shown: [], deps: undefined as never }
  const queue = [...picks]
  f.deps = {
    run: async (args) => {
      f.calls.push(args)

      return run
    },
    apply: async (args) => {
      f.applied.push(args)

      return { code: 0, stdout: 'Wrote team.recipe.toml (recipe team 0.1.0).\n', stderr: '' }
    },
    askBasics: async () => basics ?? undefined,
    decide:    async (question) => {
      f.decided.push(question.level)

      return deciders.shift() ?? 'fixed'
    },
    askChoice: async () => choices.shift(),
    pick:      async (question) => {
      f.asked.push(question)

      return queue.shift()
    },
    confirm: async (_request, command, file) => {
      f.shown.push({ command, file })

      return yes
    },
  }

  return f
}

describe('parseRecipeLevels and questionFor', () => {
  it('reads the levels and offers after a language only the plugins that fit it', () => {
    const levels = parseRecipeLevels(LEVELS.stdout)
    const lint = levels.levels[1]
    expect(questionFor(levels, lint, '')?.options.map((p) => p.name)).toEqual(['eslint', 'ruff'])
    expect(questionFor(levels, lint, 'python')?.options.map((p) => p.name)).toEqual(['ruff'])
    expect(questionFor(levels, levels.levels[2], 'python')).toMatchObject({ multi: true, options: [{ name: 'slices' }] })
    expect(questionFor(levels, lint, 'ruby')).toBeUndefined()
  })

  it('refuses text that is not a list of levels', () => {
    expect(() => parseRecipeLevels('[]')).toThrow(TypeError)
    expect(() => parseRecipeLevels('nope')).toThrow(SyntaxError)
  })
})

describe('recipeArguments and the live checks', () => {
  it('builds the command, joining the plugins of a level that takes several', () => {
    expect(recipeArguments({ name: 'team', version: '1.2.0', extends: 'base.recipe.toml', fixed: { language: ['python'], structure: ['layers', 'slices'], lint: [] } }))
      .toEqual(['recipe', 'new', '--name', 'team', '--recipe-version', '1.2.0', '--extends', 'base.recipe.toml', '--fixed', 'language=python', '--fixed', 'structure=layers,slices'])
    expect(recipeArguments({ name: 'team', version: '', extends: '', fixed: {} })).toEqual(['recipe', 'new', '--name', 'team'])
  })

  it('refuses a question with a colon and an answer the command cannot read', () => {
    expect(checkChoiceQuestion('Which: one?')).toContain('colon')
    expect(checkChoiceQuestion('Which one?')).toBeUndefined()
    expect(checkChoiceAnswer('a=b', [])).toContain('without')
    expect(checkChoiceAnswer('Python', ['Python'])).toContain('Two answers')
    expect(checkChoiceAnswer('Python', ['node'])).toBeUndefined()
  })

  it('refuses the names and versions the command refuses', () => {
    expect(checkRecipeName('Bad Name')).toContain('lowercase')
    expect(checkRecipeName('team-node')).toBeUndefined()
    expect(checkRecipeVersion('1.0')).toContain('major.minor.patch')
    expect(checkRecipeVersion('')).toBeUndefined()
  })
})

describe('newRecipe', () => {
  it('reads the levels from smith, shows the command and the file, and writes only on a yes', async () => {
    // language python, lint ruff, structure left unset (nothing ticked).
    const f = fake([['python'], ['ruff'], []])
    const outcome = await newRecipe(f.deps)
    expect(f.calls).toEqual([['recipe', 'new', '--list-levels', '--json']])
    expect(f.asked.map((q) => [q.level, q.multi, q.language])).toEqual([['language', false, ''], ['lint', false, 'python'], ['structure', true, 'python']])
    expect(f.asked[1].options.map((p) => p.name)).toEqual(['ruff'])
    expect(f.shown).toEqual([{ command: 'smith recipe new --name team --recipe-version 0.1.0 --fixed language=python --fixed lint=ruff', file: 'team.recipe.toml' }])
    expect(f.applied).toEqual([['recipe', 'new', '--name', 'team', '--recipe-version', '0.1.0', '--fixed', 'language=python', '--fixed', 'lint=ruff']])
    expect(outcome).toMatchObject({ kind: 'written', file: 'team.recipe.toml' })
  })

  it('lets the users choose a level: the question and the plugin of each answer become --chosen', async () => {
    // language: chosen by the users; lint: fixed (all plugins are offered, as no language is fixed); structure: ticks.
    const choice: RecipeChoice = { question: 'Which language?', answers: [{ answer: 'JavaScript', plugin: 'node' }, { answer: 'Python', plugin: 'python' }] }
    const f = fake([[], []], true, BASICS, LEVELS, ['chosen'], [choice])
    await newRecipe(f.deps)
    expect(f.decided).toEqual(['language', 'lint', 'structure'])
    expect(f.asked.map((q) => q.level)).toEqual(['lint', 'structure'])
    expect(f.asked[0].options.map((p) => p.name)).toEqual(['eslint', 'ruff'])
    expect(f.shown[0].command).toContain('--chosen "language=Which language?:JavaScript=node,Python=python"')
    expect(f.applied[0]).toEqual(['recipe', 'new', '--name', 'team', '--recipe-version', '0.1.0', '--chosen', 'language=Which language?:JavaScript=node,Python=python'])
  })

  it('lets the users choose a level that takes any number of plugins, an answer giving several joined by +', async () => {
    const choice: RecipeChoice = { question: 'Which layout?', answers: [{ answer: 'simple', plugin: 'layers' }, { answer: 'both', plugin: 'layers+slices' }] }
    const f = fake([['node'], []], true, BASICS, LEVELS, ['fixed', 'chosen'], [choice])
    await newRecipe(f.deps)
    expect(f.decided).toEqual(['language', 'structure'])
    expect(f.applied[0]).toContain('structure=Which layout?:simple=layers,both=layers+slices')
  })

  it('asks nothing more when the person cancels the choice', async () => {
    const f = fake([], true, BASICS, LEVELS, ['chosen'], [undefined])
    expect(await newRecipe(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.applied).toEqual([])
  })

  it('asks who decides a level of two plugins or more, and none for a level with one plugin', async () => {
    const f = fake([['node'], [], ['layers']])
    await newRecipe(f.deps)
    expect(f.decided).toEqual(['language', 'structure'])
  })

  it('ticks several plugins of an additive level', async () => {
    const f = fake([['node'], [], ['layers', 'slices']])
    await newRecipe(f.deps)
    expect(f.applied[0]).toContain('structure=layers,slices')
  })

  it('writes nothing when the person says no', async () => {
    const f = fake([['python'], [], []], false)
    expect(await newRecipe(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.applied).toEqual([])
  })

  it('writes nothing when the person cancels a question or the basics', async () => {
    const f = fake([['python'], undefined])
    expect(await newRecipe(f.deps)).toEqual({ kind: 'cancelled' })
    expect(f.shown).toEqual([])
    expect(f.applied).toEqual([])
    expect(await newRecipe(fake([], true, null).deps)).toEqual({ kind: 'cancelled' })
  })

  it('returns the words of smith when it refuses the levels or the recipe', async () => {
    expect(await newRecipe(fake([], true, undefined, { code: 1, stdout: '', stderr: '× no such flag: --list-levels\n' }).deps)).toEqual({ kind: 'refused', message: 'no such flag: --list-levels' })
    expect(await newRecipe(fake([], true, undefined, { code: 0, stdout: 'plain text', stderr: '' }).deps)).toMatchObject({ kind: 'refused' })
    const f = fake([['python'], [], []])
    f.deps.apply = async () => ({ code: 1, stdout: '', stderr: '× team.recipe.toml exists: left as it is\n' })
    expect(await newRecipe(f.deps)).toEqual({ kind: 'refused', message: 'team.recipe.toml exists: left as it is' })
  })
})
