<script>
  import Router from 'svelte-spa-router';
  import Sidebar from './lib/components/Sidebar.svelte';
  import WorkflowList from './routes/WorkflowList.svelte';
  import Editor from './routes/Editor.svelte';
  import RunList from './routes/RunList.svelte';
  import RunView from './routes/RunView.svelte';
  import ResultsBrowser from './routes/ResultsBrowser.svelte';

  const routes = {
    '/': WorkflowList,
    '/workflow/:name': Editor,
    '/runs': RunList,
    '/run/:id': RunView,
    '/results': ResultsBrowser,
  };
</script>

<Sidebar />
<main class="main-area">
  <Router {routes} />
</main>
