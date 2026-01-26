import { useState } from 'react'
import Markdown from 'react-markdown'
import { Prism as SyntaxHighlighter } from 'react-syntax-highlighter'
import { oneDark } from 'react-syntax-highlighter/dist/esm/styles/prism'

interface AgentDetail {
  name: string
  description: string
  model: string
  tools: string[]
  skills?: string[]
  dependencies?: string[]
  namespace?: string
  instructions: string
}

interface Props {
  agent: AgentDetail
}

type ViewMode = 'html' | 'code'

function AgentPanel({ agent }: Props) {
  const [viewMode, setViewMode] = useState<ViewMode>('html')

  return (
    <div style={styles.panel}>

      <div style={styles.meta}>
        <div style={styles.field}>
          <span style={styles.fieldLabel}>Model</span>
          <span style={styles.fieldValue}>{agent.model}</span>
        </div>
        {agent.namespace && (
          <div style={styles.field}>
            <span style={styles.fieldLabel}>Namespace</span>
            <span style={styles.fieldValue}>{agent.namespace}</span>
          </div>
        )}
      </div>

      <div style={styles.section}>
        <h4 style={styles.sectionTitle}>Description</h4>
        <p style={styles.description}>{agent.description}</p>
      </div>

      <div style={styles.section}>
        <h4 style={styles.sectionTitle}>Tools</h4>
        <div style={styles.badges}>
          {agent.tools.map(tool => (
            <span key={tool} style={styles.badge}>{tool}</span>
          ))}
        </div>
      </div>

      {agent.dependencies && agent.dependencies.length > 0 && (
        <div style={styles.section}>
          <h4 style={styles.sectionTitle}>Dependencies</h4>
          <div style={styles.badges}>
            {agent.dependencies.map(dep => (
              <span key={dep} style={styles.badge}>{dep}</span>
            ))}
          </div>
        </div>
      )}

      <div style={styles.instructionsSection}>
        <div style={styles.instructionsHeader}>
          <h4 style={styles.sectionTitle}>Instructions</h4>
          <div style={styles.viewToggle}>
            <button
              onClick={() => setViewMode('html')}
              style={{
                ...styles.toggleBtn,
                ...(viewMode === 'html' ? styles.toggleBtnActive : {}),
              }}
            >
              HTML
            </button>
            <button
              onClick={() => setViewMode('code')}
              style={{
                ...styles.toggleBtn,
                ...(viewMode === 'code' ? styles.toggleBtnActive : {}),
              }}
            >
              Code
            </button>
          </div>
        </div>
        {viewMode === 'html' ? (
          <div className="markdown-content" style={styles.markdownContent}>
            <Markdown
              components={{
                code({ className, children, ...props }) {
                  const match = /language-(\w+)/.exec(className || '')
                  const isInline = !match && !String(children).includes('\n')
                  return isInline ? (
                    <code className={className} style={styles.inlineCode} {...props}>
                      {children}
                    </code>
                  ) : (
                    <SyntaxHighlighter
                      style={oneDark}
                      language={match ? match[1] : 'text'}
                      PreTag="div"
                      customStyle={{
                        margin: '0.5em 0',
                        borderRadius: '4px',
                        fontSize: '12px',
                      }}
                    >
                      {String(children).replace(/\n$/, '')}
                    </SyntaxHighlighter>
                  )
                },
              }}
            >
              {agent.instructions}
            </Markdown>
          </div>
        ) : (
          <pre style={styles.instructions}>{agent.instructions}</pre>
        )}
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  panel: {
    padding: 16,
    display: 'flex',
    flexDirection: 'column',
    gap: 14,
  },
  meta: {
    display: 'flex',
    gap: 16,
  },
  field: {
    display: 'flex',
    flexDirection: 'column',
    gap: 2,
  },
  fieldLabel: {
    fontSize: 11,
    color: '#888',
    textTransform: 'uppercase' as const,
  },
  fieldValue: {
    fontSize: 14,
    color: '#fff',
  },
  section: {
    display: 'flex',
    flexDirection: 'column',
    gap: 6,
  },
  sectionTitle: {
    fontSize: 12,
    color: '#888',
    textTransform: 'uppercase' as const,
    letterSpacing: 0.5,
  },
  description: {
    fontSize: 13,
    color: '#ccc',
    lineHeight: 1.5,
  },
  badges: {
    display: 'flex',
    flexWrap: 'wrap' as const,
    gap: 6,
  },
  badge: {
    fontSize: 12,
    padding: '3px 8px',
    borderRadius: 4,
    background: '#0f3460',
    color: '#8ab4f8',
  },
  instructions: {
    fontSize: 12,
    color: '#ccc',
    background: '#0d1b36',
    padding: 12,
    borderRadius: 6,
    whiteSpace: 'pre-wrap' as const,
    lineHeight: 1.5,
    flex: 1,
    overflow: 'auto',
  },
  instructionsSection: {
    display: 'flex',
    flexDirection: 'column',
    gap: 6,
    flex: 1,
    minHeight: 0,
  },
  instructionsHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    flexShrink: 0,
  },
  viewToggle: {
    display: 'flex',
    gap: 2,
    background: '#0d1b36',
    borderRadius: 4,
    padding: 2,
  },
  toggleBtn: {
    fontSize: 11,
    padding: '4px 10px',
    border: 'none',
    borderRadius: 3,
    background: 'transparent',
    color: '#888',
    cursor: 'pointer',
  },
  toggleBtnActive: {
    background: '#1a3a6e',
    color: '#8ab4f8',
  },
  markdownContent: {
    fontSize: 13,
    color: '#ccc',
    background: '#0d1b36',
    padding: 12,
    borderRadius: 6,
    lineHeight: 1.6,
    flex: 1,
    overflow: 'auto',
  },
  inlineCode: {
    background: '#0a1628',
    padding: '0.15em 0.4em',
    borderRadius: 3,
    fontFamily: "'SF Mono', Monaco, monospace",
    fontSize: '0.9em',
  },
}

export default AgentPanel
