import { useState, useRef, useEffect } from 'react'

interface Position {
  x: number
  y: number
}

interface View {
  id: string
  name: string
  teams: string[]
  node_positions: Record<string, Record<string, Position>>
  left_width: number
  right_width: number
  created_at: string
  updated_at: string
}

interface Props {
  currentView: View | null
  onLoadView: (view: View) => void
  onSaveView: (name: string) => Promise<void>
  onUpdateView: () => Promise<void>
  onDeleteView: (id: string) => Promise<void>
  hasUnsavedChanges: boolean
}

function ViewsDropdown({
  currentView,
  onLoadView,
  onSaveView,
  onUpdateView,
  onDeleteView,
  hasUnsavedChanges,
}: Props) {
  const [isOpen, setIsOpen] = useState(false)
  const [views, setViews] = useState<View[]>([])
  const [showSaveDialog, setShowSaveDialog] = useState(false)
  const [newViewName, setNewViewName] = useState('')
  const [loading, setLoading] = useState(false)
  const dropdownRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    // Close dropdown when clicking outside
    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsOpen(false)
        setShowSaveDialog(false)
      }
    }
    document.addEventListener('mousedown', handleClickOutside)
    return () => document.removeEventListener('mousedown', handleClickOutside)
  }, [])

  useEffect(() => {
    if (isOpen) {
      loadViews()
    }
  }, [isOpen])

  const loadViews = async () => {
    try {
      const res = await fetch('/api/views')
      if (res.ok) {
        const data = await res.json()
        setViews(data || [])
      }
    } catch (err) {
      console.error('Failed to load views:', err)
    }
  }

  const handleSave = async () => {
    if (!newViewName.trim()) return
    setLoading(true)
    try {
      await onSaveView(newViewName.trim())
      setNewViewName('')
      setShowSaveDialog(false)
      await loadViews()
    } finally {
      setLoading(false)
    }
  }

  const handleUpdate = async () => {
    setLoading(true)
    try {
      await onUpdateView()
      await loadViews()
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = async (id: string) => {
    if (!confirm('Delete this view?')) return
    setLoading(true)
    try {
      await onDeleteView(id)
      await loadViews()
    } finally {
      setLoading(false)
    }
  }

  return (
    <div ref={dropdownRef} style={styles.container}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        style={styles.trigger}
      >
        <span style={styles.icon}>📐</span>
        <span>{currentView?.name || 'Views'}</span>
        {hasUnsavedChanges && <span style={styles.unsavedDot}>●</span>}
        <span style={styles.arrow}>▼</span>
      </button>

      {isOpen && (
        <div style={styles.dropdown}>
          {showSaveDialog ? (
            <div style={styles.saveDialog}>
              <input
                type="text"
                value={newViewName}
                onChange={(e) => setNewViewName(e.target.value)}
                placeholder="View name..."
                style={styles.input}
                autoFocus
                onKeyDown={(e) => {
                  if (e.key === 'Enter') handleSave()
                  if (e.key === 'Escape') setShowSaveDialog(false)
                }}
              />
              <div style={styles.dialogButtons}>
                <button
                  onClick={() => setShowSaveDialog(false)}
                  style={styles.cancelBtn}
                >
                  Cancel
                </button>
                <button
                  onClick={handleSave}
                  style={styles.saveBtn}
                  disabled={!newViewName.trim() || loading}
                >
                  Save
                </button>
              </div>
            </div>
          ) : (
            <>
              <div style={styles.actions}>
                <button
                  onClick={() => setShowSaveDialog(true)}
                  style={styles.actionBtn}
                >
                  + Save as new view
                </button>
                {currentView && hasUnsavedChanges && (
                  <button
                    onClick={handleUpdate}
                    style={styles.actionBtn}
                    disabled={loading}
                  >
                    💾 Update "{currentView.name}"
                  </button>
                )}
              </div>

              {views.length > 0 && (
                <>
                  <div style={styles.divider} />
                  <div style={styles.viewList}>
                    {views.map((view) => (
                      <div
                        key={view.id}
                        style={{
                          ...styles.viewItem,
                          ...(currentView?.id === view.id ? styles.viewItemActive : {}),
                        }}
                      >
                        <button
                          onClick={() => {
                            onLoadView(view)
                            setIsOpen(false)
                          }}
                          style={styles.viewBtn}
                        >
                          <span>{view.name}</span>
                          <span style={styles.viewMeta}>
                            {view.teams.length} team{view.teams.length !== 1 ? 's' : ''}
                          </span>
                        </button>
                        <button
                          onClick={() => handleDelete(view.id)}
                          style={styles.deleteBtn}
                          title="Delete view"
                        >
                          ×
                        </button>
                      </div>
                    ))}
                  </div>
                </>
              )}

              {views.length === 0 && (
                <div style={styles.empty}>No saved views yet</div>
              )}
            </>
          )}
        </div>
      )}
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    position: 'relative',
  },
  trigger: {
    display: 'flex',
    alignItems: 'center',
    gap: 6,
    background: '#16213e',
    border: '1px solid #333',
    borderRadius: 6,
    padding: '6px 12px',
    color: '#ccc',
    cursor: 'pointer',
    fontSize: 13,
  },
  icon: {
    fontSize: 14,
  },
  arrow: {
    fontSize: 8,
    marginLeft: 4,
  },
  unsavedDot: {
    color: '#f39c12',
    fontSize: 10,
  },
  dropdown: {
    position: 'absolute',
    top: '100%',
    left: 0,
    marginTop: 4,
    background: '#16213e',
    border: '1px solid #333',
    borderRadius: 6,
    minWidth: 240,
    zIndex: 1000,
    boxShadow: '0 4px 12px rgba(0,0,0,0.4)',
  },
  actions: {
    padding: 8,
    display: 'flex',
    flexDirection: 'column',
    gap: 4,
  },
  actionBtn: {
    background: 'none',
    border: 'none',
    padding: '8px 12px',
    color: '#4a9eff',
    cursor: 'pointer',
    textAlign: 'left',
    borderRadius: 4,
    fontSize: 13,
  },
  divider: {
    height: 1,
    background: '#333',
    margin: '4px 0',
  },
  viewList: {
    maxHeight: 300,
    overflowY: 'auto',
    padding: 4,
  },
  viewItem: {
    display: 'flex',
    alignItems: 'center',
    borderRadius: 4,
  },
  viewItemActive: {
    background: '#0f3460',
  },
  viewBtn: {
    flex: 1,
    background: 'none',
    border: 'none',
    padding: '8px 12px',
    color: '#ccc',
    cursor: 'pointer',
    textAlign: 'left',
    display: 'flex',
    flexDirection: 'column',
    gap: 2,
  },
  viewMeta: {
    fontSize: 11,
    color: '#666',
  },
  deleteBtn: {
    background: 'none',
    border: 'none',
    color: '#666',
    cursor: 'pointer',
    padding: '4px 8px',
    fontSize: 16,
  },
  saveDialog: {
    padding: 12,
  },
  input: {
    width: '100%',
    padding: '8px 12px',
    background: '#0a1628',
    border: '1px solid #333',
    borderRadius: 4,
    color: '#fff',
    fontSize: 13,
    marginBottom: 8,
  },
  dialogButtons: {
    display: 'flex',
    gap: 8,
    justifyContent: 'flex-end',
  },
  cancelBtn: {
    background: 'none',
    border: '1px solid #333',
    borderRadius: 4,
    padding: '6px 12px',
    color: '#888',
    cursor: 'pointer',
    fontSize: 12,
  },
  saveBtn: {
    background: '#4a9eff',
    border: 'none',
    borderRadius: 4,
    padding: '6px 12px',
    color: '#fff',
    cursor: 'pointer',
    fontSize: 12,
  },
  empty: {
    padding: 16,
    color: '#666',
    textAlign: 'center',
    fontSize: 13,
  },
}

export default ViewsDropdown
export type { View, Position }
