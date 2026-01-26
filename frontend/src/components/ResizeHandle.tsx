import { useCallback, useEffect, useState } from 'react'

interface Props {
  onResize: (delta: number) => void
  direction: 'left' | 'right'
}

function ResizeHandle({ onResize, direction }: Props) {
  const [isDragging, setIsDragging] = useState(false)
  const [startX, setStartX] = useState(0)

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    setIsDragging(true)
    setStartX(e.clientX)
    e.preventDefault()
  }, [])

  useEffect(() => {
    if (!isDragging) return

    const handleMouseMove = (e: MouseEvent) => {
      const delta = e.clientX - startX
      const adjustedDelta = direction === 'left' ? delta : -delta
      onResize(adjustedDelta)
      setStartX(e.clientX)
    }

    const handleMouseUp = () => {
      setIsDragging(false)
    }

    document.addEventListener('mousemove', handleMouseMove)
    document.addEventListener('mouseup', handleMouseUp)

    return () => {
      document.removeEventListener('mousemove', handleMouseMove)
      document.removeEventListener('mouseup', handleMouseUp)
    }
  }, [isDragging, startX, onResize, direction])

  return (
    <div
      onMouseDown={handleMouseDown}
      style={{
        width: 6,
        cursor: 'col-resize',
        background: isDragging ? '#4a9eff' : 'transparent',
        transition: 'background 0.15s',
        flexShrink: 0,
      }}
      onMouseEnter={(e) => {
        if (!isDragging) {
          e.currentTarget.style.background = '#333'
        }
      }}
      onMouseLeave={(e) => {
        if (!isDragging) {
          e.currentTarget.style.background = 'transparent'
        }
      }}
    />
  )
}

export default ResizeHandle
