import { createHash, createPublicKey, verify } from 'node:crypto'

/**
 * The public key smith's releases are signed with (minisign, key ID 6B2CE84DD9F901). Its private
 * half is the repository secret SMITH_SIGNING_KEY (ADR 0101).
 */
export const RELEASE_PUBLIC_KEY = 'RWQB+dlN6CxrAMNDX47n/BQK3mOTgrJ4yVIasbrbP6xRRZTQeB5yncav'

/** The DER prefix that turns 32 raw Ed25519 public key bytes into an SPKI key. */
const ED25519_SPKI = Buffer.from('302a300506032b6570032100', 'hex')

/**
 * An Ed25519 public key object from its 32 raw bytes.
 *
 * @param {Buffer} raw - The key bytes.
 * @returns {import('node:crypto').KeyObject} The key.
 */
function ed25519Key (raw) {
  return createPublicKey({ key: Buffer.concat([ED25519_SPKI, raw]), format: 'der', type: 'spki' })
}

/**
 * Checks a minisign signature (the plain Ed25519 kind smith release writes) of a message: the
 * key ID, the signature of the message, and the signature of the trusted comment.
 *
 * @param {Buffer} message - The signed file's bytes (checksums.txt).
 * @param {string} signatureText - The .minisig file's text.
 * @param {string} publicKeyText - The public key in base64, as minisign prints it.
 * @returns {boolean} True only when every check passes.
 */
export function verifyMinisign (message, signatureText, publicKeyText) {
  const lines = signatureText.split(/\r?\n/).filter(line => line !== '')
  const signatureLine = lines.find(line => !line.startsWith('untrusted comment:') && !line.startsWith('trusted comment:'))
  const trusted = lines.find(line => line.startsWith('trusted comment: '))
  const global = lines.at(-1)
  if (!signatureLine || !trusted || !global || global === signatureLine) return false
  const sig = Buffer.from(signatureLine, 'base64')
  const pub = Buffer.from(publicKeyText.trim(), 'base64')
  // Ed + key ID (8) + signature (64); Ed + key ID (8) + key (32).
  if (sig.length !== 74 || pub.length !== 42 || sig.subarray(0, 2).toString() !== 'Ed' || !sig.subarray(2, 10).equals(pub.subarray(2, 10))) return false
  const key = ed25519Key(pub.subarray(10))
  const messageSignature = sig.subarray(10)
  if (!verify(null, message, key, messageSignature)) return false
  const comment = Buffer.from(trusted.slice('trusted comment: '.length))

  return verify(null, Buffer.concat([messageSignature, comment]), key, Buffer.from(global, 'base64'))
}

/**
 * The SHA-256 a checksums.txt lists for a file.
 *
 * @param {string} checksums - The text of checksums.txt (`<sha256>  <name>` lines).
 * @param {string} name - The file name.
 * @returns {string | null} The hex digest, or null when the file is not listed.
 */
export function listedDigest (checksums, name) {
  for (const line of checksums.split(/\r?\n/)) {
    const [digest, file] = line.trim().split(/\s+\*?/, 2)
    if (file === name) return digest.toLowerCase()
  }

  return null
}

/**
 * Checks a downloaded archive against a signed checksums.txt.
 *
 * @param {object} download - What was downloaded.
 * @param {Buffer} download.archive - The archive's bytes.
 * @param {string} download.name - The archive's file name.
 * @param {Buffer} download.checksums - checksums.txt.
 * @param {string} download.signature - checksums.txt.minisig.
 * @param {string} [publicKey] - The public key; default the release key.
 * @returns {string | null} Why the archive is not trusted, or null when it is.
 */
export function checkArchive ({ archive, name, checksums, signature }, publicKey = RELEASE_PUBLIC_KEY) {
  if (!verifyMinisign(checksums, signature, publicKey)) return 'the signature of checksums.txt does not match smith\'s release key'
  const listed = listedDigest(checksums.toString('utf8'), name)
  if (listed === null) return `checksums.txt does not list ${name}`
  const actual = createHash('sha256').update(archive).digest('hex')

  return actual === listed ? null : `${name} does not match its checksum (got ${actual}, listed ${listed})`
}
