import type { LearningCue, MediaItem, PlaybackInfo, Pronunciation, ScanResult, SubtitleCue, SubtitleTrack } from '../types/media'

async function request<T>(url: string, init?: RequestInit): Promise<T> {
  const response = await fetch(url, init)
  if (!response.ok) {
    const body = (await response.json().catch(() => null)) as { error?: string; reason?: string } | null
    throw new Error(body?.error ?? body?.reason ?? `请求失败 (${response.status})`)
  }
  return response.json() as Promise<T>
}

export const mediaApi = {
  list: () => request<MediaItem[]>('/api/v1/media'),
  get: (id: number) => request<MediaItem>(`/api/v1/media/${id}`),
  scan: () => request<ScanResult>('/api/v1/media/scan', { method: 'POST' }),
  playback: (id: number) => request<PlaybackInfo>(`/api/v1/media/${id}/playback`),
  subtitles: (id: number) => request<SubtitleTrack[]>(`/api/v1/media/${id}/subtitles`),
  subtitleCues: (trackId: number) => request<SubtitleCue[]>(`/api/v1/subtitle-tracks/${trackId}/cues`),
  learningCues: (mediaId: number, englishTrackId: number, chineseTrackId: number) => {
    const query = new URLSearchParams({ englishTrackId: String(englishTrackId), chineseTrackId: String(chineseTrackId) })
    return request<LearningCue[]>(`/api/v1/media/${mediaId}/learning-cues?${query}`)
  },
  pronunciation: (word: string) => request<Pronunciation>(`/api/v1/dictionary/${encodeURIComponent(word)}`),
}
