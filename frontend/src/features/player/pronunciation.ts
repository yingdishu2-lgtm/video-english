import type { Pronunciation } from '../../types/media'

const phones: Record<string, string> = {
  AA: 'ɑ', AE: 'æ', AO: 'ɔ', AW: 'aʊ', AY: 'aɪ', EH: 'ɛ', EY: 'eɪ',
  IH: 'ɪ', IY: 'i', OW: 'oʊ', OY: 'ɔɪ', UH: 'ʊ', UW: 'u',
  B: 'b', CH: 'tʃ', D: 'd', DH: 'ð', F: 'f', G: 'ɡ', HH: 'h', JH: 'dʒ',
  K: 'k', L: 'l', M: 'm', N: 'n', NG: 'ŋ', P: 'p', R: 'ɹ', S: 's',
  SH: 'ʃ', T: 't', TH: 'θ', V: 'v', W: 'w', Y: 'j', Z: 'z', ZH: 'ʒ',
}

export function arpabetToIPA(pronunciation: string) {
  const output: string[] = []
  for (const raw of pronunciation.split('#')[0].trim().split(/\s+/)) {
    const match = raw.match(/^([A-Z]+)([012])?$/)
    if (!match) continue
    const [, phone, stress] = match
    let ipa: string | undefined
    if (phone === 'AH') ipa = stress === '0' ? 'ə' : 'ʌ'
    else if (phone === 'ER') ipa = stress === '0' ? 'ɚ' : 'ɝ'
    else ipa = phones[phone]
    if (!ipa) continue
    const stressMark = stress === '1' ? 'ˈ' : stress === '2' ? 'ˌ' : ''
    output.push(stressMark + ipa)
  }
  return output.join('')
}

export async function localPronunciation(input: string): Promise<Pronunciation> {
  const word = input.toLowerCase().replaceAll('’', "'")
  const { dictionary } = await import('cmu-pronouncing-dictionary')
  const arpabet = dictionary[word]
  const phonetic = arpabet ? `/${arpabetToIPA(arpabet)}/` : undefined
  return {
    word,
    phonetic,
    source: 'CMU Pronouncing Dictionary（美式）',
    found: Boolean(phonetic),
    cached: true,
  }
}
