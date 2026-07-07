import zh from './zh'
import en from './en'

export type MessageSchema = typeof zh

export const messages = {
  zh,
  en,
} as const

export type Locale = keyof typeof messages
