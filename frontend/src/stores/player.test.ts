import { beforeEach, describe, expect, it } from 'vitest'
import { usePlayerStore } from './player'

describe('player store', () => {
  beforeEach(() => usePlayerStore.getState().reset())

  it('tracks playback time and resets transient state', () => {
    usePlayerStore.getState().setCurrentMs(1250)
    usePlayerStore.getState().setDurationMs(9000)
    usePlayerStore.getState().setPlaybackRate(1.5)
    expect(usePlayerStore.getState()).toMatchObject({ currentMs: 1250, durationMs: 9000, playbackRate: 1.5 })

    usePlayerStore.getState().reset()
    expect(usePlayerStore.getState()).toMatchObject({ currentMs: 0, durationMs: 0, playbackRate: 1 })
  })
})

