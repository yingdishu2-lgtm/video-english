import { describe, expect, it } from 'vitest'
import { arpabetToIPA } from './pronunciation'

describe('arpabetToIPA', () => {
  it('converts vowels, consonants and lexical stress to IPA', () => {
    expect(arpabetToIPA('HH AH0 L OW1')).toBe('həlˈoʊ')
    expect(arpabetToIPA('G R AE1 V AH0 T IY0')).toBe('ɡɹˈævəti')
  })
})
