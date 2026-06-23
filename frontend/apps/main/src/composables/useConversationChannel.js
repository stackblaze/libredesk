import { Mail, MessageSquare, Globe } from 'lucide-vue-next'

const CHANNEL_META = {
  email: { label: 'Email', icon: Mail },
  livechat: { label: 'Chat', icon: MessageSquare }
}

/** Map a conversation's inbox_channel to a friendly label + icon. */
export function channelMeta (channel) {
  if (CHANNEL_META[channel]) return CHANNEL_META[channel]
  const label = channel ? channel.charAt(0).toUpperCase() + channel.slice(1) : ''
  return { label, icon: Globe }
}
