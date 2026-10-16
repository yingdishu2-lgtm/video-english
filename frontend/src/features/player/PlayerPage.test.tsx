// @vitest-environment jsdom

import '@testing-library/jest-dom/vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { cleanup, fireEvent, render, screen, waitFor } from '@testing-library/react'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { MemoryRouter, Route, Routes } from 'react-router-dom'
import { usePlayerStore } from '../../stores/player'
import { PlayerPage } from './PlayerPage'

const apiMocks = vi.hoisted(() => ({
  get: vi.fn(),
  playback: vi.fn(),
  subtitles: vi.fn(),
  learningCues: vi.fn(),
  pronunciation: vi.fn(),
  localPronunciation: vi.fn(),
}))

class ResizeObserverMock {
  static instances: ResizeObserverMock[] = []

  readonly observed = new Set<Element>()
  private readonly callback: ResizeObserverCallback

  constructor(callback: ResizeObserverCallback) {
    this.callback = callback
    ResizeObserverMock.instances.push(this)
  }

  observe = vi.fn((target: Element) => {
    this.observed.add(target)
  })

  unobserve = vi.fn((target: Element) => {
    this.observed.delete(target)
  })

  disconnect = vi.fn(() => {
    this.observed.clear()
  })

  trigger(targets: Element[] = [...this.observed]) {
    const entries = targets.map((target) => ({ target })) as ResizeObserverEntry[]
    this.callback(entries, this as unknown as ResizeObserver)
  }
}

vi.mock('../../api/client', () => ({
  mediaApi: {
    get: apiMocks.get,
    playback: apiMocks.playback,
    subtitles: apiMocks.subtitles,
    learningCues: apiMocks.learningCues,
    pronunciation: apiMocks.pronunciation,
  },
}))

vi.mock('./pronunciation', () => ({ localPronunciation: apiMocks.localPronunciation }))

