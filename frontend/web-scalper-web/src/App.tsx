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
  const [isIntern, setIsIntern] = useState(false)
  

const handleQuery = (event: React.ChangeEvent<HTMLInputElement>) => {
  setQuery(event.currentTarget.value);
}

const handleIsIntern = (event: React.ChangeEvent<HTMLInputElement>) => {
  setIsIntern(event.currentTarget.checked);
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
        body: JSON.stringify({ query, isIntern }),
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
        <span>IsIntern
        <input type="checkbox" className='url-input' checked={isIntern} onChange={handleIsIntern} />
        </span>
      </div>
      
      <form onSubmit={handleSubmit} className="search-form">
        <input
          type="text"
          value={query}
          onChange={handleQuery}
          placeholder="Napisz rodzaj stanowiska"
          className="url-input"
        />
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