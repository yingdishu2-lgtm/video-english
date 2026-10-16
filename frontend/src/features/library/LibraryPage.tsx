import { useMemo, useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { mediaApi } from '../../api/client'
import { Icon } from '../../components/Icon'
import type { MediaItem } from '../../types/media'
import './library.css'

const EMPTY_MEDIA_ITEMS: MediaItem[] = []

function formatDuration(ms: number) {
  const totalSeconds = Math.max(0, Math.round(ms / 1000))
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60

  if (hours > 0) return `${hours}:${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`
  return `${minutes}:${String(seconds).padStart(2, '0')}`
}

function formatFileSize(bytes: number) {
  if (!Number.isFinite(bytes) || bytes <= 0) return '未知大小'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const index = Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), units.length - 1)
  const value = bytes / 1024 ** index
  return `${value.toFixed(value >= 10 || index === 0 ? 0 : 1)} ${units[index]}`
}

function getMediaFacts(item: MediaItem) {
  const audio = item.streams.filter((stream) => stream.type === 'audio')
  const subtitles = item.streams.filter((stream) => stream.type === 'subtitle')
  const resolution = item.width > 0 && item.height > 0 ? `${item.width}×${item.height}` : '未知分辨率'
  return { audio, subtitles, resolution }
}

function searchableText(item: MediaItem) {
  return [item.title, item.filename, item.path, item.container, item.videoCodec,
    ...item.streams.flatMap((stream) => [stream.codec, stream.language, stream.title])]
    .filter(Boolean).join(' ').toLocaleLowerCase()
}

function selectFeatured(items: MediaItem[]) {
  const nonDemo = items.filter((item) => !/\b(demo|sample|test)\b/i.test(`${item.title} ${item.filename} ${item.path}`))
  const candidates = nonDemo.length > 0 ? nonDemo : items
  return [...candidates].sort((a, b) => b.durationMs - a.durationMs)[0]
}

function MediaArtwork({ item, featured = false }: { item: MediaItem; featured?: boolean }) {
  const [failed, setFailed] = useState(false)

  return (
    <div className={`library-artwork${featured ? ' library-artwork-featured' : ''}`}>
      {!failed ? (
        <img src={`/api/v1/media/${item.id}/poster?v=${encodeURIComponent(item.updatedAt)}`} alt={`${item.title} 的影片画面`}
          loading={featured ? 'eager' : 'lazy'} onError={() => setFailed(true)} />
      ) : (
        <div className="library-artwork-fallback" role="img" aria-label={`${item.title} 暂无影片画面`}>
          <Icon name="video" size={featured ? 32 : 24} />
          <span>暂无画面</span>
        </div>
      )}
    </div>
  )
}

function MediaCard({ item }: { item: MediaItem }) {
  const { audio, subtitles, resolution } = getMediaFacts(item)
  return (
    <article className="library-card">
      <Link className="library-card-artwork" to={`/watch/${item.id}`} aria-label={`播放 ${item.title}`}>
        <MediaArtwork key={`${item.id}-${item.updatedAt}`} item={item} />
        <span className="library-card-play" aria-hidden="true"><Icon name="play" size={18} /></span>
        <span className="library-card-duration">{formatDuration(item.durationMs)}</span>
      </Link>
      <div className="library-card-copy">
        <h3><Link to={`/watch/${item.id}`}>{item.title}</Link></h3>
        <p className="library-card-summary">
          <span>{resolution}</span><span>{audio.length} 条音轨</span><span>{subtitles.length} 条字幕</span>
        </p>
        <details className="library-details">
          <summary>媒体信息 <Icon name="chevron-right" size={14} /></summary>
          <dl>
            <div><dt>文件</dt><dd>{item.filename}</dd></div>
            <div><dt>路径</dt><dd>{item.path}</dd></div>
            <div><dt>格式</dt><dd>{item.container.toUpperCase()} · {item.videoCodec.toUpperCase()}</dd></div>
            <div><dt>大小</dt><dd>{formatFileSize(item.fileSize)}</dd></div>
          </dl>
        </details>
      </div>
    </article>
  )
}

