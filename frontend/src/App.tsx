import { useState, useEffect, useCallback, useRef } from 'react'
import Canvas from './components/Canvas'
import TeamSelector from './components/TeamSelector'
import AgentPanel from './components/AgentPanel'
import FloatingPanel from './components/FloatingPanel'
import ResizeHandle from './components/ResizeHandle'
import ViewsDropdown, { View, Position } from './components/ViewsDropdown'

interface TeamSummary {
  name: string
  version: string
  spec_dir: string
  repo_name: string
  num_agents: number
}

interface AgentDetail {
  name: string
  description: string
  model: string
  icon?: string
  tools: string[]
  skills?: string[]
  dependencies?: string[]
  namespace?: string
  instructions: string
}

// Map model names to header colors (matches AgentNode background colors)
function getModelColor(model?: string): string {
  const modelColors: Record<string, string> = {
    opus: '#e74c3c',    // red
    sonnet: '#3498db',  // blue
    haiku: '#2ecc71',   // green
  }
  return modelColors[model || ''] || '#0f3460'
}

interface OpenAgent {
  key: string // teamName/agentName - unique identifier
  teamName: string
  agent: AgentDetail
}

const MIN_SIDEBAR_WIDTH = 180
const MAX_LEFT_WIDTH = 400
const DEFAULT_LEFT_WIDTH = 240

