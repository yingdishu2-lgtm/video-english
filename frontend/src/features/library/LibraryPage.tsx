import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { mediaApi } from '../../api/client'
import type { MediaItem } from '../../types/media'

function formatDuration(ms: number) {
  const totalSeconds = Math.round(ms / 1000)
  const hours = Math.floor(totalSeconds / 3600)
  const minutes = Math.floor((totalSeconds % 3600) / 60)
  const seconds = totalSeconds % 60
  return [hours, minutes, seconds].map((part) => String(part).padStart(2, '0')).join(':')
}

function MediaRow({ item }: { item: MediaItem }) {
  const audio = item.streams.filter((stream) => stream.type === 'audio').length
  const subtitles = item.streams.filter((stream) => stream.type === 'subtitle').length
  return (
    <article className="media-row">
      <div className="media-icon" aria-hidden="true"><span>▶</span></div>
      <div className="media-main">
        <h2>{item.title}</h2>
        <p>{item.path}</p>
        <div className="tag-row">
          <span>{item.container.toUpperCase()}</span>
          <span>{item.videoCodec.toUpperCase()}</span>
          {item.width > 0 && <span>{item.width}×{item.height}</span>}
        </div>
      </div>
      <dl className="media-stats">
        <div><dt>时长</dt><dd>{formatDuration(item.durationMs)}</dd></div>
        <div><dt>音轨</dt><dd>{audio}</dd></div>
        <div><dt>字幕</dt><dd>{subtitles}</dd></div>
      </dl>
      <Link className="button button-primary" to={`/watch/${item.id}`}>开始播放</Link>
    </article>
  )
}

export function LibraryPage() {
  const queryClient = useQueryClient()
  const media = useQuery({ queryKey: ['media'], queryFn: mediaApi.list })
  const scan = useMutation({
    mutationFn: mediaApi.scan,
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['media'] }),
  })
  return (
    <section className="page library-page">
      <div className="page-heading">
        <div>
          <p className="eyebrow">LOCAL LIBRARY</p>
          <h1>媒体库</h1>
          <p>从本地媒体目录发现影片并分析音视频轨道。</p>
        </div>
        <button className="button button-primary" onClick={() => scan.mutate()} disabled={scan.isPending}>
          {scan.isPending ? '正在扫描…' : '扫描媒体目录'}
        </button>
      </div>
      {scan.data && <div className="notice success">扫描完成：新增 {scan.data.added}，更新 {scan.data.updated}，未变 {scan.data.unchanged}，失败 {scan.data.failed}</div>}
      {scan.error && <div className="notice error">{scan.error.message}</div>}
      {media.isLoading && <div className="empty-state">正在读取媒体库…</div>}
      {media.error && <div className="notice error">{media.error.message}</div>}
      {media.data?.length === 0 && (
        <div className="empty-state">
          <strong>媒体库还是空的</strong>
          <p>将 MP4、M4V、MOV、MKV 或 WebM 文件放入项目的 <code>media/</code> 目录，然后点击扫描。</p>
        </div>
      )}
      <div className="media-list">{media.data?.map((item) => <MediaRow item={item} key={item.id} />)}</div>
    </section>
  )
}

