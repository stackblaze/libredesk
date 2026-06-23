import { useMediaQuery } from '@vueuse/core'

/** Matches the shadcn sidebar mobile breakpoint (768px). */
export function useIsMobileLayout () {
  return useMediaQuery('(max-width: 768px)')
}
