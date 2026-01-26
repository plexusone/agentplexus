interface TeamSummary {
  name: string
  version: string
  repo_name: string
  num_agents: number
}

interface Props {
  teams: TeamSummary[]
  selected: string[]
  onToggle: (name: string) => void
  width: number
}

function TeamSelector({ teams, selected, onToggle, width }: Props) {
  return (
    <div style={{ ...styles.sidebar, width }}>
      <h2 style={styles.title}>Agent Teams</h2>
      <div style={styles.list}>
        {teams.map(team => {
          const isSelected = selected.includes(team.name)
          return (
            <button
              key={team.name}
              onClick={() => onToggle(team.name)}
              style={{
                ...styles.item,
                ...(isSelected ? styles.itemActive : {}),
              }}
            >
              <div style={styles.itemHeader}>
                <span style={{
                  ...styles.checkbox,
                  ...(isSelected ? styles.checkboxChecked : {}),
                }}>
                  {isSelected ? '✓' : ''}
                </span>
                <div style={styles.teamName}>{team.name}</div>
              </div>
              <div style={styles.teamMeta}>
                {team.repo_name} · {team.num_agents} agents
              </div>
            </button>
          )
        })}
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  sidebar: {
    width: 240,
    borderRight: '1px solid #333',
    padding: 16,
    display: 'flex',
    flexDirection: 'column',
    gap: 12,
    background: '#16213e',
  },
  title: {
    fontSize: 14,
    textTransform: 'uppercase',
    letterSpacing: 1,
    color: '#888',
  },
  list: {
    display: 'flex',
    flexDirection: 'column',
    gap: 4,
  },
  item: {
    background: 'none',
    border: '1px solid transparent',
    borderRadius: 6,
    padding: '8px 12px',
    textAlign: 'left' as const,
    cursor: 'pointer',
    color: '#ccc',
  },
  itemActive: {
    background: '#0f3460',
    borderColor: '#4a9eff',
    color: '#fff',
  },
  itemHeader: {
    display: 'flex',
    alignItems: 'center',
    gap: 8,
  },
  checkbox: {
    width: 16,
    height: 16,
    borderRadius: 3,
    border: '1px solid #555',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: 10,
    color: '#fff',
    flexShrink: 0,
  },
  checkboxChecked: {
    background: '#4a9eff',
    borderColor: '#4a9eff',
  },
  teamName: {
    fontSize: 14,
    fontWeight: 600,
  },
  teamMeta: {
    fontSize: 11,
    color: '#888',
    marginTop: 2,
    marginLeft: 24,
  },
}

export default TeamSelector
