/**
 * The arguments to run smith with to create a workspace in the folder the
 * terminal opens in. `smith new --here` asks its questions, lists what it
 * will create and refuses a folder that has files, so the click writes nothing.
 */
export function toNewArguments (): string[] {
  return ['new', '--here']
}
