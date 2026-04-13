package esearch

const (
	IndexEvents    = "glitch-events"
	IndexSummaries = "glitch-summaries"
	IndexPipelines = "glitch-pipelines"
	IndexInsights  = "glitch-insights"
)

const EventsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "type":      { "type": "keyword" },
      "source":    { "type": "keyword" },
      "repo":      { "type": "keyword" },
      "author":    { "type": "keyword" },
      "message":   { "type": "text" },
      "body":      { "type": "text" },
      "metadata":  { "type": "object", "enabled": false },
      "timestamp": { "type": "date" }
    }
  }
}`

const SummariesMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "scope":     { "type": "keyword" },
      "date":      { "type": "date" },
      "summary":   { "type": "text" },
      "timestamp": { "type": "date" }
    }
  }
}`

const PipelinesMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "name":      { "type": "keyword" },
      "status":    { "type": "keyword" },
      "stdout":    { "type": "text" },
      "model":     { "type": "keyword" },
      "provider":  { "type": "keyword" },
      "timestamp": { "type": "date" }
    }
  }
}`

const InsightsMapping = `{
  "settings": { "number_of_shards": 1, "number_of_replicas": 0 },
  "mappings": {
    "properties": {
      "type":           { "type": "keyword" },
      "pattern":        { "type": "text" },
      "recommendation": { "type": "text" },
      "timestamp":      { "type": "date" }
    }
  }
}`

// AllIndices returns a map of index name → mapping JSON for all managed indices.
func AllIndices() map[string]string {
	return map[string]string{
		IndexEvents:    EventsMapping,
		IndexSummaries: SummariesMapping,
		IndexPipelines: PipelinesMapping,
		IndexInsights:  InsightsMapping,
	}
}
