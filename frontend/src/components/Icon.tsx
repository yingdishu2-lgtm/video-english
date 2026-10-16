import type { SVGProps } from 'react'

const paths = {
  library: <><rect x="3" y="3" width="7" height="18" rx="2" /><path d="M14 4h5a2 2 0 0 1 2 2v13a2 2 0 0 1-2 2h-5V4ZM6.5 7h0M6 17h1" /></>,
  play: <path d="m9 5 10 7-10 7V5Z" />,
  scan: <><path d="M20 7v5h-5M4 17v-5h5" /><path d="M6 7a7 7 0 0 1 11.7-1.7L20 8M4 16l2.3 2.7A7 7 0 0 0 18 17" /></>,
  search: <><circle cx="10.5" cy="10.5" r="6.5" /><path d="m16 16 5 5" /></>,
  folder: <path d="M3 7a2 2 0 0 1 2-2h5l2 3h7a2 2 0 0 1 2 2v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V7Z" />,
  video: <><rect x="3" y="4" width="18" height="16" rx="3" /><path d="m10 9 5 3-5 3V9Z" /></>,
  captions: <><rect x="2" y="4" width="20" height="16" rx="3" /><path d="M10 10a2.5 2.5 0 1 0 0 4M18 10a2.5 2.5 0 1 0 0 4" /></>,
  audio: <><path d="M4 13v-1a8 8 0 0 1 16 0v1" /><rect x="3" y="12" width="4" height="8" rx="2" /><rect x="17" y="12" width="4" height="8" rx="2" /></>,
  'arrow-left': <path d="m11 5-7 7 7 7M4 12h16" />,
  'chevron-right': <path d="m9 5 7 7-7 7" />,
  settings: <><path d="M4 7h16M4 17h16" /><circle cx="9" cy="7" r="3" /><circle cx="15" cy="17" r="3" /></>,
  close: <path d="m6 6 12 12M18 6 6 18" />,
  check: <path d="m5 12 4 4L19 6" />,
} as const

export type IconName = keyof typeof paths

export function Icon({ name, size = 20, ...props }: SVGProps<SVGSVGElement> & { name: IconName; size?: number }) {
  return <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.75" strokeLinecap="round" strokeLinejoin="round" aria-hidden="true" {...props}>{paths[name]}</svg>
}
