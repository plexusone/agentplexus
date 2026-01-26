import { useState, useRef, useCallback, useEffect, ReactNode } from 'react'

interface Props {
  title: string
  children: ReactNode
  onClose: () => void
  defaultWidth?: number
  defaultHeight?: number
  minWidth?: number
  minHeight?: number
  offsetIndex?: number // For staggering multiple panels
  headerColor?: string // Custom header background color
}

interface Position {
  x: number
  y: number
}

interface Size {
  width: number
  height: number
}

function FloatingPanel({
  title,
  children,
  onClose,
  defaultWidth = 420,
  defaultHeight = 500,
  minWidth = 150,  // ~10em minimum
  minHeight = 150, // ~10em minimum
  offsetIndex = 0,
  headerColor = '#0f3460',
}: Props) {
  const [isMinimized, setIsMinimized] = useState(false)
  // Stagger panels by 30px each
  const offset = offsetIndex * 30
  const [position, setPosition] = useState<Position>(() => ({
    x: window.innerWidth - defaultWidth - 24 - offset,
    y: window.innerHeight - defaultHeight - 24 - offset,
  }))
  const [size, setSize] = useState<Size>({ width: defaultWidth, height: defaultHeight })
  const [isDragging, setIsDragging] = useState(false)
  const [isResizing, setIsResizing] = useState<'nw' | 'se' | false>(false)
  const dragStartRef = useRef<{ x: number; y: number; posX: number; posY: number } | null>(null)
  const resizeStartRef = useRef<{ x: number; y: number; width: number; height: number; posX: number; posY: number } | null>(null)
  const panelRef = useRef<HTMLDivElement>(null)

  const HEADER_HEIGHT = 40

  // Constrain position: header must stay visible vertically,
  // at least header height visible horizontally, body can scroll off
  const constrainPosition = useCallback((x: number, y: number, w: number) => {
    // Vertical: header cannot go off top or bottom
    const minY = 0
    const maxY = window.innerHeight - HEADER_HEIGHT

    // Horizontal: at least header height worth of panel must be visible
    const minX = -w + HEADER_HEIGHT
    const maxX = window.innerWidth - HEADER_HEIGHT

    return {
      x: Math.max(minX, Math.min(x, maxX)),
      y: Math.max(minY, Math.min(y, maxY)),
    }
  }, [])

  // Handle drag
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if ((e.target as HTMLElement).closest('button')) return
    e.preventDefault()
    setIsDragging(true)
    dragStartRef.current = {
      x: e.clientX,
      y: e.clientY,
      posX: position.x,
      posY: position.y,
    }
  }, [position])

  // Handle resize from top-left (nw)
  const handleResizeNW = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsResizing('nw')
    resizeStartRef.current = {
      x: e.clientX,
      y: e.clientY,
      width: size.width,
      height: size.height,
      posX: position.x,
      posY: position.y,
    }
  }, [size, position])

  // Handle resize from bottom-right (se)
  const handleResizeSE = useCallback((e: React.MouseEvent) => {
    e.preventDefault()
    e.stopPropagation()
    setIsResizing('se')
    resizeStartRef.current = {
      x: e.clientX,
      y: e.clientY,
      width: size.width,
      height: size.height,
      posX: position.x,
      posY: position.y,
    }
  }, [size, position])

  useEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (isDragging && dragStartRef.current) {
        const dx = e.clientX - dragStartRef.current.x
        const dy = e.clientY - dragStartRef.current.y
        const newPos = constrainPosition(
          dragStartRef.current.posX + dx,
          dragStartRef.current.posY + dy,
          size.width
        )
        setPosition(newPos)
      }
      if (isResizing && resizeStartRef.current) {
        const ref = resizeStartRef.current
        if (isResizing === 'nw') {
          // Top-left resize: grow up and left
          const dx = ref.x - e.clientX
          const dy = ref.y - e.clientY
          const newWidth = Math.max(minWidth, ref.width + dx)
          const newHeight = Math.max(minHeight, ref.height + dy)
          setSize({ width: newWidth, height: newHeight })
          setPosition(constrainPosition(
            ref.posX - (newWidth - ref.width),
            ref.posY - (newHeight - ref.height),
            newWidth
          ))
        } else if (isResizing === 'se') {
          // Bottom-right resize: grow down and right
          const dx = e.clientX - ref.x
          const dy = e.clientY - ref.y
          const newWidth = Math.max(minWidth, ref.width + dx)
          const newHeight = Math.max(minHeight, ref.height + dy)
          setSize({ width: newWidth, height: newHeight })
        }
      }
    }

    const handleMouseUp = () => {
      setIsDragging(false)
      setIsResizing(false)
      dragStartRef.current = null
      resizeStartRef.current = null
    }

    if (isDragging || isResizing) {
      document.addEventListener('mousemove', handleMouseMove)
      document.addEventListener('mouseup', handleMouseUp)
      return () => {
        document.removeEventListener('mousemove', handleMouseMove)
        document.removeEventListener('mouseup', handleMouseUp)
      }
    }
  }, [isDragging, isResizing, constrainPosition, size, minWidth, minHeight])

  // Re-constrain on window resize
  useEffect(() => {
    const handleResize = () => {
      setPosition(prev => constrainPosition(prev.x, prev.y, size.width))
    }
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [constrainPosition, size])

  return (
    <div
      ref={panelRef}
      style={{
        ...styles.panel,
        left: position.x,
        top: position.y,
        width: size.width,
        height: isMinimized ? 40 : size.height,
      }}
    >
      {/* Resize handle (top-left corner) */}
      {!isMinimized && (
        <div
          style={styles.resizeHandleNW}
          onMouseDown={handleResizeNW}
        />
      )}

      {/* Resize handle (bottom-right corner) */}
      {!isMinimized && (
        <div
          style={styles.resizeHandleSE}
          onMouseDown={handleResizeSE}
        />
      )}

      {/* Header */}
      <div
        style={{
          ...styles.header,
          background: headerColor,
          cursor: isDragging ? 'grabbing' : 'grab',
        }}
        onMouseDown={handleMouseDown}
      >
        <span style={styles.title}>{title}</span>
        <div style={styles.controls}>
          <button
            onClick={() => setIsMinimized(!isMinimized)}
            style={styles.controlBtn}
            title={isMinimized ? 'Expand' : 'Minimize'}
          >
            {isMinimized ? '▢' : '—'}
          </button>
          <button
            onClick={onClose}
            style={styles.controlBtn}
            title="Close"
          >
            ×
          </button>
        </div>
      </div>

      {/* Content */}
      {!isMinimized && (
        <div style={styles.content}>
          {children}
        </div>
      )}
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  panel: {
    position: 'fixed',
    background: '#16213e',
    border: '1px solid #333',
    borderRadius: 8,
    boxShadow: '0 8px 32px rgba(0, 0, 0, 0.4)',
    display: 'flex',
    flexDirection: 'column',
    zIndex: 1000,
    overflow: 'hidden',
  },
  resizeHandleNW: {
    position: 'absolute',
    top: 0,
    left: 0,
    width: 16,
    height: 16,
    cursor: 'nw-resize',
    zIndex: 10,
  },
  resizeHandleSE: {
    position: 'absolute',
    bottom: 0,
    right: 0,
    width: 16,
    height: 16,
    cursor: 'se-resize',
    zIndex: 10,
    // Visual indicator
    borderRight: '2px solid rgba(255,255,255,0.3)',
    borderBottom: '2px solid rgba(255,255,255,0.3)',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '8px 12px',
    background: '#0f3460',
    borderBottom: '1px solid #333',
    userSelect: 'none',
    flexShrink: 0,
  },
  title: {
    fontSize: 14,
    fontWeight: 600,
    color: '#fff',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
    whiteSpace: 'nowrap',
  },
  controls: {
    display: 'flex',
    gap: 4,
  },
  controlBtn: {
    background: 'none',
    border: 'none',
    color: '#888',
    fontSize: 16,
    width: 24,
    height: 24,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    cursor: 'pointer',
    borderRadius: 4,
  },
  content: {
    flex: 1,
    minHeight: 0, // Required for overflow to work in flex container
    overflow: 'auto',
  },
}

export default FloatingPanel
