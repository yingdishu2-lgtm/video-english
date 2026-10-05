import { describe, expect, it } from 'vitest'
import { tokenizeEnglish } from './tokenize'

describe('tokenizeEnglish', () => {
  it('keeps punctuation and contractions while marking words', () => {
    expect(tokenizeEnglish("What's here?")).toEqual([
      { text: "What's", word: true },
      { text: ' ', word: false },
      { text: 'here', word: true },
      { text: '?', word: false },
    ])
  })
})
