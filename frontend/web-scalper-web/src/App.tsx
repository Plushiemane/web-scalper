import { useState, useMemo } from 'react'
import './App.css'

interface Job {
  title: string
  link: string
}

function App() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [query, setQuery] = useState('')
  const [localization, setLocalization] = useState('Remote')
  const [ETCategories, setEtCategories] = useState<number[]>([])
  
interface PositionLevel {
  etcode : number
  name : string
}

const PositionLevels : PositionLevel[] = [
  {etcode: 1, name: "praktykant/stażysta"},
  {etcode: 2, name: "pracownik fizyczny"},
  {etcode: 3, name: "asystent"},
  {etcode: 4, name: "Mid"},
  {etcode: 5, name: "kierownik"},
  {etcode: 6, name: "dyrektor"},
  {etcode: 17, name: "junior"},
  {etcode: 18, name: "Senior"},
  {etcode: 19, name: "Ekspert"},
  {etcode: 20, name: "Menedżer"},
  {etcode: 21, name: "prezes"}
]


const handleQuery = (event: React.ChangeEvent<HTMLInputElement>) => {
  setQuery(event.currentTarget.value);
}


  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setLoading(true)
    setError('')
    setJobs([])

  
    try {
      const response = await fetch('http://localhost:8080/jobs', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ query, ETCategories }),
      })

      if (!response.ok) {
        throw new Error('Failed to fetch jobs')
      }
      const data = await response.json()
      setJobs(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred')
    } finally {
      setLoading(false)
    }
  }
const [filter, setFilter] = useState('')
console.log(ETCategories);
const handleFilter = (event: React.ChangeEvent<HTMLInputElement>) => {
  setFilter(event.currentTarget.value);
}

const filtered = useMemo(() => jobs.filter(job => job.title.includes(filter)), [jobs, filter]) ?? [];
console.log(filter)

  return (
    <div className="app">
      <h1>Job Scalper</h1>
      <div style={{margin:"2%"}}>
        <input type="text" className='url-input' value={filter} onChange={handleFilter} placeholder='Filtruj oferty'/>
      </div>
      <div style={{margin:"2%"}}>
        {PositionLevels.map((level) => (
          <label key={level.etcode} style={{marginRight:"10px"}}>
            <input
              type="checkbox"
              value={level.name}
              checked={ETCategories.includes(level.etcode)}
              onChange={(e) => {
                if (e.target.checked) {
                  setEtCategories([...ETCategories, level.etcode]);
                } else {
                  setEtCategories(ETCategories.filter((cat) => cat !== level.etcode));
                }
              }}

            />
            {level.name}
          </label>
        ))}
      </div>
      
      <form onSubmit={handleSubmit} className="search-form">
        <input
          type="text"
          value={query}
          onChange={handleQuery}
          placeholder="Napisz rodzaj stanowiska"
          className="url-input"
        />
        <input type='text' value={localization}/>
        <button type="submit" disabled={loading}>
          {loading ? 'Loading...' : 'Search Jobs'}
        </button>
      </form>

      {error && <p className="error">{error}</p>}

      {jobs.length > 0 && filter == "" && (
        <div className="jobs-grid">
          {jobs.map((job, index) => (
            <div key={index} className="job-card">
              <h3>{job.title}</h3>
              <a href={job.link} target="_blank" rel="noopener noreferrer">
                View Job →
              </a>
            </div>
          ))}
        </div>
      )}
      {filter != "" && filtered.length > 0 && (
        <div className="jobs-grid">
          {filtered.map((job, index) => (
            <div key={index} className="job-card">
              <h3>{job.title}</h3>
              <a href={job.link} target="_blank" rel="noopener noreferrer">
                View Job →
              </a>
            </div>
          ))}
        </div>

      )}

      {!loading && jobs.length === 0 && query && (
        <p className="no-results">No jobs found</p>
      )}
    </div>
  )
}

export default App