import { useEffect, useMemo, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link, useParams } from 'react-router-dom'
import { mediaApi } from '../../api/client'
import { usePlayerStore } from '../../stores/player'
import type { LearningCue } from '../../types/media'
import { localPronunciation } from './pronunciation'
import { tokenizeEnglish } from './tokenize'

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
  const [englishTrackId, setEnglishTrackId] = useState<number>()
  const [chineseTrackId, setChineseTrackId] = useState<number>()
  const [wordSelection, setWordSelection] = useState<{ word: string; cue: LearningCue }>()
  const item = useQuery({ queryKey: ['media', id], queryFn: () => mediaApi.get(id), enabled: Number.isInteger(id) })
  const playback = useQuery({ queryKey: ['playback', id], queryFn: () => mediaApi.playback(id), enabled: Number.isInteger(id), retry: false })
  const tracks = useQuery({ queryKey: ['subtitles', id], queryFn: () => mediaApi.subtitles(id), enabled: Number.isInteger(id) })
  const cues = useQuery({
    queryKey: ['learning-cues', id, englishTrackId, chineseTrackId],
    queryFn: () => mediaApi.learningCues(id, englishTrackId!, chineseTrackId!),
    enabled: englishTrackId !== undefined && chineseTrackId !== undefined,
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
    const chineseValid = supported.some((track) => track.id === chineseTrackId)
    const english = supported.find((track) => /^(en|eng)$/i.test(track.language ?? '') || /english/i.test(track.title ?? ''))
      ?? supported[0]
    const chinese = supported.find((track) => /^(zh|zho|chi|chs|cht)$/i.test(track.language ?? '') || /chinese|中文|繁體|简体/i.test(track.title ?? ''))
      ?? supported.find((track) => track.id !== english?.id)
    if (!englishValid) setEnglishTrackId(english?.id)
    if (!chineseValid) setChineseTrackId(chinese?.id)
  }, [chineseTrackId, englishTrackId, tracks.data])
  useEffect(() => {
    activeRowRef.current?.scrollIntoView({ block: 'center', behavior: 'smooth' })
  }, [activeCue?.id])

  function changeRate(value: number) {
    setPlaybackRate(value)
    if (videoRef.current) videoRef.current.playbackRate = value
  }

  if (!Number.isInteger(id)) return <section className="page"><div className="notice error">无效的媒体编号</div></section>
  return (
    <section className="player-page">
      <div className="player-header">
        <Link to="/" className="back-link">← 返回媒体库</Link>
        <div><h1>{item.data?.title ?? '载入影片…'}</h1><p>{item.data?.path}</p></div>
      </div>
      <div className="player-layout">
        <div className="video-pane">
          <div className="video-stage">
            {playback.data?.url ? (
              <video
                ref={videoRef}
                src={playback.data.url}
                controls
                preload="metadata"
                onTimeUpdate={(event) => setCurrentMs(Math.round(event.currentTarget.currentTime * 1000))}
                onLoadedMetadata={(event) => setDurationMs(Math.round(event.currentTarget.duration * 1000))}
              />
            ) : playback.isLoading ? <p>正在准备播放…</p> : <p>{playback.error?.message ?? '当前媒体暂不支持直放'}</p>}
          </div>
          <div className="control-strip">
            <span>{formatTime(currentMs)} / {formatTime(durationMs || item.data?.durationMs || 0)}</span>
            <label>倍速
              <select value={playbackRate} onChange={(event) => changeRate(Number(event.target.value))}>
                {[0.75, 1, 1.25, 1.5, 2].map((rate) => <option value={rate} key={rate}>{rate}×</option>)}
              </select>
            </label>
          </div>
        </div>
        <aside className="subtitle-pane">
          <div className="subtitle-heading">
            <div>
              <p className="eyebrow">SUBTITLE TIMELINE</p>
              <h2>字幕时间轴</h2>
            </div>
            <div className="track-pair">
            <label className="track-select">英文轨
              <select value={englishTrackId ?? ''} onChange={(event) => setEnglishTrackId(Number(event.target.value))}>
                {tracks.data?.map((track) => (
                  <option value={track.id} key={track.id} disabled={!track.supported}>
                    {track.language?.toUpperCase() || 'UND'} · {track.title || `Stream ${track.streamIndex ?? 'external'}`}{!track.supported ? '（不支持）' : ''}
                  </option>
                ))}
              </select>
            </label>
            <label className="track-select">中文轨
              <select value={chineseTrackId ?? ''} onChange={(event) => setChineseTrackId(Number(event.target.value))}>
                {tracks.data?.map((track) => (
                  <option value={track.id} key={track.id} disabled={!track.supported || track.id === englishTrackId}>
                    {track.language?.toUpperCase() || 'UND'} · {track.title || `Stream ${track.streamIndex ?? 'external'}`}{!track.supported ? '（不支持）' : ''}
                  </option>
                ))}
              </select>
            </label>
            </div>
          </div>
          <div className="subtitle-status">
            {tracks.isLoading && '正在读取字幕轨…'}
            {tracks.error && tracks.error.message}
            {cues.isLoading && '正在提取并对齐双语字幕…'}
            {cues.data && `${cues.data.length} 条双语字幕 · 点击单词查看 IPA`}
            {cues.error && cues.error.message}
          </div>
          <div className="subtitle-list">
            {cues.data?.map((cue) => {
              const active = cue.id === activeCue?.id
              return (
                <div
                  ref={active ? activeRowRef : undefined}
                  className={`subtitle-row${active ? ' active' : ''}`}
                  key={cue.id}
                  role="button"
                  tabIndex={0}
                  onClick={() => { if (videoRef.current) videoRef.current.currentTime = cue.startMs / 1000 }}
                  onKeyDown={(event) => {
                    if ((event.key === 'Enter' || event.key === ' ') && videoRef.current) videoRef.current.currentTime = cue.startMs / 1000
                  }}
                >
                  <time>{formatTime(cue.startMs)}</time>
                  <span className="cue-copy">
                    <span className="cue-en">{tokenizeEnglish(cue.en).map((token, index) => token.word ? (
                      <button
                        className="word-token"
                        key={`${token.text}-${index}`}
                        type="button"
                        onClick={(event) => { event.stopPropagation(); setWordSelection({ word: token.text, cue }) }}
                      >{token.text}</button>
                    ) : <span key={`${token.text}-${index}`}>{token.text}</span>)}</span>
                    {cue.zh && <span className="cue-zh">{cue.zh}</span>}
                  </span>
                </div>
              )
            })}
          </div>
          {wordSelection && (
            <div className="dictionary-card" role="dialog" aria-label={`${wordSelection.word} 的发音`}>
              <button className="dictionary-close" type="button" onClick={() => setWordSelection(undefined)} aria-label="关闭">×</button>
              <p className="eyebrow">PRONUNCIATION</p>
              <h3>{wordSelection.word}</h3>
              {pronunciation.isLoading && <p className="dictionary-state">正在查询音标…</p>}
              {pronunciation.data?.found && <p className="phonetic">{pronunciation.data.phonetic || '暂无书面音标'}</p>}
              {pronunciation.data?.audioUrl && <audio controls preload="none" src={pronunciation.data.audioUrl} />}
              {pronunciation.data && !pronunciation.data.found && <p className="dictionary-state">词典暂未收录这个词形。</p>}
              {pronunciation.error && <p className="dictionary-state error">音标查询暂时不可用：{pronunciation.error.message}</p>}
              <div className="word-context"><span>{wordSelection.cue.en}</span>{wordSelection.cue.zh && <small>{wordSelection.cue.zh}</small>}</div>
              {pronunciation.data && <small className="dictionary-source">音标来源：{pronunciation.data.source}{pronunciation.data.cached ? ' · 本地可用' : ''}</small>}
            </div>
          )}
        </aside>
      </div>
    </section>
  )
}
