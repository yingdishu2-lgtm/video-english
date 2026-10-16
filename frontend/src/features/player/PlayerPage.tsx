import { useEffect, useMemo, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import { mediaApi } from '../../api/client'
import { usePlayerStore } from '../../stores/player'
import type { LearningCue } from '../../types/media'
import { localPronunciation } from './pronunciation'
import { tokenizeEnglish } from './tokenize'
import { Icon } from '../../components/Icon'

function formatTime(ms: number) {
  const total = Math.floor(ms / 1000)
  const minutes = Math.floor(total / 60)
  const seconds = total % 60
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
}

function activeCueAt(cues: LearningCue[], currentMs: number) {
  let low = 0
  let high = cues.length - 1
  let candidate = -1
  while (low <= high) {
    const middle = Math.floor((low + high) / 2)
    if (cues[middle].startMs <= currentMs) {
      candidate = middle
      low = middle + 1
    } else {
      high = middle - 1
    }
  }
  if (candidate >= 0 && currentMs < cues[candidate].endMs) return cues[candidate]
  return undefined
}

export function PlayerPage() {
  const id = Number(useParams().mediaId)
  const videoRef = useRef<HTMLVideoElement>(null)
  const activeRowRef = useRef<HTMLDivElement>(null)
  const subtitleListRef = useRef<HTMLDivElement>(null)
  const wordTriggerRef = useRef<HTMLButtonElement>(null)
  const dictionaryCloseRef = useRef<HTMLButtonElement>(null)
  const [followPlayback, setFollowPlayback] = useState(true)
  const [englishTrackId, setEnglishTrackId] = useState<number>()
  const [chineseTrackId, setChineseTrackId] = useState<number>()
  const [wordSelection, setWordSelection] = useState<{ word: string; cue: LearningCue }>()
  const item = useQuery({ queryKey: ['media', id], queryFn: () => mediaApi.get(id), enabled: Number.isInteger(id) })
  const playback = useQuery({ queryKey: ['playback', id], queryFn: () => mediaApi.playback(id), enabled: Number.isInteger(id), retry: false })
  const tracks = useQuery({ queryKey: ['subtitles', id], queryFn: () => mediaApi.subtitles(id), enabled: Number.isInteger(id) })
  const cues = useQuery({
    queryKey: ['learning-cues', id, englishTrackId, chineseTrackId],
    queryFn: () => mediaApi.learningCues(id, englishTrackId!, chineseTrackId!),
    enabled: englishTrackId !== undefined && chineseTrackId !== undefined && englishTrackId !== chineseTrackId,
    staleTime: Number.POSITIVE_INFINITY,
  })
  const pronunciation = useQuery({
    queryKey: ['pronunciation', wordSelection?.word.toLowerCase()],
    queryFn: async () => {
      const local = await localPronunciation(wordSelection!.word)
      return local.found ? local : mediaApi.pronunciation(wordSelection!.word)
    },
    enabled: wordSelection !== undefined,
    staleTime: Number.POSITIVE_INFINITY,
    retry: false,
  })
  const { currentMs, durationMs, playbackRate, setCurrentMs, setDurationMs, setPlaybackRate, reset } = usePlayerStore()
  const activeCue = useMemo(() => activeCueAt(cues.data ?? [], currentMs), [cues.data, currentMs])

  useEffect(() => reset, [reset])
  useEffect(() => {
    if (!tracks.data?.length) return
    const supported = tracks.data.filter((track) => track.supported)
    const englishValid = supported.some((track) => track.id === englishTrackId)
    const chineseValid = supported.some((track) => track.id === chineseTrackId && track.id !== englishTrackId)
    const english = supported.find((track) => /^(en|eng)$/i.test(track.language ?? '') || /english/i.test(track.title ?? ''))
      ?? supported[0]
    const selectedEnglishId = englishValid ? englishTrackId : english?.id
    const chinese = supported.find((track) => track.id !== selectedEnglishId && (/^(zh|zho|chi|chs|cht)$/i.test(track.language ?? '') || /chinese|中文|繁體|简体/i.test(track.title ?? '')))
      ?? supported.find((track) => track.id !== selectedEnglishId)
    if (!englishValid) setEnglishTrackId(english?.id)
    if (!chineseValid) setChineseTrackId(chinese?.id)
  }, [chineseTrackId, englishTrackId, tracks.data])
  useEffect(() => {
    const list = subtitleListRef.current
    const row = activeRowRef.current
    if (!list) return
    let listHeight = list.clientHeight
    let listWidth = list.clientWidth
    let rowHeight = row?.getBoundingClientRect().height

    const updateEdgeSpace = () => {
      // Leave enough scroll range to center even the first and last cue.
      list.style.setProperty('--subtitle-edge-space', `${cues.data?.length ? list.clientHeight / 2 : 0}px`)
    }

    const centerActiveRow = (behavior: ScrollBehavior) => {
      if (!followPlayback || wordSelection || !row || !list.clientHeight) return
      const listBounds = list.getBoundingClientRect()
      const rowBounds = row.getBoundingClientRect()
      // A cue taller than the viewport starts at the top so its opening stays readable.
      const offset = rowBounds.top - listBounds.top - list.clientTop - Math.max(0, (list.clientHeight - rowBounds.height) / 2)
      if (Math.abs(offset) <= 1) return
      list.scrollTo({ top: list.scrollTop + offset, behavior })
    }

    updateEdgeSpace()
    centerActiveRow(window.matchMedia('(prefers-reduced-motion: reduce)').matches ? 'auto' : 'smooth')

    const observer = new ResizeObserver(() => {
      const nextListHeight = list.clientHeight
      const nextListWidth = list.clientWidth
      const nextRowHeight = row?.getBoundingClientRect().height
      if (nextListHeight === listHeight && nextListWidth === listWidth && nextRowHeight === rowHeight) return
      listHeight = nextListHeight
      listWidth = nextListWidth
      rowHeight = nextRowHeight
      updateEdgeSpace()
      centerActiveRow('auto')
    })
    observer.observe(list)
    if (row) observer.observe(row)
    return () => observer.disconnect()
  }, [activeCue?.id, cues.data?.length, followPlayback, wordSelection])
  useEffect(() => {
    if (!wordSelection) return
    const trigger = wordTriggerRef.current
    dictionaryCloseRef.current?.focus({ preventScroll: true })
    function onEscape(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.preventDefault()
        setWordSelection(undefined)
      }
    }
    document.addEventListener('keydown', onEscape)
    return () => {
      document.removeEventListener('keydown', onEscape)
      if (trigger?.isConnected) trigger.focus({ preventScroll: true })
    }
  }, [wordSelection])

  function changeRate(value: number) {
    setPlaybackRate(value)
    if (videoRef.current) videoRef.current.playbackRate = value
  }

  function seekCue(cue: LearningCue) {
    if (!videoRef.current) return
    videoRef.current.currentTime = cue.startMs / 1000
    setCurrentMs(cue.startMs)
    setFollowPlayback(true)
  }

  if (!Number.isInteger(id)) return <section className="page"><div className="notice error">无效的媒体编号</div></section>
  return (
    <section className="player-page">
      <div className="player-header">
        <Link to="/" className="back-link"><Icon name="arrow-left" size={18} /><span>返回媒体库</span></Link>
        <div className="player-title"><h1 title={item.data?.title}>{item.data?.title ?? '载入影片…'}</h1><p title={item.data?.path}>{item.data?.path}</p></div>
      </div>
      <div className="player-layout">
        <div className="video-pane">
          <div className="video-stage">
            {playback.data?.url ? (
              <video
                ref={videoRef}
                src={playback.data.url}
                controls
                playsInline
                preload="metadata"
                onTimeUpdate={(event) => setCurrentMs(Math.round(event.currentTarget.currentTime * 1000))}
                onLoadedMetadata={(event) => {
                  setDurationMs(Math.round(event.currentTarget.duration * 1000))
                  event.currentTarget.playbackRate = playbackRate
                }}
              />
            ) : playback.isLoading ? <p>正在准备播放…</p> : <p>{playback.error?.message ?? '当前媒体暂不支持直放'}</p>}
          </div>
          <div className="control-strip">
            <span className="playback-time">{formatTime(currentMs)} <span>/ {formatTime(durationMs || item.data?.durationMs || 0)}</span></span>
            <label>倍速
              <select value={playbackRate} onChange={(event) => changeRate(Number(event.target.value))}>
                {[0.75, 1, 1.25, 1.5, 2].map((rate) => <option value={rate} key={rate}>{rate}×</option>)}
              </select>
            </label>
          </div>
        </div>
        <aside className="subtitle-pane" aria-labelledby="subtitle-title">
          <div className="subtitle-heading">
            <h2 id="subtitle-title"><Icon name="captions" size={20} />字幕时间轴</h2>
            <button
              type="button"
              className="follow-button"
              aria-label="自动跟随当前字幕"
              aria-pressed={followPlayback}
              onClick={() => setFollowPlayback(!followPlayback)}
            >{followPlayback && <Icon name="check" size={16} />}{followPlayback ? '跟随播放中' : '回到当前句'}</button>
          </div>
          <div className="subtitle-tools">
            <details className="track-settings">
              <summary><Icon name="settings" size={18} />字幕轨道</summary>
            <div className="track-pair">
            <label className="track-select">英文轨
              <select disabled={!tracks.data?.some((track) => track.supported)} value={englishTrackId ?? ''} onChange={(event) => {
                const nextEnglishId = Number(event.target.value)
                if (nextEnglishId === chineseTrackId) setChineseTrackId(englishTrackId)
                setEnglishTrackId(nextEnglishId)
              }}>
                <option value="" disabled>选择英文字幕</option>
                {tracks.data?.map((track) => (
                  <option value={track.id} key={track.id} disabled={!track.supported}>
                    {track.language?.toUpperCase() || 'UND'} · {track.title || `Stream ${track.streamIndex ?? 'external'}`}{!track.supported ? '（不支持）' : ''}
                  </option>
                ))}
              </select>
            </label>
            <label className="track-select">中文轨
              <select disabled={!tracks.data?.some((track) => track.supported && track.id !== englishTrackId)} value={chineseTrackId ?? ''} onChange={(event) => setChineseTrackId(Number(event.target.value))}>
                <option value="" disabled>选择中文字幕</option>
                {tracks.data?.map((track) => (
                  <option value={track.id} key={track.id} disabled={!track.supported || track.id === englishTrackId}>
                    {track.language?.toUpperCase() || 'UND'} · {track.title || `Stream ${track.streamIndex ?? 'external'}`}{!track.supported ? '（不支持）' : ''}
                  </option>
                ))}
              </select>
            </label>
            </div>
            </details>
          <div className="subtitle-status" role="status">
            {tracks.isLoading && '正在读取字幕轨…'}
            {tracks.error && tracks.error.message}
            {cues.isLoading && '正在提取并对齐双语字幕…'}
            {cues.data && `${cues.data.length} 条双语字幕 · 点词查看 IPA`}
            {cues.error && cues.error.message}
          </div>
          </div>
          <div
            ref={subtitleListRef}
            className="subtitle-list"
            role="list"
            aria-label="双语字幕"
            onWheel={() => setFollowPlayback(false)}
            onTouchMove={() => setFollowPlayback(false)}
            onPointerDown={(event) => { if (event.target === event.currentTarget) setFollowPlayback(false) }}
            onKeyDown={(event) => {
              if (['ArrowUp', 'ArrowDown', 'PageUp', 'PageDown', 'Home', 'End'].includes(event.key)) setFollowPlayback(false)
            }}
          >
            {tracks.isSuccess && (englishTrackId === undefined || chineseTrackId === undefined) && (
              <div className="subtitle-empty">
                <strong>需要英文和中文字幕</strong>
                <p>在“字幕轨道”中选择两种字幕。若影片没有对应字幕，可添加外挂字幕后重新扫描媒体库。</p>
                <Link to="/" className="back-link">返回媒体库</Link>
              </div>
            )}
            {cues.data?.length === 0 && <p className="subtitle-empty">所选字幕暂时没有可显示的台词，请尝试其他字幕轨。</p>}
            {cues.data?.map((cue) => {
              const active = cue.id === activeCue?.id
              return (
                <div
                  ref={active ? activeRowRef : undefined}
                  className={`subtitle-row${active ? ' active' : ''}`}
                  key={cue.id}
                  role="listitem"
                  aria-current={active ? 'true' : undefined}
                >
                  <button type="button" className="cue-seek" aria-label={`跳转到 ${formatTime(cue.startMs)}`} onClick={(event) => { event.stopPropagation(); seekCue(cue) }}>
                    <time>{formatTime(cue.startMs)}</time>
                  </button>
                  <span className="cue-copy">
                    <span className="cue-en">{tokenizeEnglish(cue.en).map((token, index) => token.word ? (
                      <button
                        className="word-token"
                        key={`${token.text}-${index}`}
                        type="button"
                        aria-label={`查看 ${token.text} 的发音`}
                        onClick={(event) => {
                          event.stopPropagation()
                          wordTriggerRef.current = event.currentTarget
                          setWordSelection({ word: token.text, cue })
                        }}
                      >{token.text}</button>
                    ) : <span key={`${token.text}-${index}`}>{token.text}</span>)}</span>
                    {cue.zh && <button type="button" className="cue-zh" aria-label={`从 ${formatTime(cue.startMs)} 播放：${cue.zh}`} onClick={() => seekCue(cue)}>{cue.zh}</button>}
                  </span>
                </div>
              )
            })}
          </div>
          {wordSelection && (
            <div className="dictionary-card" role="dialog" aria-modal="false" aria-labelledby="dictionary-title" aria-describedby="dictionary-context">
              <button ref={dictionaryCloseRef} className="dictionary-close" type="button" onClick={() => setWordSelection(undefined)} aria-label="关闭发音面板">
                <Icon name="close" />
              </button>
              <h3 id="dictionary-title">{wordSelection.word}</h3>
              {pronunciation.isLoading && <p className="dictionary-state">正在查询音标…</p>}
              {pronunciation.data?.found && <p className="phonetic">{pronunciation.data.phonetic || '暂无书面音标'}</p>}
              {pronunciation.data?.audioUrl && <audio controls preload="none" src={pronunciation.data.audioUrl} />}
              {pronunciation.data && !pronunciation.data.found && <p className="dictionary-state">词典暂未收录这个词形。</p>}
              {pronunciation.error && <p className="dictionary-state error">音标查询暂时不可用：{pronunciation.error.message}</p>}
              <div className="word-context" id="dictionary-context"><span>{wordSelection.cue.en}</span>{wordSelection.cue.zh && <small>{wordSelection.cue.zh}</small>}</div>
              {pronunciation.data && <small className="dictionary-source">音标来源：{pronunciation.data.source}{pronunciation.data.cached ? ' · 本地可用' : ''}</small>}
            </div>
          )}
        </aside>
      </div>
    </section>
  )
}
