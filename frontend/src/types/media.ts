export interface MediaStream {
  id: number
  index: number
  type: 'video' | 'audio' | 'subtitle'
  codec: string
  language?: string
  title?: string
  channels?: number
  sampleRate?: number
  width?: number
  height?: number
  isDefault: boolean
  isForced: boolean
}

export interface MediaItem {
  id: number
  path: string
  filename: string
  title: string
  container: string
  durationMs: number
  videoCodec: string
  width: number
  height: number
  fileSize: number
  updatedAt: string
  streams: MediaStream[]
}

export interface ScanResult {
  added: number
  updated: number
  unchanged: number
  failed: number
}

export interface PlaybackInfo {
  mode: 'direct' | 'remux' | 'unsupported'
  status: 'ready' | 'unsupported'
  url?: string
  audioStream?: number
  reason?: string
}

export interface SubtitleTrack {
  id: number
  mediaId: number
  sourceType: 'embedded' | 'external'
  streamIndex?: number
  format: string
  language?: string
  title?: string
  externalPath?: string
  offsetMs: number
  supported: boolean
  processed: boolean
}

export interface SubtitleCue {
  id: number
  trackId: number
  sequence: number
  startMs: number
  endMs: number
  text: string
  cleanText: string
}

export interface LearningCue {
  id: number
  sequence: number
  startMs: number
  endMs: number
  en: string
  zh?: string
  enCueIds: number[]
  zhCueIds: number[]
}

export interface Pronunciation {
  word: string
  phonetic?: string
  audioUrl?: string
  source: string
  found: boolean
  cached: boolean
}
