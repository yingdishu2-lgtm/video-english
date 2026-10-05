export interface EnglishToken {
  text: string
  word: boolean
}

const wordPattern = /[A-Za-z]+(?:['’][A-Za-z]+)*/g

export function tokenizeEnglish(text: string): EnglishToken[] {
  const tokens: EnglishToken[] = []
  let cursor = 0
  for (const match of text.matchAll(wordPattern)) {
    const index = match.index ?? 0
    if (index > cursor) tokens.push({ text: text.slice(cursor, index), word: false })
    tokens.push({ text: match[0], word: true })
    cursor = index + match[0].length
  }
  if (cursor < text.length) tokens.push({ text: text.slice(cursor), word: false })
  return tokens
}
