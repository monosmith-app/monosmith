import { inflateRawSync } from 'node:zlib'

/**
 * One file's bytes from a zip archive, found through the central directory. Only what smith's
 * release archives use is read: stored and deflated entries, no zip64, no encryption.
 *
 * @param {Buffer} zip - The archive's bytes.
 * @param {(name: string) => boolean} wanted - Picks the entry by its path in the archive.
 * @returns {Buffer | null} The entry's bytes, or null when no entry is wanted.
 * @throws {Error} When the archive is not a zip this reader can read.
 */
export function readEntry (zip, wanted) {
  const end = zip.lastIndexOf(Buffer.from([0x50, 0x4B, 0x05, 0x06]))
  if (end === -1) throw new Error('not a zip archive')
  const count = zip.readUInt16LE(end + 10)
  let at = zip.readUInt32LE(end + 16)
  for (let i = 0; i < count; i++) {
    if (zip.readUInt32LE(at) !== 0x02014B50) throw new Error('a broken zip central directory')
    const method = zip.readUInt16LE(at + 10)
    const compressed = zip.readUInt32LE(at + 20)
    const nameLength = zip.readUInt16LE(at + 28)
    const extra = zip.readUInt16LE(at + 30)
    const comment = zip.readUInt16LE(at + 32)
    const local = zip.readUInt32LE(at + 42)
    const name = zip.subarray(at + 46, at + 46 + nameLength).toString('utf8')
    at += 46 + nameLength + extra + comment
    if (!wanted(name)) continue
    const start = local + 30 + zip.readUInt16LE(local + 26) + zip.readUInt16LE(local + 28)
    const data = zip.subarray(start, start + compressed)
    if (method === 0) return Buffer.from(data)
    if (method === 8) return inflateRawSync(data)
    throw new Error(`zip method ${method} is not supported`)
  }

  return null
}
