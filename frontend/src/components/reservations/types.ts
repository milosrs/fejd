export type Tone =
  | "orange"
  | "lime"
  | "blue"
  | "violet"
  | "rose"
  | "teal"
  | "amber"
  | "sky"

export const TONES: Tone[] = [
  "orange",
  "lime",
  "blue",
  "violet",
  "rose",
  "teal",
  "amber",
  "sky",
]

export const TONE_BAR: Record<Tone, string> = {
  orange: "bg-orange-500",
  lime: "bg-lime-400",
  blue: "bg-blue-500",
  violet: "bg-violet-500",
  rose: "bg-rose-500",
  teal: "bg-teal-500",
  amber: "bg-amber-500",
  sky: "bg-sky-500",
}

export const TONE_DOT: Record<Tone, string> = {
  orange: "bg-orange-400",
  lime: "bg-lime-300",
  blue: "bg-blue-400",
  violet: "bg-violet-400",
  rose: "bg-rose-400",
  teal: "bg-teal-400",
  amber: "bg-amber-400",
  sky: "bg-sky-400",
}

export function toneFor(text: string): Tone {
  let hash = 0
  for (let i = 0; i < text.length; i++) {
    hash = (hash * 31 + text.charCodeAt(i)) >>> 0
  }
  return TONES[hash % TONES.length]
}

export interface ReservationTag {
  id: string
  label: string
  tone?: Tone
}

export interface ScheduleEvent {
  id: string
  title: string
  subtitle?: string
  start: Date
  end: Date
  tone?: Tone
  attendees?: string[]
  extraAttendees?: number
  meetLink?: string
}
