<script>
  import Router from 'svelte-spa-router'
  import WorkflowList from './routes/WorkflowList.svelte'
  import Editor from './routes/Editor.svelte'
  import RunList from './routes/RunList.svelte'
  import RunView from './routes/RunView.svelte'
  import ResultsBrowser from './routes/ResultsBrowser.svelte'

  const routes = {
    '/': WorkflowList,
    '/workflow/:name': Editor,
    '/runs': RunList,
    '/run/:id': RunView,
    '/results': ResultsBrowser,
  }
</script>

<div class="shell">
  <nav class="topnav">
    <a href="#/">gl1tch</a>
    <a href="#/">Workflows</a>
    <a href="#/runs">Runs</a>
    <a href="#/results">Results</a>
  </nav>
  <main>
    <Router {routes} />
  </main>
</div>
