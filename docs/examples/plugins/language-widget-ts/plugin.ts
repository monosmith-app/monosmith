#!/usr/bin/env node
// A smith language plugin for "widget" projects (a directory with a widget.json),
// in TypeScript on @monosmith/plugin-sdk. Started by smith (SMITH_PLUGIN=1) it
// serves the plugin protocol on stdin and stdout; started by a person, or by one
// of the tasks it contributes, it is a small command line tool.
import { serve } from '@monosmith/plugin-sdk'
import { runTool } from './tool.ts'
import { createWidgetPlugin } from './widget.ts'

if (process.env.SMITH_PLUGIN === '1') {
  await serve(createWidgetPlugin())
} else {
  try {
    runTool(process.argv.slice(2))
  } catch (error) {
    console.error(error instanceof Error ? error.message : String(error))
    process.exitCode = 1
  }
}
