<script>
  import Router from 'svelte-spa-router';
  import ActivityBar from './lib/components/ActivityBar.svelte';
  import WorkflowList from './routes/WorkflowList.svelte';
  import WorkflowDetail from './routes/WorkflowDetail.svelte';
  import Settings from './routes/Settings.svelte';

  const routes = {
    '/': WorkflowList,
    '/workflow/:name': WorkflowDetail,
    '/settings': Settings,
  };
</script>

<ActivityBar />
<main class="main-area">
  <Router {routes} />
</main>