function LibraryEmpty() {
  return (
    <div className="library-empty">
      <span className="library-empty-icon" aria-hidden="true"><Icon name="folder" size={30} /></span>
      <strong>媒体库还是空的</strong>
      <p>将 MP4、M4V、MOV、MKV 或 WebM 文件放入项目的 <code>media/</code> 目录，然后扫描媒体目录。</p>
    </div>
  )
}

export function LibraryPage() {
  const [query, setQuery] = useState('')
  const queryClient = useQueryClient()
  const media = useQuery({ queryKey: ['media'], queryFn: mediaApi.list })
  const scan = useMutation({
    mutationFn: mediaApi.scan,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['media'] }),
  })
  const items = media.data ?? EMPTY_MEDIA_ITEMS
  const featured = useMemo(() => selectFeatured(items), [items])
  const normalizedQuery = query.trim().toLocaleLowerCase()
  const filteredItems = useMemo(
    () => normalizedQuery ? items.filter((item) => searchableText(item).includes(normalizedQuery)) : items,
    [items, normalizedQuery],
  )

  return (
    <section className="page library-page">
      <header className="library-header">
        <div>
          <h1>媒体库</h1>
          <p>选择一部影片，在真实对白与双语字幕中练习英语。</p>
        </div>
        <button className="button button-secondary library-scan" type="button" onClick={() => scan.mutate()} disabled={scan.isPending}>
          <Icon name="scan" size={18} />{scan.isPending ? '正在扫描…' : '扫描媒体目录'}
        </button>
      </header>

      {scan.data && <div className="library-notice library-notice-success" role="status">
        <Icon name="check" size={18} /><span>扫描完成：新增 {scan.data.added}，更新 {scan.data.updated}，未变 {scan.data.unchanged}，失败 {scan.data.failed}</span>
      </div>}
      {scan.error && <div className="library-notice library-notice-error" role="alert">
        <span>扫描失败：{scan.error.message}</span>
        <button className="button button-secondary" type="button" onClick={() => scan.mutate()}>重新扫描</button>
      </div>}

      {media.isLoading && <div className="library-loading" role="status">
        <span className="library-loading-mark" aria-hidden="true" /><strong>正在读取媒体库</strong><p>正在整理本地影片与音视频轨道…</p>
      </div>}
      {media.error && <div className="library-fetch-error" role="alert">
        <span className="library-empty-icon" aria-hidden="true"><Icon name="video" size={28} /></span>
        <strong>无法读取媒体库</strong><p>{media.error.message}</p>
        <button className="button button-secondary" type="button" onClick={() => media.refetch()}>重新加载</button>
      </div>}
      {!media.isLoading && !media.error && items.length === 0 && <LibraryEmpty />}

      {!media.isLoading && !media.error && featured && <>
        <article className="library-featured">
          <MediaArtwork key={`${featured.id}-${featured.updatedAt}`} item={featured} featured />
          <div className="library-featured-shade" aria-hidden="true" />
          <div className="library-featured-copy">
            <h2>{featured.title}</h2>
            <p>{getMediaFacts(featured).resolution} · {formatDuration(featured.durationMs)} · {featured.container.toUpperCase()}</p>
            <Link className="button button-primary library-featured-action" to={`/watch/${featured.id}`}>
              <Icon name="play" size={18} />开始播放
            </Link>
          </div>
        </article>

        <div className="library-toolbar">
          <label className="library-search">
            <span className="library-visually-hidden">搜索媒体库</span><Icon name="search" size={18} />
            <input type="search" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="搜索影片、文件或编码" />
            {query && <button type="button" onClick={() => setQuery('')} aria-label="清除搜索"><Icon name="close" size={16} /></button>}
          </label>
          <p aria-live="polite">{normalizedQuery ? `${filteredItems.length} 个结果` : `共 ${items.length} 部影片`}</p>
        </div>

        {filteredItems.length > 0 ? (
          <div className="library-grid">{filteredItems.map((item) => <MediaCard item={item} key={item.id} />)}</div>
        ) : (
          <div className="library-no-results">
            <Icon name="search" size={26} /><strong>没有找到相关影片</strong><p>换一个标题、文件名或编码关键词试试。</p>
            <button className="button button-secondary" type="button" onClick={() => setQuery('')}>清除搜索</button>
          </div>
        )}
      </>}
    </section>
  )
}
