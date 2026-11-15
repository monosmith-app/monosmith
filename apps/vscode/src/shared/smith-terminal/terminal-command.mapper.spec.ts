import { describe, expect, it } from 'vitest'
import { shellFamily, terminalLaunch, toTerminalCommand } from './terminal-command.mapper.ts'

describe('shellFamily', () => {
  it('tells the shells apart by their executable', () => {
    expect(shellFamily(String.raw`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`)).toBe('powershell')
    expect(shellFamily('/usr/bin/pwsh')).toBe('powershell')
    expect(shellFamily(String.raw`C:\Windows\System32\cmd.exe`)).toBe('cmd')
    expect(shellFamily('/bin/zsh')).toBe('posix')
    expect(shellFamily('')).toBe('posix')
  })
})

describe('terminalLaunch', () => {
  it('shows smith new --here and puts the folder of smith on the PATH', () => {
    expect(terminalLaunch('/tmp/ws/node_modules/.bin/smith', ['new', '--here'], '/bin/bash')).toEqual({ line: 'smith new --here', directory: '/tmp/ws/node_modules/.bin' })
    expect(terminalLaunch(String.raw`C:\ws\node_modules\.bin\smith.cmd`, ['up'], 'cmd.exe')).toEqual({ line: 'smith up', directory: String.raw`C:\ws\node_modules\.bin` })
  })

  it('keeps the full quoted path for a smith with another name or with no folder', () => {
    expect(terminalLaunch('/opt/smith-2/bin/smith-beta', ['up'], '/bin/bash').line).toBe('\'/opt/smith-2/bin/smith-beta\' up')
    expect(terminalLaunch('smith', ['up'], '/bin/bash')).toEqual({ line: '\'smith\' up' })
  })
})

describe('toTerminalCommand', () => {
  it('quotes the executable for PowerShell', () => {
    expect(toTerminalCommand(String.raw`C:\Program Files\smith\smith.exe`, ['up', 'eslint'], 'pwsh.exe')).toBe(String.raw`& 'C:\Program Files\smith\smith.exe' up eslint`)
  })

  it('quotes the executable for cmd', () => {
    expect(toTerminalCommand(String.raw`C:\smith\smith.cmd`, ['up'], 'cmd.exe')).toBe(String.raw`"C:\smith\smith.cmd" up`)
  })

  it('quotes the executable for a POSIX shell, even one with a quote in its path', () => {
    expect(toTerminalCommand("/home/o'neil/smith", ['up'], '/bin/bash')).toBe(String.raw`'/home/o'\''neil/smith' up`)
  })
})
