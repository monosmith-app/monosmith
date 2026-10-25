#!/usr/bin/env node
// A smith structure plugin in TypeScript on @monosmith/plugin-sdk: it adds a
// layout-check task to widget projects and makes their lint wait for it.
//
// Started by smith (SMITH_PLUGIN=1) it serves the plugin protocol on stdin and
// stdout. Started with `layout-check`, which is the command of the task it
// contributes, it checks the folder it runs in. The SDK frames the messages,
// derives the capabilities from the methods of the plugin (layout.ts) and keeps
// stdout for the protocol: anything meant for a human goes to stderr (console.error).
import { serve } from '@monosmith/plugin-sdk'
import { createLayoutPlugin, findJunkDrawers } from './layout.ts'

if (process.env.SMITH_PLUGIN === '1') {
  await serve(createLayoutPlugin())
} else if (process.argv[2] === 'layout-check') {
  const found = findJunkDrawers('.')
  if (found.length > 0) {
    console.error(`folders named for nothing: ${found.join(', ')} (name them for what they do)`)
    process.exitCode = 1
  } else {
    console.log('no junk-drawer folders')
  }
} else {
  console.error('usage: plugin.ts layout-check (smith runs this program as a plugin)')
  process.exitCode = 1
}