function App() {
  const [teams, setTeams] = useState<TeamSummary[]>([])
  const [selectedTeams, setSelectedTeams] = useState<string[]>([])
  const [openAgents, setOpenAgents] = useState<OpenAgent[]>([])
  const [leftWidth, setLeftWidth] = useState(DEFAULT_LEFT_WIDTH)

  // View state
  const [currentView, setCurrentView] = useState<View | null>(null)
  const [nodePositions, setNodePositions] = useState<Record<string, Record<string, Position>>>({})
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false)
  const initialLoadRef = useRef(true)

  const handleLeftResize = useCallback((delta: number) => {
    setLeftWidth(w => Math.min(MAX_LEFT_WIDTH, Math.max(MIN_SIDEBAR_WIDTH, w + delta)))
    setHasUnsavedChanges(true)
  }, [])

  useEffect(() => {
    fetch('/api/teams')
      .then(res => res.json())
      .then(data => {
        setTeams(data || [])
        if (data && data.length > 0 && initialLoadRef.current) {
          setSelectedTeams([data[0].name])
          initialLoadRef.current = false
        }
      })
      .catch(err => console.error('Failed to load teams:', err))
  }, [])

  const handleToggleTeam = (name: string) => {
    setSelectedTeams(prev => {
      if (prev.includes(name)) {
        if (prev.length === 1) return prev
        return prev.filter(t => t !== name)
      } else {
        return [...prev, name]
      }
    })
    setHasUnsavedChanges(true)
  }

  const handleNodeClick = async (teamName: string, agentName: string) => {
    const key = `${teamName}/${agentName}`

    // Don't open if already open
    if (openAgents.some(a => a.key === key)) {
      return
    }

    try {
      const res = await fetch(`/api/agents/${teamName}/${agentName}`)
      if (res.ok) {
        const detail = await res.json()
        setOpenAgents(prev => [...prev, { key, teamName, agent: detail }])
      }
    } catch (err) {
      console.error('Failed to load agent:', err)
    }
  }

  const handleCloseAgent = useCallback((key: string) => {
    setOpenAgents(prev => prev.filter(a => a.key !== key))
  }, [])

  const handleNodePositionsChange = useCallback((teamName: string, positions: Record<string, Position>) => {
    setNodePositions(prev => ({
      ...prev,
      [teamName]: positions,
    }))
    setHasUnsavedChanges(true)
  }, [])

  // View management
  const handleLoadView = (view: View) => {
    setCurrentView(view)
    setSelectedTeams(view.teams)
    setNodePositions(view.node_positions || {})
    setLeftWidth(view.left_width)
    setHasUnsavedChanges(false)
  }

  const handleSaveView = async (name: string) => {
    const viewData = {
      name,
      teams: selectedTeams,
      node_positions: nodePositions,
      left_width: leftWidth,
      right_width: 0, // Deprecated: floating panel doesn't use this
    }
    const res = await fetch('/api/views', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(viewData),
    })
    if (res.ok) {
      const newView = await res.json()
      setCurrentView(newView)
      setHasUnsavedChanges(false)
    }
  }

  const handleUpdateView = async () => {
    if (!currentView) return
    const viewData = {
      name: currentView.name,
      teams: selectedTeams,
      node_positions: nodePositions,
      left_width: leftWidth,
      right_width: 0, // Deprecated: floating panel doesn't use this
    }
    const res = await fetch(`/api/views/${currentView.id}`, {
      method: 'PUT',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(viewData),
    })
    if (res.ok) {
      const updatedView = await res.json()
      setCurrentView(updatedView)
      setHasUnsavedChanges(false)
    }
  }

  const handleDeleteView = async (id: string) => {
    const res = await fetch(`/api/views/${id}`, { method: 'DELETE' })
    if (res.ok) {
      if (currentView?.id === id) {
        setCurrentView(null)
      }
    }
  }

  return (
    <>
      <TeamSelector
        teams={teams}
        selected={selectedTeams}
        onToggle={handleToggleTeam}
        width={leftWidth}
      />
      <ResizeHandle direction="left" onResize={handleLeftResize} />
      <div style={styles.canvasContainer}>
        <div style={styles.header}>
          <ViewsDropdown
            currentView={currentView}
            onLoadView={handleLoadView}
            onSaveView={handleSaveView}
            onUpdateView={handleUpdateView}
            onDeleteView={handleDeleteView}
            hasUnsavedChanges={hasUnsavedChanges}
          />
        </div>
        <div style={{
          ...styles.canvasArea,
          ...(selectedTeams.length > 1 ? styles.canvasGrid : {}),
        }}>
          {selectedTeams.length === 0 ? (
            <div style={styles.empty}>
              <p>Select an agent team to visualize</p>
            </div>
          ) : (
            selectedTeams.map((teamName) => (
              <div key={teamName} style={styles.canvasWrapper}>
                <div style={styles.teamLabel}>{teamName}</div>
                <Canvas
                  teamName={teamName}
                  onNodeClick={(agentName) => handleNodeClick(teamName, agentName)}
                  savedPositions={nodePositions[teamName]}
                  onPositionsChange={(positions) => handleNodePositionsChange(teamName, positions)}
                />
              </div>
            ))
          )}
        </div>
      </div>
      {openAgents.map((openAgent, index) => (
        <FloatingPanel
          key={openAgent.key}
          title={`${openAgent.agent.name} (${openAgent.teamName})`}
          onClose={() => handleCloseAgent(openAgent.key)}
          defaultWidth={420}
          defaultHeight={520}
          offsetIndex={index}
          headerColor={getModelColor(openAgent.agent.model)}
        >
          <AgentPanel agent={openAgent.agent} />
        </FloatingPanel>
      ))}
    </>
  )
}

const styles: Record<string, React.CSSProperties> = {
  canvasContainer: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
    overflow: 'hidden',
  },
  header: {
    padding: '8px 12px',
    borderBottom: '1px solid #333',
    background: '#1a1a2e',
    display: 'flex',
    alignItems: 'center',
  },
  canvasArea: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    overflow: 'hidden',
  },
  canvasGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(2, 1fr)',
    gridTemplateRows: 'repeat(2, 1fr)',
    gap: 2,
  },
  canvasWrapper: {
    flex: 1,
    display: 'flex',
    flexDirection: 'column',
    position: 'relative',
    minHeight: 200,
    border: '1px solid #333',
  },
  teamLabel: {
    position: 'absolute',
    top: 8,
    left: 8,
    zIndex: 10,
    background: 'rgba(26, 26, 46, 0.9)',
    padding: '4px 10px',
    borderRadius: 4,
    fontSize: 12,
    fontWeight: 600,
    color: '#4a9eff',
    border: '1px solid #333',
  },
  empty: {
    flex: 1,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    color: '#666',
    fontSize: 16,
  },
}

export default App
