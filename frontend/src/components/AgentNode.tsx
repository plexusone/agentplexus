import { Handle, Position } from '@xyflow/react'
import { Brain, Sparkles, Zap, Bot, Shield, Database, Globe, Code, Server, Cloud } from 'lucide-react'
import type { LucideIcon } from 'lucide-react'

interface AgentNodeData {
  label: string
  description: string
  model: string
  icon?: string
  tools: string[]
}

const modelConfig: Record<string, { color: string; bg: string }> = {
  opus: { color: '#fff', bg: '#e74c3c' },
  sonnet: { color: '#fff', bg: '#3498db' },
  haiku: { color: '#fff', bg: '#2ecc71' },
}

// Lucide icon mapping for common icon names
const lucideIcons: Record<string, LucideIcon> = {
  brain: Brain,
  sparkles: Sparkles,
  zap: Zap,
  bot: Bot,
  shield: Shield,
  database: Database,
  globe: Globe,
  code: Code,
  server: Server,
  cloud: Cloud,
}

// Parse icon specifier: "brandkit:aws", "lucide:shield", or "aws" (infer brandkit)
function parseIconSpec(icon: string): { type: 'brandkit' | 'lucide' | 'model'; name: string } {
  if (!icon) {
    return { type: 'model', name: '' }
  }
  if (icon.startsWith('brandkit:')) {
    return { type: 'brandkit', name: icon.slice(9) }
  }
  if (icon.startsWith('lucide:')) {
    return { type: 'lucide', name: icon.slice(7) }
  }
  // Default: check if it's a known lucide icon, otherwise assume brandkit
  if (lucideIcons[icon.toLowerCase()]) {
    return { type: 'lucide', name: icon.toLowerCase() }
  }
  return { type: 'brandkit', name: icon }
}

const ModelIcon = ({ model }: { model: string }) => {
  const props = { size: 18, strokeWidth: 2 }
  switch (model) {
    case 'opus':
      return <Brain {...props} />
    case 'sonnet':
      return <Sparkles {...props} />
    case 'haiku':
      return <Zap {...props} />
    default:
      return <Bot {...props} />
  }
}

const LucideIcon = ({ name }: { name: string }) => {
  const Icon = lucideIcons[name] || Bot
  return <Icon size={18} strokeWidth={2} />
}

const BrandkitIcon = ({ name }: { name: string }) => {
  // Use white variant for dark UI backgrounds
  const iconUrl = `/api/icons/${name}/white`
  return (
    <img
      src={iconUrl}
      alt={name}
      style={{ width: 24, height: 24 }}
      onError={(e) => {
        // Hide on error, fallback will show
        (e.target as HTMLImageElement).style.display = 'none'
      }}
    />
  )
}

function AgentNode({ data }: { data: AgentNodeData }) {
  const config = modelConfig[data.model] || { color: '#fff', bg: '#666' }
  const iconSpec = parseIconSpec(data.icon || '')

  const renderIcon = () => {
    switch (iconSpec.type) {
      case 'brandkit':
        return <BrandkitIcon name={iconSpec.name} />
      case 'lucide':
        return <LucideIcon name={iconSpec.name} />
      case 'model':
      default:
        return <ModelIcon model={data.model} />
    }
  }

  return (
    <div style={styles.wrapper}>
      <div style={{ ...styles.node, background: config.bg }}>
        <Handle type="target" position={Position.Left} style={styles.handle} />
        <div style={{ color: config.color, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          {renderIcon()}
        </div>
        <Handle type="source" position={Position.Right} style={styles.handle} />
      </div>
      <span style={styles.label}>{data.label}</span>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  wrapper: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    gap: 6,
  },
  node: {
    width: 48,
    height: 48,
    borderRadius: 8,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    cursor: 'pointer',
    transition: 'all 0.15s ease',
    boxShadow: '0 2px 6px rgba(0,0,0,0.4)',
    border: '2px solid rgba(255,255,255,0.1)',
  },
  label: {
    fontSize: 11,
    fontWeight: 500,
    color: '#ccc',
    textAlign: 'center' as const,
    maxWidth: 100,
    lineHeight: 1.2,
    wordWrap: 'break-word' as const,
  },
  handle: {
    background: '#4a9eff',
    width: 8,
    height: 8,
    border: '2px solid #1a1a2e',
  },
}

export default AgentNode