function renderPlayer() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false, gcTime: 0 } },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={['/watch/7']}>
        <Routes>
          <Route path="/watch/:mediaId" element={<PlayerPage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

describe('PlayerPage mobile learning interactions', () => {
  beforeEach(() => {
    ResizeObserverMock.instances = []
    usePlayerStore.getState().reset()
    apiMocks.get.mockResolvedValue({
      id: 7,
      path: '/media/lesson.mp4',
      filename: 'lesson.mp4',
      title: 'Lesson',
      container: 'mp4',
      durationMs: 3_000,
      videoCodec: 'h264',
      width: 1280,
      height: 720,
      fileSize: 1_024,
      updatedAt: '2026-10-15T00:00:00Z',
      streams: [],
    })
    apiMocks.playback.mockResolvedValue({ mode: 'direct', status: 'ready', url: '/media/lesson.mp4' })
    apiMocks.subtitles.mockResolvedValue([
      { id: 11, mediaId: 7, sourceType: 'embedded', streamIndex: 1, format: 'srt', language: 'en', title: 'English', offsetMs: 0, supported: true, processed: true },
      { id: 12, mediaId: 7, sourceType: 'embedded', streamIndex: 2, format: 'srt', language: 'zh', title: '中文', offsetMs: 0, supported: true, processed: true },
    ])
    apiMocks.learningCues.mockResolvedValue([
      { id: 21, sequence: 1, startMs: 0, endMs: 1_000, en: 'Hello there.', zh: '你好。', enCueIds: [31], zhCueIds: [41] },
      { id: 22, sequence: 2, startMs: 1_000, endMs: 2_000, en: 'Keep learning.', zh: '继续学习。', enCueIds: [32], zhCueIds: [42] },
    ])
    apiMocks.localPronunciation.mockResolvedValue({
      word: 'Hello',
      phonetic: 'həˈloʊ',
      source: 'local',
      found: true,
      cached: true,
    })
    apiMocks.pronunciation.mockResolvedValue({
      word: 'Hello',
      source: 'remote',
      found: false,
      cached: false,
    })
    Object.defineProperty(HTMLElement.prototype, 'scrollIntoView', {
      configurable: true,
      value: vi.fn(),
    })
    vi.stubGlobal('matchMedia', vi.fn().mockReturnValue({
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    }))
    vi.stubGlobal('ResizeObserver', ResizeObserverMock)
  })

  afterEach(() => {
    cleanup()
    vi.restoreAllMocks()
    vi.unstubAllGlobals()
  })

  it('keeps word lookup separate from explicit cue controls and enables inline video', async () => {
    const { container } = renderPlayer()
    const word = await screen.findByRole('button', { name: '查看 Hello 的发音' })
    const video = container.querySelector('video')

    expect(video).not.toBeNull()
    expect(video).toHaveAttribute('playsinline')
    expect(video?.currentTime).toBe(0)

    word.focus()
    fireEvent.keyDown(word, { key: 'Enter' })
    expect(video?.currentTime).toBe(0)

    fireEvent.click(word)
    expect(await screen.findByRole('dialog', { name: 'Hello' })).toBeInTheDocument()
    expect(video?.currentTime).toBe(0)

    fireEvent.click(container.querySelectorAll('.cue-en')[1])
    expect(video?.currentTime).toBe(0)

    fireEvent.click(screen.getByRole('button', { name: '跳转到 00:01' }))
    expect(video?.currentTime).toBe(1)
  })

  it('swaps the subtitle tracks without requesting the same track twice', async () => {
    renderPlayer()
    const englishTrack = await screen.findByLabelText('英文轨')
    const chineseTrack = screen.getByLabelText('中文轨')

    await waitFor(() => {
      expect(englishTrack).toHaveValue('11')
      expect(chineseTrack).toHaveValue('12')
      expect(apiMocks.learningCues).toHaveBeenCalledWith(7, 11, 12)
    })

    fireEvent.change(englishTrack, { target: { value: '12' } })
    await waitFor(() => {
      expect(englishTrack).toHaveValue('12')
      expect(chineseTrack).toHaveValue('11')
      expect(apiMocks.learningCues).toHaveBeenCalledWith(7, 12, 11)
    })
    for (const [, englishTrackId, chineseTrackId] of apiMocks.learningCues.mock.calls) {
      expect(englishTrackId).not.toBe(chineseTrackId)
    }
  })

  it('scrolls only the subtitle list and pauses following during manual browsing', async () => {
    const { container } = renderPlayer()
    await screen.findByRole('button', { name: '查看 Keep 的发音' })
    const list = container.querySelector<HTMLElement>('.subtitle-list')
    const video = container.querySelector('video')
    const listScrollTo = vi.fn()

    expect(list).not.toBeNull()
    expect(video).not.toBeNull()
    Object.defineProperty(list!, 'clientHeight', { configurable: true, value: 240 })
    Object.defineProperty(list!, 'scrollHeight', { configurable: true, value: 600 })
    Object.defineProperty(list!, 'scrollTop', { configurable: true, writable: true, value: 0 })
    list!.getBoundingClientRect = () => ({
      top: 100,
      right: 400,
      bottom: 340,
      left: 0,
      width: 400,
      height: 240,
      x: 0,
      y: 100,
      toJSON: () => ({}),
    })
    for (const row of container.querySelectorAll<HTMLElement>('[role="listitem"]')) {
      row.getBoundingClientRect = () => ({
        top: 400,
        right: 400,
        bottom: 760,
        left: 0,
        width: 400,
        height: 360,
        x: 0,
        y: 400,
        toJSON: () => ({}),
      })
    }
    list!.scrollTo = listScrollTo

    video!.currentTime = 1.1
    fireEvent.timeUpdate(video!)
    await waitFor(() => expect(listScrollTo).toHaveBeenCalled())
    expect(listScrollTo).toHaveBeenLastCalledWith({ top: 300, behavior: 'smooth' })

    const follow = screen.getByRole('button', { name: '自动跟随当前字幕' })
    expect(follow).toHaveAttribute('aria-pressed', 'true')
    fireEvent.wheel(list!)
    expect(follow).toHaveAttribute('aria-pressed', 'false')
    expect(follow).toHaveTextContent('回到当前句')

    const callsWhilePaused = listScrollTo.mock.calls.length
    video!.currentTime = 0.5
    fireEvent.timeUpdate(video!)
    await waitFor(() => expect(usePlayerStore.getState().currentMs).toBe(500))
    expect(listScrollTo).toHaveBeenCalledTimes(callsWhilePaused)

    fireEvent.click(follow)
    expect(follow).toHaveAttribute('aria-pressed', 'true')
    await waitFor(() => expect(listScrollTo.mock.calls.length).toBeGreaterThan(callsWhilePaused))
  })

  it('centers an active cue even when the whole row is already visible', async () => {
    const { container } = renderPlayer()
    await screen.findByRole('button', { name: '查看 Keep 的发音' })
    const list = container.querySelector<HTMLElement>('.subtitle-list')
    const video = container.querySelector('video')
    const rows = container.querySelectorAll<HTMLElement>('[role="listitem"]')
    const listScrollTo = vi.fn()

    expect(list).not.toBeNull()
    expect(video).not.toBeNull()
    Object.defineProperty(list!, 'clientHeight', { configurable: true, value: 240 })
    Object.defineProperty(list!, 'scrollTop', { configurable: true, writable: true, value: 0 })
    list!.getBoundingClientRect = () => ({
      top: 100,
      right: 400,
      bottom: 340,
      left: 0,
      width: 400,
      height: 240,
      x: 0,
      y: 100,
      toJSON: () => ({}),
    })
    rows[1].getBoundingClientRect = () => ({
      top: 250,
      right: 400,
      bottom: 310,
      left: 0,
      width: 400,
      height: 60,
      x: 0,
      y: 250,
      toJSON: () => ({}),
    })
    list!.scrollTo = listScrollTo

    video!.currentTime = 1.1
    fireEvent.timeUpdate(video!)

    await waitFor(() => {
      expect(listScrollTo).toHaveBeenLastCalledWith({ top: 60, behavior: 'smooth' })
    })
    expect(list!.style.getPropertyValue('--subtitle-edge-space')).toBe('120px')
  })

  it('recenters the same cue after a resize but stays put while following is paused', async () => {
    const { container } = renderPlayer()
    await screen.findByRole('button', { name: '查看 Keep 的发音' })
    const list = container.querySelector<HTMLElement>('.subtitle-list')
    const video = container.querySelector('video')
    const rows = container.querySelectorAll<HTMLElement>('[role="listitem"]')
    const listScrollTo = vi.fn()
    let listHeight = 240
    let listWidth = 400
    let rowTop = 190
    let rowHeight = 60

    expect(list).not.toBeNull()
    expect(video).not.toBeNull()
    Object.defineProperty(list!, 'clientHeight', { configurable: true, get: () => listHeight })
    Object.defineProperty(list!, 'clientWidth', { configurable: true, get: () => listWidth })
    Object.defineProperty(list!, 'scrollTop', { configurable: true, writable: true, value: 20 })
    list!.getBoundingClientRect = () => ({
      top: 100,
      right: listWidth,
      bottom: 100 + listHeight,
      left: 0,
      width: listWidth,
      height: listHeight,
      x: 0,
      y: 100,
      toJSON: () => ({}),
    })
    rows[1].getBoundingClientRect = () => ({
      top: rowTop,
      right: 400,
      bottom: rowTop + rowHeight,
      left: 0,
      width: 400,
      height: rowHeight,
      x: 0,
      y: rowTop,
      toJSON: () => ({}),
    })
    list!.scrollTo = listScrollTo

    video!.currentTime = 1.1
    fireEvent.timeUpdate(video!)
    await waitFor(() => expect(usePlayerStore.getState().currentMs).toBe(1_100))
    expect(listScrollTo).not.toHaveBeenCalled()

    const observer = await waitFor(() => {
      const match = ResizeObserverMock.instances.find((instance) => instance.observed.has(list!))
      expect(match).toBeDefined()
      return match!
    })
    listHeight = 200
    rowTop = 220
    observer.trigger([list!])

    expect(listScrollTo).toHaveBeenLastCalledWith({ top: 70, behavior: 'auto' })
    expect(list!.style.getPropertyValue('--subtitle-edge-space')).toBe('100px')

    const callsBeforeWidthChange = listScrollTo.mock.calls.length
    listWidth = 320
    rowTop = 200
    observer.trigger([list!])
    expect(listScrollTo).toHaveBeenCalledTimes(callsBeforeWidthChange + 1)
    expect(listScrollTo).toHaveBeenLastCalledWith({ top: 50, behavior: 'auto' })

    const follow = screen.getByRole('button', { name: '自动跟随当前字幕' })
    fireEvent.wheel(list!)
    expect(follow).toHaveAttribute('aria-pressed', 'false')
    const callsWhilePaused = listScrollTo.mock.calls.length
    const pausedObserver = await waitFor(() => {
      const match = [...ResizeObserverMock.instances].reverse().find((instance) => instance.observed.has(list!))
      expect(match).toBeDefined()
      expect(match).not.toBe(observer)
      return match!
    })

    listHeight = 160
    rowHeight = 80
    pausedObserver.trigger([list!, rows[1]])
    expect(listScrollTo).toHaveBeenCalledTimes(callsWhilePaused)
  })

  it('moves focus into the pronunciation panel and restores it after Escape', async () => {
    renderPlayer()
    const word = await screen.findByRole('button', { name: '查看 Hello 的发音' })

    fireEvent.click(word)
    const dialog = await screen.findByRole('dialog', { name: 'Hello' })
    const close = screen.getByRole('button', { name: '关闭发音面板' })
    expect(dialog).toHaveAttribute('aria-modal', 'false')
    expect(close).toHaveFocus()

    fireEvent.keyDown(document, { key: 'Escape' })
    await waitFor(() => expect(dialog).not.toBeInTheDocument())
    expect(word).toHaveFocus()
  })
})
