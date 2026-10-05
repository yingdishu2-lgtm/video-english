import { create } from 'zustand'

interface PlayerState {
  currentMs: number
  durationMs: number
  playbackRate: number
  setCurrentMs: (value: number) => void
  setDurationMs: (value: number) => void
  setPlaybackRate: (value: number) => void
  reset: () => void
}

export const usePlayerStore = create<PlayerState>((set) => ({
  currentMs: 0,
  durationMs: 0,
  playbackRate: 1,
  setCurrentMs: (currentMs) => set({ currentMs }),
  setDurationMs: (durationMs) => set({ durationMs }),
  setPlaybackRate: (playbackRate) => set({ playbackRate }),
  reset: () => set({ currentMs: 0, durationMs: 0, playbackRate: 1 }),
}))

