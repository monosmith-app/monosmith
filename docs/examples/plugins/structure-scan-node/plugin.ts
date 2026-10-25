#!/usr/bin/env node
// A smith structure plugin in TypeScript on @monosmith/plugin-sdk whose doctor
// walks the files of widget projects: it stops when smith cancels the request
// and it takes four scans at the same time. It only speaks the plugin protocol.
import { serve } from '@monosmith/plugin-sdk'
import { createScanPlugin } from './scan.ts'

if (process.env.SMITH_PLUGIN === '1') {
  // WIDGET_SCAN_DELAY_MS is a pause before each file, which stands in for a slow disk so a test can cancel a scan in the middle.
  const delay = Number(process.env.WIDGET_SCAN_DELAY_MS)
  await serve(createScanPlugin(Number.isFinite(delay) && delay > 0 ? delay : 0))
} else {
  console.error('plugin.ts is a smith plugin: smith starts it and speaks the plugin protocol on its stdin and stdout')
  process.exitCode = 1
}
