import { useEffect, useCallback, useState, useRef } from 'react'
import {
  ReactFlow,
  Node,
  Edge,
  Background,
  Controls,
  useNodesState,
  useEdgesState,
  ConnectionLineType,
  MarkerType,
  NodeChange,
} from '@xyflow/react'
import '@xyflow/react/dist/style.css'
import AgentNode from './AgentNode'

interface Position {
  x: number
  y: number
}

interface Props {
  teamName: string | null
  onNodeClick: (agentName: string) => void
  savedPositions?: Record<string, Position>
  onPositionsChange?: (positions: Record<string, Position>) => void
}

interface GraphResponse {
  nodes: {
    id: string
    label: string
    description: string
    model: string
    icon?: string
    tools: string[]
    position: { x: number; y: number }
  }[]
  edges: {
    id: string
    source: string
    target: string
  }[]
}

const nodeTypes = { agent: AgentNode }

function Canvas({ teamName, onNodeClick, savedPositions, onPositionsChange }: Props) {
  const [nodes, setNodes, onNodesChange] = useNodesState<Node>([])
  const [edges, setEdges, onEdgesChange] = useEdgesState<Edge>([])
  const [loading, setLoading] = useState(false)
  const positionsRef = useRef<Record<string, Position>>({})
  const appliedSavedPositionsRef = useRef(false)

  useEffect(() => {
    if (!teamName) return
    setLoading(true)
    appliedSavedPositionsRef.current = false

    fetch(`/api/teams/${teamName}/graph`)
      .then(res => res.json())
      .then((data: GraphResponse) => {
        const flowNodes: Node[] = data.nodes.map(n => {
          // Use saved position if available
          const pos = savedPositions?.[n.id] || n.position
          positionsRef.current[n.id] = pos
          return {
            id: n.id,
            type: 'agent',
            position: pos,
            data: {
              label: n.label,
              description: n.description,
              model: n.model,
              icon: n.icon,
              tools: n.tools,
            },
          }
        })

        const flowEdges: Edge[] = data.edges.map(e => ({
          id: e.id,
          source: e.source,
          target: e.target,
          type: 'smoothstep',
          animated: true,
          style: { stroke: '#4a9eff', strokeWidth: 2 },
          markerEnd: {
            type: MarkerType.ArrowClosed,
            color: '#4a9eff',
          },
        }))

        setNodes(flowNodes)
        setEdges(flowEdges)
        setLoading(false)
        appliedSavedPositionsRef.current = true
      })
      .catch(err => {
        console.error('Failed to load graph:', err)
        setLoading(false)
      })
  }, [teamName, savedPositions, setNodes, setEdges])

  const handleNodesChange = useCallback((changes: NodeChange<Node>[]) => {
    onNodesChange(changes)

    // Track position changes
    let hasPositionChange = false
    for (const change of changes) {
      if (change.type === 'position' && change.position) {
        positionsRef.current[change.id] = change.position
        hasPositionChange = true
      }
    }

    // Only report changes after initial load and if positions actually changed
    if (hasPositionChange && appliedSavedPositionsRef.current && onPositionsChange) {
      onPositionsChange({ ...positionsRef.current })
    }
  }, [onNodesChange, onPositionsChange])

  const handleNodeClick = useCallback((_: React.MouseEvent, node: Node) => {
    onNodeClick(node.data.label as string)
  }, [onNodeClick])

  if (!teamName) {
    return (
      <div style={styles.empty}>
        <p>Select an agent team to visualize</p>
      </div>
    )
  }

  if (loading) {
    return (
      <div style={styles.empty}>
        <p>Loading...</p>
      </div>
    )
  }

  return (
    <div style={styles.canvas}>
      <ReactFlow
        nodes={nodes}
        edges={edges}
        onNodesChange={handleNodesChange}
        onEdgesChange={onEdgesChange}
        onNodeClick={handleNodeClick}
        nodeTypes={nodeTypes}
        connectionLineType={ConnectionLineType.SmoothStep}
        fitView={!savedPositions}
        fitViewOptions={{ padding: 0.2 }}
        style={{ background: '#1a1a2e' }}
      >
        <Background color="#333" gap={20} />
        <Controls />
      </ReactFlow>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  canvas: {
    flex: 1,
    height: '100%',
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

export default Canvas
